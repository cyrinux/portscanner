package database

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

type redisSentinelDatabase struct {
	mutex  sync.RWMutex
	client *redis.Client
}

// createRedisSentinelDatabase creates the redis database
func createRedisSentinelDatabase(ctx context.Context, conf config.DBConfig) (Database, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:    conf.Redis.SentinelServers,
		MasterName:       conf.Redis.MasterName,
		Password:         conf.Redis.Password,
		SentinelPassword: conf.Redis.SentinelPassword,
		DB:               conf.Redis.Database,
		MaxRetries:       10,
		MinIdleConns:     10,
	})
	_, err := client.Ping(ctx).Result() // makes sure database is connected
	if err != nil {
		return nil, &CreateDatabaseError{}
	}
	return &redisSentinelDatabase{client: client}, nil
}

func (r *redisSentinelDatabase) Set(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, err := r.client.Set(ctx, key, value, retention).Result()
	if err != nil {
		return generateError("set", err)
	}
	return key, nil
}

// func (r *redisSentinelDatabase) WatchAndSet(ctx context.Context, key string, value string, retention time.Duration) (string, error) {
// 	r.mutex.Lock()
// 	defer r.mutex.Unlock()
// 	_, err := r.client.Set(ctx, key, value, retention).Result()
// 	if err != nil {
// 		return generateError("set", err)
// 	}
// 	return key, nil
// }

func (r *redisSentinelDatabase) Get(ctx context.Context, key string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return generateError("get", err)
	}
	return value, nil
}

func (r *redisSentinelDatabase) GetAll(ctx context.Context, key string) ([]string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	arr := make([]string, 0)
	iter := r.client.Scan(ctx, 0, key, 0).Iterator()
	for iter.Next(ctx) {
		arr = append(arr, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return arr, nil
}

func (r *redisSentinelDatabase) Delete(ctx context.Context, key string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return generateError("delete", err)
	}
	return key, nil
}
