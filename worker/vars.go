package worker

import (
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/logger"
	"google.golang.org/grpc/keepalive"
	"os"
	"time"
)

var (
	appConfig   = config.GetConfig()
	log         = logger.NewConsole(appConfig.Logger.Debug)
	hostname, _ = os.Hostname()
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             1 * time.Second,  // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}
