# TODO

- use s6-overlay as init to start TOR https://github.com/just-containers/s6-overlay
- envoy config https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples
- use default grpc code and status?
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"
- try some basic auth https://github.com/grpc/grpc-go/issues/106
- get worker/consumer state on server start or register the last server state in redis
- Use go-cron for reccuring scan tasks (crontal in ParamsScannerRequest)
- scan nmap export to defectdojo
- generate k8s deploy files with kompose

# Done

- split large network range in several smaller and publish them as new tasks
- include nmap vulners https://github.com/vulnersCom/nmap-vulners
- a new scan should create a main scan which contains sub tasks
- remove a task in queue with DeleteScan
- test and config redis in cluster https://github.com/bitnami/bitnami-docker-redis#setting-up-replication
- well handle context, to be able to cancel a nmap scan on control-c for example
- add rpc service to StartConsume and StopConsume the queue
- Reports metrics with prometheus or github.com/rs/xstats
