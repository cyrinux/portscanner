# gRPC nmap scanner

# Requirement

- docker-compose
- docker

# Architecture

## gRPC

![](graphviz.svg)

# Build

```bash
$ make
```

# Server

Start the server as root with

```bash
$ make
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
