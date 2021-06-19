package scanner

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
)

type Server struct {
}

func (s *Server) Scan(ctx context.Context, in *Scanner) (*ReturnAllPorts, error) {
	log.Printf("Starting %s scan of host: %s, port: %s", in.Protocol, in.Hosts, in.Ports)

	result := StartNmapScan(in)
	portList := []*Port{}
	var port = new(Port)

	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		for _, p := range host.Ports {
			port = &Port{
				Host:        host.Addresses[0].Addr,
				Id:          fmt.Sprintf("%v", p.ID),
				ServiceName: string(p.Service.Name),
				Protocol:    string(p.Protocol),
				State:       p.State.State,
			}
			portList = append(portList, port)
		}
	}

	log.Printf("Nmap done: %d hosts up scanned for %d ports in %3f seconds\n", len(result.Hosts), len(portList), result.Stats.Finished.Elapsed)

	return &ReturnAllPorts{Port: portList}, nil
}
