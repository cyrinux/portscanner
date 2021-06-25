package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	rmq "github.com/adjust/rmq/v3"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
)

const (
	prefetchLimit      = 1000
	pollDuration       = 100 * time.Millisecond
	pollDurationPushed = 5000 * time.Millisecond
	reportBatchSize    = 10000
	consumeDuration    = time.Second
)

type Worker struct {
	ctx    context.Context
	config config.Config
}

// NewWorker create a new worker and init the database connection
func NewWorker(config config.Config) *Worker {
	return &Worker{
		config: config,
		ctx:    context.TODO(),
	}
}

// StartWorker start a scanner worker
func (w *Worker) StartWorker() {

	config := w.config

	numConsumers, err := strconv.ParseInt(
		config.NumConsumers, 10, 0,
	)
	if err != nil {
		panic(err)
	}

	// open tasks queues and connection
	incomingQueue, pushQueue, connection := openQueues(config)

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

	// returned messages rejected/delayed back to the incoming queue
	go func() {
		for {
			returned, _ := incomingQueue.ReturnRejected(math.MaxInt64)
			if returned > 0 {
				log.Printf("Returned reject message %v", returned)
			}
			time.Sleep(1 * time.Second)
		}
	}()

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
	hostname, _ := os.Hostname()
	name := fmt.Sprintf("%s-consumer-%s-%d", queue, hostname, tag)
	log.Printf("New consumer: %s\n", name)
	return name, &Consumer{
		name:   name,
		count:  0,
		before: time.Now(),
		ctx:    context.Background(),
		config: config,
	}
}

func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Printf("start consume %s", payload)
	time.Sleep(consumeDuration)

	consumer.count++
	if consumer.count%reportBatchSize == 0 {
		duration := time.Since(consumer.before)
		consumer.before = time.Now()
		perSecond := time.Second / (duration / reportBatchSize)
		log.Printf("%s consumed %d %s %d", consumer.name, consumer.count, payload, perSecond)
	}

	var request *proto.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &request)

	deferTime := time.Unix(int64(request.DeferDuration), 0).Unix()
	if deferTime >= time.Now().Unix() {
		if err := delivery.Reject(); err != nil {
			log.Printf("failed to reject %s: %s", payload, err)
		} else {
			log.Printf("delayed %s, this is too early", payload)
		}
	} else {
		newEngine := engine.NewEngine(consumer.config)
		key, result, err := newEngine.StartNmapScan(consumer.ctx, request)
		if err != nil {
			log.Printf("Scan %v %v: %v", key, result, err)
		}

		key, scanResult, _ := engine.ParseScanResult(key, result)

		scannerResponse := &proto.ScannerResponse{
			HostResult: scanResult,
		}
		scanResultJSON, err := json.Marshal(scannerResponse)
		if err != nil {
			log.Printf("failed to parse result: %s", err)
		}
		_, err = consumer.config.DB.Set(
			consumer.ctx, key, string(scanResultJSON),
			time.Duration(request.GetRetentionTime())*time.Second,
		)
		if err != nil {
			log.Printf("failed to insert result: %s", err)
		}

		if consumer.count%reportBatchSize > 0 {
			if err := delivery.Ack(); err != nil {
				log.Printf("failed to ack %s: %s", payload, err)
			} else {
				log.Printf("acked %s", payload)
			}
		} else { // reject one per batch
			if err := delivery.Reject(); err != nil {
				log.Printf("failed to reject %s: %s", payload, err)
			} else {
				log.Printf("rejected %s", payload)
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

// openQueues open the broker queues
func openQueues(config config.Config) (rmq.Queue, rmq.Queue, rmq.Connection) {
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnection(
		config.RmqDbName, "tcp", config.RmqServer, 1, errChan,
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
