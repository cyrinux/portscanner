package scanner

import (
	"encoding/json"
	rmq "github.com/adjust/rmq/v3"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engines"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"log"
	"time"
)

const (
	shouldLog = true
)

type Server struct {
	db database.Database
}

// NewServer create a new server and init the database connection
func NewServer(db database.Database) *Server {
	return &Server{db}
}

// DeleteScan delelee a scan from the database
func (s *Server) DeleteScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	_, err := s.db.Delete(in.Key)
	return generateResponse(in.Key, nil, err)
}

// GetScan fetch a scan from the database by key id
func (s *Server) GetScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	var scannerResponse proto.ScannerResponse

	scanResult, err := s.db.Get(in.Key)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	err = json.Unmarshal([]byte(scanResult), &scannerResponse)
	if err != nil {
		return generateResponse(in.Key, nil, err)
	}

	return generateResponse(in.Key, &scannerResponse, nil)
}

// Scan function prepare a nmap scan
func (s *Server) StartScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {

	guid := xid.New()

	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	result, err := engines.StartNmapScan(in)
	if err != nil || result == nil {
		return generateResponse("", nil, err)
	}

	scanResult, _ := engines.ParseScanResult(result)

	scannerResponse := &proto.ScannerResponse{
		HostResult: scanResult,
	}
	scanResultJSON, err := json.Marshal(scannerResponse)
	if err != nil {
		return generateResponse("", nil, err)
	}
	_, err = s.db.Set(guid.String(), string(scanResultJSON), time.Duration(in.GetRetentionTime())*time.Second)
	if err != nil {
		return generateResponse("", nil, err)
	}

	return generateResponse(guid.String(), scannerResponse, err)
}

// Scan function prepare a nmap scan
func (s *Server) StartAsyncScan(ctx context.Context, in *proto.ParamsScannerRequest) (*proto.ServerResponse, error) {
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnection("scannerqueue", "tcp", "localhost:6379", 1, errChan)
	taskQueue, err := connection.OpenQueue("tasks")

	guid := xid.New()

	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	// create task
	taskBytes, err := json.Marshal(in)
	if err != nil {
		return generateResponse(guid.String(), nil, err)
	}
	err = taskQueue.PublishBytes(taskBytes)
	if err != nil {
		return generateResponse(guid.String(), nil, err)
	}

	return generateResponse(guid.String(), nil, err)

	// result, err := StartNmapScan(in)
	// if err != nil || result == nil {
	// 	return generateResponse("", nil, err)
	// }

	// scanResult, _ := ParseScanResult(result)

	// scannerResponse := &proto.ScannerResponse{
	// 	HostResult: scanResult,
	// }
	// scanResultJSON, err := json.Marshal(scannerResponse)
	// if err != nil {
	// 	return generateResponse("", nil, err)
	// }
	// _, err = s.db.Set(guid.String(), string(scanResultJSON), time.Duration(in.GetRetentionTime())*time.Second)
	// if err != nil {
	// 	return generateResponse("", nil, err)
	// }

	// return generateResponse(guid.String(), nil, err)
}

// generateResponse generate the response for the grpc return
func generateResponse(key string, value *proto.ScannerResponse, err error) (*proto.ServerResponse, error) {
	if err != nil {
		return &proto.ServerResponse{Success: false, Key: "", Value: value, Error: err.Error()}, nil
	}
	return &proto.ServerResponse{Success: true, Key: key, Value: value, Error: ""}, nil
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
