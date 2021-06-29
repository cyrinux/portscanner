package config

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Config struct {
	RMQ struct {
		NumConsumers string `default:"5"`
		Name         string `default:"scanner"`
		Server       string `required:"true"`
		Password     string `default:""`
	}
	DBConfig struct {
		Driver   string `default:"redis"`
		Name     string `default:"scanner"`
		Server   string `default:"redis"`
		Password string `default:""`
	}
	ControllerServer string `default:"server:9000"`
}

func GetConfig(ctx context.Context) Config {
	// Init config
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal().Err(errors.Wrap(err, "Unable to process config"))
	}

	return config
}
