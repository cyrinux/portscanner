package util

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
)

// RedisConnect return a redis client based on the config
func RedisConnect(ctx context.Context, config config.Config) *redis.Client {
	// Connect to redis for the locker
	redisClient := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     config.RmqServer,
		Password: config.RmqDbPassword,
		DB:       0,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return redisClient
}
