package ntp

import (
	"fmt"
	"sync"
	"time"
)

type ShadowForkSyncTimer struct {
	lastTime    time.Time
	lastNtpTime time.Time
	currentTime time.Time
	mut         sync.RWMutex
}

// NewShadowForkSyncTimer -
func NewShadowForkSyncTimer(initialTime time.Time, increaseChan chan uint64) *ShadowForkSyncTimer {
	sfst := &ShadowForkSyncTimer{lastTime: initialTime, currentTime: initialTime, lastNtpTime: time.Now()}
	go func(sfst *ShadowForkSyncTimer) {
		setRound := false
		for {
			select {
			//case <-time.After(time.Second * 2):
			//	log.Debug("NewShadowForkSyncTimer from time.After")
			case round := <-increaseChan:
				log.Debug("NewShadowForkSyncTimer from chan", "round", round)
				if !setRound {
					sfst.lastTime = sfst.lastTime.Add(time.Second * time.Duration(6*(round-1)))
					sfst.currentTime = sfst.lastTime
					log.Info("NewShadowForkSyncTimer set lastTime", "lastTime", sfst.lastTime, "round", round, "currentTime", sfst.currentTime)
					setRound = true
				}
			}
			log.Debug("NewShadowForkSyncTimer increase time")
			sfst.IncreaseTime(time.Second * 6)
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
	diff := time.Now().Sub(sfst.lastNtpTime)
	if diff > time.Millisecond*5990 {
		diff = 0
	}
	sfst.currentTime = sfst.lastTime.Add(diff)
	//	log.Info("ShadowForkSyncTimer CurrentTime", "diff", diff, "lastNtpTime", sfst.lastNtpTime, "lastTime", sfst.lastTime, "currentTime", sfst.currentTime)
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
	sfst.lastNtpTime = time.Now()
	sfst.lastTime = sfst.lastTime.Add(duration)
	sfst.currentTime = sfst.lastTime
	log.Info("ShadowForkSyncTimer IncreaseTime", "lastTime", sfst.lastTime)
}
