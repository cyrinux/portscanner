package config

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/database"
	"log"
	"os"
)

type Config struct {
	NumConsumers     string
	RmqDbName        string
	RmqServer        string
	RmqDbPassword    string
	DB               database.Database
	ControllerServer string
}

func GetConfig(ctx context.Context) Config {

	config := Config{
		NumConsumers:     os.Getenv("RMQ_CONSUMERS"),
		RmqServer:        os.Getenv("RMQ_DB_SERVER"),
		RmqDbPassword:    os.Getenv("RMQ_DB_PASSWORD"),
		RmqDbName:        os.Getenv("RMQ_DB_NAME"),
		ControllerServer: os.Getenv("CONTROLLER_SERVER"),
	}

	if config.NumConsumers == "0" {
		config.NumConsumers = "5"
	}

	dbConfig := database.DBConfig{
		DBServer:   os.Getenv("DB_SERVER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBDriver:   os.Getenv("DB_DRIVER"),
		DBName:     os.Getenv("DB_NAME"),
	}

	db, err := database.Factory(ctx, dbConfig)
	if err != nil {
		log.Println(err)
	}

	config.DB = db

	return config
}
