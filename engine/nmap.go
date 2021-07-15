package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	nmap "github.com/Ullaakut/nmap/v2"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/logger"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/pkg/errors"
)

var (
	confLogger = config.GetConfig().Logger
	log        = logger.New(confLogger.Debug, confLogger.Pretty)
)

// Engine define a scanner engine
type Engine struct {
	ctx    context.Context
	db     database.Database
	State  pb.ScannerResponse_Status
	config config.NMAPConfig
}

// NewEngine create a new nmap engine
func NewEngine(ctx context.Context, db database.Database, conf config.NMAPConfig) *Engine {
	return &Engine{ctx: ctx, db: db, config: conf}
}

func (e *Engine) parseParams(s *pb.ParamsScannerRequest) ([]string, []string, []nmap.Option, error) {
	hostsList := strings.Split(s.Hosts, ",")
	ports := s.Ports

	var options = []nmap.Option{
		nmap.WithVerbosity(3),
		nmap.WithTargets(hostsList...),
		nmap.WithTimingTemplate(nmap.Timing(s.ScanSpeed)),
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
		options = append(options, nmap.WithScriptUpdateDB())
	}

	if s.ServiceScripts != "" {
		options = append(options, nmap.WithScriptUpdateDB())
		options = append(options, nmap.WithServiceInfo())
		for _, script := range strings.Split(s.ServiceScripts, ",") {
			options = append(options, nmap.WithScripts(script))

		}
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

	return hostsList, portsList, options, nil
}

// StartNmapScan start a nmap scan
func (e *Engine) StartNmapScan(params *pb.ParamsScannerRequest) (*pb.ParamsScannerRequest, *nmap.Run, error) {
	scannerResponse := []*pb.ScannerResponse{
		{Status: pb.ScannerResponse_RUNNING},
	}
	scannerMainResponse := pb.ScannerMainResponse{Request: params, Key: params.Key, Response: scannerResponse}

	resultJSONJSON, err := json.Marshal(&scannerMainResponse)
	if err != nil {
		return params, nil, err
	}

	_, err = e.db.Set(e.ctx, params.Key, string(resultJSONJSON), params.RetentionDuration.AsDuration())
	if err != nil {
		return params, nil, err
	}

	// int32s in seconds
	timeout := time.Duration(params.Timeout) * time.Second

	// define scan context
	ctx, cancel := context.WithTimeout(e.ctx, timeout)
	defer cancel()

	// parse all input options
	hosts, ports, options, err := e.parseParams(params)
	if err != nil {
		return params, nil, err
	}

	// add context to nmap options
	options = append(options, nmap.WithContext(ctx))

	// create a nmap scanner
	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		return params, nil, err
	}

	// nmap args
	nmapArgs := fmt.Sprintf("%s", scanner.Args())
	params.NmapParams = nmapArgs

	log.Info().Msgf("starting scan %s of host: %s, port: %s, timeout: %v, retention: %vs, args: %v",
		params.Key,
		hosts,
		ports,
		timeout,
		params.RetentionDuration.Seconds,
		nmapArgs,
	)

	// Function to listen and print the progress
	progress := make(chan float32, 1)
	go func() {
		var previous float32
		for p := range progress {
			if p > previous {
				log.Debug().Msgf("scan %s : %v%% - host: %s, port: %s, timeout: %v, retention: %vs",
					params.Key,
					p,
					hosts,
					ports,
					timeout,
					params.RetentionDuration.Seconds,
				)
				previous = p
			} else {
				continue
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	result, warnings, err := scanner.RunWithProgress(progress)
	if err != nil {
		return params, nil, err
	}

	if warnings != nil {
		log.Warn().Msgf("nmap warnings: %v", warnings)
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
	portList := []*pb.Port{}
	resultJSON := []*pb.HostResult{}
	totalPorts := 0
	if result == nil || len(result.Hosts) == 0 {
		err := errors.New("scan timeout or not result")
		log.Error().Err(err).Msg("")
		return nil, err
	}
	for _, host := range result.Hosts {
		var osversion string
		if len(host.Addresses) == 0 {
			continue
		}
		if len(host.OS.Matches) > 0 {
			fp := host.OS.Matches[0]
			osversion = fmt.Sprintf("name: %v, accuracy: %v%%", fp.Name, fp.Accuracy)
		}

		for _, ip := range host.Addresses {
			address := ip.Addr
			hostResult := &pb.Host{
				Address:   address,
				OsVersion: osversion,
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

				newPort := &pb.Port{
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
