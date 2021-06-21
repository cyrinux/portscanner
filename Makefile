all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpcnmapscanner .

proto:
	docker run -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -f scanner/scanner.proto -l go
	sudo chown -R $(shell id -u):$(shell id -g) ./gen
	mv gen/pb-go/scanner/scanner.pb.go ./scanner
	rm ./gen -rf

build-docker: proto
	docker build -t grpcnmapscanner .

server:
	sudo ./grpcnmapscanner -grpc

testscan:
	docker run -d  --name mongo-scanner  -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=secret mongo
	grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'1.1.1.1,8.8.8.8',ports:'80,U:53,T:443,22,T:8040-8080'"
	grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'scanme.nmap.org',ports:'U:53,U:500'"
