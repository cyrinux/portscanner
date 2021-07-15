package helpers

import (
	"context"

	"github.com/cyrinux/grpcnmapscanner/config"
	redis "github.com/go-redis/redis/v8"
)

// NewRedisClient return a sentinel redis client from RMQ config params
func NewRedisClient(ctx context.Context, conf config.Config) (*redis.Client, error) {
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:    conf.RMQ.Redis.SentinelServers,
		MasterName:       conf.RMQ.Redis.MasterName,
		Password:         conf.RMQ.Redis.Password,
		SentinelPassword: conf.RMQ.Redis.SentinelPassword,
		DB:               conf.RMQ.Database,
		MaxRetries:       10,
	})
	return redisClient, nil
}
