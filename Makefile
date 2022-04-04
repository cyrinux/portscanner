export DOCKER_BUILDKIT=1
export GRPCUSER=user1
export GRPCPASS=secret1

all: up

up: build-docker
	@docker network create my-network --subnet 10.190.33.0/24 || true
	docker-compose up -d

logs: up
	docker-compose logs -f --tail=50

build: generate
	 @echo Build grpcnmapscanner
	 @GOARCH=amd64 GOOS=linux go build -mod=vendor -ldflags="-w -s" -o grpcnmapscanner .

generate: deps vendor
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

.PHONY: graphviz
graphviz:
	@protodot -inc vendor,proto -src proto/v1/backend.proto -output graphviz
	@dot -Tpng ~/protodot/generated/graphviz.dot -o graphviz.png
	@dot -Tsvg ~/protodot/generated/graphviz.dot -o graphviz.svg
	@xdg-open graphviz.png

clean:
	rm -f proto/*/*.pb.go grpcnmapscanner

.PHONY: proto
proto: generate

deps:
	@echo Fetching go deps
	@go install github.com/seamia/protodot@latest
	@go install github.com/golang/protobuf/protoc-gen-go@latest

proto-docker:
	@echo Generating protobuf code from docker
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/service.proto
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/backend.proto
	@docker run --rm -e UID=$(shell id -u) -e GID=$(shell id -g) -v `pwd`:/defs namely/protoc-all -l go -f proto/auth_service.proto
	@sudo chown -R $(shell id -u):$(shell id -g) ./gen
	@mv gen/pb-go/proto/*.pb.go ./proto
	@rm ./gen -rf

testscan:
	./bin/scanner StartAsyncScan '{"targets":"scanme.nmap.org","fast_mode":true}'

.PHONY: cert
cert:
	cd cert; ./gen.sh; cd ..

chaos-kill:
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock gaiaadm/pumba --label com.docker.compose.project=grpcnmapscanner -l info --random --interval 30s kill

chaos-loss:
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock gaiaadm/pumba --label com.docker.compose.project=grpcnmapscanner -l info --random --interval 30s netem --tc-image gaiadocker/iproute2 --duration 15s loss --percent 30

chaos-delay:
	docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock gaiaadm/pumba --label com.docker.compose.project=grpcnmapscanner -l info --random --interval 30s netem --tc-image gaiadocker/iproute2 --duration 15s delay --time 5000

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
