package broker

import (
	"context"

	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/go-redis/redis/v8"
)

// Broker represent a RMQ broker
type Broker struct {
	Incoming   rmq.Queue
	Pushed     rmq.Queue
	Connection rmq.Connection
	tasktype   string
}

// NewBroker open the broker queues
func NewBroker(ctx context.Context, tasktype string, config config.Config, redisClient *redis.Client) Broker {
	errChan := make(chan error, 10)
	go RmqLogErrors(errChan)

	connection, err := rmq.OpenConnectionWithRedisClient(
		config.RMQ.Name, redisClient, errChan,
	)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("can't open RMQ connection")
	}

	queueIncomingName := fmt.Sprintf("%s-incoming", tasktype)
	queueIncoming, err := connection.OpenQueue(queueIncomingName)
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Stack().Err(err).Msgf("can't open %s queue", queueIncomingName)
	}

	queuePushName := fmt.Sprintf("%s-rejected", tasktype)
	queuePushed, err := connection.OpenQueue(queuePushName)
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Stack().Err(err).Msgf("can't open %s queue", queuePushName)
	}

	queueIncoming.SetPushQueue(queuePushed)

	return Broker{Incoming: queueIncoming, Pushed: queuePushed, Connection: connection, tasktype: tasktype}
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
