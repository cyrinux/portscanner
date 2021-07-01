package worker

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// Worker define the worker struct
type Worker struct {
	sync.Mutex
	ctx         context.Context
	broker      broker.Broker
	config      config.Config
	locker      *redislock.Client
	redisClient *redis.Client
	state       proto.ScannerServiceControl
	consumers   []*Consumer
	db          database.Database
}

// NewWorker create a new worker and init the database connection
func NewWorker(config config.Config) *Worker {
	ctx := context.Background()

	// Storage database init
	db, err := database.Factory(context.TODO(), config)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	redisClient := util.RedisConnect(context.TODO(), config)
	broker := broker.NewBroker(context.TODO(), config, redisClient)
	locker := redislock.New(redisClient)
	consumers := make([]*Consumer, 0)

	return &Worker{
		config:      config,
		ctx:         ctx,
		broker:      broker,
		locker:      locker,
		consumers:   consumers,
		db:          db,
		redisClient: redisClient,
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
				if serviceControl.State == 1 && worker.state.State != 1 { //proto.ScannerServiceControl_START
					worker.state.State = proto.ScannerServiceControl_START
					worker.startConsuming()
				} else if serviceControl.State == 2 && worker.state.State != 2 { //proto.ScannerServiceControl_STOP
					worker.state.State = proto.ScannerServiceControl_STOP
					worker.stopConsuming()
				}
			}
		}
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(queue rmq.Queue) {
	log.Info().Msg("Starting the returner")
	go func() {
		for {
			// Try to obtain lock.
			lock, err := worker.locker.Obtain(worker.ctx, "returner", 10*time.Second, nil)
			if err != nil && err != redislock.ErrNotObtained {
				log.Error().Err(err).Msg("can't obtain returner lock")
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

func (worker *Worker) startConsuming() {
	numConsumers := worker.config.RMQ.NumConsumers
	numConsumers++                    // we got one consumer for the returned, lets add 1 more
	prefetchLimit := numConsumers + 1 // prefetchLimit need to be > numConsumers

	worker.broker = broker.NewBroker(context.TODO(), worker.config, worker.redisClient)

	err := worker.broker.Incoming.StartConsuming(prefetchLimit, pollDuration)
	if err != nil {
		log.Error().Stack().Err(err).Msg("queue incoming consume error")
	}

	err = worker.broker.Pushed.StartConsuming(prefetchLimit, pollDurationPushed)
	if err != nil {
		log.Error().Stack().Err(err).Msg("queue pushed consume error")
	}

	for i := 0; i < int(numConsumers)+1; i++ {
		tag, consumer := NewConsumer(worker.db, i, "incoming")
		if _, err := worker.broker.Incoming.AddConsumer(tag, consumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
		consumer.state.State = proto.ScannerServiceControl_START

		// store consumer pointer to the worker struct
		worker.consumers = append(worker.consumers, consumer)

		tag, consumer = NewConsumer(worker.db, i, "push")
		if _, err := worker.broker.Pushed.AddConsumer(tag, consumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	worker.startReturner(worker.broker.Incoming)
}

func (worker *Worker) stopConsuming() {
	log.Info().Msg("Stop consumming...")
	for _, consumer := range worker.consumers {
		if consumer.engine != nil {
			log.Info().Msgf("cancelling consumer %v", consumer.name)
			consumer.state.State = worker.state.State
			consumer.cancel()
		}
	}
	<-worker.broker.Incoming.StopConsuming()
	<-worker.broker.Pushed.StopConsuming()
	<-worker.broker.Connection.StopAllConsuming() // wait for all Consume() calls to finish
}
