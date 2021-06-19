package scanner

import (
	"context"
	nmap "github.com/Ullaakut/nmap/v2"
	"log"
	"strings"
	"time"
)

func StartNmapScan(s *Scanner) *nmap.Run {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	hosts := strings.Split(s.Hosts, ",")
	ports := strings.Split(s.Ports, ",")

	var options = []nmap.Option{
		nmap.WithContext(ctx),
		nmap.WithTargets(hosts...),
		nmap.WithPorts(ports...),
	}

	if strings.ToLower(s.Protocol) == "udp" {
		options = append(options, nmap.WithUDPScan())
	} else {
		options = append(options, nmap.WithAggressiveScan())
	}

	scanner, err := nmap.NewScanner(options...)

	if err != nil {
		log.Fatalf("unable to create nmap scanner: %v", err)
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		log.Fatalf("unable to run nmap scan: %v", err)
	}

	if warnings != nil {
		log.Printf("Warnings: \n %v", warnings)
	}

	return result
}
