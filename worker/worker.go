package worker

import (
	"context"
	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/go-redis/redis/v8"
	// "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Worker define the worker struct
type Worker struct {
	ctx         context.Context
	broker      *Broker
	config      config.Config
	locker      *redislock.Client
	redisClient *redis.Client
	state       proto.ScannerServiceControl
	db          database.Database
}

// Broker represent a RMQ broker
type Broker struct {
	incoming   rmq.Queue
	pushed     rmq.Queue
	connection rmq.Connection
}

// NewWorker create a new worker and init the database connection
func NewWorker(config config.Config, ctx context.Context) *Worker {
	// Storage database init
	db, err := database.Factory(context.Background(), config)
	if err != nil {
		log.Fatal().Err(err)
	}

	redisClient := util.RedisConnect(ctx, config)
	broker := NewBroker(ctx, config, *redisClient)
	locker := redislock.New(redisClient)
	return &Worker{
		config:      config,
		ctx:         ctx,
		broker:      broker,
		redisClient: redisClient,
		locker:      locker,
		db:          db,
	}
}

// handleSignal handle the worker exit signal
func (worker *Worker) handleSignal() {
	// open tasks queues and connection
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()
	worker.stopConsuming() // wait for all Consume() calls to finish
}

// StartWorker start a scanner worker
func (worker *Worker) StartWorker() {
	// handle exit signal
	go worker.handleSignal()

	// start the worker on boot
	worker.state.State = proto.ScannerServiceControl_START
	worker.startConsuming()
	// watch the control server and stop/start service
	worker.StreamControlService()
}

// StreamControlService return the workers status and control them
func (worker *Worker) StreamControlService() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(
		worker.config.ControllerServer,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		log.Fatal().Msgf("could not connect to controller: %s", err)
	}
	defer conn.Close()

	client := proto.NewScannerServiceClient(conn)
	getState := &proto.ScannerServiceControl{State: 0}
	for {
		// wait before try to reconnect
		reconnectTime := 5 * time.Second
		log.Info().Msgf("trying to connect in %v to server control", reconnectTime)
		time.Sleep(reconnectTime)
		stream, err := client.StreamServiceControl(worker.ctx, getState)
		if err != nil {
			break
		}
		log.Info().Msg("connected to server control")

		for {
			// cpu cooling
			time.Sleep(500 * time.Millisecond)

			serviceControl, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			if serviceControl == nil {
				continue
			} else {
				if serviceControl.State == 1 { //proto.ScannerServiceControl_START
					worker.state.State = proto.ScannerServiceControl_START
				} else if serviceControl.State == 2 { //proto.ScannerServiceControl_STOP
					worker.state.State = proto.ScannerServiceControl_STOP
				}
			}
		}
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(queue rmq.Queue) {
	log.Print("Starting the returner")
	go func() {
		for {
			// Try to obtain lock.
			lock, err := worker.locker.Obtain(worker.ctx, "returner", 10*time.Second, nil)
			if err != nil && err != redislock.ErrNotObtained {
				log.Error().Err(err)
			} else if err != redislock.ErrNotObtained {
				// Sleep and check the remaining TTL.
				if ttl, err := lock.TTL(worker.ctx); err != nil {
					log.Error().Msgf("Returner error: %v: ttl: %v", err, ttl)
				} else if ttl > 0 {
					// Yay, I still have my lock!
					returned, _ := queue.ReturnRejected(returnerLimit)
					if returned > 0 {
						log.Error().Msgf("Returner success requeue %v tasks messages to incoming", returned)
					}
					lock.Refresh(worker.ctx, 5*time.Second, nil)
				}
			}
			// cpu cooling
			time.Sleep(1 * time.Second)
		}
	}()
}

// NewBroker open the broker queues
func NewBroker(
	ctx context.Context, config config.Config, redisClient redis.Client) *Broker {
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnectionWithRedisClient(
		config.RMQ.Name, &redisClient, errChan,
	)
	if err != nil {
		log.Fatal().Err(err)
	}

	queueIncoming, err := connection.OpenQueue("tasks")
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Err(err)
	}

	queuePushed, err := connection.OpenQueue("tasks-rejected")
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		log.Fatal().Err(err)
	}

	queueIncoming.SetPushQueue(queuePushed)

	return &Broker{incoming: queueIncoming, pushed: queuePushed, connection: connection}
}

func (worker *Worker) startConsuming() {
	numConsumers, err := strconv.ParseInt(
		worker.config.RMQ.NumConsumers, 10, 0,
	)
	if err != nil {
		log.Fatal().Err(err)
	}
	numConsumers++                    // we got one consumer for the returned, lets add 1 more
	prefetchLimit := numConsumers + 1 // prefetchLimit need to be > numConsumers

	err = worker.broker.incoming.StartConsuming(prefetchLimit, pollDuration)
	if err != nil {
		log.Error().Err(err)
	}

	err = worker.broker.pushed.StartConsuming(prefetchLimit, pollDurationPushed)
	if err != nil {
		log.Error().Err(err)
	}

	for i := 0; i < int(numConsumers)+1; i++ {
		tag, consumer := NewConsumer(worker, i, "incoming")
		if _, err := worker.broker.incoming.AddConsumer(tag, consumer); err != nil {
			log.Error().Err(err)
		}
		tag, consumer = NewConsumer(worker, i, "push")
		if _, err := worker.broker.pushed.AddConsumer(tag, consumer); err != nil {
			log.Error().Err(err)
		}
	}

	worker.startReturner(worker.broker.incoming)
}

func (worker *Worker) stopConsuming() {
	log.Info().Msg("Stop consumming...")
	<-worker.broker.incoming.StopConsuming()
	<-worker.broker.pushed.StopConsuming()
	<-worker.broker.connection.StopAllConsuming() // wait for all Consume() calls to finish
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
