package scanner

import (
	"encoding/json"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"strings"
	"sync"
	"time"
)

// Server define the grpc server struct
type Server struct {
	sync.Mutex
	ctx    context.Context
	config config.Config
	queue  rmq.Queue
	err    error
	state  proto.ScannerServiceControl
	db     database.Database
}

// Listen start the grpc server
func Listen(allConfig config.Config) {
	log.Info().Msg("Prepare to serve the gRPC api")
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal().Err(err).Msg("Can't listen on tcp/9000")
	}
	srv := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))

	reflection.Register(srv)
	proto.RegisterScannerServiceServer(srv, NewServer(allConfig))
	if e := srv.Serve(listener); e != nil {
		log.Fatal().Err(err).Msg("Can't serve the gRPC service")
	}
}

// NewServer create a new server and init the database connection
func NewServer(config config.Config) *Server {
	ctx := context.Background()

	errChan := make(chan error, 10) //TODO: arbitrary, to be change
	go rmqLogErrors(errChan)

	// Storage database init
	db, err := database.Factory(context.Background(), config)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't open the database")
	}

	// Broker init
	connection, err := rmq.OpenConnectionWithRedisClient(
		config.RMQ.Name,
		util.RedisConnect(ctx, config),
		errChan,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("RMQ connection error")
	}

	queue, err := connection.OpenQueue("tasks")
	if err != nil {
		log.Fatal().Err(err).Msg("RMQ queue open error")
	}
	return &Server{
		ctx:    ctx,
		config: config,
		queue:  queue,
		db:     db,
		err:    err,
	}
}

// DeleteScan delele a scan from the database
func (server *Server) DeleteScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	_, err := server.db.Delete(ctx, in.Key)
	return generateResponse(in.Key, nil, err)
}

// StreamServiceControl control the service
func (server *Server) StreamServiceControl(in *proto.ScannerServiceControl, stream proto.ScannerService_StreamServiceControlServer) error {
	for {
		if err := stream.Send(&server.state); err != nil {
			log.Error().Msgf("Send error %s", err)
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// ServiceControl control the service
func (server *Server) ServiceControl(ctx context.Context, in *proto.ScannerServiceControl) (*proto.ScannerServiceControl, error) {
	if in.GetState() == proto.ScannerServiceControl_UNKNOWN {
		return &proto.ScannerServiceControl{State: server.state.State}, nil
	}

	if server.state.State != in.GetState() {
		server.state.State = in.GetState()
	}

	return &proto.ScannerServiceControl{State: server.state.State}, nil
}

// GetScan return the engine scan resultt
func (server *Server) GetScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	var scannerResponse proto.ScannerResponse

	scanResult, err := server.db.Get(ctx, in.Key)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	err = json.Unmarshal([]byte(scanResult), &scannerResponse)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	return generateResponse(in.Key, &scannerResponse, nil)
}

// StartScan function prepare a nmap scan
func (server *Server) StartScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {
	// sanitize
	request := parseParamsScannerRequest(in)

	// we start the scan
	newEngine := engine.NewEngine(server.config, server.db)

	scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}

	key, scanResult, err := newEngine.StartNmapScan(ctx, request)
	if err != nil || scanResult == nil {
		return generateResponse(key, nil, err)
	}

	key, scanParsedResult, err := engine.ParseScanResult(key, scanResult)
	if err != nil {
		return generateResponse(key, nil, err)
	}

	scanResultJSON, _ := json.Marshal(scanParsedResult)

	// and write the response to the database
	_, err = server.db.Set(
		ctx, key,
		string(scanResultJSON),
		time.Duration(request.GetRetentionTime())*time.Second,
	)
	if err != nil {
		return generateResponse(key, &scannerResponse, err)
	}

	scannerResponse = proto.ScannerResponse{
		HostResult: scanParsedResult,
		Status:     proto.ScannerResponse_OK,
	}
	return generateResponse(key, &scannerResponse, err)
}

// StartAsyncScan function prepare a nmap scan
func (server *Server) StartAsyncScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {

	request := parseParamsScannerRequest(in)

	// and write the response to the database
	scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_QUEUED}
	scanResponseJSON, _ := json.Marshal(&scannerResponse)
	log.Info().Msgf("Receive async task order: %+v", request)
	_, err := server.db.Set(ctx, request.Key, string(scanResponseJSON), 0)
	if err != nil {
		log.Print(err)
	}

	// create scan task
	taskScanBytes, err := json.Marshal(request)
	if err != nil {
		scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}
		return generateResponse(request.Key, &scannerResponse, err)
	}
	err = server.queue.PublishBytes(taskScanBytes)
	if err != nil {
		scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}
		return generateResponse(request.Key, &scannerResponse, err)
	}

	scannerResponse = proto.ScannerResponse{Status: proto.ScannerResponse_QUEUED}
	return generateResponse(request.Key, &scannerResponse, err)
}

// generateResponse generate the response for the grpc return
func generateResponse(key string, value *proto.ScannerResponse, err error) (*proto.ServerResponse, error) {
	if err != nil {
		return &proto.ServerResponse{
			Success: false,
			Key:     key,
			Value:   value,
			Error:   err.Error(),
		}, nil
	}
	return &proto.ServerResponse{
		Success: true,
		Key:     key,
		Value:   value,
		Error:   "",
	}, nil
}

// rmqLogErrors display the rmq errors log
func rmqLogErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Error().Msgf("heartbeat error (limit): %s", err.RedisErr)
			} else {
				log.Error().Msgf("heartbeat error: %v", err.RedisErr)
			}
		case *rmq.ConsumeError:
			log.Error().Msgf("consume error: %v", err.RedisErr)
		case *rmq.DeliveryError:
			log.Error().Msgf("delivery error: %v %v", err.Delivery, err.RedisErr)
		default:
			log.Error().Msgf("other error: %v", err.Error)
		}
	}
}

func parseParamsScannerRequest(request *proto.ParamsScannerRequest) *proto.ParamsScannerRequest {

	// if the Key is not forced, we generate one unique
	guid := xid.New()
	if request.Key == "" {
		request.Key = guid.String()
	}

	// If timeout < 10s, fallback to 1h
	if request.Timeout < 30 {
		request.Timeout = 60 * 60
	}

	// replace all whitespaces
	request.Hosts = strings.ReplaceAll(request.Hosts, " ", "")
	request.Ports = strings.ReplaceAll(request.Ports, " ", "")

	// add the defer duration to the current unix timestamp
	request.DeferDuration = time.Now().Add(time.Duration(request.DeferDuration) * time.Second).Unix()

	return request
}
