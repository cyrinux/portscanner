package worker

import (
	"context"
	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"io"
	"time"
)

// Worker define the worker struct
type Worker struct {
	ctx         context.Context
	broker      broker.Broker
	config      config.Config
	locker      *redislock.Client
	redisClient *redis.Client
	state       proto.ScannerServiceControl
	consumers   []*Consumer
	db          database.Database
	name        string
}

// NewWorker create a new worker and init the database connection
func NewWorker(ctx context.Context, config config.Config, name string) *Worker {
	log.Info().Msgf("starting worker %s", name)

	// Storage database init
	db, err := database.Factory(context.Background(), config)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	redisClient := util.RedisConnect(context.Background(), config)
	broker := broker.NewBroker(context.Background(), name, config, redisClient)
	locker := redislock.New(redisClient)
	consumers := make([]*Consumer, 0)

	return &Worker{
		config:      config,
		name:        name,
		ctx:         ctx,
		broker:      broker,
		locker:      locker,
		consumers:   consumers,
		db:          db,
		redisClient: redisClient,
	}
}

// StartWorker start a scanner worker
func (worker *Worker) StartWorker() {

	// start the worker on boot
	worker.startConsuming()

	// watch the control server and stop/start service
	worker.StreamControlService()
}

// StreamControlService return the workers status and control them
func (worker *Worker) StreamControlService() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(
		worker.config.Global.ControllerServer,
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
		log.Debug().Msgf("trying to connect in %v to server control", reconnectTime)
		time.Sleep(reconnectTime)
		stream, err := client.StreamServiceControl(worker.ctx, getState)
		if err != nil {
			break
		}
		log.Debug().Msg("connected to server control")

		for {
			serviceControl, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			if serviceControl.State == 1 && worker.state.State != 1 { //proto.ScannerServiceControl_START
				worker.state.State = proto.ScannerServiceControl_START
				worker.startConsuming()
			} else if serviceControl.State == 2 && worker.state.State != 2 { //proto.ScannerServiceControl_STOP
				worker.state.State = proto.ScannerServiceControl_STOP
				worker.StopConsuming()
			}

			// cpu cooling
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(queue rmq.Queue) {
	log.Info().Msg("starting the returner routine")
	go func() {
		conf := worker.config
		for {
			// Try to obtain lock.
			lock, err := worker.locker.Obtain(worker.ctx, "returner", 10*time.Second, nil)
			if err != nil && err != redislock.ErrNotObtained {
				log.Error().Stack().Err(err).Msg("can't obtain returner lock")
			} else if err != redislock.ErrNotObtained {
				// Sleep and check the remaining TTL.
				if ttl, err := lock.TTL(worker.ctx); err != nil {
					log.Error().Stack().Err(err).Msgf("returner error, ttl: %v", ttl)
				} else if ttl > 0 {
					// Yay, I still have my lock!
					returned, _ := queue.ReturnRejected(conf.RMQ.ReturnerLimit)
					if returned > 0 {
						log.Info().Msgf("returner success requeue %v tasks messages to incoming", returned)
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
	conf := worker.config
	numConsumers := conf.RMQ.NumConsumers
	prefetchLimit := numConsumers + 1 // prefetchLimit need to be > numConsumers
	log.Info().Msgf("start consuming %s with %v consumers...", worker.name, numConsumers)

	worker.broker = broker.NewBroker(context.TODO(), worker.name, worker.config, worker.redisClient)

	err := worker.broker.Incoming.StartConsuming(prefetchLimit, conf.RMQ.PollDuration)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s queue incoming consume error", worker.name)
	}

	err = worker.broker.Pushed.StartConsuming(prefetchLimit, conf.RMQ.PollDurationPushed)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s queue pushed consume error", worker.name)
	}

	numConsumers += 2 // we got one consumer for the returned, lets add 2 more
	for i := 0; i < int(numConsumers); i++ {
		tag, consumer := NewConsumer(worker.ctx, worker.db, i, worker.name, "incoming")
		if _, err := worker.broker.Incoming.AddConsumer(tag, consumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
		consumer.state.State = proto.ScannerServiceControl_START

		// store consumer pointer to the worker struct
		worker.consumers = append(worker.consumers, consumer)

		tag, consumer = NewConsumer(worker.ctx, worker.db, i, worker.name, "rejected")
		if _, err := worker.broker.Pushed.AddConsumer(tag, consumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}

		// worker.consumers = append(worker.consumers, consumer) //TODO or not?
	}

	worker.startReturner(worker.broker.Incoming)
}

// StopConsuming stop consumer messages on the broker
func (worker *Worker) StopConsuming() {
	log.Info().Msgf("stop consuming %s...", worker.name)
	for _, consumer := range worker.consumers {
		if consumer.engine != nil {
			log.Info().Msgf("%s cancelling consumer %v", worker.name, consumer.name)
			consumer.state.State = worker.state.State
			consumer.cancel()
		}
	}
	<-worker.broker.Incoming.StopConsuming()
	<-worker.broker.Pushed.StopConsuming()
	<-worker.broker.Connection.StopAllConsuming() // wait for all Consume() calls to finish
}
