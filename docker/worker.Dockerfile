# syntax=docker/dockerfile:1

FROM golang:1.16 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpcnmapscanner .

FROM alpine:latest as worker
MAINTAINER  Cyril Levis <grpcnmapscanner@levis.name>
LABEL Name=grpcnmapscanner-worker Version=0.0.1
RUN apk --no-cache add ca-certificates nmap nmap-scripts && rm -f /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/grpcnmapscanner /app/grpcnmapscanner
CMD ["/app/grpcnmapscanner", "-worker"]
