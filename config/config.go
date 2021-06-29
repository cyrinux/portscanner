package config

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	DB struct {
		Driver   string `default:"redis"`
		Server   string `default:"redis:6379"`
		Name     string `default:"scanner"`
		Password string `default:""`
	}
	RMQ struct {
		Server       string `default:"redis:6379"`
		Name         string `default:"scanner"`
		Password     string `default:""`
		NumConsumers int64  `default:"5" split_words:"true"`
	}
	ControllerServer string `default:"server:9000" split_words:"true"`
}

func GetConfig(ctx context.Context) Config {
	// Init config
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal().Err(errors.Wrap(err, "Unable to process config"))
	}
	fmt.Printf("%+v", config)

	return config
}
