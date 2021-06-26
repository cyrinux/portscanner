package database

import (
	"context"
	"time"
)

type DBConfig struct {
	DBDriver   string
	DBName     string
	DBServer   string
	DBPassword string
}

// Database abstraction
type Database interface {
	Set(ctx context.Context, key string, value string, retention time.Duration) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) (string, error)
}

// Factory looks up acording to the databaseName the database implementation
func Factory(ctx context.Context, config DBConfig) (Database, error) {
	switch config.DBDriver {
	case "redis":
		return createRedisDatabase(ctx, config)
	default:
		return nil, &NotImplementedDatabaseError{config.DBDriver}
	}
}
