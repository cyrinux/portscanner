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
	Delete(ctx context.Context, key string) (string, error)
}

// Factory looks up acording to the databaseName the database implementation
func Factory(ctx context.Context, config config.Config) (Database, error) {
	switch config.DBConfig.Driver {
	case "redis":
		return createRedisDatabase(ctx, config)
	default:
		return nil, &NotImplementedDatabaseError{config.DBConfig.Driver}
	}
}
