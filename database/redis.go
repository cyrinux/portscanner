package database

import (
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis"
	"log"
	"os"
	"time"
)

type redisDatabase struct {
	client *redis.Client
}

// CreateRedisDatabase creates the redis database
func createRedisDatabase(config config.Config) (Database, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.DBServer,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := client.Ping().Result() // makes sure database is connected
	if err != nil {
		return nil, &CreateDatabaseError{}
	}
	return &redisDatabase{client: client}, nil
}

func (r *redisDatabase) Set(key string, value string, retention time.Duration) (string, error) {
	_, err := r.client.Set(key, value, retention).Result()
	if err != nil {
		return generateError("set", err)
	}
	return key, nil
}

func (r *redisDatabase) Get(key string) (string, error) {
	value, err := r.client.Get(key).Result()
	log.Printf("DEBUG %v - %v - %v", key, value, err)
	if err != nil {
		return generateError("get", err)
	}
	return value, nil

}
func (r *redisDatabase) Delete(key string) (string, error) {
	_, err := r.client.Del(key).Result()
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
