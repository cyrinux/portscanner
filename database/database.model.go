package database

import (
	"time"
)

type DBConfig struct {
	DBDriver string
	DBName   string
	DBServer string
}

// Database abstraction
type Database interface {
	Set(key string, value string, retention time.Duration) (string, error)
	Get(key string) (string, error)
	Delete(key string) (string, error)
}

// Factory looks up acording to the databaseName the database implementation
func Factory(config DBConfig) (Database, error) {
	switch config.DBDriver {
	case "redis":
		return createRedisDatabase(config)
	default:
		return nil, &NotImplementedDatabaseError{config.DBDriver}
	}
}
