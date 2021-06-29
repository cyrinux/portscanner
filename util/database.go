package util

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// RedisConnect return a redis client based on the config
func RedisConnect(ctx context.Context, config config.Config) *redis.Client {
	// Connect to redis for the locker
	redisClient := redis.NewClient(&redis.Options{
		Network:    "tcp",
		Addr:       config.RMQ.Server,
		Password:   config.RMQ.Password,
		DB:         0,
		MaxRetries: 5,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Err(err)
	}

	return redisClient
}
