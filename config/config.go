package config

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/database"
	// "github.com/go-redis/redis/v8"
	"log"
	"os"
)

type Config struct {
	DBDriver      string
	DBServer      string
	NumConsumers  string
	RmqDbName     string
	RmqServer     string
	RmqDbPassword string
	DB            database.Database
	// RedisClient   redis.Client
}

func GetConfig(ctx context.Context) Config {

	config := Config{
		NumConsumers:  os.Getenv("RMQ_CONSUMERS"),
		RmqServer:     os.Getenv("RMQ_DB_SERVER"),
		RmqDbPassword: os.Getenv("RMQ_DB_PASSWORD"),
		RmqDbName:     os.Getenv("RMQ_DB_NAME"),
	}

	if config.NumConsumers == "0" {
		config.NumConsumers = "5"
	}

	dbConfig := database.DBConfig{
		DBServer:   os.Getenv("DB_SERVER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBDriver:   os.Getenv("DB_DRIVER"),
	}

	db, err := database.Factory(ctx, dbConfig)
	if err != nil {
		log.Println(err)
	}

	config.DB = db

	// // Connect to redis for the locker
	// redisClient := *redis.NewClient(&redis.Options{
	// 	Network:  "tcp",
	// 	Addr:     config.RmqServer,
	// 	Password: config.RmqDbPassword,
	// 	DB:       0,
	// })
	// _, err = redisClient.Ping(ctx).Result()
	// if err != nil {
	// 	panic(err)
	// }
	// defer redisClient.Close()

	// config.RedisClient = redisClient

	return config
}
