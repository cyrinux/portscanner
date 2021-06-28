# syntax=docker/dockerfile:1

FROM golang:1.16 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest as scanner
MAINTAINER  Cyril Levis <grpcnmapscanner@levis.name>
LABEL Name=grpcnmapscanner Version=0.0.2
RUN apk --no-cache add ca-certificates nmap nmap-scripts && rm -f /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/grpcnmapscanner /app/grpcnmapscanner
EXPOSE 9000
ENTRYPOINT ["/app/grpcnmapscanner"]
