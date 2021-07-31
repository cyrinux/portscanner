package mock

import (
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
)

type MockEngine struct {
	State   pb.ScannerResponse_Status
	Started *pb.ParamsScannerRequest

	StartImpl func(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error)
}

func (e *MockEngine) Start(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
	e.Started = params
	return e.StartImpl(params, async)
}

func (e *MockEngine) SetState(state pb.ScannerResponse_Status) {
	e.State = state
}

func (e *MockEngine) GetState() pb.ScannerResponse_Status {
	return e.State
}
