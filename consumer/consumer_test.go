package consumer

import (
	"context"
	"encoding/json"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/mock"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"testing"
	"time"
)

var (
	req1 = pb.ParamsScannerRequest{
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
	req2 = pb.ParamsScannerRequest{
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
	req3 = pb.ParamsScannerRequest{
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
	req4 = pb.ParamsScannerRequest{
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

func init() {
	before := time.Now().Add(time.Minute * -10)
	after := time.Now().Add(time.Minute * 10)

	req1.Targets = "scanme.nmap.org"
	req1.DeferTime = timestamppb.New(before)

	req2.Targets = ""
	req2.DeferTime = timestamppb.New(before)

	req3.Targets = "levis.name,maximbaz.com"
	req3.DeferTime = timestamppb.New(before)

	req4.Targets = "levis.name,maximbaz.com"
	req4.DeferTime = timestamppb.New(after)
}

func Test_CreateMockLock_Create_New_Locker(t *testing.T) {
	ctx := context.Background()
	conf := config.GetConfig()
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	assert.NotNil(t, mockLock, "Should not be nil")

	db := mock.MockDatabase{
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}
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
	tag, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	assert.NotEmpty(t, tag, "Tag should not be empty")
	assert.NotNil(t, consumer, "Consumer should not be nil")
}

func Test_consumeNow_ValidateConsumingOneTarget(t *testing.T) {
	ctx := context.Background()
	conf := config.GetConfig()
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req1)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}
	engine := &mock.MockEngine{
		StartImpl: func(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
			var hostResults []*pb.HostResult
			targets := strings.Split(params.Targets, ",")
			for _, t := range targets {
				hostResults = append(hostResults, &pb.HostResult{Host: &pb.Host{Address: t}})
			}
			return hostResults, nil
		},
	}
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_START
	consumer.Consume(&delivery)
	assert.Equal(t, 1, len(db.Contents), "it must be one")
	assert.Equal(t, `{"response":[{"host_result":[{"host":{"address":"scanme.nmap.org"}}]}]}`, db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], "it must returb an engine json response")
	assert.Equal(t, 1, delivery.Acks, "it must return one acknowledged msg")
	assert.Equal(t, &req1, engine.Started, "engine need to take requests params as params engine")

}

func Test_consumeNow_ValidateConsumingZeroTarget(t *testing.T) {
	ctx := context.Background()
	conf := config.GetConfig()
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req2)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}

	engine := &mock.MockEngine{
		StartImpl: func(params *pb.ParamsScannerRequest, async bool) ([]*pb.HostResult, error) {
			return nil, errors.New("my engine return an error")
		},
	}
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_START
	consumer.Consume(&delivery)
	assert.Equal(t, 1, len(db.Contents), "it must be one")
	assert.Equal(t, 1, delivery.Acks, "it must return one acks msg")
	assert.Equal(t, &req2, engine.Started, "engine need to take requests params as params engine")
	assert.True(t, strings.Contains(db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], `"status":4`), "engine need to got the error state")
}

func Test_Cancel_CancelSuccesfull(t *testing.T) {
	ctx := context.Background()
	conf := config.GetConfig()
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req3)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}
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
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_START
	consumer.Consume(&delivery)
	err := consumer.Cancel()
	assert.NoError(t, err, "consuming one target return should return nil")
	assert.Equal(t, 1, len(db.Contents), "it must be one")
	assert.Equal(t, 1, delivery.Acks, "it must return one acknowledged msg")
	assert.Equal(t, &req3, engine.Started, "engine need to take requests params as params engine")
	assert.True(t, strings.Contains(db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], `"status":5`), "engine need to got the cancel state")
}

func Test_Cancel_CancelDbError(t *testing.T) {
	ctx := context.Background()
	conf := config.GetConfig()
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req3)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return "", errors.New("failed to insert failed result")
		},
	}
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
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_START
	consumer.Consume(&delivery)
	err := consumer.Cancel()
	assert.EqualError(t, err, "failed to insert failed result", "consuming one target return should return nil")
	assert.Equal(t, 1, len(db.Contents), "it must be one")
	assert.Equal(t, 1, delivery.Acks, "it must return one acknowledged msg")
	assert.Equal(t, &req3, engine.Started, "engine need to take requests params as params engine")
	assert.True(t, strings.Contains(db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], `"status":5`), "engine need to got the cancel state")
}

func Test_Consumer_ConsumerIsStop(t *testing.T) {
	ctx := context.Background()
	conf := config.GetConfig()
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req3)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}
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
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_STOP
	consumer.Consume(&delivery)
	assert.Equal(t, 0, len(db.Contents), "it must be zero")
	assert.Equal(t, 1, delivery.Rejected, "it must return one rejected msg")
	assert.Nil(t, engine.Started, "engine need to take requests params as params engine")
	assert.Equal(t, db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], "")
}

func Test_Consumer_PayloadIsDefered(t *testing.T) {
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			return true, nil
		},
	}
	ctx := context.Background()
	conf := config.GetConfig()
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req4)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}
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
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_START
	consumer.Consume(&delivery)
	assert.Equal(t, 0, len(db.Contents), "it must be zero")
	assert.Equal(t, 1, delivery.Rejected, "it must return one rejected msg")
	assert.Nil(t, engine.Started, "engine need to take requests params as params engine")
	assert.Equal(t, db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], "")
}

func Test_Consumer_CantTakeALock(t *testing.T) {
	mockLock := &mock.MockLocker{
		ObtainImpl: func(ctx context.Context, key string, ttl time.Duration) (bool, error) {
			// return false, errors.New("lock already taken")
			return false, nil
		},
	}
	ctx := context.Background()
	conf := config.GetConfig()
	delivery := mock.MockDelivery{
		PayloadImpl: func() string {
			reqJSON, _ := json.Marshal(&req4)
			return string(reqJSON)
		},
	}
	db := mock.MockDatabase{
		Contents: map[string]string{},
		SetImpl: func(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
			return key, nil
		},
	}
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
	_, consumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	consumer.State.State = pb.ServiceStateValues_START
	consumer.Consume(&delivery)
	assert.Equal(t, 0, len(db.Contents), "it must be one")
	assert.Equal(t, 1, delivery.Rejected, "it must return one rejected msg")
	assert.Nil(t, engine.Started, "engine need to take requests params as params engine")
	assert.Equal(t, db.Contents["9c68bb98-2ba9-412a-a6b1-06963b408e84"], "")
}
