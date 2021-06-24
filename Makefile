all: up

up: build-docker
	docker-compose up -d --build --scale worker=3
	docker-compose logs -f --tail=50

build:
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -installsuffix cgo -o grpcnmapscanner .

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: proto
proto:
	@echo Generating protobuf code from docker
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -f proto/service.proto -l go
	@sudo chown -R $(shell id -u):$(shell id -g) ./gen
	@mv gen/pb-go/proto/service.pb.go ./proto
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
	protodot -src proto/service.proto -output graphviz
	dot -Tpng ~/protodot/generated/graphviz.dot -o graphviz.png
	dot -Tsvg ~/protodot/generated/graphviz.dot -o graphviz.svg
	xdg-open graphviz.png

