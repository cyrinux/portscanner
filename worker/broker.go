package worker

import (
	"context"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// Broker represent a RMQ broker
type Broker struct {
	config     config.Config
	incoming   rmq.Queue
	pushed     rmq.Queue
	connection rmq.Connection
}

// NewBroker open the broker queues
func newBroker(ctx context.Context, config config.Config, redisClient *redis.Client) *Broker {
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnectionWithRedisClient(
		config.RMQ.Name, redisClient, errChan,
	)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("can't open RMQ connection")
	}

	queueIncoming, err := connection.OpenQueue("tasks")
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Stack().Err(err).Msg("can't open tasks queue")
	}

	queuePushed, err := connection.OpenQueue("tasks-rejected")
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Stack().Err(err).Msg("can't open tasks-rejected queue")
	}

	queueIncoming.SetPushQueue(queuePushed)

	return &Broker{incoming: queueIncoming, pushed: queuePushed, connection: connection, config: config}
}

// RmqLogErrors display the rmq errors log
func rmqLogErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Error().Msgf("heartbeat error (limit): %v", err.RedisErr)
			} else {
				log.Error().Msgf("heartbeat error: %v", err.RedisErr)
			}
		case *rmq.ConsumeError:
			log.Error().Msgf("consume error: %v", err.RedisErr)
		case *rmq.DeliveryError:
			log.Error().Msgf("delivery error: %v %v", err.Delivery, err.RedisErr)
		default:
			log.Error().Msgf("other error: %v", err.Error)
		}
	}
}
