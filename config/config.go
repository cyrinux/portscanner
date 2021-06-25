package config

import (
	"context"
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

func GetConfig(ctx context.Context) Config {

	config := Config{
		NumConsumers: os.Getenv("RMQ_CONSUMERS"),
		RmqServer:    os.Getenv("RMQ_DB_SERVER"),
		RmqDbName:    os.Getenv("RMQ_DB_NAME"),
	}

	if config.NumConsumers == "0" {
		config.NumConsumers = "5"
	}

	dbConfig := database.DBConfig{
		DBServer: os.Getenv("DB_SERVER"),
		DBDriver: os.Getenv("DB_DRIVER"),
	}

	db, err := database.Factory(ctx, dbConfig)
	if err != nil {
		panic(err)
	}

	config.DB = db

	return config
}

func GetDBConfig() database.DBConfig {

	dbConfig := database.DBConfig{
		DBServer: os.Getenv("DB_SERVER"),
		DBDriver: os.Getenv("DB_DRIVER"),
	}

	return dbConfig
}
