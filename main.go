package main

import (
	"fmt"

	"flag"
	"github.com/cyrinux/grpcnmapscanner/server"
)

func main() {

	isServer := flag.Bool("server", false, "enable server service")
	isWorker := flag.Bool("worker", false, "enable worker service")
	flag.Parse()

	if *isWorker {
		fmt.Println("I'm a worker")

	}

	if *isServer {
		server.StartServer()
	}

}
