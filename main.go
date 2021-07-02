package main

import (
	"flag"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/server"
	"github.com/cyrinux/grpcnmapscanner/worker"
)

func main() {
	isServer := flag.Bool("server", false, "start the gRPC server")
	isWorker := flag.Bool("worker", false, "start the worker")
	allConfig := config.GetConfig()
	flag.Parse()

	if *isServer {
		server.Listen(allConfig)
	}

	if *isWorker {
		workerNMAP := worker.NewWorker(allConfig, "nmap")
		workerNMAP.StartWorker()
	}
}
