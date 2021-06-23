package engines

import (
	"context"
	"fmt"
	nmap "github.com/Ullaakut/nmap/v2"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"log"
	"strings"
	"time"
)

// StartNmapScan start a nmap scan
func StartNmapScan(s *proto.ParamsScannerRequest) (*nmap.Run, error) {
	timeout := time.Duration(s.Timeout) * time.Second
	retention := time.Duration(s.RetentionTime) * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	hosts := strings.Split(s.Hosts, ",")
	ports := s.Ports

	var options = []nmap.Option{nmap.WithContext(ctx),
		nmap.WithTargets(hosts...),
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

	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		return nil, err
	}

	log.Printf("Starting scan of host: %s, port: %s, timeout: %v, retention: %v",
		hosts,
		ports,
		timeout,
		retention,
	)

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, err
	}
	for _, tb := range result.TaskBegin {
		log.Printf("%v", tb.Time)
	}
	for _, te := range result.TaskEnd {
		log.Printf("%v", te.Time)
	}

	if warnings != nil {
		log.Printf("warnings: %v", warnings)
	}

	log.Printf("Nmap done: %d/%d hosts up scanned in %3f seconds\n",
		result.Stats.Hosts.Up,
		result.Stats.Hosts.Total,
		result.Stats.Finished.Elapsed,
	)

	return result, err
}

func ParseScanResult(result *nmap.Run) ([]*proto.HostResult, error) {
	portList := []*proto.Port{}
	scanResult := []*proto.HostResult{}
	totalPorts := 0
	if len(result.Hosts) == 0 || result == nil {
		log.Println("Scan timeout.")
		return nil, nil
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
			Address:   address,
			OsVersion: &osversion,
			State:     &host.Status.Reason,
		}
		for _, p := range host.Ports {
			version := &proto.PortVersion{
				ExtraInfos:  &p.Service.ExtraInfo,
				LowVersion:  &p.Service.LowVersion,
				HighVersion: &p.Service.HighVersion,
				Product:     &p.Service.Product,
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
	return scanResult, nil
}
