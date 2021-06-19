all: build

build: proto
	go build

proto:
	@docker run -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -f scanner/scanner.proto -l go
	sudo chown -R $(shell id -u):$(shell id -g) ./gen
	@mv gen/pb-go/scanner/scanner.pb.go ./scanner
	@rm ./gen -rf

server:
	./grpcnmapscanner &

testscan: server
	grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'1.1.1.1,8.8.8.8',ports:'80,53,443,22,T:8040-8080'"
	grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'scanme.nmap.org',ports:'53,500',protocol:'udp'"
