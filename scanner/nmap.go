package scanner

import (
	"context"
	nmap "github.com/Ullaakut/nmap/v2"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"log"
	"strings"
	"time"
)

// StartNmapScan start a nmap scan
func StartNmapScan(s *proto.SetScannerRequest) (*nmap.Run, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout)*time.Second)
	defer cancel()

	hosts := strings.Split(s.Hosts, ",")
	ports := strings.Split(s.Ports, ",")
	isUDPScan := strings.Contains(s.Ports, "U:")
	isSCTPScan := strings.Contains(s.Ports, "S:")
	isTCPScan := strings.Contains(s.Ports, "T:")
	isTCPScan = !strings.Contains(s.Ports, ":")

	var options = []nmap.Option{
		nmap.WithContext(ctx),
		nmap.WithTargets(hosts...),
		nmap.WithPorts(ports...),
	}

	if isUDPScan {
		options = append(options, nmap.WithUDPScan())
	}

	if isTCPScan {
		options = append(options, nmap.WithTCPXmasScan())
	}

	if isSCTPScan {
		options = append(options, nmap.WithSCTPInitScan())
	}

	if s.GetServiceOsDetection() {
		options = append(options, nmap.WithOSDetection())
	}

	if s.GetServiceVersionDetection() {
		options = append(options, nmap.WithServiceInfo())
	}

	if s.GetServiceDefaultScripts() {
		options = append(options, nmap.WithDefaultScript())
	}

	if s.GetWithAggressiveScan() {
		options = append(options, nmap.WithAggressiveScan())
	} else if s.GetWithNullScan() {
		options = append(options, nmap.WithTCPNullScan())
	} else if s.GetWithSynScan() {
		options = append(options, nmap.WithSYNScan())
	}
	scanner, err := nmap.NewScanner(options...)
	if err != nil {
		log.Printf("unable to create nmap scanner: %v", err)
	}
	result, warnings, err := scanner.Run()
	if err != nil {
		log.Printf("unable to run nmap scan: %v", err)
	}
	if warnings != nil {
		log.Printf("warnings: \n %v", warnings)
	}
	return result, err
}
