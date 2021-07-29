package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Ullaakut/nmap/v2"
	rmq "github.com/adjust/rmq/v4"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/locker"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"time"
)

var (
	conf        = config.GetConfig().Logger
	log         = logger.New(conf.Debug, conf.Pretty)
	hostname, _ = os.Hostname()
)

// Consumer define a broker consumer
type Consumer struct {
	ctx       context.Context
	db        database.Database
	Name      string
	cancel    context.CancelFunc
	Engine    *engine.Engine
	Locker    locker.MyLockerInterface
	State     pb.ServiceStateValues
	conf      config.Config
	taskType  string
	Success   chan int64
	Failed    chan int64
	request   *pb.ParamsScannerRequest
	startTime *timestamppb.Timestamp
	endTime   *timestamppb.Timestamp
}

// New create a new consumer
func New(
	ctx context.Context,
	db database.Database,
	tag int,
	taskType string,
	conf config.NMAPConfig,
	queue string,
	locker locker.MyLockerInterface) (string, *Consumer) {

	name := fmt.Sprintf("%s-consumer-%s-%s-%d", taskType, queue, hostname, tag)
	log.Info().Msgf("new: %s", name)
	ctx, cancel := context.WithCancel(ctx)
	newEngine := engine.New(ctx, db, conf, locker)

	return name, &Consumer{
		ctx:      ctx,
		Name:     name,
		cancel:   cancel,
		db:       db,
		Engine:   newEngine,
		Locker:   locker,
		conf:     config.GetConfig(),
		taskType: taskType,
		Success:  make(chan int64),
		Failed:   make(chan int64),
	}
}

// Cancel is a function trigger on consumer context cancel
func (consumer *Consumer) Cancel() {
	log.Debug().Msgf("%s cancelled, writing state to database", consumer.Name)
	// if scan fail or cancelled, mark task as cancel
	consumer.Engine.State = pb.ScannerResponse_CANCEL
	scannerResponse := &pb.ScannerResponse{
		Status: pb.ScannerResponse_CANCEL,
	}
	scanResultJSON, err := json.Marshal(scannerResponse)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s failed to parse failed result", consumer.Name)
	}

	if consumer.request != nil {
		_, err = consumer.db.Set(
			consumer.ctx,
			consumer.request.Key,
			string(scanResultJSON),
			consumer.request.DeferDuration.AsDuration(),
		)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to insert failed result", consumer.Name)
		}
	}

	time.Sleep(1000 * time.Millisecond)

	consumer.cancel()
}

// Consume consume the message tasks on the redis broker
func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()

	var request *pb.ParamsScannerRequest
	err := json.Unmarshal([]byte(payload), &request)
	if err != nil {
		return
	}

	log.Debug().Msgf("%s: consumer state: %v", consumer.Name, consumer.State.State)

	if consumer.State.State != pb.ServiceStateValues_START {
		log.Info().Msgf("%s: start consume %s", consumer.Name, payload)
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to requeue %s: %s", consumer.Name, payload, err)
		} else {
			log.Debug().Msgf("%s: requeue %s, worker are stop", consumer.Name, payload)
		}
		return
	}

	if request.DeferTime.AsTime().Unix() <= time.Now().Unix() {
		consumer.consumeNow(delivery, request, payload)
	} else {
		if err := delivery.Reject(); err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to reject %s", consumer.Name, payload)
		} else {
			log.Debug().Msgf("%s: delayed %s, this is too early", consumer.Name, payload)
		}
	}

	time.Sleep(consumer.conf.RMQ.ConsumeDuration)
}

func (consumer *Consumer) markTaskCancelled(params *pb.ParamsScannerRequest, result *nmap.Run) error {
	var scannerMainResponse pb.ScannerMainResponse
	var scannerResponses []*pb.ScannerResponse

	if consumer.Engine.State != pb.ScannerResponse_CANCEL {
		// scan failed
		consumer.Failed <- 1
		// if scan fail or cancelled, mark task as cancel

		smrJSON, err := consumer.db.Get(consumer.ctx, consumer.request.Key)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to get main response, key: %s", consumer.Name, consumer.request.Key)
			return err
		}
		err = json.Unmarshal([]byte(smrJSON), &scannerMainResponse)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to read response from json", consumer.Name)
			return err
		}
		scannerResponses = scannerMainResponse.Response
		scannerResponses = append(scannerResponses, &pb.ScannerResponse{
			StartTime:         consumer.startTime,
			Status:            pb.ScannerResponse_ERROR,
			EndTime:           consumer.endTime,
			RetentionDuration: params.RetentionDuration},
		)
		scannerMainResponse = pb.ScannerMainResponse{
			Key:      params.Key,
			Request:  params,
			Response: scannerResponses,
		}
		scanResultJSON, err := json.Marshal(&scannerMainResponse)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s failed to parse failed result", consumer.Name)
			return err
		}
		_, err = consumer.db.Set(
			consumer.ctx,
			params.Key,
			string(scanResultJSON),
			params.RetentionDuration.AsDuration(),
		)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("%s: failed to insert failed result", consumer.Name)
			return err
		}

	}

	return nil
}

