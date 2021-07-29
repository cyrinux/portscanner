package worker

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/client"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/consumer"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/helpers"
	"github.com/cyrinux/grpcnmapscanner/locker"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/go-redis/redis/v8"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	conf        = config.GetConfig().Logger
	log         = logger.New(conf.Debug, conf.Pretty)
	hostname, _ = os.Hostname()

	kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             1 * time.Second,  // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	grpcStreamRetryParams = []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(1 * time.Second)),
	}

	grpcUnaryRetryParams = []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(1 * time.Second)),
		grpc_retry.WithMax(10),
	}
)

// Worker define the worker struct
type Worker struct {
	ctx         context.Context
	name        string
	broker      *broker.Broker
	conf        config.Config
	redisClient *redis.Client
	locker      locker.MyLockerInterface
	state       *pb.ServiceStateValues
	consumers   []*consumer.Consumer
	db          database.Database
	grpcClient  pb.BackendServiceClient
	returned    chan int64
}

// NewWorker create a new worker and init the database connection
func NewWorker(ctx context.Context, conf config.Config, name string) *Worker {
	log.Info().Msgf("%s worker is starting", name)

	var err error

	// Storage database connection init, dedicated context to keep access to the database
	wait := 1000 * time.Millisecond
	var db database.Database
	for {
		db, err = database.Factory(context.Background(), conf)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s cannot connect to main database, retrying in %v...", name, wait)
			time.Sleep(wait)
		} else {
			log.Info().Msgf("%s connected to main database", name)
			break
		}
	}

	// redis, redisLock
	redisClient := helpers.NewRedisClient(ctx, conf).Connect()

	// distributed lock - with redis
	myLocker := locker.CreateRedisLock(redisClient)

	// grpc conn
	wait = 1000 * time.Millisecond
	var cc *grpc.ClientConn
	for {
		cc, err = connectToServer(ctx, conf, name)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s cannot connect to the control server, retrying in %v...", name, wait)
			time.Sleep(wait)
		} else {
			log.Info().Msgf("%s connected to the control server", name)
			break
		}
	}

	state := &pb.ServiceStateValues{State: pb.ServiceStateValues_UNKNOWN}

	return &Worker{
		name:        name,
		conf:        conf,
		ctx:         ctx,
		broker:      new(broker.Broker),
		locker:      myLocker,
		consumers:   make([]*consumer.Consumer, 0),
		db:          db,
		state:       state,
		grpcClient:  pb.NewBackendServiceClient(cc),
		redisClient: redisClient,
		returned:    make(chan int64),
	}
}

// connectToServer connect to the server
func connectToServer(ctx context.Context, conf config.Config, name string) (*grpc.ClientConn, error) {
	// tls client config
	tlsCredentials, err := loadTLSCredentials(
		conf.Backend.CAFile,
		conf.Backend.ClientCertFile,
		conf.Backend.ClientKeyFile,
	)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("can't load TLS credentials")
		return nil, err
	}

	cc1, err := grpc.DialContext(
		ctx,
		conf.BackendServer,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(grpcUnaryRetryParams...)),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(grpcStreamRetryParams...)),
	)

	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s could not connect to server %s", name, conf.BackendServer)
		return nil, err
	}
	log.Debug().Msgf("%s cc1: connected to the control server: %s", name, conf.BackendServer)

	authClient := client.NewAuthClient(cc1, conf.Backend.Username, conf.Backend.Password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), conf.Backend.RefreshDuration)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("cannot create auth interceptor to %s", conf.BackendServer)
		return nil, err
	}

	cc2, err := grpc.DialContext(
		ctx,
		conf.BackendServer,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)

	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s could not connect to control server: %s", name, conf.BackendServer)
		return nil, err
	}
	log.Debug().Msgf("%s cc2: connected to the control server: %s", name, conf.BackendServer)

	return cc2, nil
}

// StartWorker start a scanner worker
func (worker *Worker) StartWorker() {

	// start the worker on boot
	worker.startConsuming()

	// watch the control server and stop/start service
	worker.StreamServiceControl()
}

