# syntax=docker/dockerfile:1

FROM golang:1.16 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates nmap nmap-scripts
WORKDIR /app
COPY --from=builder /app/grpcnmapscanner /app/grpcnmapscanner
EXPOSE 9000
CMD ["/app/grpcnmapscanner"]
