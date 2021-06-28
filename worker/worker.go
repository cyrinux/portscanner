package worker

import (
	"context"
	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/util"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"log"
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
}

// Broker represent a RMQ broker
type Broker struct {
	incoming   rmq.Queue
	pushed     rmq.Queue
	connection rmq.Connection
}

// NewWorker create a new worker and init the database connection
func NewWorker(config config.Config) *Worker {
	ctx := context.Background()
	redisClient := util.RedisConnect(ctx, config)
	broker := NewBroker(ctx, config, *redisClient)
	locker := redislock.New(redisClient)
	return &Worker{
		config:      config,
		ctx:         ctx,
		broker:      broker,
		redisClient: redisClient,
		locker:      locker,
	}
}

// StartWorker start a scanner worker
func (worker *Worker) StartWorker() {
	// open tasks queues and connection
	worker.state.State = proto.ScannerServiceControl_UNKNOWN

	// manage signals
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT)
		defer signal.Stop(signals)

		<-signals // wait for signal
		go func() {
			<-signals // hard exit on second signal (in case shutdown gets stuck)
			os.Exit(1)
		}()
		<-worker.broker.connection.StopAllConsuming() // wait for all Consume() calls to finish
	}()

	worker.ControlService()
}

// ControlService return the workers status and control them
func (worker *Worker) ControlService() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial("server:9000", grpc.WithInsecure())
	if err != nil {
		log.Printf("could not connect: %s", err)
		panic(err)
	}
	defer conn.Close()

	client := proto.NewScannerServiceClient(conn)
	state := proto.ScannerServiceControl{State: 0}
	var response *proto.ScannerServiceControl
	for {
		response, err = client.ServiceControl(worker.ctx, &state)
		if err != nil {
			log.Print(err)
		} else {
			if response.State == proto.ScannerServiceControl_START && worker.state.State != proto.ScannerServiceControl_START {
				log.Print("from stop/unknown to start")
				worker.state.State = proto.ScannerServiceControl_START
				worker.startConsuming()
			} else if response.State == proto.ScannerServiceControl_FORCESTART {
				log.Print("to force start")
				worker.state.State = proto.ScannerServiceControl_START
				worker.startConsuming()
			} else if response.State == proto.ScannerServiceControl_STOP && worker.state.State == proto.ScannerServiceControl_START {
				log.Print("from start to stop")
				worker.state.State = proto.ScannerServiceControl_STOP
			} else {
				log.Print("no state change")
			}
		}
		time.Sleep(2 * time.Second)
	}
}

// RmqLogErrors display the rmq errors log
func rmqLogErrors(errChan <-chan error) {
	for err := range errChan {
		switch err := err.(type) {
		case *rmq.HeartbeatError:
			if err.Count == rmq.HeartbeatErrorLimit {
				log.Print("heartbeat error (limit): ", err)
			} else {
				log.Print("heartbeat error: ", err)
			}
		case *rmq.ConsumeError:
			log.Print("consume error: ", err)
		case *rmq.DeliveryError:
			log.Print("delivery error: ", err.Delivery, err)
		default:
			log.Print("other error: ", err)
		}
	}
}

// Locker help to lock some tasks
func (worker *Worker) startReturner(queue rmq.Queue) {
	go func() {
		for {
			// Try to obtain lock.
			lock, err := worker.locker.Obtain(worker.ctx, "returner", 10*time.Second, nil)
			if err != nil && err != redislock.ErrNotObtained {
				log.Fatalln(err)
			} else if err != redislock.ErrNotObtained {
				// Sleep and check the remaining TTL.
				if ttl, err := lock.TTL(worker.ctx); err != nil {
					log.Fatalf("Returner error: %v: ttl: %v", err, ttl)
				} else if ttl > 0 {
					// Yay, I still have my lock!
					returned, _ := queue.ReturnRejected(returnerLimit)
					if returned > 0 {
						log.Printf("Returned reject message %v", returned)
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
		config.RmqDbName, &redisClient, errChan,
	)
	if err != nil {
		panic(err)
	}

	incomingQueue, err := connection.OpenQueue("tasks")
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		panic(err)
	}

	pushQueue, err := connection.OpenQueue("tasks-rejected")
	if err != nil && err != rmq.ErrorAlreadyConsuming {
		panic(err)
	}

	incomingQueue.SetPushQueue(pushQueue)

	return &Broker{incoming: incomingQueue, pushed: pushQueue, connection: connection}
}

func (worker *Worker) startConsuming() {
	numConsumers, err := strconv.ParseInt(
		worker.config.NumConsumers, 10, 0,
	)
	if err != nil {
		panic(err)
	}

	err = worker.broker.incoming.StartConsuming(prefetchLimit, pollDuration)
	if err != nil {
		log.Print(err)
	}
	err = worker.broker.pushed.StartConsuming(prefetchLimit, pollDurationPushed)
	if err != nil {
		log.Print(err)
	}

	for i := 0; i < int(numConsumers); i++ {
		tag, consumer := NewConsumer(worker, i, "incoming")
		if _, err := worker.broker.incoming.AddConsumer(tag, consumer); err != nil {
			log.Printf("%v\n", err)
		}
		tag, consumer = NewConsumer(worker, i, "push")
		if _, err := worker.broker.pushed.AddConsumer(tag, consumer); err != nil {
			log.Printf("%v\n", err)
		}
	}

	// start the returner
	worker.startReturner(worker.broker.incoming)

}

func (worker *Worker) stopConsuming() {
	<-worker.broker.incoming.StopConsuming()
	<-worker.broker.pushed.StopConsuming()
	<-worker.broker.connection.StopAllConsuming() // wait for all Consume() calls to finish
}
