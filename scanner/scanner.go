package scanner

import (
	"encoding/json"
	rmq "github.com/adjust/rmq/v3"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"log"
	"strings"
	"time"
)

const (
	shouldLog = true
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
	_, err := s.config.DB.Delete(in.Key)
	return generateResponse(in.Key, nil, err)
}

// GetScan fetch a scan from the database by key id
func (s *Server) GetScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	var scannerResponse proto.ScannerResponse

	scanResult, err := s.config.DB.Get(in.Key)
	log.Printf("%v", scanResult)
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

	request.Hosts = strings.ReplaceAll(request.Hosts, " ", "")
	request.Ports = strings.ReplaceAll(request.Ports, " ", "")

	return request
}

// Scan function prepare a nmap scan
func (s *Server) StartScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {

	// sanitize
	in = parseParamsScannerRequest(in)

	// we start the scan
	newEngine := engine.NewEngine(s.config)

	scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}

	key, scanResult, err := newEngine.StartNmapScan(ctx, in)
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
	_, err = s.config.DB.Set(key, string(scanResultJSON), time.Duration(in.GetRetentionTime())*time.Second)
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

	connection, err := rmq.OpenConnection(s.config.RmqDbName, "tcp", s.config.RmqServer, 1, errChan)
	taskQueue, err := connection.OpenQueue("tasks")

	in = parseParamsScannerRequest(in)

	// create scan task
	taskScanBytes, err := json.Marshal(in)
	if err != nil {
		scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}
		return generateResponse(in.Key, &scannerResponse, err)
	}
	err = taskQueue.PublishBytes(taskScanBytes)
	if err != nil {
		scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_ERROR}
		return generateResponse(in.Key, &scannerResponse, err)
	}

	scannerResponse := proto.ScannerResponse{Status: proto.ScannerResponse_QUEUED}
	return generateResponse(in.Key, &scannerResponse, err)
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
