# TODO

- test and config redis in cluster https://github.com/bitnami/bitnami-docker-redis#setting-up-replication
- add rpc service to StartConsume and StopConsume the queue
- well handle context, to be able to cancel a nmap scan on control-c for example
- Reports metrics with prometheus or github.com/rs/xstats
- Use go-cron for reccuring scan tasks (crontal in ParamsScannerRequest)
- scan nmap export to defectdojo

# Done

- remove a task in queue with DeleteScan
