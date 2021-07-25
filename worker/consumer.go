package worker

import (
	"context"
	"encoding/json"
	"fmt"
	rmq "github.com/adjust/rmq/v4"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

// Consumer define a broker consumer
type Consumer struct {
	name     string
	cancel   context.CancelFunc
	ctx      context.Context
	db       database.Database
	engine   *engine.Engine
	locker   *redislock.Client
	state    pb.ServiceStateValues
	conf     config.Config
	taskType string
	success  chan int64
	failed   chan int64
	request  *pb.ParamsScannerRequest
}

// NewConsumer create a new consumer
func NewConsumer(
	ctx context.Context,
	db database.Database,
	tag int,
	taskType string,
	conf config.NMAPConfig,
	queue string,
	locker *redislock.Client) (string, *Consumer) {

	name := fmt.Sprintf("%s-consumer-%s-%s-%d", taskType, queue, hostname, tag)
	log.Info().Msgf("new: %s", name)
	ctx, cancel := context.WithCancel(ctx)
	engine := engine.New(ctx, db, conf, locker)

	return name, &Consumer{
		ctx:      ctx,
		name:     name,
		cancel:   cancel,
		db:       db,
		engine:   engine,
		conf:     config.GetConfig(),
		taskType: taskType,
		success:  make(chan int64),
		failed:   make(chan int64),
	}
}

// onCancel is a function trigger on consumer context cancel
func (consumer *Consumer) Cancel() {
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

	if consumer.request != nil {
		_, err = consumer.db.Set(
			consumer.ctx,
			consumer.request.Key,
			string(scanResultJSON),
			consumer.request.DeferDuration.AsDuration(),
		)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to insert failed result", consumer.name)
		}
	}

	consumer.cancel()
}

// Consume consume the message tasks on the redis broker
func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()

	var request *pb.ParamsScannerRequest
	json.Unmarshal([]byte(payload), &request)

	log.Debug().Msgf("%s: consumer state: %v", consumer.name, consumer.state.State)
	if consumer.state.State != pb.ServiceStateValues_START {
		log.Info().Msgf("%s: start consume %s", consumer.name, payload)
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to requeue %s: %s", consumer.name, payload, err)
		} else {
			log.Debug().Msgf("%s: requeue %s, worker are stop", consumer.name, payload)
		}
		return
	}

	if request.DeferTime.AsTime().Unix() <= time.Now().Unix() {
		consumer.consumeNow(delivery, request, payload)
	} else {
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to reject %s", consumer.name, payload)
		} else {
			log.Debug().Msgf("%s: delayed %s, this is too early", consumer.name, payload)
		}
	}

	time.Sleep(consumer.conf.RMQ.ConsumeDuration)
}

// consumeNow really consume the message
func (consumer *Consumer) consumeNow(delivery rmq.Delivery, request *pb.ParamsScannerRequest, payload string) {
	consumer.request = request

	var smr pb.ScannerMainResponse
	var srs []*pb.ScannerResponse

	startTime := timestamppb.Now()
	params, result, err := consumer.engine.Start(request, true)
	endTime := timestamppb.Now()

	if err != nil && consumer.engine.State != pb.ScannerResponse_CANCEL {
		// scan failed
		consumer.failed <- 1
		// if scan fail or cancelled, mark task as cancel
		log.Error().Stack().Err(err).Msgf("%s: scan %s: %v", consumer.name, params.Key, result)
		smrJSON, err := consumer.db.Get(consumer.ctx, request.Key)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to get main response, key: %s", consumer.name, request.Key)
		}
		err = json.Unmarshal([]byte(smrJSON), &smr)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to read response from json", consumer.name)
		}
		srs = smr.Response
		srs = append(srs, &pb.ScannerResponse{
			StartTime:         startTime,
			Status:            pb.ScannerResponse_ERROR,
			EndTime:           endTime,
			RetentionDuration: params.RetentionDuration},
		)
		smr = pb.ScannerMainResponse{
			Key:      params.Key,
			Request:  params,
			Response: srs,
		}
		scanResultJSON, err := json.Marshal(&smr)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to parse failed result", consumer.name)
		}
		_, err = consumer.db.Set(
			consumer.ctx,
			params.Key,
			string(scanResultJSON),
			params.RetentionDuration.AsDuration(),
		)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to insert failed result", consumer.name)
		}
	}

	// if scan is cancel, result will be nil and we can't
	// parse the result
	if err == nil && result != nil {
		// success scan
		consumer.success <- 1

		scanResult, err := engine.ParseScanResult(result)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s can't parse the scan result, key: %s", consumer.name, request.Key)
		}
		smrJSON, err := consumer.db.Get(consumer.ctx, request.Key)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to get main response, key: %s", consumer.name, request.Key)
		}
		err = json.Unmarshal([]byte(smrJSON), &smr)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to read response from json", consumer.name)
		}
		srs = smr.Response
		childKey := fmt.Sprintf("%s-%s", params.Key, request.Targets)
		srs = append(srs, &pb.ScannerResponse{
			Key:               childKey,
			StartTime:         startTime,
			HostResult:        scanResult,
			EndTime:           endTime,
			Status:            pb.ScannerResponse_OK,
			RetentionDuration: params.RetentionDuration},
		)
		smr.Response = srs

		// change child scan status to OK
		for i := range smr.Response {
			if smr.Response[i].Key == childKey {
				smr.Response[i].Status = pb.ScannerResponse_OK
			}
		}

		scanResultJSON, err := json.Marshal(&smr)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to parse result", consumer.name)
		}
		_, err = consumer.db.Set(
			consumer.ctx, params.Key, string(scanResultJSON), request.RetentionDuration.AsDuration(),
		)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to insert result", consumer.name)
		}
	}
	if err := delivery.Ack(); err != nil {
		log.Error().Stack().Err(err).Msgf("%s: failed to ack :%v", consumer.name, payload)
	} else {
		log.Info().Msgf("%s: acked %v", consumer.name, payload)
	}
}
