package worker

import (
	"context"
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"log"
	"os"
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

// Consumer define a broker consumer
type Consumer struct {
	name   string
	count  int
	before time.Time
	ctx    context.Context
	config config.Config
}

// NewConsumer create a new consumer
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
