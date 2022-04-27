package logging

import (
	"context"
	"sync"
	"time"
)

type lifeSpanner struct {
	mutDuration sync.RWMutex
	duration    time.Duration

	notifyChan   chan struct{}
	resetChan    chan struct{}
	cancel       context.CancelFunc
	checkHandler func() bool
}

func newLifeSpanner(notifyChan chan struct{}, checkHandler func() bool, initialDuration time.Duration) *lifeSpanner {
	spanner := &lifeSpanner{
		duration:     initialDuration,
		notifyChan:   notifyChan,
		resetChan:    make(chan struct{}),
		checkHandler: checkHandler,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	spanner.cancel = cancelFunc
	go spanner.process(ctx)

	return spanner
}

func (spanner *lifeSpanner) reset() {
	spanner.resetChan <- struct{}{}
}

func (spanner *lifeSpanner) resetDuration(newDuration time.Duration) {
	spanner.mutDuration.Lock()
	spanner.duration = newDuration
	spanner.mutDuration.Unlock()

	spanner.reset()
}

func (spanner *lifeSpanner) getDuration() time.Duration {
	spanner.mutDuration.RLock()
	defer spanner.mutDuration.RUnlock()

	return spanner.duration
}

func (spanner *lifeSpanner) process(ctx context.Context) {
	timer := time.NewTimer(spanner.getDuration())

	defer timer.Stop()

	for {
		timer.Reset(spanner.getDuration())

		select {
		case <-ctx.Done():
			log.Debug("closing lifeSpanner.process go routine")
			return
		case <-timer.C:
			if spanner.checkHandler() {
				spanner.notifyChan <- struct{}{}
			}
		case <-spanner.resetChan: // will cause an iteration which will reset the timer automatically
		}
	}
}

func (spanner *lifeSpanner) close() {
	spanner.cancel()
}
