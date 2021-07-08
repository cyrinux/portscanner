export DOCKER_BUILDKIT=1

all: up

up: build-docker
	docker network create my-network --subnet 10.190.33.0/24 || true
	docker-compose up -d --scale worker=2
	docker-compose logs -f --tail=50

build:
	 CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -mod vendor -installsuffix cgo -o grpcnmapscanner .

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: proto
proto:
	@echo Generating protobuf code from docker
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/service.proto
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/backend.proto
	@sudo chown -R $(shell id -u):$(shell id -g) ./gen
	@mv gen/pb-go/proto/*.pb.go ./proto
	@rm ./gen -rf

build-docker: proto
	docker-compose build

.PHONY: server
server:
	./grpcnmapscanner -server

.PHONY: worker
worker:
	./grpcnmapscanner -worker

testscan: up
	docker-compose run --rm client grpc_cli ls server:9000
	docker-compose run --rm client grpc_cli call server:9000 scanner.ScannerService.StartScan "hosts:'1.1.1.1,8.8.8.8',ports:'80,U:53,T:443,22,T:8040-8080'"
	docker-compose run --rm client grpc_cli call server:9000 scanner.ScannerService.StartAsyncScan "hosts:'scanme.nmap.org',fast_mode:true'"
	docker-compose run --rm client grpc_cli call server:9000 scanner.ScannerService.StartAsyncScan "hosts:'1.1.1.1,8.8.8.8',ports:'80,U:53,T:443,22,T:8040-8080'"
	docker-compose logs -f --tail=50

graphviz:
	protodot -src proto/backend.proto -output graphviz
	dot -Tpng ~/protodot/generated/graphviz.dot -o graphviz.png
	dot -Tsvg ~/protodot/generated/graphviz.dot -o graphviz.svg
	xdg-open graphviz.png

cobra:
	protoc \
		-I. \
		--gofast_out=plugins=grpc:. \
		--cobra_out=plugins=client:. \
		proto/*.proto

clean:
	rm -f proto/*.pb.go grpcnmapscanner

deps:
	go get github.com/gogo/protobuf/protoc-gen-gofast
	go get github.com/Zenithar/go-protoc-gen-cobra
