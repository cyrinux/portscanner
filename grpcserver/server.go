package grpcserver

import (
	"fmt"
	"log"
	"net"

	"github.com/cyrinux/grpcnmapscanner/scanner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// StartServer start the gRPC server
func StartServer() {
	fmt.Println("Go gRPC NMAP Scanner!")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := scanner.Server{}
	grpcServer := grpc.NewServer()

	scanner.RegisterScannerServiceServer(grpcServer, &s)
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
