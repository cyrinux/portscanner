package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/scanner"
	"github.com/cyrinux/grpcnmapscanner/worker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func main() {

	isServer := flag.Bool("server", false, "start the gRPC server")
	isWorker := flag.Bool("worker", false, "start the worker")
	allConfig := config.GetConfig(context.Background())
	flag.Parse()

	if *isServer {
		fmt.Println("Prepare to serve the gRPC api")
		listener, err := net.Listen("tcp", ":9000")
		if err != nil {
			panic(err) // The port may be on use
		}
		srv := grpc.NewServer()

		reflection.Register(srv)
		proto.RegisterScannerServiceServer(srv, scanner.NewServer(allConfig))
		if e := srv.Serve(listener); e != nil {
			panic(err)
		}
	}

	if *isWorker {
		worker := worker.NewWorker(allConfig)
		worker.StartWorker()
	}

}
