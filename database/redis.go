package database

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisDatabase struct {
	client *redis.Client
}

// createRedisDatabase creates the redis database
func createRedisDatabase(ctx context.Context, conf config.DBConfig) (Database, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Server,
		Password:     conf.Redis.Password,
		DB:           conf.Redis.Database,
		MaxRetries:   10,
		MinIdleConns: 10,
	})
	_, err := client.Ping(ctx).Result() // makes sure database is connected
	if err != nil {
		return nil, &CreateDatabaseError{}
	}
	return redisDatabase{client: client}, nil
}

func (r redisDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	_, err := r.client.Set(ctx, key, value, retention).Result()
	if err != nil {
		return generateError("set", err)
	}
	return key, nil
}

func (r redisDatabase) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return generateError("get", err)
	}
	return value, nil
}

func (r redisDatabase) GetAll(ctx context.Context, key string) ([]string, error) {
	arr := make([]string, 1000)
	iter := r.client.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		arr = append(arr, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return arr, nil
}

func (r redisDatabase) Delete(ctx context.Context, key string) (string, error) {
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return generateError("delete", err)
	}
	return key, nil
}
