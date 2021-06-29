package engine

import (
	"context"
	"encoding/json"
	"fmt"
	nmap "github.com/Ullaakut/nmap/v2"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/proto"
	// "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type Engine struct {
	config config.Config
}

// NewEngine create a new nmap engine and init the database connection
func NewEngine(config config.Config) *Engine {
	return &Engine{config}
}

func parseParamsScannerRequestNmapOptions(ctx context.Context, s *proto.ParamsScannerRequest) ([]string, []string, []nmap.Option, error) {

	hostsList := strings.Split(s.Hosts, ",")
	ports := s.Ports

	var options = []nmap.Option{
		nmap.WithVerbosity(3),
		nmap.WithTargets(hostsList...),
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
func (e *Engine) StartNmapScan(ctx context.Context, s *proto.ParamsScannerRequest) (string, *nmap.Run, error) {

	scannerResponse := &proto.ScannerResponse{
		Status: proto.ScannerResponse_RUNNING,
	}

	scanResultJSON, err := json.Marshal(scannerResponse)
	if err != nil {
		return s.Key, nil, err
	}
	_, err = e.config.DB.Set(ctx, s.Key, string(scanResultJSON), time.Duration(s.GetRetentionTime())*time.Second)
	if err != nil {
		return s.Key, nil, err
	}

	// int32s in seconds
	timeout := time.Duration(s.Timeout) * time.Second
	retention := time.Duration(s.RetentionTime) * time.Second

	// define scan context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

	log.Info().Msgf("Starting scan of host: %s, port: %s, timeout: %v, retention: %v",
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
			previous = p
			if p < previous {
				continue
			} else {
				log.Info().Msgf("Running scan of host: %s, port: %s, timeout: %v, retention: %v: %v %%",
					hosts,
					ports,
					timeout,
					retention,
					p,
				)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	result, warnings, err := scanner.RunWithProgress(progress)
	if err != nil {
		return s.Key, nil, err
	}

	if warnings != nil {
		log.Warn().Msgf("Nmap warnings: %v", warnings)
	}

	log.Info().Msgf("Nmap done: %d/%d hosts up scanned in %3f seconds",
		result.Stats.Hosts.Up,
		result.Stats.Hosts.Total,
		result.Stats.Finished.Elapsed,
	)

	return s.Key, result, err
}

// ParseScanResult parse the nmap result and create the expected output
func ParseScanResult(key string, result *nmap.Run) (string, []*proto.HostResult, error) {
	portList := []*proto.Port{}
	scanResult := []*proto.HostResult{}
	totalPorts := 0
	if len(result.Hosts) == 0 || result == nil {
		log.Error().Msg("Scan timeout or no result")
		return key, nil, nil
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
		for _, adr := range host.Addresses {
			address := adr.Addr
			hostResult := &proto.Host{
				Address:   address,
				OsVersion: osversion,
				State:     host.Status.Reason,
			}
			for _, p := range host.Ports {
				version := &proto.PortVersion{
					ExtraInfos:  p.Service.ExtraInfo,
					LowVersion:  p.Service.LowVersion,
					HighVersion: p.Service.HighVersion,
					Product:     p.Service.Product,
				}
				newPort := &proto.Port{
					PortId:      fmt.Sprintf("%v", p.ID),
					ServiceName: p.Service.Name,
					Protocol:    p.Protocol,
					State:       p.State.Reason,
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
	return key, scanResult, nil
}