// StreamServiceControl return the workers status and control them
func (worker *Worker) StreamServiceControl() {
	wait := 1000 * time.Millisecond

	getState := &pb.ServiceStateValues{State: pb.ServiceStateValues_UNKNOWN}
	for {
		log.Debug().Msgf("%s trying to connect in %v to server control", worker.name, wait)
		stream, err := worker.grpcClient.StreamServiceControl(worker.ctx, getState)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get stream connection")
			break
		}
		log.Info().Msgf("%s connected to server control", worker.name)

		for range time.Tick(500 * time.Millisecond) {

			serviceControl, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Error().Stack().Err(err).Msg("can't get service control state ")
				break
			}

			if serviceControl.State == 1 && worker.state.State != 1 { //pb.ServiceStateValues_START
				worker.state.State = pb.ServiceStateValues_START
				worker.startConsuming()
			} else if serviceControl.State == 2 && worker.state.State != 2 { //pb.ServiceStateValues_STOP
				worker.state.State = pb.ServiceStateValues_STOP
				worker.StopConsuming()
			}

		}

		time.Sleep(wait)
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(returner *broker.Returner) {
	wait := 500 * time.Millisecond
	log.Info().Msgf("starting the returner routine on %s", worker.name)
	lockKey := "returner"
	for {
		// Try to obtain lock.
		ok, err := worker.locker.Obtain(worker.ctx, lockKey, 10*time.Second)

		if err != nil {
			log.Error().Stack().Err(err).Msg("returner can't obtain lock")
		} else if ok {
			defer func(locker locker.MyLockerInterface, ctx context.Context, key string) {
				err = locker.Release(ctx, key)
				if err != nil {
					log.Error().Stack().Err(err).Msg("can't defer lock")
				}
			}(worker.locker, worker.ctx, lockKey)
			// Sleep and check the remaining TTL.
			if ttl, err := worker.locker.TTL(worker.ctx, lockKey); err != nil {
				log.Error().Stack().Err(err).Msgf("returner error, ttl: %v", ttl)
			} else if ttl > 0 {
				// Yay, I still have my lock!
				ret, err := returner.Queue.ReturnRejected(worker.conf.RMQ.ReturnerLimit)
				if err != nil {
					log.Error().Stack().Err(err).Msg("error while returning message")
				}
				if ret > 0 {
					log.Info().Msgf("returner success requeue %d tasks messages to incoming", ret)
					// prometheus returned stats
					returner.Returned <- ret
				}
				err = worker.locker.Refresh(worker.ctx, lockKey, 5*time.Second)
				if err != nil {
					log.Error().Stack().Err(err).Msg("can't refresh lock")
				}
			}
		}
		// cpu cooling
		time.Sleep(wait)
	}
}

func (worker *Worker) startConsuming() *Worker {
	conf := worker.conf

	numConsumers := conf.RMQ.NumConsumers
	prefetchLimit := numConsumers + 1 // prefetchLimit need to be > numConsumers
	log.Info().Msgf("start consuming %s with %d consumers...", worker.name, numConsumers)

	worker.broker = broker.New(worker.ctx, worker.name, conf.RMQ, worker.redisClient)

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
		tag, incomingConsumer := consumer.New(
			worker.ctx,
			worker.db,
			i,
			worker.name,
			worker.conf.NMAP,
			"incoming",
			worker.locker,
		)
		if _, err := worker.broker.Incoming.AddConsumer(tag, incomingConsumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
		incomingConsumer.State.State = pb.ServiceStateValues_START

		// store consumer pointer to the worker struct
		worker.consumers = append(worker.consumers, incomingConsumer)

		// start prometheus collector
		go worker.collectConsumerStats(incomingConsumer.Success, incomingConsumer.Failed, worker.returned)

		tag, returnerConsumer := consumer.New(worker.ctx, worker.db, i, worker.name, worker.conf.NMAP, "rejected", worker.locker)
		if _, err := worker.broker.Pushed.AddConsumer(tag, returnerConsumer); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}

	go worker.startReturner(
		&broker.Returner{
			Queue:    worker.broker.Incoming,
			Returned: worker.returned,
		},
	)

	return worker
}

// StopConsuming stop consumer messages on the broker
func (worker *Worker) StopConsuming() *Worker {
	log.Info().Msgf("%s stop consuming...", worker.name)
	for _, c := range worker.consumers {
		if c.Engine != nil {
			log.Info().Msgf("%s cancelling consumer %s", worker.name, c.Name)
			c.State.State = worker.state.State
			c.Cancel()
		}
	}

	<-worker.broker.Incoming.StopConsuming()
	<-worker.broker.Pushed.StopConsuming()
	<-worker.broker.Connection.StopAllConsuming() // wait for all Consume() calls to finish

	return worker
}

// collectConsumerStats manage the tasks prometheus counters
func (worker *Worker) collectConsumerStats(success chan int64, failed chan int64, returned chan int64) {
	wait := 1000 * time.Millisecond
	for {
		stream, err := worker.grpcClient.StreamTasksStatus(worker.ctx)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get consumer stats")
			break
		}
		log.Info().Msgf("%s stats collector connected to server control", worker.name)

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
			time.Sleep(wait)
		}

		// wait before try to reconnect
		log.Debug().Msgf("%s trying to reconnect in %v to server control", worker.name, wait)
		time.Sleep(wait)
	}
}

// authMethods allowed grpc endpoints
func authMethods() map[string]bool {
	const backendServicePath = "/proto.BackendService/"

	return map[string]bool{
		backendServicePath + "StreamServiceControl": true,
		backendServicePath + "StreamTasksStatus":    true,
	}
}
