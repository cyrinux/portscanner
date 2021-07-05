package worker

import (
	"context"
	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	redis "github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"os"
	"sync"
	"time"
)

var (
	conf        = config.GetConfig()
	log         = logger.New(conf.Logger.Debug, conf.Logger.Pretty)
	hostname, _ = os.Hostname()
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             1 * time.Second,  // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// Worker define the worker struct
type Worker struct {
	sync.Mutex
	ctx         context.Context
	broker      broker.Broker
	config      config.Config
	locker      *redislock.Client
	redisClient *redis.Client
	state       pb.ScannerServiceControl
	consumers   []*Consumer
	db          database.Database
	name        string
	grpcServer  *grpc.ClientConn
	returned    chan int64
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
	consumers := make([]*Consumer, 1)

	grpcServer, err := grpc.Dial(
		config.Global.ControllerServer,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		log.Error().Stack().Err(err).Msg("could not connect to server")
	}

	return &Worker{
		config:      config,
		name:        name,
		ctx:         ctx,
		broker:      broker,
		locker:      locker,
		consumers:   consumers,
		db:          db,
		redisClient: redisClient,
		grpcServer:  grpcServer,
		returned:    make(chan int64),
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
	client := pb.NewScannerServiceClient(worker.grpcServer)
	getState := &pb.ScannerServiceControl{State: 0}
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
			if serviceControl.State == 1 && worker.state.State != 1 { //pb.ScannerServiceControl_START
				worker.state.State = pb.ScannerServiceControl_START
				worker.startConsuming()
			} else if serviceControl.State == 2 && worker.state.State != 2 { //pb.ScannerServiceControl_STOP
				worker.state.State = pb.ScannerServiceControl_STOP
				worker.StopConsuming()
			}

			// cpu cooling
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(queue rmq.Queue, returned chan int64) {
	log.Info().Msg("starting the returner routine")
	conf := worker.config
	for {
		// Try to obtain lock.
		lock, err := worker.locker.Obtain(worker.ctx, "returner", 10*time.Second, nil)
		if err != nil && err != redislock.ErrNotObtained {
			log.Error().Stack().Err(err).Msg("returner can't obtain lock")
		} else if err != redislock.ErrNotObtained {
			// Sleep and check the remaining TTL.
			if ttl, err := lock.TTL(worker.ctx); err != nil {
				log.Error().Stack().Err(err).Msgf("returner error, ttl: %v", ttl)
			} else if ttl > 0 {
				// Yay, I still have my lock!
				r, _ := queue.ReturnRejected(conf.RMQ.ReturnerLimit)
				if r > 0 {
					log.Info().Msgf("returner success requeue %v tasks messages to incoming", r)

					returned <- r
				}
				lock.Refresh(worker.ctx, 5*time.Second, nil)
			}
		}
		// cpu cooling
		time.Sleep(500 * time.Millisecond)
	}
}

func (worker *Worker) startConsuming() {
	conf := worker.config
	numConsumers := conf.RMQ.NumConsumers
	prefetchLimit := numConsumers + 1 // prefetchLimit need to be > numConsumers
	log.Info().Msgf("start consuming %s with %v consumers...", worker.name, numConsumers)

	worker.broker = broker.NewBroker(context.Background(), worker.name, worker.config, worker.redisClient)

	err := worker.broker.Incoming.StartConsuming(prefetchLimit, conf.RMQ.PollDuration)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s queue incoming consume error", worker.name)
	}

	err = worker.broker.Pushed.StartConsuming(prefetchLimit, conf.RMQ.PollDurationPushed)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s queue pushed consume error", worker.name)
	}

	numConsumers += 1 // we got one consumer for the returned, lets add 2 more
	for i := 0; i < int(numConsumers); i++ {
		tag, consumer := NewConsumer(worker.ctx, worker.db, i, worker.name, "incoming")
		if _, err := worker.broker.Incoming.AddConsumer(tag, consumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
		consumer.state.State = pb.ScannerServiceControl_START

		// store consumer pointer to the worker struct
		worker.consumers = append(worker.consumers, consumer)

		// start prometheus collector
		go worker.collectConsumerStats(consumer.success, consumer.failed, worker.returned)

		tag, consumer = NewConsumer(worker.ctx, worker.db, i, worker.name, "rejected")
		if _, err := worker.broker.Pushed.AddConsumer(tag, consumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	go worker.startReturner(worker.broker.Incoming, worker.returned)
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

// collectConsumerStats manage the tasks prometheus counters
func (worker *Worker) collectConsumerStats(success chan int64, failed chan int64, returned chan int64) {
	client := pb.NewScannerServiceClient(worker.grpcServer)
	for {
		stream, err := client.StreamTasksStatus(worker.ctx)
		if err != nil {
			break
		}
		log.Debug().Msg("connected to server control")
		for {
			var s, f, r int64
			select {
			case s = <-success:
				// log.Debug().Msgf("DEBUG success %+v", s)
			case f = <-failed:
				// log.Debug().Msgf("DEBUG failed %+v", f)
			case r = <-returned:
				// log.Debug().Msgf("DEBUG returned %+v", r)
			default:
			}
			err = stream.Send(&pb.TasksStatus{Success: s, Failed: f, Returned: r})
			if err != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		// wait before try to reconnect
		reconnectTime := 5 * time.Second
		log.Debug().Msgf("trying to reconnect in %v to server control", reconnectTime)
		time.Sleep(reconnectTime)
	}
}
