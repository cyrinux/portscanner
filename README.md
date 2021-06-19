# gRPC nmap scanner

# Requirement

- golang
- nmap
- grpc_cli

# Build

```bash
$ make
```

# Server

Start the server as root with

```bash
$ sudo make server
```

or with docker

```bash
$ docker build -t grpcnmapscanner:latest .
$ docker run --cap-add=NET_RAW --rm -p 127.0.0.1:9000:9000 -ti grpcnmapscanner:latest
```

# Client

Run a test scan with

```bash
$ make testscan
```

Or

```bash
$ grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'google.com',ports:'80,443'"
```
