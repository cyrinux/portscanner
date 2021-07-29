package consumer

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CreateMockLock_Create_New_Locker(t *testing.T) {
	mockLock := mock.CreateMockLock()
	assert.NotNil(t, mockLock, "Should not be nil")

	ctx := context.Background()
	conf := config.GetConfig()
	db, _ := mock.CreateMockDatabase(ctx)
	tag, incConsumer := New(ctx, db, 1, "nmap", conf.NMAP, "incoming", mockLock)
	assert.NotEmpty(t, tag, "Tag should not be empty")
	assert.NotNil(t, incConsumer, "Consumer should not be nil")
}
