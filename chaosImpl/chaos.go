package chaosImpl

import (
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")
var startTime time.Time

func init() {
	startTime = time.Now()
}
