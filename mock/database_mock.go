package mock

import (
	"context"
	"encoding/json"
	// "github.com/cyrinux/grpcnmapscanner/database"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"time"
)

type mockDatabase struct {
	Contents map[string]string
}

// CreateMockDatabase creates the mock database
func CreateMockDatabase(ctx context.Context) (mockDatabase, error) {
	contents := map[string]string{}
	return mockDatabase{Contents: contents}, nil
}

func (r mockDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	r.Contents[key] = value
	return key, nil
}

func (r mockDatabase) Get(ctx context.Context, key string) (string, error) {
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

func (r mockDatabase) GetAll(ctx context.Context, key string) ([]string, error) {
	arr := make([]string, 1000)
	return arr, nil
}

func (r mockDatabase) Delete(ctx context.Context, key string) (string, error) {
	return key, nil
}
