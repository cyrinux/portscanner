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
	Database         int      `default:"0"`
	MasterName       string   `default:"mymaster" split_words:"true"`
	Password         string   `default:""`
	MasterPassword   string   `default:"" split_words:"true" `
	SentinelPassword string   `default:"" split_words:"true"`
	SentinelServers  []string `default:"redis-sentinel:26379,redis-sentinel:26380,redis-sentinel:26381" split_words:"true"`
	Server           string   `default:"redis:6379"`
}

// DBConfig is the global database struct
type DBConfig struct {
	Driver string `default:"redis-sentinel"`
	Redis  Redis
}

type NMAPConfig struct {
	TorServer string `default:"socks4://tor:9050" split_words:"true"`
}

// RMQConfig  contains the redis broker config
type RMQConfig struct {
	Name               string        `default:"broker"`
	Database           int           `default:"0"`
	NumConsumers       int64         `default:"5" split_words:"true"`
	ReturnerLimit      int64         `default:"200" split_words:"true"`
	PollDuration       time.Duration `default:"100ms" split_words:"true"`
	PollDurationPushed time.Duration `default:"5000ms" split_words:"true"`
	ConsumeDuration    time.Duration `default:"1000ms" split_words:"true"`
	Redis              Redis
}

// LoggerConfig contains the logger config
type LoggerConfig struct {
	Debug  bool `default:"false"`
	Pretty bool `default:"true"`
}

// PrometheusConfig contains the prometheus config
type PrometheusConfig struct {
	Server string `default:"prometheus:8140"`
}

// GlobalConfig contains some others params
type Global struct {
	BackendServer      string `default:"server:9001" split_words:"true"`
	FrontendListenPort int    `default:"9000" split_words:"true"`
	BackendListenPort  int    `default:"9001" split_words:"true"`
}

// JWT contains JWT config
type JWT struct {
	SecretKey     string        `default:"secret" split_words:"true"`
	TokenDuration time.Duration `default:"15m" split_words:"true"`
}

// Config is the main global config struct
type Config struct {
	DB         DBConfig
	RMQ        RMQConfig
	NMAP       NMAPConfig
	Logger     LoggerConfig
	Prometheus PrometheusConfig
	Global
	Backend  GRPCBackend
	Frontend GRPCFrontend
}

// GRPCBackendServer is the GRPC backend endpoint config
type GRPCBackend struct {
	CAFile         string `default:"/etc/scanner/cert/ca-cert.pem" split_words:"true"`
	ServerCertFile string `default:"/etc/scanner/cert/server-cert.pem" split_words:"true"`
	ServerKeyFile  string `default:"/etc/scanner/cert/server-key.pem" split_words:"true"`
	ClientCertFile string `default:"/etc/scanner/cert/client-worker-cert.pem" split_words:"true"`
	ClientKeyFile  string `default:"/etc/scanner/cert/client-worker-key.pem" split_words:"true"`
	JWT            JWT
}

// GRPCBackendServer is the GRPC backend endpoint config
type GRPCFrontend struct {
	CAFile         string `default:"/etc/scanner/cert/ca-cert.pem" split_words:"true"`
	ServerCertFile string `default:"/etc/scanner/cert/server-cert.pem" split_words:"true"`
	ServerKeyFile  string `default:"/etc/scanner/cert/server-key.pem" split_words:"true"`
	JWT            JWT
}

// GetConfig get the configuration
func GetConfig() Config {

	// Init config
	var conf Config

	if err := envconfig.Process("", &conf); err != nil {
		log.Fatal().Stack().Err(errors.Wrap(err, "unable to process config"))
	}

	return conf
}
