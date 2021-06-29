package config

import (
	"context"
	"github.com/cyrinux/grpcnmapscanner/database"
	// "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type Config struct {
	RmqNumConsumers  string
	RmqDbName        string
	RmqServer        string
	RmqDbPassword    string
	DB               database.Database
	ControllerServer string
}

func GetConfig(ctx context.Context) Config {

	config := Config{
		RmqServer:        os.Getenv("RMQ_DB_SERVER"),
		RmqDbPassword:    os.Getenv("RMQ_DB_PASSWORD"),
		RmqDbName:        os.Getenv("RMQ_DB_NAME"),
		ControllerServer: os.Getenv("CONTROLLER_SERVER"),
	}

	rmqNumConsumers := "5"
	if tmpRmqNumConsumers, ok := os.LookupEnv("RMQ_CONSUMERS"); ok {
		if tmpRmqNumConsumers == "0" {
			config.RmqNumConsumers = rmqNumConsumers
		}
		config.RmqNumConsumers = tmpRmqNumConsumers
	}

	dbConfig := database.DBConfig{
		DBServer:   os.Getenv("DB_SERVER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBDriver:   os.Getenv("DB_DRIVER"),
		DBName:     os.Getenv("DB_NAME"),
	}

	db, err := database.Factory(ctx, dbConfig)
	if err != nil {
		log.Fatal().Err(err)
	}

	config.DB = db

	return config
}
