package logging

import (
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("api/shared/logging")

const thresholdMinAPICallDurationToLog = 200 * time.Millisecond

// LogAPIActionDurationIfNeeded will log the duration of an action triggered by an API call if it's above a threshold
func LogAPIActionDurationIfNeeded(startTime time.Time, action string) {
	duration := time.Since(startTime)
	if duration > thresholdMinAPICallDurationToLog {
		log.Debug(fmt.Sprintf("%s took %s", action, duration))
	}
}
