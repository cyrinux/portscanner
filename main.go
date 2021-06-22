package main

import (
	"fmt"
	"github.com/cyrinux/grpcnmapscanner/proto"
	"github.com/cyrinux/grpcnmapscanner/scanner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func main() {
	fmt.Println("Prepare to serve")

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err) // The port may be on use
	}
	grpcServer := grpc.NewServer()

	reflection.Register(grpcServer)
	proto.RegisterScannerServiceServer(grpcServer, &scanner.Server{})
	if e := grpcServer.Serve(listener); e != nil {
		panic(err)
	}

}
