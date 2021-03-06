package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/cyrinux/grpcnmapscanner/auth"
	"github.com/cyrinux/grpcnmapscanner/broker"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/engine"
	"github.com/cyrinux/grpcnmapscanner/helpers"
	"github.com/cyrinux/grpcnmapscanner/locker"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/google/uuid"
	"github.com/mikioh/ipaddr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	stateStatusKey = "worker-state"
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
	locker      locker.MyLockerInterface
	taskType    string
	tasksStatus TasksStatus
	jwtManager  *auth.JWTManager
	userStore   *auth.InMemoryUserStore
}

// NewServer create a new server and init the database connection
func NewServer(ctx context.Context, conf config.Config, taskType string) *Server {
	errChan := make(chan error)
	go broker.RmqLogErrors(errChan)

	//redis
	redisClient := helpers.NewRedisClient(ctx, conf).Connect()

	// distributed lock - with redis
	myLocker := locker.CreateRedisLock(redisClient)

	// Broker init nmap queue
	myBroker := broker.New(ctx, taskType, conf.RMQ, redisClient)

	// clean the queues
	go myBroker.Cleaner()

	// Storage database init
	db, err := database.Factory(ctx, conf)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("can't open the database")
	}

	// start the rmq stats to prometheus
	go brokerStatsToProm(myBroker, taskType)

	// allocate the jwtManager
	jwtManager := auth.NewJWTManager(
		conf.Backend.JWT.SecretKey,
		conf.Backend.JWT.TokenDuration,
	)

	// seed the users
	userStore := auth.NewInMemoryUserStore()
	err = seedUsers(userStore)
	if err != nil {
		log.Fatal().Stack().Err(err).Msgf("cannot seed users")
	} else {
		log.Debug().Msg("database users seeded")
	}

	return &Server{
		ctx:        ctx,
		taskType:   taskType,
		config:     conf,
		broker:     *myBroker,
		db:         db,
		err:        err,
		locker:     myLocker,
		jwtManager: jwtManager,
		userStore:  userStore,
	}
}

// Listen start the frontend grpc server
func Listen(ctx context.Context, conf config.Config) error {
	log.Info().Msg("prepare to serve the gRPC api")

	// Start the server
	server := NewServer(ctx, conf, "nmap")

	// Update the server service state
	go getServiceControl(server)

	// start the auth server
	authServer := auth.NewAuthServer(server.userStore, server.jwtManager)

	// Serve the frontend
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := func(server *Server, authServer pb.AuthServiceServer) error {
			frontendSrv, frontendListener, err := server.getGRPCServerAndListener(conf.FrontendListenPort)
			if err != nil {

				log.Error().Stack().Err(err).Msg("")
			}
			pb.RegisterAuthServiceServer(frontendSrv, authServer)
			pb.RegisterScannerServiceServer(frontendSrv, server)
			if err = frontendSrv.Serve(*frontendListener); err != nil {
				wg.Done()
			}
			return err
		}(server, authServer)
		if err != nil {
			log.Fatal().Err(err)
		}
	}()

	// Serve the backend
	wg.Add(1)
	go func() {
		err := func(server *Server, authServer pb.AuthServiceServer) error {
			backendSrv, backendListener, err := server.getGRPCServerAndListener(conf.BackendListenPort)
			if err != nil {

				return err
			}
			pb.RegisterAuthServiceServer(backendSrv, authServer)
			pb.RegisterBackendServiceServer(backendSrv, server)

			if err = backendSrv.Serve(*backendListener); err != nil {
				wg.Done()
			}
			return err
		}(server, authServer)
		if err != nil {
			log.Fatal().Err(err)
		}
	}()

	wg.Wait()
	log.Info().Msg("backend and fronted stopped")

	return nil
}

func (server *Server) getGRPCServerAndListener(listenPort int) (*grpc.Server, *net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", listenPort))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "can't start server, can't listen on tcp/%d", listenPort)
	}

	tlsCredentials, err := auth.LoadTLSCredentials(
		server.config.Frontend.CAFile,
		server.config.Frontend.ServerCertFile,
		server.config.Frontend.ServerKeyFile,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "can't load TLS credentials")
	}

	interceptor := auth.NewAuthInterceptor(server.jwtManager, accessibleRoles())
	srv := grpc.NewServer(
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

			srv.GracefulStop()

			<-server.ctx.Done()
		}
	}()

	reflection.Register(srv)

	return srv, &listener, nil
}

