package config

import (
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/database"
	"os"
)

type Config struct {
	DBDriver     string
	DBServer     string
	NumConsumers string
	RmqDbName    string
	RmqServer    string
	DB           database.Database
}

func GetConfig() Config {
	dbDriver, ok := os.LookupEnv("DB_DRIVER")
	if !ok {
		fmt.Println("DB_DRIVER is not present, fallback on redis")
		dbDriver = "redis"
	}

	db, err := database.Factory(dbDriver)
	if err != nil {
		panic(err)
	}

	config := Config{
		NumConsumers: os.Getenv("RMQ_CONSUMERS"),
		RmqServer:    os.Getenv("RMQ_DB_SERVER"),
		RmqDbName:    os.Getenv("RMQ_DB_NAME"),
		DBServer:     os.Getenv("DB_SERVER"),
		DBDriver:     os.Getenv("DB_DRIVER"),
		DB:           db,
	}

	return config
}
