package scanner

import (
	"encoding/json"
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"time"
)

type Server struct {
	db database.Database
}

func NewServer(db database.Database) *Server {
	return &Server{db}
}

func (s *Server) DeleteScan(ctx context.Context, in *proto.GetScannerRequest) (*proto.ServerResponse, error) {
	_, err := s.db.Delete(in.Key)
	return generateResponse(in.Key, nil, err)
}

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

	portList := []*proto.Port{}
	scanResult := []*proto.HostResult{}
	totalPorts := 0

	result, err := StartNmapScan(in)
	if err != nil || result == nil {
		return generateResponse("", nil, err)
	}

	for _, host := range result.Hosts {
		var osversion string
		if len(host.Addresses) == 0 {
			continue
		}
		if len(host.OS.Matches) > 0 {
			fp := host.OS.Matches[0]
			osversion = fmt.Sprintf("name: %v, accuracy: %v%%", fp.Name, fp.Accuracy)
		}
		address := host.Addresses[0].Addr
		hostResult := &proto.Host{
			Address:   address,
			OsVersion: &osversion,
			State:     &host.Status.Reason,
		}
		for _, p := range host.Ports {
			version := &proto.PortVersion{
				ExtraInfos:  &p.Service.ExtraInfo,
				LowVersion:  &p.Service.LowVersion,
				HighVersion: &p.Service.HighVersion,
				Product:     &p.Service.Product,
			}
			newPort := &proto.Port{
				PortId:      fmt.Sprintf("%v", p.ID),
				ServiceName: p.Service.Name,
				Protocol:    p.Protocol,
				State:       p.State.Reason,
				Version:     version,
			}
			portList = append(portList, newPort)
		}
		totalPorts += len(portList)

		scan := &proto.HostResult{
			Host:  hostResult,
			Ports: portList,
		}

		scanResult = append(scanResult, scan)
	}

	scannerResponse := &proto.ScannerResponse{
		HostResult: scanResult,
		// StartTime:  fmt.Sprintf("%v", &result.TaskBegin[1]),
		// EndTime:    fmt.Sprintf("%v", &result.TaskEnd[1]),
	}
	// fmt.Print(result.TaskBegin[0])
	// fmt.Print(result.TaskEnd[0])
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

func generateResponse(key string, value *proto.ScannerResponse, err error) (*proto.ServerResponse, error) {
	if err != nil {
		return &proto.ServerResponse{Success: false, Key: "", Value: value, Error: err.Error()}, nil
	}
	return &proto.ServerResponse{Success: true, Key: key, Value: value, Error: ""}, nil
}
