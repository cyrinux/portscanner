package scanner

import (
	"fmt"
	"log"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/net/context"
)

type Server struct {
}

func (s *Server) GetScan(ctx context.Context, in *Task) (*ScanResults, error) {
	scanResults, err := GetTaskResult(in)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	// if scanResults != nil {
	// 	log.Printf("%v\n", scanResults)
	// 	log.Printf("%v\n", scanResults)
	// 	log.Printf("%+v\n", &scanResults)
	// 	log.Printf("%v\n", &scanResults)
	// } else {
	// 	log.Printf("%v\n", scanResults)
	// 	log.Printf("%v\n", scanResults)
	// 	log.Printf("%+v\n", &scanResults)
	// 	log.Printf("%v\n", &allScanResrint
	// 	log.Print("NUL !")
	log.Printf("%v\n", &*scanResults)
	// 	log.Printf("%v\n", &scanResults)
	// }
	return scanResults, err

	// return &ScanResults{
	// 	HostResult:  scanResults.HostResult,
	// 	CreatedDate: scanResults.CreatedDate,
	// 	Guid:        scanResults.Guid,
	// }, nil
}

// Scan function prepare a nmap scan
func (s *Server) Scan(ctx context.Context, in *Scanner) (*ScanResults, error) {
	scanId := xid.New()

	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	log.Printf("Starting scan of host: %s, port: %s, timeout: %v", in.Hosts, in.Ports, in.Timeout)

	portList := []*Port{}
	scanResults := []*HostResult{}
	totalPorts := 0

	result, err := StartNmapScan(in)
	if err != nil || result == nil {
		return &ScanResults{
			HostResult: nil,
			Guid:       scanId.String(),
		}, err
	}

	for _, host := range result.Hosts {
		osversion := "unknown"
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}
		if len(host.OS.Matches) > 0 {
			fp := host.OS.Matches[0]
			osversion = fmt.Sprintf("name: %v, accuracy: %v%%", fp.Name, fp.Accuracy)
		}
		address := host.Addresses[0].Addr
		hostResult := &Host{
			Address:   address,
			OsVersion: &osversion,
		}
		for _, p := range host.Ports {
			version := &PortVersion{
				ExtraInfos:  &p.Service.ExtraInfo,
				LowVersion:  &p.Service.LowVersion,
				HighVersion: &p.Service.HighVersion,
				Product:     &p.Service.Product,
			}
			newPort := &Port{
				PortId:      fmt.Sprintf("%v", p.ID),
				ServiceName: p.Service.Name,
				Protocol:    p.Protocol,
				State:       p.State.State,
				Version:     version,
			}
			portList = append(portList, newPort)
		}
		totalPorts += len(portList)

		scanResult := &HostResult{
			Host:  hostResult,
			Ports: portList,
		}

		scanResults = append(scanResults, scanResult)
		mongoTask := bson.M{"_id": scanId.String(), "result": scanResult}
		if _, err = InsertDbResult(&mongoTask); err != nil {
			return nil, err
		}

	}

	log.Printf("Nmap done: %d hosts up scanned for %d ports in %3f seconds\n", result.Stats.Hosts.Up, totalPorts, result.Stats.Finished.Elapsed)

	return &ScanResults{
		HostResult: scanResults,
		Guid:       scanId.String(),
	}, nil
}
