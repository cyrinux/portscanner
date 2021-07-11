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
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/pkg/errors"
)

var (
	confLogger = config.GetConfig().Logger
	log        = logger.New(confLogger.Debug, confLogger.Pretty)
)

// Engine define a scanner engine
type Engine struct {
	ctx   context.Context
	db    database.Database
	State proto.ScannerResponse_Status
}

// NewEngine create a new nmap engine
func NewEngine(ctx context.Context, db database.Database) *Engine {
	return &Engine{ctx: ctx, db: db}
}

func parseParamsScannerRequestNmapOptions(
	ctx context.Context,
	s *proto.ParamsScannerRequest) ([]string, []string, []nmap.Option, error) {

	hostsList := strings.Split(s.Hosts, ",")
	ports := s.Ports

	var options = []nmap.Option{
		nmap.WithVerbosity(3),
		nmap.WithTargets(hostsList...),
		nmap.WithTimingTemplate(nmap.Timing(s.ScanSpeed)),
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
func (engine *Engine) StartNmapScan(s *proto.ParamsScannerRequest) (string, *nmap.Run, error) {
	scannerResponse := &proto.ScannerResponse{
		Status: proto.ScannerResponse_RUNNING,
	}

	scanResultJSON, err := json.Marshal(scannerResponse)
	if err != nil {
		return s.Key, nil, err
	}
	_, err = engine.db.Set(engine.ctx, s.Key, string(scanResultJSON), time.Duration(s.GetRetentionTime())*time.Second)
	if err != nil {
		return s.Key, nil, err
	}

	// int32s in seconds
	timeout := time.Duration(s.Timeout) * time.Second
	retention := time.Duration(s.RetentionTime) * time.Second

	// define scan context
	ctx, cancel := context.WithTimeout(engine.ctx, timeout)
	defer cancel()

	// parse all input options
	hosts, ports, options, err := parseParamsScannerRequestNmapOptions(ctx, s)
	if err != nil {
		return s.Key, nil, err
	}

	// add context to nmap options
	options = append(options, nmap.WithContext(ctx))

	// create a nmap scanner
	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		return s.Key, nil, err
	}

	log.Info().Msgf("starting scan %s of host: %s, port: %s, timeout: %v, retention: %v",
		s.Key,
		hosts,
		ports,
		timeout,
		retention,
	)

	// Function to listen and print the progress
	progress := make(chan float32, 1)
	go func() {
		var previous float32
		for p := range progress {
			if p < previous {
				continue
			} else {
				previous = p
				log.Debug().Msgf("scan %s : %v%% - host: %s, port: %s, timeout: %v, retention: %v",
					s.Key,
					p,
					hosts,
					ports,
					timeout,
					retention,
				)
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	result, warnings, err := scanner.RunWithProgress(progress)
	if err != nil {
		return s.Key, nil, err
	}

	if warnings != nil {
		log.Warn().Msgf("nmap warnings: %v", warnings)
	}

	log.Info().Msgf("nmap done: %d/%d hosts up scanned in %3f seconds",
		result.Stats.Hosts.Up,
		result.Stats.Hosts.Total,
		result.Stats.Finished.Elapsed,
	)

	return s.Key, result, nil
}

// ParseScanResult parse the nmap result and create the expected output
func ParseScanResult(result *nmap.Run) ([]*proto.HostResult, error) {
	portList := []*proto.Port{}
	scanResult := []*proto.HostResult{}
	totalPorts := 0
	if result == nil || len(result.Hosts) == 0 {
		err := errors.New("scan timeout or not result")
		log.Error().Err(err).Msg("")
		return nil, err
	}
	// fmt.Printf("%+v", result)
	for _, host := range result.Hosts {

		// fmt.Printf("Post: %+v\n", result.PostScripts)
		// fmt.Printf("Pre: %+v\n", result.PreScripts)

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
			hostResult := &proto.Host{
				Address:   address,
				OsVersion: &osversion,
				State:     host.Status.Reason,
			}

			for _, port := range host.Ports {
				scripts := make([]*proto.Script, 0)
				for _, v := range port.Scripts {
					// split some scripts result (vulners)
					scriptOutputArray := strings.Split(v.Output, "\n    \t")

					output := make([]string, 0)
					for _, line := range scriptOutputArray {
						line = strings.ReplaceAll(line, "\t", "  ")
						line = strings.TrimSpace(line)
						output = append(output, line)
					}

					scripts = append(
						scripts,
						&proto.Script{
							Id:     v.ID,
							Output: output,
						},
					)
				}

				version := &proto.PortVersion{
					ExtraInfos:  port.Service.ExtraInfo,
					LowVersion:  port.Service.LowVersion,
					HighVersion: port.Service.HighVersion,
					Product:     port.Service.Product,
					Scripts:     scripts,
				}

				newPort := &proto.Port{
					PortId:      fmt.Sprintf("%v", port.ID),
					ServiceName: &port.Service.Name,
					Protocol:    port.Protocol,
					State:       port.State.Reason,
					Version:     version,
				}
				portList = append(portList, newPort)
			}
			totalPorts += len(portList)

			scan := &proto.HostResult{
				Host:  hostResult,
				Ports: portList,
			}

			scanResult = append(scanResult, scan)
		}
	}
	return scanResult, nil
}
