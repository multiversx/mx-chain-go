package ntp

import (
	"fmt"
	"github.com/multiversx/mx-chain-go/process"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// roundDuration represents the standard duration of a blockchain round
	roundDuration = 6 * time.Second
	// maxTimeDrift is the maximum allowed drift before resetting (round duration - 10ms)
	maxTimeDrift = roundDuration - 10*time.Millisecond
	// autoAdvanceInterval is the interval for automatic time advancement
	autoAdvanceInterval = 1 * time.Second
)

type ShadowForkSyncTimer struct {
	lastTime    time.Time
	lastNtpTime time.Time
	currentTime time.Time
	mut         sync.RWMutex
	closeChan   chan struct{}
	closed      atomic.Bool
	wg          sync.WaitGroup
	handler     process.RoundHandler
}

// NewShadowForkSyncTimer creates a new shadow fork sync timer with controlled time progression
func NewShadowForkSyncTimer(initialTime time.Time, increaseChan chan uint64) *ShadowForkSyncTimer {
	sfst := &ShadowForkSyncTimer{
		lastTime:    initialTime,
		currentTime: initialTime,
		lastNtpTime: time.Now(),
		closeChan:   make(chan struct{}),
	}

	sfst.wg.Add(1)
	go sfst.timeLoop(increaseChan)

	return sfst
}

// timeLoop handles the time progression logic
func (sfst *ShadowForkSyncTimer) timeLoop(increaseChan chan uint64) {
	defer sfst.wg.Done()

	setRound := false
	ticker := time.NewTicker(autoAdvanceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sfst.closeChan:
			log.Debug("ShadowForkSyncTimer: shutting down time loop")
			return

		case <-ticker.C:
			log.Debug("ShadowForkSyncTimer: auto-advancing time")
			sfst.IncreaseTime(roundDuration)

		case round := <-increaseChan:
			if round == 0 {
				log.Warn("ShadowForkSyncTimer: received invalid round 0, ignoring")
				continue
			}

			log.Debug("ShadowForkSyncTimer: received round update", "round", round)

			if !setRound {
				sfst.mut.Lock()
				roundOffset := roundDuration * time.Duration(round)
				sfst.lastTime = sfst.lastTime.Add(roundOffset)
				sfst.currentTime = sfst.lastTime
				sfst.mut.Unlock()

				log.Info("ShadowForkSyncTimer: initialized to round",
					"round", round,
					"lastTime", sfst.lastTime,
					"currentTime", sfst.currentTime)
				setRound = true
			}
		}
	}
}

// Close stops the timer and releases resources
func (sfst *ShadowForkSyncTimer) Close() error {
	if sfst.closed.Swap(true) {
		// Already closed
		return nil
	}

	close(sfst.closeChan)
	sfst.wg.Wait()

	log.Debug("ShadowForkSyncTimer: closed successfully")
	return nil
}

// StartSyncingTime -
func (sfst *ShadowForkSyncTimer) StartSyncingTime() {
	return
}

// ClockOffset -
func (sfst *ShadowForkSyncTimer) ClockOffset() time.Duration {
	return time.Duration(0)
}

// FormattedCurrentTime -
func (sfst *ShadowForkSyncTimer) FormattedCurrentTime() string {
	t := sfst.CurrentTime()
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
	return str
}

// CurrentTime returns the current simulated time
func (sfst *ShadowForkSyncTimer) CurrentTime() time.Time {
	sfst.mut.Lock()
	defer sfst.mut.Unlock()

	diff := time.Since(sfst.lastNtpTime)
	if diff > maxTimeDrift {
		diff = 0
	}

	sfst.currentTime = sfst.lastTime.Add(diff)
	return sfst.currentTime
}

// IsInterfaceNil returns true if there is no value under the interface
func (sfst *ShadowForkSyncTimer) IsInterfaceNil() bool {
	return sfst == nil
}

func (sfst *ShadowForkSyncTimer) SetRoundHandler(handler process.RoundHandler) {
	sfst.mut.Lock()
	defer sfst.mut.Unlock()

	sfst.handler = handler
	log.Debug("ShadowForkSyncTimer: round handler set", "handler", handler)
}

// IncreaseTime advances the simulated time by the specified duration
func (sfst *ShadowForkSyncTimer) IncreaseTime(duration time.Duration) {
	if duration < 0 {
		log.Warn("ShadowForkSyncTimer: attempted to increase time by negative duration", "duration", duration)
		return
	}

	sfst.mut.Lock()
	defer sfst.mut.Unlock()

	sfst.lastNtpTime = time.Now()
	sfst.lastTime = sfst.lastTime.Add(duration)
	sfst.currentTime = sfst.lastTime

	log.Info("ShadowForkSyncTimer: time advanced",
		"duration", duration,
		"newTime", sfst.lastTime)
}
