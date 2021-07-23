package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/auth"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/helpers"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/mikioh/ipaddr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

var (
	confLogger = config.GetConfig().Logger
	log        = logger.New(confLogger.Debug, confLogger.Pretty)
	opsSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "scanner_processed_ops_success",
		Help: "The total number of success tasks",
	})
	opsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "scanner_processed_ops_failed",
		Help: "The total number of failed tasks",
	})
	opsReturned = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_ops_returned",
		Help: "The total number of returned tasks",
	})
	opsReady = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_queue_ready",
		Help: "The total number of ready tasks",
	})
	opsRejected = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_size_queue_rejected",
		Help: "The total number of rejected tasks",
	})
	workersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_workers_count",
		Help: "The total number of scanner workers",
	})
	consumersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "scanner_consumers_count",
		Help: "The total number of scanner consumers",
	})

	kaep = keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}
	kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     3600 * time.Second, // If a client is idle for 3600 seconds, send a GOAWAY
		MaxConnectionAge:      1800 * time.Second, // If any connection is alive for more than 1800 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,    // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,    // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               2 * time.Second,    // Wait 1 second for the ping ack before assuming the connection is dead
	}
)

// TasksStatus define the status of all tasks
type TasksStatus struct {
	success  int64 //tassks success
	failed   int64 //tasks failure
	returned int64 //tasks returned
	// ready    int64 //queue tasks ready
	// rejected int64 //queue tasks rejected
}

// Server define the grpc server struct
type Server struct {
	ctx         context.Context
	config      config.Config
	db          database.Database
	err         error
	broker      broker.Broker
	State       pb.ServiceStateValues
	locker      *redislock.Client
	taskType    string
	tasksStatus TasksStatus
	jwtManager  *auth.JWTManager
}

// NewServer create a new server and init the database connection
func NewServer(ctx context.Context, conf config.Config, taskType string) *Server {
	errChan := make(chan error)
	go broker.RmqLogErrors(errChan)

	//redis
	redisClient := helpers.NewRedisClient(ctx, conf).Connect()

	// distributed lock - with redis
	locker := redislock.New(redisClient)

	// Broker init nmap queue
	brker := broker.New(ctx, taskType, conf.RMQ, redisClient)

	// clean the queues
	go brker.Cleaner()
	log.Debug().Msg("DEBUUUUGG 3")

	// Storage database init
	db, err := database.Factory(ctx, conf)
	log.Debug().Msg("DEBUUUUGG 4444")
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("can't open the database")
	}

	log.Debug().Msg("DEBUUUUGGGG 7")
	// start the rmq stats to prometheus
	go brokerStatsToProm(brker, taskType)

	log.Debug().Msg("DEBUUUUGGG 8")
	// allocate the jwtManager
	var jwtManager = new(auth.JWTManager)

	log.Debug().Msg("DEBUUUUGGG 9")
	return &Server{
		ctx:        ctx,
		taskType:   taskType,
		config:     conf,
		broker:     *brker,
		db:         db,
		err:        err,
		locker:     locker,
		jwtManager: jwtManager,
	}
}

