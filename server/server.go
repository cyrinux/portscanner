package server

import (
	"encoding/json"
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	conf       = config.GetConfig()
	log        = logger.New(conf.Logger.Debug, conf.Logger.Pretty)
	opsSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "scanner_processed_ops_success",
		Help: "The total number of success tasks",
	})
	opsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "scanner_processed_ops_failed",
		Help: "The total number of failed tasks",
	})
	opsReturned = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_ops_returned",
		Help: "The total number of returned tasks",
	})
	opsReady = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_queue_ready",
		Help: "The total number of ready tasks tasks",
	})
	opsRejected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_queue_rejected",
		Help: "The total number of rejected tasks tasks",
	})
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     3600 * time.Second, // If a client is idle for 3600 seconds, send a GOAWAY
	MaxConnectionAge:      1800 * time.Second, // If any connection is alive for more than 1800 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 5 * time.Second,    // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,    // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               2 * time.Second,    // Wait 1 second for the ping ack before assuming the connection is dead
}

// TasksStatus define the status of all tasks
type TasksStatus struct {
	success  int64 //tassks success
	failed   int64 //tasks failure
	returned int64 //tasks returned
	ready    int64 //queue tasks ready
	rejected int64 //queue tasks rejected
}

// Server define the grpc server struct
type Server struct {
	ctx         context.Context
	config      config.Config
	db          database.Database
	err         error
	broker      broker.Broker
	state       pb.ScannerServiceControl
	tasktype    string
	tasksStatus TasksStatus
}

// Listen start the grpc server
func Listen(ctx context.Context, allConfig config.Config) error {
	log.Info().Msg("prepare to serve the gRPC api")
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Error().Err(err).Msg("can't start server, can't listen on tcp/9000")
		return err
	}

	//prometheus endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	//grpc endpoint
	srv := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)

	reflection.Register(srv)
	pb.RegisterScannerServiceServer(srv, NewServer(ctx, allConfig, "nmap"))
	if err = srv.Serve(listener); err != nil {
		log.Error().Msg("can't serve the gRPC service")
		return err
	}

	return nil
}