// getScan get the scan main response
func (server *Server) getScanMainResponse(ctx context.Context, request *pb.GetScannerRequest) (*pb.ScannerMainResponse, error) {
	var smr pb.ScannerMainResponse
	dbSmr, err := server.db.Get(ctx, request.Key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(dbSmr), &smr)
	if err != nil {
		log.Error().Stack().Err(err).Msg("can't read scan result")
		return nil, err
	}

	return &smr, nil
}

// DeleteScan delete a scan from the database
func (server *Server) DeleteScan(ctx context.Context, request *pb.GetScannerRequest) (*pb.ServerResponse, error) {
	// get username from metadata
	username, role, err := server.getUsernameAndRoleFromRequest(ctx)
	if err != nil {
		return generateResponse(request.Key, nil, err)
	}

	smr, err := server.getScanMainResponse(ctx, request)
	if err != nil {
		return generateResponse(request.Key, nil, errors.New("can't get scan result or not allowed to access this resource"))
	}

	// test is the request username == the scan response username
	if role != "admin" && (smr.Request.GetUsername() == "" || smr.Request.GetUsername() != username) {
		return generateResponse(request.Key, nil, errors.New("not allowed to access this resource"))
	}

	_, err = server.db.Delete(ctx, request.Key)
	return generateResponse(request.Key, nil, err)
}

// StreamServiceControl control the service
func (server *Server) StreamServiceControl(in *pb.ServiceStateValues, stream pb.BackendService_StreamServiceControlServer) error {
	wait := 500 * time.Millisecond
	for {
		if err := stream.Send(&server.State); err != nil {
			return errors.Wrapf(err, "streamer service control send error: %v", in)
		}
		time.Sleep(wait)
	}
}

// StreamTasksStatus manage the tasks counter
func (server *Server) StreamTasksStatus(stream pb.BackendService_StreamTasksStatusServer) error {
	wait := 500 * time.Millisecond
	for {
		time.Sleep(wait)
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
		err = stream.Send(&pb.PrometheusStatus{TasksStatus: &pb.TasksStatus{
			Success: server.tasksStatus.success, Failed: server.tasksStatus.failed, Returned: server.tasksStatus.returned}},
		)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't send prometheus stats")
			// return err
		}
	}
}

// ServiceControl control the service
func (server *Server) ServiceControl(ctx context.Context, request *pb.ServiceStateValues) (*pb.ServiceStateValues, error) {

	var state pb.ServiceStateValues
	if request.GetState() == pb.ServiceStateValues_UNKNOWN {
		return &pb.ServiceStateValues{State: server.State.State}, nil
	}

	if server.State.State != request.GetState() {
		state = pb.ServiceStateValues{State: request.GetState()}
		stateJSON, err := json.Marshal(&state)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't Unmarshal state")
		}
		_, err = server.db.Set(ctx, stateStatusKey, string(stateJSON), 0)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't set service state")
		}
		server.State.State = state.State
	}
	return &pb.ServiceStateValues{State: state.State}, nil
}

// getServiceControl get the service state from database
// and update the server state
func getServiceControl(server *Server) {
	wait := 500 * time.Millisecond
	var state pb.ServiceStateValues
	for {
		time.Sleep(wait)
		stateJSON, err := server.db.Get(server.ctx, stateStatusKey)
		if err != nil || stateJSON == "" {
			log.Error().Stack().Err(err).Msg("can't get service state, setting to START")
			unknownState := pb.ServiceStateValues{State: pb.ServiceStateValues_START}
			unknownStateJSON, _ := json.Marshal(&unknownState)
			server.db.Set(server.ctx, stateStatusKey, string(unknownStateJSON), 0)
			server.State.State = unknownState.State
			continue
		}
		err = json.Unmarshal([]byte(stateJSON), &state)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't Unmarshal service state")
		}

		if server.State.State != state.State {
			log.Debug().Msgf("Server state changed to %v", state.State)
			server.State.State = state.State
		}
	}
}

