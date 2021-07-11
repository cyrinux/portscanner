package server

import (
	"encoding/json"
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/helpers"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	confLogger = config.GetConfig().Logger
	log        = logger.New(confLogger.Debug, confLogger.Pretty)
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
		Help: "The total number of ready tasks",
	})
	opsRejected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_queue_rejected",
		Help: "The total number of rejected tasks",
	})
	workersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_workers_count",
		Help: "The total number of scanner workers",
	})
	consumersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_consumers_count",
		Help: "The total number of scanner consumers",
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
	State       pb.ServiceStateValues
	tasktype    string
	tasksStatus TasksStatus
}

// ScannerListen start the frontend grpc server
func Listen(ctx context.Context, conf config.Config) {
	log.Info().Msg("prepare to serve the gRPC api")

	// Start the server
	server := NewServer(ctx, conf, "nmap")

	// Serve the frontend
	go func(s *Server) {
		frontListener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.FrontendListenPort))
		if err != nil {
			log.Fatal().Err(err).Msgf(
				"can't start frontend server, can't listen on tcp/%d", conf.FrontendListenPort,
			)
		}
		srvFrontend := grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(kaep),
			grpc.KeepaliveParams(kasp),
		)

		// graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				// sig is a ^C, handle it
				log.Info().Msg("shutting down gRPC server...")

				srvFrontend.GracefulStop()

				<-ctx.Done()
			}
		}()

		reflection.Register(srvFrontend)
		pb.RegisterScannerServiceServer(srvFrontend, server)
		if err = srvFrontend.Serve(frontListener); err != nil {
			log.Fatal().Msg("can't serve the gRPC frontend service")
		}
	}(server)

	// Serve the backend
	backendListener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.BackendListenPort))
	if err != nil {
		log.Fatal().Err(err).Msgf(
			"can't start backend server, can't listen on tcp/%d", conf.BackendListenPort,
		)
	}
	srvBackend := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
	)
	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Info().Msg("shutting down gRPC server...")

			srvBackend.GracefulStop()

			<-ctx.Done()
		}
	}()

	reflection.Register(srvBackend)
	pb.RegisterBackendServiceServer(srvBackend, server)
	if err = srvBackend.Serve(backendListener); err != nil {
		log.Fatal().Msg("can't serve the gRPC backend service")
	}
}

// brokerStatsToProm read broker stats each 2s and write to prometheus
func brokerStatsToProm(brk *broker.Broker, name string) {
	for range time.Tick(1000 * time.Millisecond) {
		stats, err := broker.GetStats(brk)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get RMQ statistics")
			continue
		}
		queue := fmt.Sprintf("%s-incoming", name)
		s := stats.QueueStats[queue]
		opsReady.Set(float64(s.ReadyCount))
		opsRejected.Set(float64(s.RejectedCount))
		consumersCount.Set(float64(s.ConsumerCount()))
		if s.ReadyCount > 0 || s.RejectedCount > 0 {
			log.Debug().Msgf("broker %s incoming queue ready: %v, rejected: %v", name, s.ReadyCount, s.RejectedCount)
		}
	}
}

// NewServer create a new server and init the database connection
func NewServer(ctx context.Context, conf config.Config, tasktype string) *Server {
	errChan := make(chan error)
	go broker.RmqLogErrors(errChan)

	//redis
	redisClient, err := helpers.NewRedisClient(ctx, conf)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	// Broker init nmap queue
	brker := broker.NewBroker(ctx, tasktype, conf.RMQ, redisClient)
	// clean the queues
	go brker.Cleaner()

	// Storage database init
	db, err := database.Factory(ctx, conf)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("can't open the database")
	}

	// start the rmq stats to prometheus
	go brokerStatsToProm(&brker, tasktype)

	return &Server{
		ctx:      ctx,
		tasktype: tasktype,
		config:   conf,
		broker:   brker,
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
func (server *Server) StreamServiceControl(in *pb.ServiceStateValues, stream pb.BackendService_StreamServiceControlServer) error {
	for {
		if err := stream.Send(&server.State); err != nil {
			log.Error().Stack().Err(err).Msgf("streamer service control send error")
			return errors.Wrap(err, "streamer service control send error")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// StreamTasksStatus manage the tasks counter
func (server *Server) StreamTasksStatus(stream pb.BackendService_StreamTasksStatusServer) error {
	for {
		promStatus, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			return err
		}
		server.tasksStatus.success += promStatus.TasksStatus.Success
		opsSuccess.Add(float64(promStatus.TasksStatus.Success))
		server.tasksStatus.failed += promStatus.TasksStatus.Failed
		opsFailed.Add(float64(promStatus.TasksStatus.Failed))
		server.tasksStatus.returned += promStatus.TasksStatus.Returned
		opsReturned.Set(float64(promStatus.TasksStatus.Returned))
		stream.Send(&pb.PrometheusStatus{TasksStatus: &pb.TasksStatus{
			Success: server.tasksStatus.success, Failed: server.tasksStatus.failed, Returned: server.tasksStatus.returned}},
		)

		time.Sleep(500 * time.Millisecond)
	}
}

// ServiceControl control the service
func (server *Server) ServiceControl(ctx context.Context, in *pb.ServiceStateValues) (*pb.ServiceStateValues, error) {
	if in.GetState() == pb.ServiceStateValues_UNKNOWN {
		return &pb.ServiceStateValues{State: server.State.State}, nil
	}

	if server.State.State != in.GetState() {
		server.State.State = in.GetState()
	}

	return &pb.ServiceStateValues{State: server.State.State}, nil
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

// GetAllScans return the engine scan result
func (server *Server) GetAllScans(ctx context.Context, in *empty.Empty) (*pb.AllServerResponses, error) {
	var scannerResponse pb.ScannerResponse
	var allScannerResponses []*pb.ScannerResponse

	allKeys, err := server.db.GetAll(ctx, "*")
	if err != nil {
		return generateArrayResponses("list", nil, err)
	}

	for _, key := range allKeys {
		scanResult, err := server.db.Get(ctx, key)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get scan result")
			continue
		}

		err = json.Unmarshal([]byte(scanResult), &scannerResponse)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't list scan result")
			return generateArrayResponses("list", nil, err)
		}

		allScannerResponses = append(allScannerResponses, &scannerResponse)
	}

	return generateArrayResponses("list", allScannerResponses, nil)
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

// generateArrayResponses generate the response for the grpc return
func generateArrayResponses(key string, responses []*pb.ScannerResponse, err error) (*pb.AllServerResponses, error) {
	if key == "list" {
		arr := make([]*pb.ServerResponse, 0)
		for _, scanResp := range responses {
			serverResponse := &pb.ServerResponse{
				Success: true,
				Key:     scanResp.Key,
				Value:   scanResp,
				Error:   "",
			}
			arr = append(arr, serverResponse)
		}
		return &pb.AllServerResponses{Response: arr}, nil
	}
	return &pb.AllServerResponses{Response: nil}, errors.Wrap(nil, "can't get all responses")
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
