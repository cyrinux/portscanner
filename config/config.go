package config

import (
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
		Password     string `default:""`
		Name         string `default:"grpcnmapscanner"`
		NumConsumers int64  `default:"5" split_words:"true"`
	}
	Logger struct {
		Debug bool `default:"false"`
	}
	ControllerServer string `default:"server:9000" split_words:"true"`
}

func GetConfig() Config {
	// Init config
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal().Err(errors.Wrap(err, "Unable to process config"))
	}

	return config
}
