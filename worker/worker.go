package worker

import (
	"context"
	"io"
	"os"
	"time"

	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	brk "github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto"
	redis "github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"strconv"
)

var (
	conf        = config.GetConfig().Logger
	log         = logger.New(conf.Debug, conf.Pretty)
	hostname, _ = os.Hostname()
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             1 * time.Second,  // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// Worker define the worker struct
type Worker struct {
	ctx         context.Context
	name        string
	broker      brk.Broker
	conf        config.Config
	locker      *redislock.Client
	redisClient *redis.Client
	state       pb.ScannerServiceControl
	consumers   []*Consumer
	db          database.Database
	grpcServer  *grpc.ClientConn
	returned    chan int64
}

// NewWorker create a new worker and init the database connection
func NewWorker(ctx context.Context, conf config.Config, name string) *Worker {
	log.Info().Msgf("%s worker is starting", name)

	// Storage database connection init, dedicated context to keep access to the database
	db, err := database.Factory(context.Background(), conf)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}

	// redis
	rmqDB, _ := strconv.ParseInt(conf.RMQ.Database, 10, 0)
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:    conf.RMQ.Redis.SentinelServers,
		MasterName:       conf.RMQ.Redis.MasterName,
		Password:         conf.RMQ.Redis.Password,
		SentinelPassword: conf.RMQ.Redis.SentinelPassword,
		DB:               int(rmqDB),
		MaxRetries:       5,
	})

	// broker - with redis
	broker := brk.NewBroker(ctx, name, conf.RMQ, redisClient)

	// distributed lock - with redis
	locker := redislock.New(redisClient)

	grpcServer, err := grpc.Dial(
		conf.Global.ControllerServer,
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(kacp),
	)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s could not connect to server", name)
	}

	return &Worker{
		conf:        conf,
		name:        name,
		ctx:         ctx,
		broker:      broker,
		locker:      locker,
		consumers:   make([]*Consumer, 0),
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
		stream, err := client.StreamServiceControl(worker.ctx, getState)
		if err != nil {
			break
		}
		log.Debug().Msgf("%s connected to server control", worker.name)

		for range time.Tick(500 * time.Millisecond) {
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
		}

		// wait before try to reconnect
		reconnectTime := 1000 * time.Millisecond
		log.Debug().Msgf("%s trying to connect in %v to server control", worker.name, reconnectTime)
		time.Sleep(reconnectTime)
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(queue rmq.Queue, returned chan int64) {
	log.Info().Msg("starting the returner routine")
	conf := worker.conf
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
				rtr, _ := queue.ReturnRejected(conf.RMQ.ReturnerLimit)
				if rtr > 0 {
					log.Info().Msgf("returner success requeue %d tasks messages to incoming", rtr)
					returned <- rtr
				}
				lock.Refresh(worker.ctx, 5*time.Second, nil)
			}
		}
		// cpu cooling
		time.Sleep(500 * time.Millisecond)
	}
}

func (worker *Worker) startConsuming() *Worker {
	conf := worker.conf
	numConsumers := conf.RMQ.NumConsumers
	prefetchLimit := numConsumers + 1 // prefetchLimit need to be > numConsumers
	log.Info().Msgf("start consuming %s with %d consumers...", worker.name, numConsumers)

	worker.broker = brk.NewBroker(worker.ctx, worker.name, conf.RMQ, worker.redisClient)

	err := worker.broker.Incoming.StartConsuming(prefetchLimit, conf.RMQ.PollDuration)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s queue incoming consume error", worker.name)
	}

	err = worker.broker.Pushed.StartConsuming(prefetchLimit, conf.RMQ.PollDurationPushed)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s queue pushed consume error", worker.name)
	}

	numConsumers++ // we got one consumer for the returned, lets add 2 more

	for i := 0; i < int(numConsumers); i++ {
		tag, incConsumer := NewConsumer(worker.ctx, worker.db, i, worker.name, "incoming")
		if _, err := worker.broker.Incoming.AddConsumer(tag, incConsumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
		incConsumer.state.State = pb.ScannerServiceControl_START

		// store consumer pointer to the worker struct
		worker.consumers = append(worker.consumers, incConsumer)

		// start prometheus collector
		go worker.collectConsumerStats(incConsumer.success, incConsumer.failed, worker.returned)

		tag, rConsumer := NewConsumer(worker.ctx, worker.db, i, worker.name, "rejected")
		if _, err := worker.broker.Pushed.AddConsumer(tag, rConsumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	go worker.startReturner(worker.broker.Incoming, worker.returned)

	return worker
}

// StopConsuming stop consumer messages on the broker
func (worker *Worker) StopConsuming() *Worker {
	log.Info().Msgf("%s stop consuming...", worker.name)
	for _, consumer := range worker.consumers {
		if consumer.engine != nil {
			log.Info().Msgf("%s cancelling consumer %s", worker.name, consumer.name)
			consumer.state.State = worker.state.State
			consumer.cancel()
		}
	}
	<-worker.broker.Incoming.StopConsuming()
	<-worker.broker.Pushed.StopConsuming()
	<-worker.broker.Connection.StopAllConsuming() // wait for all Consume() calls to finish

	return worker
}

// collectConsumerStats manage the tasks prometheus counters
func (worker *Worker) collectConsumerStats(success chan int64, failed chan int64, returned chan int64) {
	client := pb.NewScannerServiceClient(worker.grpcServer)
	for {
		stream, err := client.StreamTasksStatus(worker.ctx)
		if err != nil {
			break
		}
		log.Debug().Msgf("%s stats collector connected to server control", worker.name)
		for {
			var s, f, r int64
			select {
			case s = <-success:
			case f = <-failed:
			case r = <-returned:
			default:
			}
			err = stream.Send(&pb.PrometheusStatus{
				TasksStatus: &pb.TasksStatus{
					Success: s, Failed: f, Returned: r,
				},
			})
			if err != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		// wait before try to reconnect
		reconnectTime := 5 * time.Second
		log.Debug().Msgf("%s trying to reconnect in %v to server control", worker.name, reconnectTime)
		time.Sleep(reconnectTime)
	}
}
