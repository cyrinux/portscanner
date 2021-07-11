# TODO

- include nmap vulners https://github.com/vulnersCom/nmap-vulners
- get worker/consumer state on server start or register the last server state in redis
- add rpc service to StartConsume and StopConsume the queue
- Reports metrics with prometheus or github.com/rs/xstats
- Use go-cron for reccuring scan tasks (crontal in ParamsScannerRequest)
- scan nmap export to defectdojo

# Done

- remove a task in queue with DeleteScan
- test and config redis in cluster https://github.com/bitnami/bitnami-docker-redis#setting-up-replication
- well handle context, to be able to cancel a nmap scan on control-c for example
