package logging

import (
	"fmt"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("api/shared/logging")

const thresholdMinAPICallDurationToLog = 200 * time.Millisecond

// LogAPIActionDurationIfNeeded will log the duration of an action triggered by an API call if it's above a threshold
func LogAPIActionDurationIfNeeded(startTime time.Time, action string) {
	duration := time.Since(startTime)
	if duration < thresholdMinAPICallDurationToLog {
		return
	}
	if duration.Nanoseconds() == 1<<63-1 {
		// fix for this issue https://github.com/tcnksm/go-httpstat/issues/13
		// in some situations, a duration of 2562047h47m16.854775807s was printed
		return
	}

	log.Debug(fmt.Sprintf("%s took %s", action, duration))
}
