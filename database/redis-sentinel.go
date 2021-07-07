package database

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

type redisSentinelDatabase struct {
	client *redis.Client
}

// createRedisSentinelDatabase creates the redis database
func createRedisSentinelDatabase(ctx context.Context, conf config.DBConfig) (Database, error) {
	database, _ := strconv.ParseInt(conf.Redis.Database, 10, 0)
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:    conf.Redis.SentinelServers,
		MasterName:       conf.Redis.MasterName,
		Password:         conf.Redis.Password,
		SentinelPassword: conf.Redis.SentinelPassword,
		DB:               int(database),
		MaxRetries:       10,
		MinIdleConns:     10,
	})
	_, err := client.Ping(ctx).Result() // makes sure database is connected
	if err != nil {
		return nil, &CreateDatabaseError{}
	}
	return redisSentinelDatabase{client: client}, nil
}

func (r redisSentinelDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	_, err := r.client.Set(ctx, key, value, retention).Result()
	if err != nil {
		return generateError("set", err)
	}
	return key, nil
}

func (r redisSentinelDatabase) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return generateError("get", err)
	}
	return value, nil
}

func (r redisSentinelDatabase) Delete(ctx context.Context, key string) (string, error) {
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return generateError("delete", err)
	}
	return key, nil
}