// GetScan return the engine scan result
func (server *Server) GetScan(ctx context.Context, request *pb.GetScannerRequest) (*pb.ServerResponse, error) {
	// get username from metadata
	username, role, err := server.getUsernameAndRoleFromRequest(ctx)
	if err != nil {
		return generateResponse(request.Key, nil, err)
	}

	scannerMainResponse, err := server.getScanMainResponse(ctx, request)
	if err != nil {
		return generateResponse(request.Key, nil, errors.New("can't get scan result or not allowed to access this resource"))
	}

	// test is the request username == the scan response username
	if role != "admin" && (scannerMainResponse.Request.GetUsername() == "" || scannerMainResponse.Request.GetUsername() != username) {
		return generateResponse(request.Key, nil, errors.New("not allowed to access this resource"))
	}

	if scannerMainResponse.Request.BurnAfterReading && scannerMainResponse.Request.Role == "user" {
		_, err = server.db.Delete(ctx, request.Key)
		if err != nil {
			return generateResponse(request.Key, scannerMainResponse, errors.New("can't burn after reading"))
		}
	}

	return generateResponse(request.Key, scannerMainResponse, nil)
}

// GetAllScans return the engine scan result
func (server *Server) GetAllScans(_ *emptypb.Empty, stream pb.ScannerService_GetAllScansServer) error {
	allKeys, err := server.db.GetAll(server.ctx, "*")
	if err != nil {
		return err
	}

	for _, key := range allKeys {
		var scannerMainResponse pb.ScannerMainResponse
		dbSmr, err := server.db.Get(server.ctx, key)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't get scan result")
			continue
		}

		err = json.Unmarshal([]byte(dbSmr), &scannerMainResponse)
		scannerMainResponse.Key = key
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't list scan result")
			return err
		}

		response, err := generateResponse(key, &scannerMainResponse, err)
		if err != nil {
			return err
		}
		err = stream.Send(response)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't send getAllScans response")
			//			return err
		}
	}

	return nil
}

func (server *Server) getUsernameAndRoleFromRequest(ctx context.Context) (string, string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	values := md["authorization"]
	if len(values) == 0 {
		return "", "", errors.New("authorization token is not provided")
	}
	accessToken := values[0]
	claims, err := server.jwtManager.Verify(accessToken)
	if err != nil {
		return "", "", errors.Wrapf(err, "access token is invalid: %v", err)
	}
	return claims.Username, claims.Role, nil
}

