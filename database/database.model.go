package database

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"time"
)

// Database abstraction
type Database interface {
	Set(ctx context.Context, key string, value string, retention time.Duration) (string, error)
	Get(ctx context.Context, key string) (string, error)
	GetAll(ctx context.Context, key string) ([]string, error)
	Delete(ctx context.Context, key string) (string, error)
}

// Factory looks up acording to the databaseName the database implementation
func Factory(ctx context.Context, conf config.Config) (Database, error) {
	switch conf.DB.Driver {
	case "redis":
		return createRedisDatabase(ctx, conf.DB)
	case "redis-sentinel":
		return createRedisSentinelDatabase(ctx, conf.DB)
	default:
		return nil, &NotImplementedDatabaseError{conf.DB.Driver}
	}
}
