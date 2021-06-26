package worker

import (
	"context"
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	prefetchLimit      = 1000
	returnerLimit      = 1000
	pollDuration       = 100 * time.Millisecond
	pollDurationPushed = 5000 * time.Millisecond
	reportBatchSize    = 10000
	consumeDuration    = 5 * time.Second
)

var (
	hostname, _ = os.Hostname()
)

// Worker define the worker struct
type Worker struct {
	ctx    context.Context
	config config.Config
}

// NewWorker create a new worker and init the database connection
func NewWorker(config config.Config) *Worker {
	return &Worker{
		config: config,
		ctx:    context.Background(),
	}
}

// StartWorker start a scanner worker
func (worker *Worker) StartWorker() {
	config := worker.config
	ctx := worker.ctx

	numConsumers, err := strconv.ParseInt(
		config.NumConsumers, 10, 0,
	)
	if err != nil {
		panic(err)
	}

	// open tasks queues and connection
	incomingQueue, pushQueue, connection := openQueues(ctx, config, worker)

	// start the returner
	locker := redislock.New(redisConnect(ctx, worker))
	worker.startReturner(ctx, locker, incomingQueue)

	for i := 0; i < int(numConsumers); i++ {
		tag, consumer := NewConsumer(i, "incoming", config)
		if _, err := incomingQueue.AddConsumer(tag, consumer); err != nil {
			panic(err)
		}
		tag, consumer = NewConsumer(i, "push", config)
		if _, err := pushQueue.AddConsumer(tag, consumer); err != nil {
			panic(err)
		}
	}

	// manage signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()
	<-connection.StopAllConsuming() // wait for all Consume() calls to finish
}

type Consumer struct {
	name   string
	count  int
	before time.Time
	ctx    context.Context
	config config.Config
}

func NewConsumer(tag int, queue string, config config.Config) (string, *Consumer) {
	name := fmt.Sprintf("%s-consumer-%s-%d", queue, hostname, tag)
	log.Printf("New consumer: %s\n", name)
	return name, &Consumer{
		name:   name,
		count:  0,
		before: time.Now(),
		ctx:    context.TODO(),
		config: config,
	}
}

// Consume consume the message tasks on the redis broker
func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Printf("%s: start consume %s", consumer.name, payload)
	time.Sleep(consumeDuration)

	consumer.count++
	if consumer.count%reportBatchSize == 0 {
		duration := time.Since(consumer.before)
		consumer.before = time.Now()
		perSecond := time.Second / (duration / reportBatchSize)
		log.Printf("%s: consumed %d %s %d", consumer.name, consumer.count, payload, perSecond)
	}

	var request *proto.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &request)

	deferTime := time.Unix(int64(request.DeferDuration), 0).Unix()
	if deferTime >= time.Now().Unix() {
		if err := delivery.Reject(); err != nil {
			log.Printf("%s: failed to reject %s: %s", consumer.name, payload, err)
		} else {
			log.Printf("%s: delayed %s, this is too early", consumer.name, payload)
		}
	} else {
		newEngine := engine.NewEngine(consumer.config)
		key, result, err := newEngine.StartNmapScan(consumer.ctx, request)
		if err != nil {
			log.Printf("%s: scan %v %v: %v", consumer.name, key, result, err)
		}

		key, scanResult, _ := engine.ParseScanResult(key, result)

		scannerResponse := &proto.ScannerResponse{
			HostResult: scanResult,
		}
		scanResultJSON, err := json.Marshal(scannerResponse)
		if err != nil {
			log.Printf("%s: failed to parse result: %s", consumer.name, err)
		}
		_, err = consumer.config.DB.Set(
			consumer.ctx, key, string(scanResultJSON),
			time.Duration(request.GetRetentionTime())*time.Second,
		)
		if err != nil {
			log.Printf("%s: failed to insert result: %s", consumer.name, err)
		}

		if consumer.count%reportBatchSize > 0 {
			if err := delivery.Ack(); err != nil {
				log.Printf("%s: failed to ack %s: %s", consumer.name, payload, err)
			} else {
				log.Printf("%s: acked %s", consumer.name, payload)
			}
		} else { // reject one per batch
			if err := delivery.Reject(); err != nil {
				log.Printf("%s: failed to reject %s: %s", consumer.name, payload, err)
			} else {
				log.Printf("%s: rejected %s", consumer.name, payload)
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
func (worker *Worker) startReturner(ctx context.Context, locker *redislock.Client, queue rmq.Queue) {
	go func() {
		for {
			// Try to obtain lock.
			lock, err := locker.Obtain(ctx, "returner", 10*time.Second, nil)
			if err != nil && err != redislock.ErrNotObtained {
				log.Fatalln(err)
			} else if err != redislock.ErrNotObtained {
				// Sleep and check the remaining TTL.
				if ttl, err := lock.TTL(ctx); err != nil {
					log.Fatalf("Returner error: %v: ttl: %v", err, ttl)
				} else if ttl > 0 {
					// Yay, I still have my lock!
					returned, _ := queue.ReturnRejected(returnerLimit)
					if returned > 0 {
						log.Printf("Returned reject message %v", returned)
					}
					lock.Refresh(ctx, 5*time.Second, nil)
				}
			}
			// cpu cooling
			time.Sleep(1 * time.Second)
		}
	}()
}

// openQueues open the broker queues
func openQueues(
	ctx context.Context, config config.Config, worker *Worker,
) (rmq.Queue, rmq.Queue, rmq.Connection) {
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnectionWithRedisClient(
		config.RmqDbName, redisConnect(ctx, worker), errChan,
	)
	if err != nil {
		panic(err)
	}

	incomingQueue, err := connection.OpenQueue("tasks")
	if err != nil {
		panic(err)
	}
	if err := incomingQueue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		panic(err)
	}

	pushQueue, err := connection.OpenQueue("tasks-rejected")
	if err != nil {
		panic(err)
	}
	incomingQueue.SetPushQueue(pushQueue)

	if err := pushQueue.StartConsuming(prefetchLimit, pollDurationPushed); err != nil {
		panic(err)
	}

	return incomingQueue, pushQueue, connection
}

func redisConnect(ctx context.Context, worker *Worker) *redis.Client {
	config := worker.config
	// Connect to redis for the locker
	redisClient := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     config.RmqServer,
		Password: config.RmqDbPassword,
		DB:       0,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	return redisClient
}