// brokerStatsToProm read broker stats each 2s and write to prometheus
func brokerStatsToProm(brk *broker.Broker, name string) {
	for {
		stats, err := broker.GetStats(brk)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get RMQ stats")
			time.Sleep(2000 * time.Millisecond)
			continue
		}
		queue := fmt.Sprintf("%s-incoming", name)
		s := stats.QueueStats[queue]
		if s.ReadyCount > 0 || s.RejectedCount > 0 {
			opsReady.Set(float64(s.ReadyCount))
			opsRejected.Set(float64(s.RejectedCount))
			log.Debug().Msgf("broker %s incoming queue ready: %v, rejected: %v", name, s.ReadyCount, s.RejectedCount)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

// NewServer create a new server and init the database connection
func NewServer(ctx context.Context, config config.Config, tasktype string) *Server {
	errChan := make(chan error)
	go broker.RmqLogErrors(errChan)

	// Broker init nmap queue
	brk := broker.NewBroker(context.TODO(), tasktype, config, util.RedisConnect(ctx, config))

	// Storage database init
	db, err := database.Factory(context.TODO(), config)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("can't open the database")
	}

	// start the rmq stats to prometheus
	go brokerStatsToProm(&brk, "nmap")

	return &Server{
		ctx:      ctx,
		tasktype: tasktype,
		config:   config,
		broker:   brk,
		db:       db,
		err:      err,
	}
}

// DeleteScan delele a scan from the database
func (server *Server) DeleteScan(ctx context.Context, in *pb.GetScannerRequest) (*pb.ServerResponse, error) {
	_, err := server.db.Delete(ctx, in.Key)
	return generateResponse(in.Key, nil, err)
}

// StreamServiceControl control the service
func (server *Server) StreamServiceControl(in *pb.ScannerServiceControl, stream pb.ScannerService_StreamServiceControlServer) error {
	for {
		if err := stream.Send(&server.state); err != nil {
			log.Error().Stack().Err(err).Msgf("streamer service control send error")
			return errors.Wrap(err, "streamer service control send error")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// StreamTasksStatus manage the tasks counter
func (server *Server) StreamTasksStatus(stream pb.ScannerService_StreamTasksStatusServer) error {
	for {
		tasksStatus, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			return err
		}
		server.tasksStatus.success += tasksStatus.Success
		opsSuccess.Add(float64(tasksStatus.Success))
		server.tasksStatus.failed += tasksStatus.Failed
		opsFailed.Add(float64(tasksStatus.Failed))
		server.tasksStatus.returned += tasksStatus.Returned
		opsReturned.Set(float64(tasksStatus.Returned))
		stream.Send(&pb.TasksStatus{Success: server.tasksStatus.success, Failed: server.tasksStatus.failed, Returned: server.tasksStatus.returned})

		time.Sleep(500 * time.Millisecond)
	}
}

// ServiceControl control the service
func (server *Server) ServiceControl(ctx context.Context, in *pb.ScannerServiceControl) (*pb.ScannerServiceControl, error) {
	if in.GetState() == pb.ScannerServiceControl_UNKNOWN {
		return &pb.ScannerServiceControl{State: server.state.State}, nil
	}

	if server.state.State != in.GetState() {
		server.state.State = in.GetState()
	}

	return &pb.ScannerServiceControl{State: server.state.State}, nil
}

// GetScan return the engine scan result
func (server *Server) GetScan(ctx context.Context, in *pb.GetScannerRequest) (*pb.ServerResponse, error) {
	var scannerResponse pb.ScannerResponse

	scanResult, err := server.db.Get(ctx, in.Key)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	err = json.Unmarshal([]byte(scanResult), &scannerResponse)
	if err != nil {
		log.Error().Stack().Err(err).Msg("can't read scan result")
		return generateResponse(in.Key, nil, err)
	}

	return generateResponse(in.Key, &scannerResponse, nil)
}

// StartScan function prepare a nmap scan
func (server *Server) StartScan(ctx context.Context, in *pb.ParamsScannerRequest) (*pb.ServerResponse, error) {
	// sanitize
	request := parseParamsScannerRequest(in)

	// we start the scan
	newEngine := engine.NewEngine(ctx, server.db)

	scannerResponse := pb.ScannerResponse{Status: pb.ScannerResponse_ERROR}

	key, result, err := newEngine.StartNmapScan(request)
	if err != nil {
		return generateResponse(key, nil, err)
	}

	scanParsedResult, err := engine.ParseScanResult(result)
	if err != nil {
		return generateResponse(key, nil, err)
	}
	scannerResponse = pb.ScannerResponse{
		HostResult: scanParsedResult,
	}

	scanResultJSON, err := json.Marshal(&scannerResponse)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	// and write the response to the database
	_, err = server.db.Set(
		ctx, key,
		string(scanResultJSON),
		time.Duration(request.GetRetentionTime())*time.Second,
	)
	if err != nil {
		return generateResponse(key, &scannerResponse, err)
	}

	scannerResponse = pb.ScannerResponse{
		HostResult: scanParsedResult,
		Status:     pb.ScannerResponse_OK,
	}

	return generateResponse(key, &scannerResponse, nil)
}

// StartAsyncScan function prepare a nmap scan
func (server *Server) StartAsyncScan(ctx context.Context, in *pb.ParamsScannerRequest) (*pb.ServerResponse, error) {
	request := parseParamsScannerRequest(in)

	// and write the response to the database
	scannerResponse := pb.ScannerResponse{Status: pb.ScannerResponse_QUEUED}
	scanResponseJSON, _ := json.Marshal(&scannerResponse)
	log.Info().Msgf("receive async task order: %v", request)
	_, err := server.db.Set(ctx, request.Key, string(scanResponseJSON), 0)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	// create scan task
	taskScanBytes, err := json.Marshal(request)
	if err != nil {
		scannerResponse := pb.ScannerResponse{Status: pb.ScannerResponse_ERROR}
		return generateResponse(request.Key, &scannerResponse, err)
	}
	err = server.broker.Incoming.PublishBytes(taskScanBytes)
	if err != nil {
		scannerResponse := pb.ScannerResponse{Status: pb.ScannerResponse_ERROR}
		return generateResponse(request.Key, &scannerResponse, err)
	}

	scannerResponse = pb.ScannerResponse{Status: pb.ScannerResponse_QUEUED}
	return generateResponse(request.Key, &scannerResponse, nil)
}

// generateResponse generate the response for the grpc return
func generateResponse(key string, value *pb.ScannerResponse, err error) (*pb.ServerResponse, error) {
	if err != nil {
		return &pb.ServerResponse{
			Success: false,
			Key:     key,
			Value:   value,
			Error:   err.Error(),
		}, nil
	}
	return &pb.ServerResponse{
		Success: true,
		Key:     key,
		Value:   value,
		Error:   "",
	}, nil
}

func parseParamsScannerRequest(request *pb.ParamsScannerRequest) *pb.ParamsScannerRequest {

	// if the Key is not forced, we generate one unique
	guid := uuid.New()
	if request.Key == "" {
		request.Key = guid.String()
	}

	// If timeout < 10s, fallback to 1h
	if request.Timeout < 30 && request.Timeout > 0 {
		request.Timeout = 60 * 60
		// if timeout not set
	} else if request.Timeout == 0 {
		request.Timeout = 60 * 60
	}

	// replace all whitespaces
	request.Hosts = strings.ReplaceAll(request.Hosts, " ", "")
	request.Ports = strings.ReplaceAll(request.Ports, " ", "")

	// add the defer duration to the current unix timestamp
	request.DeferDuration = time.Now().Add(time.Duration(request.DeferDuration) * time.Second).Unix()

	return request
}
