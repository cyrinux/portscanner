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

	ctx := context.Background()

	if *isServer {
		scanner.Listen(allConfig, ctx)
	}

	if *isWorker {
		worker := worker.NewWorker(allConfig, ctx)
		worker.StartWorker()
	}
}
