package scanner

import (
	"context"
	nmap "github.com/Ullaakut/nmap/v2"
	"log"
	"strings"
	"time"
)

func StartNmapScan(s *Scanner) (*nmap.Run, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout)*time.Second)
	defer cancel()

	hosts := strings.Split(s.Hosts, ",")
	ports := strings.Split(s.Ports, ",")

	var options = []nmap.Option{
		nmap.WithContext(ctx),
		nmap.WithTargets(hosts...),
		nmap.WithPorts(ports...),
		nmap.WithUDPScan(),
		nmap.WithSYNScan(),
	}

	if s.GetServiceOsDetection() {
		options = append(options, nmap.WithOSDetection())
	}

	if s.GetServiceVersionDetection() {
		options = append(options, nmap.WithServiceInfo())
		options = append(options, nmap.WithDefaultScript())
	}

	if s.GetWithAggressiveScan() {
		options = append(options, nmap.WithAggressiveScan())
	} else if s.GetWithNullScan() {
		options = append(options, nmap.WithTCPNullScan())
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
