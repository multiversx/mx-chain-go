package sender

import (
	"sync"
	"time"
)

type timerWrapper struct {
	mutTimer sync.Mutex
	timer    *time.Timer
}

// CreateNewTimer will stop the existing timer and will initialize a new one
func (wrapper *timerWrapper) CreateNewTimer(duration time.Duration) {
	wrapper.mutTimer.Lock()
	wrapper.stopTimer()
	wrapper.timer = time.NewTimer(duration)
	wrapper.mutTimer.Unlock()
}

// ExecutionReadyChannel returns the chan on which the ticker will emit periodic values as to signal that
// the execution is ready to take place
func (wrapper *timerWrapper) ExecutionReadyChannel() <-chan time.Time {
	wrapper.mutTimer.Lock()
	defer wrapper.mutTimer.Unlock()

	return wrapper.timer.C
}

func (wrapper *timerWrapper) stopTimer() {
	if wrapper.timer == nil {
		return
	}

	wrapper.timer.Stop()
}

// Close will simply stop the inner timer so this component won't contain leaked resource
func (wrapper *timerWrapper) Close() {
	wrapper.mutTimer.Lock()
	defer wrapper.mutTimer.Unlock()

	wrapper.stopTimer()
}
