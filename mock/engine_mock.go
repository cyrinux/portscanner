package mock

import (
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"strings"
)

type MockEngine struct {
	State   pb.ScannerResponse_Status
	Started *pb.ParamsScannerRequest
}

func (e *MockEngine) Start(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
	e.Started = params
	var hostResults []*pb.HostResult
	targets := strings.Split(params.Targets, ",")

	for i := 0; i < len(targets); i++ {
		hostResults = append(hostResults, &pb.HostResult{Host: &pb.Host{Address: "scanme.nmap.org"}})
	}

	return hostResults, nil
}

func (e *MockEngine) SetState(state pb.ScannerResponse_Status) {
	e.State = state
}

func (e *MockEngine) GetState() pb.ScannerResponse_Status {
	return e.State
}

func CreateMockEngine() *MockEngine {
	return &MockEngine{}
}
