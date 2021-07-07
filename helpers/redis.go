package helpers

import (
	"context"

	"github.com/cyrinux/grpcnmapscanner/config"
	redis "github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"strconv"
)

// NewRedisClient return a sentinel redis client from RMQ config params
func NewRedisClient(ctx context.Context, conf config.Config) (*redis.Client, error) {
	rmqDB, err := strconv.ParseInt(conf.RMQ.Database, 10, 0)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse database number")
	}
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:    conf.RMQ.Redis.SentinelServers,
		MasterName:       conf.RMQ.Redis.MasterName,
		Password:         conf.RMQ.Redis.Password,
		SentinelPassword: conf.RMQ.Redis.SentinelPassword,
		DB:               int(rmqDB),
		MaxRetries:       10,
		MinIdleConns:     10,
	})
	return redisClient, nil
}
