package broker

import (
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/logger"
)

var (
	conf = config.GetConfig()
	log  = logger.New(conf.Logger.Debug, conf.Logger.Pretty)
)
