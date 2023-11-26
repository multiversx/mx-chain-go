package ntp

import (
	"fmt"
	"sync"
	"time"
)

type ShadowForkSyncTimer struct {
	lastTime    time.Time
	currentTime time.Time
	mut         sync.RWMutex
}

// NewShadowForkSyncTimer -
func NewShadowForkSyncTimer(initialTime time.Time, increaseChan <-chan bool) *ShadowForkSyncTimer {
	sfst := &ShadowForkSyncTimer{lastTime: initialTime, currentTime: initialTime}
	go func(sfst *ShadowForkSyncTimer) {
		for {
			select {
			//case <-time.After(time.Second * 2):
			//	log.Debug("NewShadowForkSyncTimer from time.After")
			case <-increaseChan:
				log.Debug("NewShadowForkSyncTimer from chan")
			}
			log.Debug("NewShadowForkSyncTimer increase time")
			sfst.IncreaseTime(time.Millisecond * 5990)
		}
	}(sfst)
	return sfst
}

// Close -
func (sfst *ShadowForkSyncTimer) Close() error {
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

// CurrentTime -
func (sfst *ShadowForkSyncTimer) CurrentTime() time.Time {
	sfst.mut.RLock()
	defer sfst.mut.RUnlock()
	sfst.currentTime = sfst.currentTime.Add(time.Millisecond * 1)
	return sfst.currentTime
}

// IsInterfaceNil returns true if there is no value under the interface
func (sfst *ShadowForkSyncTimer) IsInterfaceNil() bool {
	return sfst == nil
}

// IncreaseTime -
func (sfst *ShadowForkSyncTimer) IncreaseTime(duration time.Duration) {
	sfst.mut.Lock()
	defer sfst.mut.Unlock()

	sfst.lastTime = sfst.lastTime.Add(duration)
	sfst.currentTime = sfst.lastTime
	log.Info("ShadowForkSyncTimer IncreaseTime", "lastTime", sfst.lastTime)
}
