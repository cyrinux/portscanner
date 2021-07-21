package worker

import (
	// "context"
	"crypto/tls"
	"crypto/x509"
	pb "github.com/cyrinux/grpcnmapscanner/proto/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/metadata"
	// "google.golang.org/grpc/status"
)

func loadTLSCredentials(caFile, certFile, keyFile string) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, errors.New("failed to add server CA's certificate")
	}

	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

// WorkerClient is a client to call worker service RPCs
type WorkerClient struct {
	service pb.BackendServiceClient
}

// NewWorkerClient returns a new worker client
func NewWorkerClient(cc *grpc.ClientConn) *WorkerClient {
	service := pb.NewBackendServiceClient(cc)
	return &WorkerClient{service}
}