// consumeNow really consume the message
func (consumer *Consumer) consumeNow(delivery rmq.Delivery, request *pb.ParamsScannerRequest, payload string) error {
	consumer.request = request

	// fmt.Printf("%#v", request)
	// fmt.Printf("%v", request)
	// fmt.Printf("%+v", request)

	var scannerMainResponse pb.ScannerMainResponse
	var scannerResponses []*pb.ScannerResponse

	consumer.startTime = timestamppb.Now()
	params, result, err := consumer.Engine.Start(consumer.request, true)
	consumer.endTime = timestamppb.Now()

	// mark the task as cancel if fail
	if err != nil {
		log.Error().Stack().Err(err).Msgf("%s: scan %s: %v", consumer.Name, params.Key, result)
		err := consumer.markTaskCancelled(params, result)
		if err != nil {
			return err
		}
	}

	// if scan is cancel, result will be nil and we can't
	// parse the result
	if err == nil && result != nil {
		// success scan
		consumer.Success <- 1

		wait := 500 * time.Millisecond
		for {
			// cpu cooling
			time.Sleep(wait)

			// Try to obtain lock.
			lockerKey := fmt.Sprintf("consumer-%s", request.Key)
			ok, err := consumer.Locker.Obtain(consumer.ctx, lockerKey, 10*time.Second)
			if err != nil {
				log.Error().Stack().Err(err).Msg("returner can't obtain lock")
			} else if ok {
				defer consumer.Locker.Release(consumer.ctx, lockerKey)
				// Sleep and check the remaining TTL.
				if ttl, err := consumer.Locker.TTL(consumer.ctx, lockerKey); err != nil {
					log.Error().Stack().Err(err).Msgf("returner error, ttl: %v", ttl)
				} else if ttl > 0 {

					scannerMainResponseJSON, err := consumer.db.Get(consumer.ctx, request.Key)
					if err != nil {
						log.Error().Stack().Err(err).Msgf("%s failed to get main response, key: %s", consumer.Name, request.Key)
					}
					err = json.Unmarshal([]byte(scannerMainResponseJSON), &scannerMainResponse)
					if err != nil {
						log.Error().Stack().Err(err).Msgf("%s failed to read response from json", consumer.Name)
						return err
					}
					fmt.Printf("%+v", &scannerMainResponse)

					childKey := request.Key
					if (request.ProcessPerTarget || request.NetworkChuncked) && len(request.Targets) > 1 {
						childKey = fmt.Sprintf("%s-%s", params.Key, request.Targets)
					}

					scannerResponses = scannerMainResponse.Response

					scannerResponse := &pb.ScannerResponse{
						Key:               childKey,
						StartTime:         consumer.startTime,
						EndTime:           consumer.endTime,
						RetentionDuration: params.RetentionDuration,
					}
					scannerResponse.Status = pb.ScannerResponse_OK

					scanResult, err := engine.ParseScanResult(result)
					if err != nil {
						log.Error().Stack().Err(err).Msgf("%s can't parse the scan result, key: %s", consumer.Name, childKey)
						scannerResponse.Status = pb.ScannerResponse_ERROR
					}
					scannerResponse.HostResult = scanResult

					for i := range scannerResponses {
						err := consumer.Locker.Refresh(consumer.ctx, lockerKey, 1*time.Second)
						if err != nil {
							return err
						}
						if scannerResponses[i].Key == childKey {
							scannerResponses[i] = scannerResponse
							break
						}
					}

					scannerMainResponse.Response = scannerResponses

					scanResultJSON, err := json.Marshal(&scannerMainResponse)
					if err != nil {
						log.Error().Stack().Err(err).Msgf("%s failed to parse result", consumer.Name)
					}

					err = consumer.Locker.Refresh(consumer.ctx, lockerKey, 2*time.Second)
					if err != nil {
						return err
					}

					_, err = consumer.db.Set(
						consumer.ctx, params.Key, string(scanResultJSON), request.RetentionDuration.AsDuration(),
					)
					if err != nil {
						log.Error().Stack().Err(err).Msgf("%s: failed to insert result", consumer.Name)
					}
				}
				if err := delivery.Ack(); err != nil {
					log.Error().Stack().Err(err).Msgf("%s: failed to ack :%v", consumer.Name, payload)
				} else {
					log.Info().Msgf("%s: was acked %v", consumer.Name, payload)
				}

				break
			}

		}
	}

	return nil
}
