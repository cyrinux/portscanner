package worker

import (
	"context"
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v3"
	"github.com/amyangfei/redlock-go/v2/redlock"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	prefetchLimit = 1000
	pollDuration  = 100 * time.Millisecond

	reportBatchSize = 10000
	consumeDuration = time.Second
	shouldLog       = true
)

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
func (w *Worker) StartWorker() {
	lockMgr, err := redlock.NewRedLock([]string{w.config.RmqServer})
	if err != nil {
		log.Printf("Unable to connect redlock: %v", err)
	}

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

	if err := pushQueue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		panic(err)
	}

	for i := 0; i < int(numConsumersInt); i++ {
		nameIncoming := fmt.Sprintf("consumer incoming %d", i)
		if _, err := incomingQueue.AddConsumer(nameIncoming, NewConsumer(i, "incoming", w.config)); err != nil {
			panic(err)
		}
		namePush := fmt.Sprintf("consumer push %d", i)
		if _, err := pushQueue.AddConsumer(namePush, NewConsumer(i, "push", w.config)); err != nil {
			panic(err)
		}
	}

	// returned messages rejected/delayed back to the incoming queue
	go func() {
		for {
			expiry, err := lockMgr.Lock(w.ctx, "tasks_returner", 15*time.Second)
			if err != nil {
				log.Printf("Returner loop: %v", err)
			} else {
				log.Printf("Returner loop: took the lock, expiry: %v", expiry)
			}
			returned, _ := incomingQueue.ReturnRejected(1000)
			if returned > 0 {
				log.Printf("Returned reject message %v", returned)

			}
			lockMgr.UnLock(w.ctx, "tasks_returner")
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
		lockMgr.UnLock(w.ctx, "tasks_returner")
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

func NewConsumer(tag int, queue string, config config.Config) *Consumer {
	hostname, _ := os.Hostname()
	name := fmt.Sprintf("%s-consumer-%s-%d", queue, hostname, tag)
	log.Printf("New consumer: %s\n", name)
	return &Consumer{
		name:   name,
		count:  0,
		before: time.Now(),
		ctx:    context.TODO(),
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
		_, err = consumer.config.DB.Set(key, string(scanResultJSON), time.Duration(request.GetRetentionTime())*time.Second)
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