// Listen start the frontend grpc server
func Listen(ctx context.Context, conf config.Config) {
	log.Info().Msg("prepare to serve the gRPC api")

	// Start the server
	server := NewServer(ctx, conf, "nmap")

	// Serve the frontend
	go func(s *Server) {
		frontListener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.FrontendListenPort))
		if err != nil {
			log.Fatal().Err(err).Msgf(
				"can't start frontend server, can't listen on tcp/%d", conf.FrontendListenPort,
			)
		}

		tlsCredentials, err := auth.LoadTLSCredentials(
			s.config.Frontend.CAFile,
			s.config.Frontend.ServerCertFile,
			s.config.Frontend.ServerKeyFile,
		)
		if err != nil {
			log.Fatal().Msgf("cannot load TLS credentials: %v", err)
		}

		interceptor := auth.NewAuthInterceptor(server.jwtManager, accessibleRoles())
		srvFrontend := grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(kaep),
			grpc.KeepaliveParams(kasp),
			grpc.Creds(tlsCredentials),
			grpc.UnaryInterceptor(interceptor.Unary()),
			grpc.StreamInterceptor(interceptor.Stream()),
		)

		// graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for range c {
				// sig is a ^C, handle it
				log.Info().Msg("shutting down gRPC server...")

				srvFrontend.GracefulStop()

				<-ctx.Done()
			}
		}()

		reflection.Register(srvFrontend)

		userStore := auth.NewInMemoryUserStore()
		err = seedUsers(userStore)
		if err != nil {
			log.Fatal().Msgf("cannot seed users: ", err)
		} else {
			log.Debug().Msg("users seeded")
		}

		jwtManager := auth.NewJWTManager(secretKey, tokenDuration)
		authServer := auth.NewAuthServer(userStore, jwtManager)
		pb.RegisterAuthServiceServer(srvFrontend, authServer)

		pb.RegisterScannerServiceServer(srvFrontend, server)

		if err = srvFrontend.Serve(frontListener); err != nil {
			log.Fatal().Msg("can't serve the gRPC frontend service")
		}
	}(server)

	// Serve the backend
	backendListener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.BackendListenPort))
	if err != nil {
		log.Fatal().Err(err).Msgf(
			"can't start backend server, can't listen on tcp/%d", conf.BackendListenPort,
		)
	}

	tlsCredentials, err := auth.LoadTLSCredentials(
		conf.Backend.CAFile,
		conf.Backend.ServerCertFile,
		conf.Backend.ServerKeyFile,
	)

	if err != nil {
		log.Fatal().Msgf("cannot load TLS credentials: %v", err)
	}

	interceptor := auth.NewAuthInterceptor(server.jwtManager, accessibleRoles())
	srvBackend := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.Creds(tlsCredentials),
		grpc.StreamInterceptor(interceptor.Stream()),
		grpc.UnaryInterceptor(interceptor.Unary()),
	)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			// sig is a ^C, handle it
			log.Info().Msg("shutting down gRPC server...")
			srvBackend.GracefulStop()
			<-ctx.Done()
		}
	}()

	reflection.Register(srvBackend)

	userStore := auth.NewInMemoryUserStore()
	err = seedUsers(userStore)
	if err != nil {
		log.Fatal().Msgf("cannot seed users: ", err)
	} else {
		log.Debug().Msg("users seeded")
	}

	jwtManager := auth.NewJWTManager(secretKey, tokenDuration)
	authServer := auth.NewAuthServer(userStore, jwtManager)
	pb.RegisterAuthServiceServer(srvBackend, authServer)

	pb.RegisterBackendServiceServer(srvBackend, server)
	if err = srvBackend.Serve(backendListener); err != nil {
		log.Fatal().Msg("can't serve the gRPC backend service")
	}
}

// DeleteScan delele a scan from the database
func (server *Server) DeleteScan(ctx context.Context, request *pb.GetScannerRequest) (*pb.ServerResponse, error) {
	_, err := server.db.Delete(ctx, request.Key)
	return generateResponse(request.Key, nil, err)
}

// StreamServiceControl control the service
func (server *Server) StreamServiceControl(request *pb.ServiceStateValues, stream pb.BackendService_StreamServiceControlServer) error {
	for {
		if err := stream.Send(&server.State); err != nil {
			log.Error().Stack().Err(err).Msgf("streamer service control send error")
			return errors.Wrap(err, "streamer service control send error")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// StreamTasksStatus manage the tasks counter
func (server *Server) StreamTasksStatus(stream pb.BackendService_StreamTasksStatusServer) error {
	for {
		promStatus, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			return err
		}
		server.tasksStatus.success += promStatus.TasksStatus.Success
		opsSuccess.Add(float64(promStatus.TasksStatus.Success))
		server.tasksStatus.failed += promStatus.TasksStatus.Failed
		opsFailed.Add(float64(promStatus.TasksStatus.Failed))
		server.tasksStatus.returned += promStatus.TasksStatus.Returned
		opsReturned.Set(float64(promStatus.TasksStatus.Returned))
		stream.Send(&pb.PrometheusStatus{TasksStatus: &pb.TasksStatus{
			Success: server.tasksStatus.success, Failed: server.tasksStatus.failed, Returned: server.tasksStatus.returned}},
		)

		time.Sleep(500 * time.Millisecond)
	}
}

// ServiceControl control the service
func (server *Server) ServiceControl(ctx context.Context, request *pb.ServiceStateValues) (*pb.ServiceStateValues, error) {
	if request.GetState() == pb.ServiceStateValues_UNKNOWN {
		return &pb.ServiceStateValues{State: server.State.State}, nil
	}

	if server.State.State != request.GetState() {
		server.State.State = request.GetState()
	}

	return &pb.ServiceStateValues{State: server.State.State}, nil
}

// GetScan return the engine scan result
func (server *Server) GetScan(ctx context.Context, request *pb.GetScannerRequest) (*pb.ServerResponse, error) {
	var smr pb.ScannerMainResponse

	dbSmr, err := server.db.Get(ctx, request.Key)
	if err != nil {
		return generateResponse(request.Key, nil, err)
	}

	err = json.Unmarshal([]byte(dbSmr), &smr)
	if err != nil {
		log.Error().Stack().Err(err).Msg("can't read scan result")
		return generateResponse(request.Key, nil, err)
	}

	return generateResponse(request.Key, &smr, nil)
}

// GetAllScans return the engine scan result
func (server *Server) GetAllScans(in *empty.Empty, stream pb.ScannerService_GetAllScansServer) error {
	allKeys, err := server.db.GetAll(server.ctx, "*")
	if err != nil {
		return err
	}

	for _, key := range allKeys {
		var smr pb.ScannerMainResponse
		dbSmr, err := server.db.Get(server.ctx, key)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get scan result")
			continue
		}

		err = json.Unmarshal([]byte(dbSmr), &smr)
		smr.Key = key
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't list scan result")
			return err
		}

		response, err := generateResponse(key, &smr, err)
		if err != nil {
			return err
		}
		stream.Send(response)
	}

	return nil
}

