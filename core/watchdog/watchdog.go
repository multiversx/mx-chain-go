package watchdog

import (
	"bytes"
	"runtime/pprof"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
)

var log = logger.GetOrCreate("watchdog")

type watchdog struct {
	alarmScheduler      core.TimersScheduler
	chanStopNodeProcess chan endProcess.ArgEndProcess
}

// NewWatchdog creates a new instance of WatchdogTimer
func NewWatchdog(
	alarmScheduler core.TimersScheduler,
	chanStopNodeProcess chan endProcess.ArgEndProcess,
) (core.WatchdogTimer, error) {
	if check.IfNil(alarmScheduler) {
		return nil, ErrNilAlarmScheduler
	}
	if chanStopNodeProcess == nil {
		return nil, ErrNilEndProcessChan
	}

	return &watchdog{
		alarmScheduler:      alarmScheduler,
		chanStopNodeProcess: chanStopNodeProcess,
	}, nil
}

// Set sets the given alarm
func (w *watchdog) Set(callback func(alarmID string), duration time.Duration, alarmID string) {
	w.alarmScheduler.Add(callback, duration, alarmID)
}

// SetDefault sets the default alarm with the specified duration.
// When the default alarm expires, the goroutines stack traces will be logged, and the node will gracefully close.
func (w *watchdog) SetDefault(duration time.Duration, watchdogID string) {
	w.alarmScheduler.Add(w.defaultWatchdogExpiry, duration, watchdogID)
}

func (w *watchdog) defaultWatchdogExpiry(watchdogID string) {
	buffer := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buffer, 1)
	if err != nil {
		log.Error("could not dump goroutines")
	}

	log.Error("watchdog alarm has expired", "alarm", watchdogID)
	log.Warn(buffer.String())

	arg := endProcess.ArgEndProcess{
		Reason:      "alarm " + watchdogID + " has expired",
		Description: "the " + watchdogID + " is stuck",
	}
	w.chanStopNodeProcess <- arg
}

// Stop stops the alarm with the specified ID
func (w *watchdog) Stop(alarmID string) {
	w.alarmScheduler.Cancel(alarmID)
}

// Reset resets the alarm with the given ID
func (w *watchdog) Reset(alarmID string) {
	w.alarmScheduler.Reset(alarmID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (w *watchdog) IsInterfaceNil() bool {
	return w == nil
}
