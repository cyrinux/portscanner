package mock

import (
	"context"
	"encoding/json"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"time"
)

type MockDatabase struct {
	Contents map[string]string
	SetImpl  func(ctx context.Context, key string, value string, retention time.Duration) (string, error)
}

func (r MockDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	r.Contents[key] = value
	return r.SetImpl(ctx, key, value, retention)
}

func (r MockDatabase) Get(ctx context.Context, key string) (string, error) {
	hostResult := &pb.HostResult{
		Host: &pb.Host{
			Address: "scanme.nmap.org",
		},
	}
	hostResults := []*pb.HostResult{}
	hostResults = append(hostResults, hostResult)

	sr := pb.ScannerResponse{
		HostResult: hostResults,
	}
	srs := []*pb.ScannerResponse{&sr}
	smr := pb.ScannerMainResponse{Response: srs}

	smrJSON, err := json.Marshal(&smr)
	return string(smrJSON), err
}

func (r MockDatabase) GetAll(ctx context.Context, key string) ([]string, error) {
	arr := make([]string, 1000)
	return arr, nil
}

func (r MockDatabase) Delete(ctx context.Context, key string) (string, error) {
	return key, nil
}
