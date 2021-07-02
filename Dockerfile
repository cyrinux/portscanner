# syntax=docker/dockerfile:1

FROM golang:1.16 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest as scanner
MAINTAINER  Cyril Levis <grpcnmapscanner@levis.name>
LABEL Name=grpcnmapscanner Version=0.0.3
RUN apk --no-cache add ca-certificates nmap nmap-scripts && rm -f /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/grpcnmapscanner /usr/local/bin/grpcnmapscanner
EXPOSE 9000 6060
ENTRYPOINT ["grpcnmapscanner"]
