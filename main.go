package main

import (
	"fmt"

	"flag"
	grpc "github.com/cyrinux/grpcnmapscanner/grpcserver"
)

func main() {

	isServer := flag.Bool("server", false, "enable server service")
	isWorker := flag.Bool("worker", false, "enable worker service")
	flag.Parse()

	if *isWorker {
		fmt.Println("I'm a worker")

	}

	if *isServer {
		grpc.StartServer()
	}

}
