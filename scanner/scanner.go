package scanner

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
	"time"
)

type Server struct {
}

func (s *Server) Scan(ctx context.Context, in *Scanner) (*AllScanResults, error) {
	createdDate := time.Now()
	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	log.Printf("Starting scan of host: %s, port: %s, timeout: %v", in.Hosts, in.Ports, in.Timeout)

	portList := []*Port{}
	allScanResults := []*ScanResult{}

	result, err := StartNmapScan(in)
	if err != nil {
		return &AllScanResults{Scanresult: nil}, err
	}

	for _, host := range result.Hosts {
		var osversion string = "unknown"
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}
		if len(host.OS.Matches) > 0 {
			fp := host.OS.Matches[0]
			osversion = fmt.Sprintf("name: %v, accuracy: %v%%", fp.Name, fp.Accuracy)
		}
		hostResult := &Host{Address: host.Addresses[0].Addr, OsVersion: &osversion}

		for _, p := range host.Ports {
			port := &Port{
				Host:        hostResult,
				Number:      fmt.Sprintf("%v", p.ID),
				ServiceName: string(p.Service.Name),
				Protocol:    string(p.Protocol),
				State:       p.State.State,
			}
			portList = append(portList, port)
		}

		var scanResult = &ScanResult{Port: portList, Host: hostResult}
		allScanResults = append(allScanResults, scanResult)
	}

	log.Printf("Nmap done: %d hosts up scanned for %d ports in %3f seconds\n", len(result.Hosts), len(portList), result.Stats.Finished.Elapsed)

	return &AllScanResults{Scanresult: allScanResults, CreatedDate: createdDate.String()}, nil
}
