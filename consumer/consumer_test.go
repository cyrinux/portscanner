package consumer

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/mock"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var (
	req = pb.ParamsScannerRequest{
		Key:                     "9c68bb98-2ba9-412a-a6b1-06963b408e84",
		Ports:                   "443,80,22",
		Timeout:                 3600,
		ServiceVersionDetection: true,
		OsDetection:             true,
		ServiceDefaultScripts:   true,
		FastMode:                true,
		ScanSpeed:               5,
		ProcessPerTarget:        true,
		OpenOnly:                false,
		NetworkChuncked:         true,
		Username:                "user1",
		Role:                    "user",
	}
)

func Test_CreateMockLock_Create_New_Locker(t *testing.T) {
	mockLock := mock.CreateMockLock()
	assert.NotNil(t, mockLock, "Should not be nil")

	ctx := context.Background()
	conf := config.GetConfig()
	db, _ := mock.CreateMockDatabase(ctx)
	engine := &mock.MockEngine{
		StartImpl: func(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
			var hostResults []*pb.HostResult
			targets := strings.Split(params.Targets, ",")

			for i := 0; i < len(targets); i++ {
				hostResults = append(hostResults, &pb.HostResult{Host: &pb.Host{Address: "scanme.nmap.org"}})
			}

			return hostResults, nil
		},
	}
	tag, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	assert.NotEmpty(t, tag, "Tag should not be empty")
	assert.NotNil(t, incConsumer, "Consumer should not be nil")
}

// func Test_consumeNow_Validate_Consuming_zero_target(t *testing.T) {
// 	mockLock := mock.CreateMockLock()
// 	ctx := context.Background()
// 	conf := config.GetConfig()

// 	engine := mock.CreateMockEngine()
// 	rmqMock := mock.CreateMockRMQ()
// 	db, _ := mock.CreateMockDatabase(ctx)

// 	_, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)

// 	req.Targets = ""
// 	err := incConsumer.consumeNow(&rmqMock, &req, "")
// 	assert.NotNil(t, err, "consumerNow return should NOT return nil")
// }

func Test_consumeNow_Validate_Consuming_one_target(t *testing.T) {
	mockLock := mock.CreateMockLock()
	ctx := context.Background()
	conf := config.GetConfig()
	rmqMock := mock.CreateMockRMQ()
	db, _ := mock.CreateMockDatabase(ctx)

	engine := &mock.MockEngine{
		StartImpl: func(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
			var hostResults []*pb.HostResult
			targets := strings.Split(params.Targets, ",")

			for i := 0; i < len(targets); i++ {
				hostResults = append(hostResults, &pb.HostResult{Host: &pb.Host{Address: "scanme.nmap.org"}})
			}

			return hostResults, nil
		},
	}
	_, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)

	req.Targets = "levis.name"
	err := incConsumer.consumeNow(&rmqMock, &req, "")
	assert.Nil(t, err, "consuming one target return should return nil")

	assert.Equal(t, 1, len(db.Contents), "it must be one")
	assert.Equal(t, `{"response":[{"host_result":[{"host":{"address":"scanme.nmap.org"}}]}]}`, db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], "it must returb an engine json response")
	assert.Equal(t, 1, rmqMock.Acks, "it must return one acknowledged msg")
	assert.Equal(t, &req, engine.Started, "engine need to take requests params as params engine")

}

func Test_consumeNow_Validate_Consuming_zero_target(t *testing.T) {
	mockLock := mock.CreateMockLock()
	ctx := context.Background()
	conf := config.GetConfig()
	rmqMock := mock.CreateMockRMQ()
	db, _ := mock.CreateMockDatabase(ctx)

	engine := &mock.MockEngine{
		StartImpl: func(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
			return nil, errors.New("my engine return an error")
		},
	}
	_, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)

	req.Targets = "levis.name"
	err := incConsumer.consumeNow(&rmqMock, &req, "")
	assert.NoError(t, err, "consuming one target return should return nil")
	assert.Equal(t, 1, len(db.Contents), "it must be one")
	assert.True(t, strings.Contains(db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], `"status":4`))
	assert.Equal(t, 1, rmqMock.Acks, "it must return one acknowledged msg")
	assert.Equal(t, &req, engine.Started, "engine need to take requests params as params engine")

}

// func Test_consumeNow_Validate_Consuming_two_targets(t *testing.T) {
// 	mockLock := mock.CreateMockLock()
// 	ctx := context.Background()
// 	conf := config.GetConfig()

// 	engine := mock.CreateMockEngine()
// 	rmqMock := mock.CreateMockRMQ()
// 	db, _ := mock.CreateMockDatabase(ctx)

// 	_, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)

// 	req.Targets = "levis.name,scanme.nmap.org"
// 	err := incConsumer.consumeNow(&rmqMock, &req, "")
// 	assert.Nil(t, err, "consumerNow return should return nil")
// }
