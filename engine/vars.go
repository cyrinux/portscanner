package engine

import (
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/logger"
)

var (
	appConfig = config.GetConfig()
	log       = logger.New(appConfig.Logger.Debug, appConfig.Logger.Pretty)
)
