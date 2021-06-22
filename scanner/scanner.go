package scanner

import (
	"fmt"
	"log"

	"bytes"
	"encoding/json"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/rs/xid"
	"golang.org/x/net/context"
)

type Server struct {
}

var DB database.Database

func main() database.Database {
	databaseImplementation := "redis"
	DB, err := database.Factory(databaseImplementation)
	if err != nil {
		panic(err)
	}
	return DB
}

func (s *Server) GetScan(ctx context.Context, in *proto.GetScannerResponse) (*proto.ServerResponse, error) {
	// scanResult, err := db.Get(in)
	// if err != nil {
	// 	return nil, err
	// }
	return generateResponse("", nil)
}

// Scan function prepare a nmap scan
func (s *Server) StartScan(ctx context.Context, in *proto.SetScannerRequest) (*proto.ServerResponse, error) {
	scanId := xid.New()

	if in.Timeout < 10 {
		in.Timeout = 60 * 5
	}

	log.Printf("Starting scan of host: %s, port: %s, timeout: %v", in.Hosts, in.Ports, in.Timeout)

	portList := []*proto.Port{}
	scanResult := []*proto.HostResult{}
	totalPorts := 0

	result, err := StartNmapScan(in)
	if err != nil || result == nil {
		return generateResponse("empty scan", err)
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
		hostResult := &proto.Host{
			Address:   address,
			OsVersion: &osversion,
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
				State:       p.State.State,
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

	scannerResponse := &proto.ScannerResponse{HostResult: scanResult}

	scanResultJSON, err := json.Marshal(scannerResponse)
	prettyprint(scanResultJSON)
	if err != nil {
		return generateResponse("Can't convert as json", err)
	}
	_, err = DB.Set(scanId.String(), string(scanResultJSON))
	if err != nil {
		return generateResponse("Can't store in redis", err)
	}

	log.Printf("Nmap done: %d hosts up scanned for %d ports in %3f seconds\n", result.Stats.Hosts.Up, totalPorts, result.Stats.Finished.Elapsed)

	return generateResponse(string(scanResultJSON), err)
}

func generateResponse(value string, err error) (*proto.ServerResponse, error) {
	if err != nil {
		return &proto.ServerResponse{Success: false, Value: value, Error: err.Error()}, nil
	}
	return &proto.ServerResponse{Success: true, Value: value, Error: ""}, nil
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	fmt.Printf("%#v", string(out.Bytes()))
	return out.Bytes(), err
}
