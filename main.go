//go:generate go run -mod=vendor git.rootprojects.org/root/go-gitver/v2 --package version --outfile ./version/xversion.go
//go:generate protoc -I/usr/include -I. --go_out=plugins=grpc:. ./proto/service.proto ./proto/backend.proto

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/server"
	"github.com/cyrinux/grpcnmapscanner/version"
	"github.com/cyrinux/grpcnmapscanner/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	isServer := flag.Bool("server", false, "Start the gRPC server")
	isWorker := flag.Bool("worker", false, "Start the worker")
	wantProfiler := flag.Bool("pprof", false, "Start pprof profiler")
	wantPrometheus := flag.Bool("prometheus", false, "Expose prometheus stats")
	flag.Parse()

	if *showVersion {
		fmt.Print(version.Show())
		os.Exit(0)
	}

	ctx := context.Background()

	if *wantProfiler {
		go profilerListen()
	}
	if *wantPrometheus {
		go prometheusListen()
	}

	allConfig := config.GetConfig()

	log.Print(version.Show())

	if *isServer {
		startServer(ctx, allConfig)
	} else if *isWorker {
		startWorker(ctx, allConfig, "nmap")
	} else {
		startServer(ctx, allConfig)
		startWorker(ctx, allConfig, "nmap")
	}
}

func handleSignalWorker(worker *worker.Worker) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
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
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)
	<-signals // hard exit on second signal (in case shutdown gets stuck)
	os.Exit(1)
}

func startServer(ctx context.Context, conf config.Config) {
	server.Listen(ctx, conf)

	go handleSignalServer()
}

func prometheusListen() {
	log.Print("starting prometheus listener")
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println(http.ListenAndServe(":2112", nil))
}

func startWorker(ctx context.Context, conf config.Config, tasktype string) {
	w := worker.NewWorker(ctx, conf, tasktype)
	w.StartWorker()

	go handleSignalWorker(w)
}

func profilerListen() {
	log.Print("starting pprof listener")
	log.Println(http.ListenAndServe(":6060", nil))
}
