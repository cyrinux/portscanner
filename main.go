package main

import (
	"context"
	"flag"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/server"
	"github.com/cyrinux/grpcnmapscanner/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func startServer(ctx context.Context, conf config.Config) {

	server.Listen(ctx, conf)

	go handleSignalServer()
}

func prometheusListen() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func startWorker(ctx context.Context, conf config.Config, tasktype string) {
	w := worker.NewWorker(ctx, conf, tasktype)
	w.StartWorker()

	go handleSignalWorker(w)
}

func main() {
	isServer := flag.Bool("server", false, "start the gRPC server")
	isWorker := flag.Bool("worker", false, "start the worker")
	wantProfiler := flag.Bool("pprof", false, "start pprof profiler")
	wantPrometheus := flag.Bool("prometheus", false, "expose prometheus stats")
	allConfig := config.GetConfig()
	flag.Parse()

	if *wantProfiler {
		go func() {
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}
	if *wantPrometheus {
		log.Print("starting prometheus")
		go prometheusListen()
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
