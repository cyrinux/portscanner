package worker

import (
	"context"
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	// "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

// Consumer define a broker consumer
type Consumer struct {
	name      string
	count     int64
	scanCount map[int]int64
	before    time.Time
	ctx       context.Context
	worker    *Worker
}

// NewConsumer create a new consumer
func NewConsumer(worker *Worker, tag int, queue string) (string, *Consumer) {
	name := fmt.Sprintf("%s-consumer-%s-%d", queue, hostname, tag)
	log.Info().Msgf("New consumer: %s", name)
	return name, &Consumer{
		ctx:       context.Background(),
		name:      name,
		count:     0,
		scanCount: make(map[int]int64),
		before:    time.Now(),
		worker:    worker,
	}
}

// Consume consume the message tasks on the redis broker
func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	time.Sleep(consumeDuration)
	payload := delivery.Payload()

	log.Debug().Msgf("%s: consumer state: %v", consumer.name, consumer.worker.state.State)
	if consumer.worker.state.State != proto.ScannerServiceControl_START {
		log.Info().Msgf("%s: start consume %s", consumer.name, payload)
		if err := delivery.Reject(); err != nil {
			log.Error().Msgf("%s: failed to requeue %s: %s", consumer.name, payload, err)
		} else {
			log.Debug().Msgf("%s: requeue %s, worker are stop", consumer.name, payload)
		}
		return
	}

	consumer.count++
	if consumer.count%reportBatchSize == 0 {
		duration := time.Since(consumer.before)
		consumer.before = time.Now()
		perSecond := time.Second / (duration / reportBatchSize)
		log.Debug().Msgf("%s: consumed %d %s %d", consumer.name, consumer.count, payload, perSecond)
	}

	var request *proto.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &request)

	deferTime := time.Unix(int64(request.DeferDuration), 0).Unix()
	if deferTime <= time.Now().Unix() {
		newEngine := engine.NewEngine(consumer.ctx, consumer.worker.config, consumer.worker.db)
		key, result, err := newEngine.StartNmapScan(request)
		if err != nil {
			log.Error().Msgf("%s: scan %v %v: %v", consumer.name, key, result, err)
		}
		// consumer.scanCount[request.ScanSpeed]+

		key, scanResult, _ := engine.ParseScanResult(key, result)

		scannerResponse := &proto.ScannerResponse{
			HostResult: scanResult,
		}
		scanResultJSON, err := json.Marshal(scannerResponse)
		if err != nil {
			log.Error().Msgf("%s: failed to parse result: %s", consumer.name, err)
		}
		_, err = consumer.worker.db.Set(
			consumer.ctx, key, string(scanResultJSON),
			time.Duration(request.GetRetentionTime())*time.Second,
		)
		if err != nil {
			log.Error().Msgf("%s: failed to insert result: %s", consumer.name, err)
		}

		if consumer.count%reportBatchSize > 0 {
			if err := delivery.Ack(); err != nil {
				log.Error().Msgf("%s: failed to ack %s: %s", consumer.name, payload, err)
			} else {
				log.Info().Msgf("%s: acked %s", consumer.name, payload)
			}
		} else { // reject one per batch
			if err := delivery.Reject(); err != nil {
				log.Error().Msgf("%s: failed to reject %s: %s", consumer.name, payload, err)
			} else {
				log.Info().Msgf("%s: rejected %s", consumer.name, payload)
			}
		}
	} else {
		if err := delivery.Reject(); err != nil {
			log.Error().Msgf("%s: failed to reject %s: %s", consumer.name, payload, err)
		} else {
			log.Debug().Msgf("%s: delayed %s, this is too early", consumer.name, payload)
		}
	}
}
