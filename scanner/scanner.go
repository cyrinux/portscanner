package scanner

import (
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strings"
	"time"
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,            // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
	MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
	MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
	Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
}

// Server define the grpc server struct
type Server struct {
	config      config.Config
	queue       rmq.Queue
	err         error
	workerState proto.ScannerServiceControl
}

// Listen start the grpc server
func Listen(allConfig config.Config) {
	fmt.Println("Prepare to serve the gRPC api")
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err) // The port may be on use
	}
	srv := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))

	reflection.Register(srv)
	proto.RegisterScannerServiceServer(srv, NewServer(allConfig))
	if e := srv.Serve(listener); e != nil {
		panic(err)
	}
}

// NewServer create a new server and init the database connection
func NewServer(config config.Config) *Server {
	ctx := context.Background()
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnectionWithRedisClient(
		config.RmqDbName,
		util.RedisConnect(ctx, config),
		errChan,
	)
	if err != nil {
		panic(err)
	}

	queue, err := connection.OpenQueue("tasks")
	if err != nil {
		panic(err)
	}

	return &Server{config: config, queue: queue, err: err}
}

// DeleteScan delele a scan from the database
func (server *Server) DeleteScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	_, err := server.config.DB.Delete(ctx, in.Key)
	return generateResponse(in.Key, nil, err)
}

// StreamServiceControl control the service
func (server *Server) StreamServiceControl(in *proto.ScannerServiceControl, stream proto.ScannerService_StreamServiceControlServer) error {
	if in.GetState() != proto.ScannerServiceControl_UNKNOWN {
		for {
			if err := stream.Send(&server.workerState); err != nil {
				return err
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

// ServiceControl control the service
func (server *Server) ServiceControl(ctx context.Context, in *proto.ScannerServiceControl) (*proto.ScannerServiceControl, error) {
	if in.GetState() == proto.ScannerServiceControl_UNKNOWN {
		log.Printf("DEBUG SERVER 1: %v", server.workerState.State)
		return &proto.ScannerServiceControl{State: server.workerState.State}, nil
	}

	if server.workerState.State != in.GetState() {
		server.workerState.State = in.GetState()
	}
	log.Printf("DEBUG SERVER 2: %v", server.workerState.State)

	return &proto.ScannerServiceControl{State: server.workerState.State}, nil
}

// GetScan return the engine scan resultt
func (server *Server) GetScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	var scannerResponse proto.ScannerResponse

	scanResult, err := server.config.DB.Get(ctx, in.Key)
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
	newEngine := engine.NewEngine(server.config)

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
	_, err = server.config.DB.Set(
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
	log.Printf("Receive async task order: %v", scanResponseJSON)
	_, err := server.config.DB.Set(ctx, request.Key, string(scanResponseJSON), 0)
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
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
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
