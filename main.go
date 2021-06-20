package main

import (
	"flag"
	// "fmt"
	grpc "github.com/cyrinux/grpcnmapscanner/grpcserver"
	// "github.com/cyrinux/grpcnmapscanner/tasks"
)

func main() {

	// isTasksServer := flag.Bool("tasksserver", false, "start tasks server")
	isGRPCServer := flag.Bool("grpc", false, "start gRPC server")
	// isWorker := flag.Bool("worker", false, "start a tasks worker")

	flag.Parse()

	// if *isWorker {
	// 	fmt.Print("I'm a worker")
	// }

	if *isGRPCServer {
		grpc.StartServer()

	}
	// if *isTasksServer {
	// 	tasks.StartServer()
	// }

}
