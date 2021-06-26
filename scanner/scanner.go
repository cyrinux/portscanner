package scanner

import (
	"encoding/json"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/go-redis/redis/v8"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"log"
	"strings"
	"time"
)

type Server struct {
	config config.Config
}

// NewServer create a new server and init the database connection
func NewServer(config config.Config) *Server {
	return &Server{config}
}

// DeleteScan delelee a scan from the database
func (s *Server) DeleteScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	_, err := s.config.DB.Delete(ctx, in.Key)
	return generateResponse(in.Key, nil, err)
}

// GetScan fetch a scan from the database by key id
func (s *Server) GetScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	var scannerResponse proto.ScannerResponse

	scanResult, err := s.config.DB.Get(ctx, in.Key)
	log.Printf("%s %v", in.Key, scanResult)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	err = json.Unmarshal([]byte(scanResult), &scannerResponse)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	return generateResponse(in.Key, &scannerResponse, nil)
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

// Scan function prepare a nmap scan
func (s *Server) StartScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {

	// sanitize
	request := parseParamsScannerRequest(in)

	// we start the scan
	newEngine := engine.NewEngine(s.config)

	scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}

	key, scanResult, err := newEngine.StartNmapScan(ctx, request)
	if err != nil || scanResult == nil {
		// and write the response to the database
		return generateResponse(key, nil, err)
	}

	key, scanParsedResult, err := engine.ParseScanResult(key, scanResult)
	if err != nil {
		// and write the response to the database
		return generateResponse(key, nil, err)
	}

	scanResultJSON, _ := json.Marshal(scanParsedResult)

	// and write the response to the database
	_, err = s.config.DB.Set(
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

// Scan function prepare a nmap scan
func (s *Server) StartAsyncScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {

	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnectionWithRedisClient(
		s.config.RmqDbName,
		redisConnect(ctx, s),
		errChan,
	)
	taskQueue, err := connection.OpenQueue("tasks")

	request := parseParamsScannerRequest(in)

	// and write the response to the database
	scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_QUEUED}
	scanResponseJSON, _ := json.Marshal(&scannerResponse)
	_, err = s.config.DB.Set(ctx, request.Key, string(scanResponseJSON), 0)
	if err != nil {
		log.Print(err)
	}

	// create scan task
	taskScanBytes, err := json.Marshal(request)
	if err != nil {
		scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}
		return generateResponse(request.Key, &scannerResponse, err)
	}
	err = taskQueue.PublishBytes(taskScanBytes)
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

func redisConnect(ctx context.Context, s *Server) *redis.Client {
	config := s.config
	// Connect to redis for the locker
	redisClient := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     config.RmqServer,
		Password: config.RmqDbPassword,
		DB:       0,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return redisClient
}
