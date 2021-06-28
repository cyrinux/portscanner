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
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

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
		log.Printf("could not connect to controller: %s", err)
		panic(err)
	}
	defer conn.Close()

	client := proto.NewScannerServiceClient(conn)
	getState := &proto.ScannerServiceControl{State: 0}
	for {
		// wait before try to reconnect
		time.Sleep(1 * time.Second)

		log.Printf("trying to connect to server control")
		stream, err := client.StreamServiceControl(worker.ctx, getState)
		if err != nil {
			continue
		}
		log.Printf("connected to server control")
		for {
			// cpu cooling
			time.Sleep(500 * time.Millisecond)

			serviceControl, err := stream.Recv()
			if err == io.EOF {
				log.Print("DEBUG 1")
				continue
			} else if err != nil {
				continue
			} else if serviceControl == nil {
				log.Print("DEBUG 3")
				continue
			} else {
				log.Printf("DEBUG 4: %v", worker.state.State)
				if serviceControl.State == proto.ScannerServiceControl_START && worker.state.State != proto.ScannerServiceControl_START {
					log.Printf("DEBUG 5")
					log.Print("from stop/unknown to start")
					worker.state.State = proto.ScannerServiceControl_START
				} else if serviceControl.State == proto.ScannerServiceControl_STOP && worker.state.State == proto.ScannerServiceControl_START {
					log.Print("DEBUG 6")
					log.Print("from start to stop")
					worker.state.State = proto.ScannerServiceControl_STOP
				}
			}
		}
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
	log.Print("DEBUGGGG starting the returner DEBBUGGG")
	go func() {
		for {
			// Try to obtain lock.
			lock, err := worker.locker.Obtain(worker.ctx, "returner", 10*time.Second, nil)
			if err != nil && err != redislock.ErrNotObtained {
				log.Fatalln(err)
			} else if err != redislock.ErrNotObtained {
				// Sleep and check the remaining TTL.
				if ttl, err := lock.TTL(worker.ctx); err != nil {
					log.Printf("Returner error: %v: ttl: %v", err, ttl)
				} else if ttl > 0 {
					// Yay, I still have my lock!
					returned, _ := queue.ReturnRejected(returnerLimit)
					if returned > 0 {
						log.Printf("Returner success requeue %v tasks messages to incoming", returned)
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

	worker.startReturner(worker.broker.incoming)
}

func (worker *Worker) stopConsuming() {
	log.Print("Stop consumming...")
	<-worker.broker.incoming.StopConsuming()
	<-worker.broker.pushed.StopConsuming()
	<-worker.broker.connection.StopAllConsuming() // wait for all Consume() calls to finish
}