// StartScan function prepare a nmap scan
func (server *Server) StartScan(ctx context.Context, params *pb.ParamsScannerRequest) (*pb.ServerResponse, error) {
	params = parseRequest(params)

	createAt := timestamppb.Now()

	// we start the scan
	newEngine := engine.New(ctx, server.db, server.config.NMAP, server.locker)
	sr := []*pb.ScannerResponse{{
		StartTime: timestamppb.Now(),
		Status:    pb.ScannerResponse_UNKNOWN,
	}}
	params, result, err := newEngine.Start(params, false)
	if err != nil {
		return generateResponse(params.Key, nil, err)
	}
	// end of scan
	sr[0].EndTime = timestamppb.Now()

	// parse result
	scanParsedResult, err := engine.ParseScanResult(result)
	if err != nil {
		return generateResponse(params.Key, nil, err)
	}
	sr[0].HostResult = scanParsedResult
	sr[0].Status = pb.ScannerResponse_OK

	// forge main response
	smr := pb.ScannerMainResponse{
		Key:      params.Key,
		CreateAt: createAt,
		Response: sr,
	}
	smrJSON, err := json.Marshal(&smr)
	if err != nil {
		log.Error().Stack().Err(err).Msg("can't create main JSON response")
	}

	// and write the response to the database
	_, err = server.db.Set(
		ctx, params.Key,
		string(smrJSON),
		params.RetentionDuration.AsDuration(),
	)
	if err != nil {
		return generateResponse(params.Key, &smr, err)
	}

	return generateResponse(params.Key, &smr, nil)
}

func splitInSubnets(targets string, n int) ([]string, error) {
	var networks []string

	_, network, err := net.ParseCIDR(targets)
	if err != nil {
		return nil, err
	}

	p := ipaddr.NewPrefix(network)
	subnetPrefixes := p.Subnets(n)

	for _, subnet := range subnetPrefixes {
		networks = append(networks, subnet.String())
	}

	return networks, nil
}

// StartAsyncScan function prepare a nmap scan
func (server *Server) StartAsyncScan(ctx context.Context, params *pb.ParamsScannerRequest) (*pb.ServerResponse, error) {

	params = parseRequest(params)
	targets := params.Targets
	createAt := timestamppb.Now()

	srs := []*pb.ScannerResponse{}
	smr := pb.ScannerMainResponse{}

	// if we want to split a task by host
	// we split host list by "," and then
	// spawn a consumer per host
	//
	// also, we split a network cidr in smaller one
	var hosts []string
	if params.NetworkChuncked || params.ProcessPerTarget {
		hosts = strings.Split(params.Targets, ",")
		if params.NetworkChuncked {
			var splittedNetworks []string
			for _, network := range hosts {
				splittedNetwork, err := splitInSubnets(network, 3)
				if err != nil {
					splittedNetworks = append(splittedNetworks, network)
				}
				splittedNetworks = append(splittedNetworks, splittedNetwork...)
			}
			hosts = splittedNetworks
		}
	} else {
		// or we use the default behavior, passing all host
		// to one nmap process
		hosts = []string{params.Targets}
	}

	// we loop of each host
	for i, host := range hosts {
		// if we split the scan, let use a key ID containing
		// the main Key and host
		subKey := params.Key
		if (params.NetworkChuncked || params.ProcessPerTarget) && len(hosts) > 1 {
			subKey = fmt.Sprintf("%s-%s", params.Key, host)
		}

		// we create the task response with queued status
		sr := &pb.ScannerResponse{
			Key:    subKey,
			Status: pb.ScannerResponse_QUEUED,
		}

		// then append it to the main task
		srs = append(srs, sr)
		// we override the Hosts param in case we split it
		// for the scan request
		params.Targets = host
		smr = pb.ScannerMainResponse{
			Key:      params.Key,
			Request:  params,
			CreateAt: createAt,
			Response: srs,
		}

		log.Info().Msgf("receive async task order: %v", params)

		smrJSON, err := json.Marshal(&smr)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't create main JSON response")
			return generateResponse(params.Key, &smr, err)
		}

		// write main response to database
		_, err = server.db.Set(ctx, params.Key, string(smrJSON), 0)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't write main response to database")
			return generateResponse(params.Key, &smr, err)
		}

		// create scan task and publish it to the RMQ broker
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			srs[i].Status = pb.ScannerResponse_ERROR
			return generateResponse(params.Key, &smr, err)
		}

		err = server.broker.Incoming.PublishBytes(paramsJSON)
		if err != nil {
			srs[i].Status = pb.ScannerResponse_ERROR
			return generateResponse(params.Key, &smr, err)
		}

	}

	params.Targets = targets
	smr = pb.ScannerMainResponse{
		Key:      params.Key,
		Request:  params,
		CreateAt: createAt,
		Response: srs,
	}
	return generateResponse(params.Key, &smr, nil)
}

