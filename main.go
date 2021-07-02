package main

import (
	"context"
	"flag"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/server"
	"github.com/cyrinux/grpcnmapscanner/worker"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func handleSignalWorker(worker *worker.Worker) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()
	<-signals              // wait for signal
	worker.StopConsuming() // wait for all Consume() calls to finish
}

func handleSignalServer() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)
	<-signals // hard exit on second signal (in case shutdown gets stuck)
	os.Exit(1)
}

func startServer(ctx context.Context, config config.Config) {
	if err := server.Listen(ctx, config); err != nil {
		os.Exit(1)
	}
	go handleSignalServer()
}

func startWorker(ctx context.Context, config config.Config, tasktype string) {
	w := worker.NewWorker(ctx, config, tasktype)
	w.StartWorker()
	go handleSignalWorker(w)
}

func main() {
	isServer := flag.Bool("server", false, "start the gRPC server")
	isWorker := flag.Bool("worker", false, "start the worker")
	wantProfiler := flag.Bool("pprof", false, "start pprof profiler")
	allConfig := config.GetConfig()
	flag.Parse()

	if *wantProfiler {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	ctx := context.Background()

	if *isServer {
		startServer(ctx, allConfig)
	} else if *isWorker {
		startWorker(ctx, allConfig, "nmap")
	} else {
		startServer(ctx, allConfig)
		startWorker(ctx, allConfig, "nmap")
	}

}
