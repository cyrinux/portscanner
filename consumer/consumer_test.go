package consumer

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/mock"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CreateMockLock_Create_New_Locker(t *testing.T) {
	mockLock := mock.CreateMockLock()
	assert.NotNil(t, mockLock, "Should not be nil")

	ctx := context.Background()
	conf := config.GetConfig()
	db, _ := mock.CreateMockDatabase(ctx)
	engine := mock.CreateMockEngine()
	tag, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)
	assert.NotEmpty(t, tag, "Tag should not be empty")
	assert.NotNil(t, incConsumer, "Consumer should not be nil")
}

func Test_consumeNow_Misc(t *testing.T) {
	mockLock := mock.CreateMockLock()
	ctx := context.Background()
	conf := config.GetConfig()
	engine := mock.CreateMockEngine()
	db, _ := mock.CreateMockDatabase(ctx)
	_, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock, engine)

	rmqMock := mock.CreateMockRMQ()

	req := pb.ParamsScannerRequest{
		Key:                     "9c68bb98-2ba9-412a-a6b1-06963b408e84",
		Targets:                 "levis.name",
		Ports:                   "443,80,22",
		Timeout:                 3600,
		ServiceVersionDetection: true,
		OsDetection:             true,
		ServiceDefaultScripts:   true,
		FastMode:                true,
		ScanSpeed:               3,
		ProcessPerTarget:        true,
		OpenOnly:                false,
		NetworkChuncked:         true,
		Username:                "user1",
		Role:                    "user",
	}

	consumed := incConsumer.consumeNow(&rmqMock, &req, "")
	assert.NoError(t, consumed, "Should be ....")

}
