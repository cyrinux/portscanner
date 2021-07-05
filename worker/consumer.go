package worker

import (
	"context"
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	pb "github.com/cyrinux/grpcnmapscanner/proto"
	"sync"
	"time"
)

// Consumer define a broker consumer
type Consumer struct {
	sync.Mutex
	name     string
	cancel   context.CancelFunc
	ctx      context.Context
	db       database.Database
	engine   *engine.Engine
	state    pb.ScannerServiceControl
	tasktype string
	success  chan int64
	failed   chan int64
}

// NewConsumer create a new consumer
func NewConsumer(
	ctx context.Context,
	db database.Database, tag int, tasktype string,
	queue string) (string, *Consumer) {

	name := fmt.Sprintf("%s-consumer-%s-%s-%d", tasktype, queue, hostname, tag)
	log.Info().Msgf("new: %s", name)
	ctx, cancel := context.WithCancel(ctx)
	engine := engine.NewEngine(ctx, db)

	return name, &Consumer{
		ctx:      ctx,
		before:   time.Now(),
		cancel:   cancel,
		count:    0,
		db:       db,
		engine:   engine,
		name:     name,
		tasktype: tasktype,
		success:  make(chan int64, 1),
		failed:   make(chan int64, 1),
	}
}

// onCancel is a function trigger on consumer context cancel
func onCancel(consumer *Consumer, request *pb.ParamsScannerRequest) {
	// waiting for cancel signal
	<-consumer.ctx.Done()
	log.Debug().Msgf("%s cancelled, writing state to database", consumer.name)
	// if scan fail or cancelled, mark task as cancel
	consumer.engine.State = pb.ScannerResponse_CANCEL
	scannerResponse := &pb.ScannerResponse{
		Status: pb.ScannerResponse_CANCEL,
	}
	scanResultJSON, err := json.Marshal(scannerResponse)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s failed to parse failed result", consumer.name)
	}
	_, err = consumer.db.Set(
		context.Background(),
		request.Key,
		string(scanResultJSON),
		time.Duration(request.GetRetentionTime())*time.Second,
	)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s: failed to insert failed result", consumer.name)
	}
}

// Consume consume the message tasks on the redis broker
func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()

	var request *pb.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &request)

	go onCancel(consumer, request)

	log.Debug().Msgf("%s: consumer state: %v", consumer.name, consumer.state.State)
	if consumer.state.State != pb.ScannerServiceControl_START {
		log.Info().Msgf("%s: start consume %s", consumer.name, payload)
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to requeue %s: %s", consumer.name, payload, err)
		} else {
			log.Debug().Msgf("%s: requeue %s, worker are stop", consumer.name, payload)
		}
		return
	}

	deferTime := time.Unix(int64(request.DeferDuration), 0).Unix()
	if deferTime <= time.Now().Unix() {
		key, result, err := consumer.engine.StartNmapScan(request)
		if err != nil && consumer.engine.State != pb.ScannerResponse_CANCEL {
			// scan failed
			consumer.failed <- 1
			// if scan fail or cancelled, mark task as cancel
			log.Error().Stack().Err(err).Msgf("%s: scan %v: %v", consumer.name, key, result)
			scannerResponse := &pb.ScannerResponse{
				Status: pb.ScannerResponse_ERROR,
			}
			scanResultJSON, err := json.Marshal(scannerResponse)
			if err != nil {
				log.Error().Stack().Err(err).Msgf("%s failed to parse failed result", consumer.name)
			}
			_, err = consumer.db.Set(
				context.Background(),
				key,
				string(scanResultJSON),
				time.Duration(request.GetRetentionTime())*time.Second,
			)
			if err != nil {
				log.Error().Stack().Err(err).Msgf("%s: failed to insert failed result", consumer.name)
			}
		} else {
			// success scan
			consumer.success <- 1
		}

		// if scan is cancel, result will be nil and we can't
		// parse the result
		if result != nil {
			scanResult, _ := engine.ParseScanResult(result)
			scannerResponse := &pb.ScannerResponse{
				HostResult: scanResult,
			}
			scanResultJSON, err := json.Marshal(scannerResponse)
			if err != nil {
				log.Error().Stack().Err(err).Msgf("%s failed to parse result", consumer.name)
			}
			_, err = consumer.db.Set(
				consumer.ctx, key, string(scanResultJSON),
				time.Duration(request.GetRetentionTime())*time.Second,
			)
			if err != nil {
				log.Error().Stack().Err(err).Msgf("%s: failed to insert result", consumer.name)
			}

		}
		if err := delivery.Ack(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to ack %s: %s", consumer.name, payload)
		} else {
			log.Info().Msgf("%s: acked %s", consumer.name, payload)
		}
	} else {
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to reject %s", consumer.name, payload)
		} else {
			log.Debug().Msgf("%s: delayed %s, this is too early", consumer.name, payload)
		}
	}
	time.Sleep(conf.RMQ.ConsumeDuration)
}
