package mock

import (
	"context"
	"encoding/json"
	"github.com/cyrinux/grpcnmapscanner/database"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"time"
)

type mockDatabase struct {
}

// CreateMockDatabase creates the mock database
func CreateMockDatabase(ctx context.Context) (database.Database, error) {
	return mockDatabase{}, nil
}

func (r mockDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	return key, nil
}

func (r mockDatabase) Get(ctx context.Context, key string) (string, error) {
	smr := pb.ScannerMainResponse{}
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
