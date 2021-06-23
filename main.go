package main

import (
	"flag"
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/database"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/scanner"
	"github.com/cyrinux/grpcnmapscanner/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
)

func main() {

	isServer := flag.Bool("server", false, "start the gRPC server")
	isWorker := flag.Bool("worker", false, "start the worker")
	flag.Parse()

	dbDriver, ok := os.LookupEnv("DB_DRIVER")
	if !ok {
		fmt.Println("DB_DRIVER is not present, fallback on redis")
		dbDriver = "redis"
	}

	if *isServer {
		fmt.Print("Prepare to serve the gRPC api")
		listener, err := net.Listen("tcp", ":9000")
		if err != nil {
			panic(err) // The port may be on use
		}
		db, err := database.Factory(dbDriver)
		if err != nil {
			panic(err)
		}
		srv := grpc.NewServer()

		reflection.Register(srv)
		proto.RegisterScannerServiceServer(srv, scanner.NewServer(db))
		if e := srv.Serve(listener); e != nil {
			panic(err)
		}
	}

	if *isWorker {
		fmt.Print("I'm a scanner worker")
		worker.StartWorker()
	}

}
