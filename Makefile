all: build

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpcnmapscanner .

.PHONY: proto
proto:
	docker run -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -f proto/service.proto -l go
	sudo chown -R $(shell id -u):$(shell id -g) ./gen
	mv gen/pb-go/proto/service.pb.go ./proto
	rm ./gen -rf

build-docker: proto
	docker build -t grpcnmapscanner .

server:
	./grpcnmapscanner -grpc

testscan:
	docker run -d  --name mongo-scanner  -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=secret mongo
	grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'1.1.1.1,8.8.8.8',ports:'80,U:53,T:443,22,T:8040-8080'"
	grpc_cli call 127.0.0.1:9000 scanner.ScannerService.Scan "hosts:'scanme.nmap.org',ports:'U:53,U:500'"

graphviz:
	protodot -src proto/service.proto -output graphviz
	dot -Tpng /home/cyril/protodot/generated/graphviz.dot -o graphviz.png
	dot -Tsvg /home/cyril/protodot/generated/graphviz.dot -o graphviz.svg
	xdg-open graphviz.png

