package worker

import "time"

const (
	returnerLimit      = 1000
	pollDuration       = 100 * time.Millisecond
	pollDurationPushed = 5000 * time.Millisecond
	reportBatchSize    = 10000
	consumeDuration    = 1 * time.Second
)
