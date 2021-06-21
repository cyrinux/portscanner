package scanner

import (
	"fmt"
	"log"
	"time"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

type Server struct {
}

func (s *Server) GetScan(ctx context.Context, in *Task) (*AllScanResults, error) {
	resultBytes, err := GetTaskResult(in)
	for r := range resultBytes {
		log.Printf("%+v\n\n", r)
	}
	err = bson.Unmarshal(resultBytes, &AllScanResults{})

	return &AllScanResults{}, err
}

// Scan function prepare a nmap scan
func (s *Server) Scan(ctx context.Context, in *Scanner) (*AllScanResults, error) {
	createdDate := time.Now()
	scanId := xid.New()

	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	log.Printf("Starting scan of host: %s, port: %s, timeout: %v", in.Hosts, in.Ports, in.Timeout)

	portList := []*Port{}
	allScanResults := []*HostResult{}
	totalPorts := 0

	result, err := StartNmapScan(in)
	if err != nil || result == nil {
		return &AllScanResults{
			HostResult:  nil,
			CreatedDate: createdDate.String(),
			Guid:        scanId.String(),
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
			Host:        hostResult,
			Ports:       portList,
			CreatedDate: time.Now().String(),
		}

		allScanResults = append(allScanResults, scanResult)
		mongoRow := bson.D{primitive.E{
			Key:   scanId.String(),
			Value: scanResult,
		}}
		_, err = InsertDbResult(&mongoRow)
	}

	log.Printf("Nmap done: %d hosts up scanned for %d ports in %3f seconds\n", result.Stats.Hosts.Up, totalPorts, result.Stats.Finished.Elapsed)

	return &AllScanResults{
		HostResult:  allScanResults,
		CreatedDate: createdDate.String(),
		Guid:        scanId.String(),
	}, nil
}