// StartScan function prepare a nmap scan
func (server *Server) StartScan(ctx context.Context, params *pb.ParamsScannerRequest) (*pb.ServerResponse, error) {

	createAt := timestamppb.Now()

	params, err := server.parseRequest(ctx, params)
	if err != nil {
		return generateResponse("unknown key", nil, err)
	}

	// we start the scan
	newEngine := engine.New(ctx, server.db, server.config.NMAP, server.locker)
	sr := []*pb.ScannerResponse{{
		StartTime: timestamppb.Now(),
		Status:    pb.ScannerResponse_UNKNOWN,
	}}
	result, err := newEngine.Start(params, false)
	if err != nil {
		return generateResponse(params.Key, nil, err)
	}
	// end of scan
	sr[0].EndTime = timestamppb.Now()
	sr[0].HostResult = result
	sr[0].Status = pb.ScannerResponse_OK

	// forge main response
	smr := pb.ScannerMainResponse{
		Key:      params.Key,
		Request:  params,
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

// splitInSubnets split a subnet in several smaller
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

	params, err := server.parseRequest(ctx, params)
	if err != nil {
		return generateResponse(params.Key, nil, err)
	}

	targets := params.Targets
	createAt := timestamppb.Now()

	var scannerResponses []*pb.ScannerResponse
	var scannerMainResponse pb.ScannerMainResponse

	// if we want to split a task by host
	// we split host list by "," and then
	// spawn a consumer per host
	//
	// also, we split a network cidr in smaller one
	var hosts []string
	if params.NetworkChuncked || params.ProcessPerTarget {
		hosts = strings.Split(params.Targets, ",")
		if params.NetworkChuncked {
			var splitNetworks []string
			for _, network := range hosts {
				splittedNetwork, err := splitInSubnets(network, 3)
				if err != nil {
					splitNetworks = append(splitNetworks, network)
				}
				splitNetworks = append(splitNetworks, splittedNetwork...)
			}
			hosts = splitNetworks
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
		scannerResponse := &pb.ScannerResponse{
			Key:    subKey,
			Status: pb.ScannerResponse_QUEUED,
		}

		// then append it to the main task
		scannerResponses = append(scannerResponses, scannerResponse)

		// we override the Hosts param in case we split it
		// for the scan request
		params.Targets = host
		scannerMainResponse = pb.ScannerMainResponse{
			Key:      params.Key,
			Request:  params,
			CreateAt: createAt,
			Response: scannerResponses,
		}

		log.Info().Msgf("receive async task order: %v", params)

		smrJSON, err := json.Marshal(&scannerMainResponse)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't create main JSON response")
			return generateResponse(params.Key, &scannerMainResponse, err)
		}

		// write main response to database
		_, err = server.db.Set(ctx, params.Key, string(smrJSON), 0)
		if err != nil {
			log.Error().Stack().Err(err).Msg("can't write main response to database")
			return generateResponse(params.Key, &scannerMainResponse, err)
		}

		// create scan task and publish it to the RMQ broker
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			scannerResponses[i].Status = pb.ScannerResponse_ERROR
			return generateResponse(params.Key, &scannerMainResponse, err)
		}

		err = server.broker.Incoming.PublishBytes(paramsJSON)
		if err != nil {
			scannerResponses[i].Status = pb.ScannerResponse_ERROR
			return generateResponse(params.Key, &scannerMainResponse, err)
		}

	}

	params.Targets = targets
	scannerMainResponse = pb.ScannerMainResponse{
		Key:      params.Key,
		Request:  params,
		CreateAt: createAt,
		Response: scannerResponses,
	}
	return generateResponse(params.Key, &scannerMainResponse, nil)
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
func (server *Server) parseRequest(ctx context.Context, request *pb.ParamsScannerRequest) (*pb.ParamsScannerRequest, error) {

	// get username from metadata
	username, role, err := server.getUsernameAndRoleFromRequest(ctx)
	if err != nil {
		return nil, err
	}
	request.Username = username
	request.Role = role

	// if the Key is forced and we are admin we use it, else we generate an uuid
	if !(request.Key != "" && request.Role == "admin") {
		guid := uuid.New()
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

	return request, nil
}

// brokerStatsToProm read broker stats each 2s and write to prometheus
func brokerStatsToProm(myBroker *broker.Broker, name string) {
	var prevReady, prevReject int64
	for range time.Tick(1000 * time.Millisecond) {
		stats, err := myBroker.GetStats()
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
	err := createUser(userStore, "admin1", "secret1", "admin")
	if err != nil {
		return err
	}
	err = createUser(userStore, "worker1", "secret1", "worker")
	if err != nil {
		return err
	}
	err = createUser(userStore, "user1", "secret1", "user")
	if err != nil {
		return err
	}
	return createUser(userStore, "user2", "secret2", "user")
}

// accessibleRoles contains the map of grant access by service
func accessibleRoles() map[string][]string {
	const backendServicePath = "/proto.BackendService/"
	const frontendServicePath = "/proto.ScannerService/"

	return map[string][]string{
		backendServicePath + "StreamServiceControl": {"worker", "admin"},
		backendServicePath + "StreamTasksStatus":    {"worker", "admin"},

		frontendServicePath + "StartAsyncScan": {"user", "admin"},
		frontendServicePath + "StartScan":      {"user", "admin"},

		frontendServicePath + "GetScan":    {"user", "admin"},
		frontendServicePath + "DeleteScan": {"user", "admin"},

		frontendServicePath + "GetAllScans":    {"admin"},
		frontendServicePath + "ServiceControl": {"admin"},
	}
}
