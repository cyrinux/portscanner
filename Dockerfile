# syntax=docker/dockerfile:1

FROM golang:alpine as builder
WORKDIR /app
COPY . .
RUN apk --no-cache add make git protobuf-dev musl-dev openssl
RUN make build

FROM alpine:latest as scanner
MAINTAINER  Cyril Levis <grpcnmapscanner@levis.name>
LABEL Name=grpcnmapscanner
RUN apk --no-cache add ca-certificates nmap nmap-scripts iputils tcptraceroute fping openssl && rm -f /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/grpcnmapscanner /usr/local/bin/grpcnmapscanner
EXPOSE 9000 9001 6060 2112
ENTRYPOINT ["grpcnmapscanner"]
