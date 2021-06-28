package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisDatabase struct {
	client *redis.Client
}

// createRedisDatabase creates the redis database
func createRedisDatabase(ctx context.Context, config DBConfig) (Database, error) {
	ctx = context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     config.DBServer,
		Password: config.DBPassword,
		DB:       0,
	})
	_, err := client.Ping(ctx).Result() // makes sure database is connected
	if err != nil {
		return nil, &CreateDatabaseError{}
	}
	return &redisDatabase{client: client}, nil
}

func (r *redisDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	_, err := r.client.Set(ctx, key, value, retention).Result()
	if err != nil {
		return generateError("set", err)
	}
	return key, nil
}

func (r *redisDatabase) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return generateError("get", err)
	}
	return value, nil
}

func (r *redisDatabase) Delete(ctx context.Context, key string) (string, error) {
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return generateError("delete", err)
	}
	return key, nil
}

func generateError(operation string, err error) (string, error) {
	if err == redis.Nil {
		return "", &OperationError{operation}
	}
	return "", &DownError{}
}
