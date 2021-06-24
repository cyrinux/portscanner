package engine

import (
	"context"
	"encoding/json"
	"fmt"
	nmap "github.com/Ullaakut/nmap/v2"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"log"
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
		portsList = strings.Split(ports, ",")
		options = append(options, nmap.WithPorts(portsList...))
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
	}

	return hostsList, portsList, options, nil
}

// StartNmapScan start a nmap scan
func (e *Engine) StartNmapScan(ctx context.Context, s *proto.ParamsScannerRequest) (string, *nmap.Run, error) {

	scannerResponse := &proto.ScannerResponse{
		Status: proto.ScannerResponse_RUNNING,
	}

	scanResultJSON, err := json.Marshal(scannerResponse)
	_, err = e.config.DB.Set(s.Key, string(scanResultJSON), time.Duration(s.GetRetentionTime())*time.Second)

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

	// Filter out hosts that don't have any open ports
	options = append(options, nmap.WithFilterHost(func(h nmap.Host) bool {
		// Filter out hosts with no open ports.
		for idx := range h.Ports {
			if h.Ports[idx].Status() == "open" {
				return true
			}
		}

		return false
	}))

	// create a nmap scanner
	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		return s.Key, nil, err
	}

	log.Printf("Starting scan of host: %s, port: %s, timeout: %v, retention: %v",
		hosts,
		ports,
		timeout,
		retention,
	)

	result, warnings, err := scanner.Run()
	if err != nil {
		return s.Key, nil, err
	}

	for _, tb := range result.TaskBegin {
		log.Printf("Task begin: %v", tb.Time)
	}
	for _, te := range result.TaskEnd {
		log.Printf("Task end: %v", te.Time)
	}

	if warnings != nil {
		log.Printf("warnings: %v", warnings)
	}

	log.Printf("Nmap done: %d/%d hosts up scanned in %3f seconds\n",
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
		log.Println("Scan timeout or no result")
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
		address := host.Addresses[0].Addr
		hostResult := &proto.Host{
			Address: address,
			// Fqdn:      fqdn,
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
	return key, scanResult, nil
}
