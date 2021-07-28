package engine

import (
	"context"
	"encoding/json"
	"fmt"
	nmap "github.com/Ullaakut/nmap/v2"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/helpers"
	"github.com/cyrinux/grpcnmapscanner/locker"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var (
	confLogger              = config.GetConfig().Logger
	log                     = logger.New(confLogger.Debug, confLogger.Pretty)
	ErrTimeoutOrUnreachable = errors.New("scan timeout, no result or network unreachable?")
)

// Engine define a scanner engine
type Engine struct {
	ctx    context.Context
	db     database.Database
	locker locker.MyLocker
	State  pb.ScannerResponse_Status
	config config.NMAPConfig
}

// ParamsParsed are the engine params parsed
type ParamsParsed struct {
	hosts    []string
	ports    []string
	options  []nmap.Option
	timeout  time.Duration
	nmapArgs []string
	params   pb.ParamsScannerRequest
}

// New create a new nmap engine
func New(ctx context.Context, db database.Database, conf config.NMAPConfig, locker locker.MyLocker) *Engine {
	return &Engine{ctx: ctx, db: db, config: conf, locker: locker}
}

// Start the scan
func (e *Engine) Start(params *pb.ParamsScannerRequest, async bool) (*pb.ParamsScannerRequest, *nmap.Run, error) {
	if async {
		return e.startAsyncScan(params)
	} else {
		return e.startScan(params)
	}
}

func (e *Engine) parseNMAPParams(s *pb.ParamsScannerRequest) (*ParamsParsed, error) {
	hostsList := strings.Split(s.GetTargets(), ",")
	ports := s.Ports

	var options = []nmap.Option{
		nmap.WithVerbosity(3),
		nmap.WithTargets(hostsList...),
		nmap.WithTimingTemplate(nmap.Timing(s.GetScanSpeed())),
	}

	if s.GetUseTor() {
		options = append(options, nmap.WithProxies(e.config.TorServer))
		options = append(options, nmap.WithSkipHostDiscovery())
	}

	portsList := strings.Split(ports, ",")
	if len(ports) == 0 && !s.GetPingOnly() {
		if s.GetFastMode() {
			options = append(options, nmap.WithFastMode())
		} else {
			ports = "1-65535"
			portsList = strings.Split(ports, ",")
			options = append(options, nmap.WithPorts(portsList...))
		}
	} else if len(ports) == 0 && s.GetPingOnly() {
		options = append(options, nmap.WithPingScan())
	} else {
		options = append(options, nmap.WithPorts(portsList...))
	}

	isUDPScan := strings.Contains(ports, "U:")
	isSCTPScan := strings.Contains(ports, "S:")
	isTCPScan := strings.Contains(ports, "T:")

	if isUDPScan {
		options = append(options, nmap.WithUDPScan())
	}

	if isTCPScan {
		options = append(options, nmap.WithSYNScan())
	}

	if isSCTPScan {
		options = append(options, nmap.WithSCTPInitScan())
	}

	if s.GetOsDetection() {
		options = append(options, nmap.WithOSDetection())
	}

	if s.GetServiceVersionDetection() {
		options = append(options, nmap.WithServiceInfo())
	}

	if s.GetServiceDefaultScripts() {
		options = append(options, nmap.WithDefaultScript())
	}

	if s.GetServiceScripts() != "" {
		options = append(options, nmap.WithServiceInfo())
		for _, script := range strings.Split(s.ServiceScripts, ",") {
			options = append(options, nmap.WithScripts(script))

		}
	}

	if s.GetOpenOnly() {
		options = append(options, nmap.WithOpenOnly())
	}

	if s.GetWithAggressiveScan() {
		options = append(options, nmap.WithTimingTemplate(nmap.TimingAggressive))
		options = append(options, nmap.WithAggressiveScan())
	} else if s.GetPingOnly() {
		options = append(options, nmap.WithPingScan())
	} else if s.GetWithNullScan() {
		options = append(options, nmap.WithTCPNullScan())
	} else if s.GetWithSynScan() {
		options = append(options, nmap.WithSYNScan())
	}

	return &ParamsParsed{hosts: hostsList, ports: portsList, options: options}, nil
}

func (e *Engine) startScan(params *pb.ParamsScannerRequest) (*pb.ParamsScannerRequest, *nmap.Run, error) {
	sr := []*pb.ScannerResponse{
		{Status: pb.ScannerResponse_RUNNING},
	}
	smr := pb.ScannerMainResponse{Key: params.Key, Response: sr}

	smrJSON, err := json.Marshal(&smr)
	if err != nil {
		return params, nil, err
	}
	_, err = e.db.Set(e.ctx, params.Key, string(smrJSON), params.RetentionDuration.AsDuration())
	if err != nil {
		return params, nil, err
	}

	params, result, err := e.run(params)
	if err != nil {
		return params, nil, err
	}
	return params, result, nil

}

func (e *Engine) startAsyncScan(params *pb.ParamsScannerRequest) (*pb.ParamsScannerRequest, *nmap.Run, error) {
	var scannerResponses []*pb.ScannerResponse
	var smr *pb.ScannerMainResponse
	var err error

	smrJSON, err := e.db.Get(e.ctx, params.Key)
	if err != nil {
		return params, nil, err
	}
	err = json.Unmarshal([]byte(smrJSON), &smr)
	if err != nil {
		return params, nil, err
	}
	scannerResponses = smr.Response

	if smr.Request != nil && smr.Request.Targets != "" {
		hosts := strings.Split(smr.Request.Targets, ",")
		hosts = helpers.MakeUnique(hosts)
		smr.Request.Targets = strings.Join(hosts, ",")
	}

	key := params.Key
	if (params.ProcessPerTarget || params.NetworkChuncked) && len(params.Targets) > 1 {
		key = fmt.Sprintf("%s-%s", params.Key, params.Targets)
	}

	for i, sr := range scannerResponses {
		if sr.Key == key {
			scannerResponses[i].Status = pb.ScannerResponse_RUNNING
			break
		}
	}

	smr.Response = scannerResponses

	smrNewJSON, err := json.Marshal(smr)
	if err != nil {
		return params, nil, err
	}
	_, err = e.db.Set(e.ctx, params.Key, string(smrNewJSON), params.RetentionDuration.AsDuration())
	if err != nil {
		return params, nil, err
	}

	params, result, err := e.run(params)
	if err != nil {
		return params, nil, err
	}
	return params, result, nil
}

func (e *Engine) run(params *pb.ParamsScannerRequest) (*pb.ParamsScannerRequest, *nmap.Run, error) {
	// parse all input options
	paramsParsed, err := e.parseNMAPParams(params)
	if err != nil {
		return params, nil, err
	}

	// int32s in seconds
	paramsParsed.timeout = time.Duration(params.Timeout) * time.Second

	// define scan context
	ctx, cancel := context.WithTimeout(e.ctx, paramsParsed.timeout)
	defer cancel()

	// add context to nmap options
	paramsParsed.options = append(paramsParsed.options, nmap.WithContext(ctx))

	// create a nmap scanner
	scanner, err := nmap.NewScanner(paramsParsed.options...)
	if err != nil {
		return params, nil, err
	}

	// nmap args
	paramsParsed.nmapArgs = scanner.Args()

	log.Info().Msgf("starting scan %s of host: %s, port: %s, args: %v",
		params.Key,
		paramsParsed.hosts,
		paramsParsed.ports,
		fmt.Sprintf("%s", paramsParsed.nmapArgs),
	)

	// Function to listen and print the progress
	progress := make(chan float32, 1)
	go progressNmap(progress, params, paramsParsed)

	result, warnings, err := scanner.RunWithProgress(progress)
	if err != nil {
		return params, nil, err
	}

	if warnings != nil {
		log.Warn().Msgf("nmap: %v", warnings)
	}

	log.Info().Msgf("nmap done: %d/%d hosts up scanned in %3f seconds",
		result.Stats.Hosts.Up,
		result.Stats.Hosts.Total,
		result.Stats.Finished.Elapsed,
	)

	return params, result, nil
}

// ParseScanResult parse the nmap result and create the expected output
func ParseScanResult(result *nmap.Run) ([]*pb.HostResult, error) {
	var portList []*pb.Port
	var resultJSON []*pb.HostResult
	var totalPorts int
	var err error
	if result == nil || len(result.Hosts) == 0 || result.NmapErrors == nil {
		log.Error().Stack().Err(ErrTimeoutOrUnreachable).Msg("")
		return nil, err
	}
	for _, host := range result.Hosts {
		var osVersion string
		if len(host.Addresses) == 0 {
			continue
		}
		if len(host.OS.Matches) > 0 {
			fp := host.OS.Matches[0]
			osVersion = fmt.Sprintf("name: %v, accuracy: %v%%", fp.Name, fp.Accuracy)
		}

		for _, ip := range host.Addresses {
			address := ip.Addr
			hostResult := &pb.Host{
				Address:   address,
				OsVersion: osVersion,
				State:     host.Status.Reason,
			}

			for _, port := range host.Ports {
				scripts := make([]*pb.Script, 0)
				for _, v := range port.Scripts {
					// split some scripts results
					scriptOutputArray := strings.Split(v.Output, "\n")

					output := make([]string, 0)
					for _, line := range scriptOutputArray {
						line = strings.ReplaceAll(line, "\t", "  ")
						line = strings.TrimSpace(line)
						output = append(output, line)
					}

					scripts = append(
						scripts,
						&pb.Script{
							Id:     v.ID,
							Output: output,
						},
					)
				}

				version := &pb.PortVersion{
					ExtraInfos:  port.Service.ExtraInfo,
					LowVersion:  port.Service.LowVersion,
					HighVersion: port.Service.HighVersion,
					Product:     port.Service.Product,
					Scripts:     scripts,
					Confidence:  int32(port.Service.Confidence),
				}

				var newPort = &pb.Port{
					PortId:      fmt.Sprintf("%v", port.ID),
					ServiceName: port.Service.Name,
					Protocol:    port.Protocol,
					State:       port.State.Reason,
					Version:     version,
				}
				portList = append(portList, newPort)
			}
			totalPorts += len(portList)

			scan := &pb.HostResult{
				Host:  hostResult,
				Ports: portList,
			}

			resultJSON = append(resultJSON, scan)
		}
	}
	return resultJSON, nil
}

func progressNmap(progress chan float32, params *pb.ParamsScannerRequest, parsedParams *ParamsParsed) {
	var previous float32
	for current := range progress {
		if current > previous {
			log.Debug().Msgf("running scan %s : %v%% - host: %s, port: %s",
				params.Key,
				current,
				parsedParams.hosts,
				parsedParams.ports,
			)
			previous = current
		} else {
			continue
		}
		time.Sleep(10000 * time.Millisecond)
	}
}
