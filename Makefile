export DOCKER_BUILDKIT=1

all: up

up: build-docker
	@docker network create my-network --subnet 10.190.33.0/24 || true
	docker-compose up -d

logs: up
	docker-compose logs -f --tail=50

build: generate
	 @echo Build grpcnmapscanner
	 @GOARCH=amd64 GOOS=linux go build -mod=vendor -ldflags="-w -s" -o grpcnmapscanner .

generate: deps
	@echo Generate go:generate protobuf and stuff
	@go generate -mod=vendor ./...

.PHONY: vendor
vendor: deps
	@echo Make vendor
	@go mod tidy
	@go mod vendor

build-docker:
	docker-compose build

.PHONY: server
server:
	./grpcnmapscanner -server

.PHONY: worker
worker:
	./grpcnmapscanner -worker

graphviz:
	@protodot -src proto/backend.proto -output graphviz
	@dot -Tpng ~/protodot/generated/graphviz.dot -o graphviz.png
	@dot -Tsvg ~/protodot/generated/graphviz.dot -o graphviz.svg
	@xdg-open graphviz.png

clean:
	rm -f proto/*.pb.go grpcnmapscanner

.PHONY: proto
proto: generate

deps:
	@echo Fetching go deps
	@go get -u github.com/golang/protobuf/proto
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@go get -u google.golang.org/grpc

proto-docker:
	@echo Generating protobuf code from docker
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/service.proto
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/backend.proto
	@sudo chown -R $(shell id -u):$(shell id -g) ./gen
	@mv gen/pb-go/proto/*.pb.go ./proto
	@rm ./gen -rf

testscan: up
	docker-compose run --rm client grpc_cli ls server:9000
	docker-compose run --rm client grpc_cli call server:9000 scanner.ScannerService.StartScan "hosts:'1.1.1.1,8.8.8.8',ports:'80,U:53,T:443,22,T:8040-8080'"
	docker-compose run --rm client grpc_cli call server:9000 scanner.ScannerService.StartAsyncScan "hosts:'scanme.nmap.org',fast_mode:true'"
	docker-compose run --rm client grpc_cli call server:9000 scanner.ScannerService.StartAsyncScan "hosts:'1.1.1.1,8.8.8.8',ports:'80,U:53,T:443,22,T:8040-8080'"
	docker-compose logs -f --tail=50
