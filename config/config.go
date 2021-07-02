package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

// Config struct
type Config struct {
	DB struct {
		Driver   string `default:"redis"`
		Server   string `default:"redis:6379"`
		Name     string `default:"scanner"`
		Password string `default:""`
	}
	RMQ struct {
		Server             string        `default:"redis:6379"`
		Password           string        `default:""`
		Name               string        `default:"grpcnmapscanner"`
		NumConsumers       int64         `default:"5" split_words:"true"`
		ReturnerLimit      int64         `default:"1000" split_words:"true"`
		ReportBatchSize    int64         `default:"10000" split_words:"true"`
		PollDuration       time.Duration `default:"100" split_words:"true"`
		PollDurationPushed time.Duration `default:"5000" split_words:"true"`
		ConsumeDuration    time.Duration `default:"1000" split_words:"true"`
	}
	Logger struct {
		Debug  bool `default:"false"`
		Pretty bool `default:"true"`
	}
	Global struct {
		ControllerServer string `default:"server:9000" split_words:"true"`
	}
}

// GetConfig get the configuration
func GetConfig() Config {

	log.Info().Msg("reading config")
	// Init config
	var conf Config

	if err := envconfig.Process("", &conf); err != nil {
		log.Fatal().Stack().Err(errors.Wrap(err, "unable to process config"))
	}
	return conf
}
