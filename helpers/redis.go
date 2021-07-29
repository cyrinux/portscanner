package helpers

import (
	"context"

	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/logger"
	redis "github.com/go-redis/redis/v8"
	"time"
)

var (
	conf = config.GetConfig().Logger
	log  = logger.New(conf.Debug, conf.Pretty)
)

// RedisClient is a redis client struct
type RedisClient struct {
	ctx    context.Context
	conf   config.Config
	client *redis.Client
}

// NewRedisClient return a sentinel redis client from RMQ config params
func NewRedisClient(ctx context.Context, conf config.Config) *RedisClient {
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:    conf.RMQ.Redis.SentinelServers,
		MasterName:       conf.RMQ.Redis.MasterName,
		Password:         conf.RMQ.Redis.Password,
		SentinelPassword: conf.RMQ.Redis.SentinelPassword,
		DB:               conf.RMQ.Database,
		MaxRetries:       3,
		MinRetryBackoff:  1 * time.Second,
		MaxRetryBackoff:  3 * time.Second,
		DialTimeout:      5 * time.Second,
	})
	return &RedisClient{ctx: ctx, client: redisClient, conf: conf}
}

// Connect is use to connect to redis
func (rc *RedisClient) Connect() *redis.Client {
	wait := 500 * time.Millisecond
	var redisClient RedisClient
	for {
		redisClient = *NewRedisClient(rc.ctx, rc.conf)
		err := redisClient.client.Ping(rc.ctx)
		if err.Err() != nil {
			log.Error().Stack().Err(err.Err()).Msgf("cannot connected to redis, retrying in %v...", wait)
			time.Sleep(wait)
		} else {
			log.Info().Msg("connected to redis database")
			break
		}
	}
	return redisClient.client
}
