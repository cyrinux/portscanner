package engine

import (
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/logger"
)

var (
	appConfig = config.GetConfig()
	log       = logger.NewConsole(appConfig.Logger.Debug)
)
