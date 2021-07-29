package mock

import (
	"github.com/cyrinux/grpcnmapscanner/engine"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
)

type MockEngine struct {
	State pb.ScannerResponse_Status
}

func (e *MockEngine) Start(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
	// hostsResult := []*pb.HostResult{}
	return nil, nil
}

func (e *MockEngine) SetState(state pb.ScannerResponse_Status) {
	e.State = state
}

func (e *MockEngine) GetState() pb.ScannerResponse_Status {
	return e.State
}

func CreateMockEngine() engine.EngineInterface {
	return &MockEngine{}
}
