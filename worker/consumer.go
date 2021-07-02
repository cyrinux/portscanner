package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
)

// Consumer define a broker consumer
type Consumer struct {
	name      string
	tasktype  string
	count     int64
	scanCount map[int]int64
	before    time.Time
	ctx       context.Context
	cancel    context.CancelFunc
	engine    *engine.Engine
	state     proto.ScannerServiceControl
	db        database.Database
}

// NewConsumer create a new consumer
func NewConsumer(
	ctx context.Context,
	db database.Database, tag int, tasktype string,
	queue string) (string, *Consumer) {

	name := fmt.Sprintf("%s-consumer-%s-%s-%d", tasktype, queue, hostname, tag)
	log.Info().Msgf("new consumer: %s", name)
	ctx, cancel := context.WithCancel(ctx)
	engine := engine.NewEngine(ctx, db)

	return name, &Consumer{
		ctx:      ctx,
		cancel:   cancel,
		name:     name,
		tasktype: tasktype,
		count:    0,
		before:   time.Now(),
		engine:   engine,
		db:       db,
	}
}

// onCancel is a function trigger on consumer context cancel
func onCancel(consumer *Consumer, request *proto.ParamsScannerRequest) {
	// waiting for cancel signal
	<-consumer.ctx.Done()
	log.Debug().Msgf("%s cancelled, writing state to database", consumer.name)
	// if scan fail or cancelled, mark task as cancel
	consumer.engine.State = proto.ScannerResponse_CANCEL
	scannerResponse := &proto.ScannerResponse{
		Status: proto.ScannerResponse_CANCEL,
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

	var request *proto.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &request)

	go onCancel(consumer, request)

	log.Debug().Msgf("%s: consumer state: %v", consumer.name, consumer.state.State)
	if consumer.state.State != proto.ScannerServiceControl_START {
		log.Info().Msgf("%s: start consume %s", consumer.name, payload)
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to requeue %s: %s", consumer.name, payload, err)
		} else {
			log.Debug().Msgf("%s: requeue %s, worker are stop", consumer.name, payload)
		}
		return
	}

	consumer.count++
	if consumer.count%conf.RMQ.ReportBatchSize == 0 {
		duration := time.Since(consumer.before)
		consumer.before = time.Now()
		perSecond := time.Second / (duration / time.Duration(conf.RMQ.ReportBatchSize))
		log.Debug().Msgf("%s: consumed %d %s %d", consumer.name, consumer.count, payload, perSecond)
	}

	deferTime := time.Unix(int64(request.DeferDuration), 0).Unix()
	if deferTime <= time.Now().Unix() {

		key, result, err := consumer.engine.StartNmapScan(request)
		if err != nil && consumer.engine.State != proto.ScannerResponse_CANCEL {
			// if scan fail or cancelled, mark task as cancel
			log.Error().Stack().Err(err).Msgf("%s: scan %v: %v", consumer.name, key, result)
			scannerResponse := &proto.ScannerResponse{
				Status: proto.ScannerResponse_ERROR,
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
		}

		// if scan is cancel, result will be nil and we can't
		// parse the result
		if result != nil {
			scanResult, _ := engine.ParseScanResult(result)
			scannerResponse := &proto.ScannerResponse{
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

		if consumer.count%conf.RMQ.ReportBatchSize > 0 {
			if err := delivery.Ack(); err != nil {
				log.Error().Stack().Err(err).Msgf("%s: failed to ack %s: %s", consumer.name, payload)
			} else {
				log.Info().Msgf("%s: acked %s", consumer.name, payload)
			}
		} else { // reject one per batch
			if err := delivery.Reject(); err != nil {
				log.Error().Stack().Err(err).Msgf("%s: failed to reject %s", consumer.name, payload)
			} else {
				log.Info().Msgf("%s: rejected %s", consumer.name, payload)
			}
		}
	} else {
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to reject %s", consumer.name, payload)
		} else {
			log.Debug().Msgf("%s: delayed %s, this is too early", consumer.name, payload)
		}
	}

	time.Sleep(conf.RMQ.ConsumeDuration * time.Millisecond)
}
