package broker

import (
	"context"

	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/logger"
	"github.com/go-redis/redis/v8"
	"time"
)

var (
	conf = config.GetConfig()
	log  = logger.New(conf.Logger.Debug, conf.Logger.Pretty)
)

// Broker represent a RMQ broker
type Broker struct {
	ctx        context.Context
	Incoming   rmq.Queue
	Pushed     rmq.Queue
	Connection rmq.Connection
	Stats      rmq.Stats
	taskType   string
}

// Returner define a returner
type Returner struct {
	Queue    rmq.Queue
	Returned chan int64
}

// New open the broker queues
func New(ctx context.Context, taskType string, conf config.RMQConfig, redisClient *redis.Client) *Broker {
	errChan := make(chan error, 10)

	go RmqLogErrors(errChan)

	if redisClient == nil {
		redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    conf.Redis.SentinelServers,
			MasterName:       conf.Redis.MasterName,
			Password:         conf.Redis.Password,
			SentinelPassword: conf.Redis.SentinelPassword,
			DB:               conf.Database,
			MaxRetries:       3,
			MinRetryBackoff:  100 * time.Millisecond,
			MaxRetryBackoff:  1 * time.Second,
		})
	}

	var connection rmq.Connection
	var err error
	wait := 500 * time.Millisecond
	for {
		connection, err = rmq.OpenConnectionWithRedisClient(
			conf.Name, redisClient, errChan,
		)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("can't open RMQ connection, retrying in %v...", wait)
			time.Sleep(wait)
		} else {
			log.Info().Msg("connected to RMQ database")
			break
		}
	}

	queueIncomingName := fmt.Sprintf("%s-incoming", taskType)
	queueIncoming, err := connection.OpenQueue(queueIncomingName)
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Stack().Err(err).Msgf("can't open %s queue", queueIncomingName)
	}

	queuePushName := fmt.Sprintf("%s-rejected", taskType)
	queuePushed, err := connection.OpenQueue(queuePushName)
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Stack().Err(err).Msgf("can't open %s queue", queuePushName)
	}

	queueIncoming.SetPushQueue(queuePushed)

	return &Broker{
		ctx:        context.Background(),
		Incoming:   queueIncoming,
		Pushed:     queuePushed,
		Connection: connection,
		taskType:   taskType,
	}
}

// RmqLogErrors display the rmq errors log
func RmqLogErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Error().Stack().Err(err).Msg("heartbeat error (limit)")
			} else {
				log.Error().Stack().Err(err).Msg("heartbeat error")
			}
		case *rmq.ConsumeError:
			log.Error().Stack().Err(err).Msg("consume error")
		case *rmq.DeliveryError:
			log.Error().Stack().Err(err).Msg("delivery error")
		default:
			log.Error().Stack().Err(err).Msg("other error")
		}
	}
}

// GetStats return RMQ broker connection stats
func (broker *Broker) GetStats() (rmq.Stats, error) {
	queues, err := broker.Connection.GetOpenQueues()
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
	}
	stats, err := broker.Connection.CollectStats(queues)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
	}

	return stats, err
}

// Cleaner clean the queues
func (broker *Broker) Cleaner() {
	cleaner := rmq.NewCleaner(broker.Connection)
	for range time.Tick(1000 * time.Millisecond) {
		returned, err := cleaner.Clean()
		if err != nil {
			log.Error().Stack().Err(err).Msgf("failed to clean")
			continue
		}
		if returned > 0 {
			log.Debug().Msgf("broker cleaned %d tasks", returned)
		}
	}
}
