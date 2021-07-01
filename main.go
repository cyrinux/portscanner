package main

import (
	"context"
	"flag"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/scanner"
	"github.com/cyrinux/grpcnmapscanner/worker"
)

func main() {
	isServer := flag.Bool("server", false, "start the gRPC server")
	isWorker := flag.Bool("worker", false, "start the worker")
	allConfig := config.GetConfig(context.Background())
	flag.Parse()

	if *isServer {
		scanner.Listen(allConfig)
	}

	if *isWorker {
		workerNMAP := worker.NewWorker(allConfig, "nmap")
		workerNMAP.StartWorker()
	}
}
