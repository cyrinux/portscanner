package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"time"
)

// Redis
type Redis struct {
	Name             string   `default:"scanner"`
	Database         string   `default:"0"`
	MasterName       string   `default:"mymasterrrr" split_words:"true"`
	Password         string   `default:""`
	MasterPassword   string   `default:"" split_words:"true" `
	SentinelPassword string   `default:"" split_words:"true"`
	SentinelServers  []string `default:"redis-sentinel:26379,redis-sentinel:26380,redis-sentinel:26381" split_words:"true"`
	Server           string   `default:"redis:6379"`
}

// DBConfig struct
type DBConfig struct {
	Driver string `default:"redis-sentinel"`
	Redis  Redis
}

type RMQConfig struct {
	Name               string        `default:"broker"`
	Database           string        `default:"0"`
	NumConsumers       int64         `default:"5" split_words:"true"`
	ReturnerLimit      int64         `default:"200" split_words:"true"`
	PollDuration       time.Duration `default:"100ms" split_words:"true"`
	PollDurationPushed time.Duration `default:"5000ms" split_words:"true"`
	ConsumeDuration    time.Duration `default:"1000ms" split_words:"true"`
	Redis              Redis
}

type LoggerConfig struct {
	Debug  bool `default:"false"`
	Pretty bool `default:"true"`
}

type PrometheusConfig struct {
	Server string `default:"prometheus:8140"`
}

type GlobalConfig struct {
	ControllerServer string `default:"server:9000" split_words:"true"`
}

type Config struct {
	DB         DBConfig
	RMQ        RMQConfig
	Logger     LoggerConfig
	Prometheus PrometheusConfig
	Global     GlobalConfig
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
