package worker

import (
	"encoding/json"
	"strconv"

	"fmt"
	rmq "github.com/adjust/rmq/v3"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engines"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	prefetchLimit = 1000
	pollDuration  = 100 * time.Millisecond

	reportBatchSize = 10000
	consumeDuration = time.Millisecond
	shouldLog       = true
)

type Worker struct {
	config config.Config
}

// NewWorker create a new worker and init the database connection
func NewWorker(config config.Config) *Worker {
	return &Worker{config}
}

// StartWorker start a scanner worker
func (w *Worker) StartWorker() {
	numConsumersInt, err := strconv.ParseInt(w.config.NumConsumers, 10, 0)
	if err != nil {
		panic(err)
	}
	errChan := make(chan error, 10)
	go rmqLogErrors(errChan)

	connection, err := rmq.OpenConnection(w.config.RmqDbName, "tcp", w.config.RmqServer, 1, errChan)
	if err != nil {
		panic(err)
	}

	queue, err := connection.OpenQueue("tasks")
	if err != nil {
		panic(err)
	}

	if err := queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		panic(err)
	}

	for i := 0; i < int(numConsumersInt); i++ {
		name := fmt.Sprintf("consumer %d", i)
		if _, err := queue.AddConsumer(name, NewConsumer(i, w.config)); err != nil {
			panic(err)
		}
	}

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
	config config.Config
}

func NewConsumer(tag int, config config.Config) *Consumer {
	name, _ := os.Hostname()
	return &Consumer{
		name:   fmt.Sprintf("consumer-%s-%d", name, tag),
		count:  0,
		before: time.Now(),
		config: config,
	}
}

func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Printf("start consume %s", payload)
	time.Sleep(consumeDuration)

	consumer.count++
	if consumer.count%reportBatchSize == 0 {
		duration := time.Now().Sub(consumer.before)
		consumer.before = time.Now()
		perSecond := time.Second / (duration / reportBatchSize)
		log.Printf("%s consumed %d %s %d", consumer.name, consumer.count, payload, perSecond)

	}

	var in *proto.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &in)
	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	key, result, err := engines.StartNmapScan(in)
	if err != nil {
		log.Printf("%v: %v", result, err)
	}

	key, scanResult, _ := engines.ParseScanResult(key, result)

	scannerResponse := &proto.ScannerResponse{
		HostResult: scanResult,
	}
	scanResultJSON, err := json.Marshal(scannerResponse)
	if err != nil {
		log.Printf("failed to parse result: %s", err)
	}
	_, err = consumer.config.DB.Set(key, string(scanResultJSON), time.Duration(in.GetRetentionTime())*time.Second)
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