// generateResponse generate the response for the grpc return
func generateResponse(key string, value *pb.ScannerMainResponse, err error) (*pb.ServerResponse, error) {
	if err != nil {
		return &pb.ServerResponse{
			Success: false,
			Key:     key,
			Value:   value,
			Error:   err.Error(),
		}, nil
	}
	return &pb.ServerResponse{
		Success: true,
		Key:     key,
		Value:   value,
		Error:   "",
	}, nil
}

// parseRequest parse, sanitize the request
func parseRequest(request *pb.ParamsScannerRequest) *pb.ParamsScannerRequest {
	// if the Key is not forced, we generate one unique
	guid := uuid.New()
	if request.Key == "" {
		request.Key = guid.String()
	}

	// If timeout < 10s, fallback to 1h
	if (request.Timeout < 30 && request.Timeout > 0) || request.Timeout == 0 {
		request.Timeout = 60 * 60
	}

	if request.RetentionDuration == nil {
		request.RetentionDuration = durationpb.New(0)
	}

	// replace all whitespaces
	request.Targets = strings.ReplaceAll(request.Targets, " ", "")
	request.Ports = strings.ReplaceAll(request.Ports, " ", "")

	// defer Duration to defer Time
	request.DeferTime = timestamppb.New(time.Now().Add(request.DeferDuration.AsDuration()))

	return request
}

// brokerStatsToProm read broker stats each 2s and write to prometheus
func brokerStatsToProm(brker *broker.Broker, name string) {
	var prevReady, prevReject int64
	for range time.Tick(1000 * time.Millisecond) {
		stats, err := brker.GetStats()
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get RMQ statistics")
			continue
		}
		queue := fmt.Sprintf("%s-incoming", name)
		s := stats.QueueStats[queue]
		opsReady.Set(float64(s.ReadyCount))
		opsRejected.Set(float64(s.RejectedCount))
		consumersCount.Set(float64(s.ConsumerCount()))
		if (s.ReadyCount > 0 || s.RejectedCount > 0) && (prevReady != s.ReadyCount || prevReject != s.RejectedCount) {
			log.Debug().Msgf("broker %s incoming queue ready: %v, rejected: %v", name, s.ReadyCount, s.RejectedCount)
		}
		prevReady = s.ReadyCount
		prevReject = s.RejectedCount
	}
}

// createUser create the users
func createUser(userStore auth.UserStore, username, password, role string) error {
	user, err := auth.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userStore.Save(user)
}

func seedUsers(userStore auth.UserStore) error {
	err := createUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}
	err = createUser(userStore, "worker1", "secret1", "worker")
	if err != nil {
		return err
	}
	return createUser(userStore, "user1", "secret", "user")
}

func accessibleRoles() map[string][]string {
	const backendServicePath = "/proto.BackendService/"
	const frontendServicePath = "/proto.ScannerService/"

	return map[string][]string{
		backendServicePath + "StreamServiceControl": {"worker", "admin"},
		backendServicePath + "StreamTasksStatus":    {"worker", "admin"},

		frontendServicePath + "StartAsyncScan": {"user"},
		frontendServicePath + "StartScan":      {"user"},

		frontendServicePath + "GetScan": {"user", "admin"},

		frontendServicePath + "GetAllScans":    {"admin"},
		frontendServicePath + "ServiceControl": {"admin"},
	}
}
