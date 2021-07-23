package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

const minFileLifeSpan = time.Second

type secondsLifeSpanner struct {
	*baseLifeSpanner
	timeSpanInSeconds time.Duration
	cancelFunc        context.CancelFunc
}

func newSecondsLifeSpanner(timeSpanInSeconds time.Duration) (*secondsLifeSpanner, error) {
	log.Info("NewSecondsLifeSpanner entered", "timespan", timeSpanInSeconds)
	if timeSpanInSeconds < minFileLifeSpan {
		return nil, fmt.Errorf("%w, provided %v", core.ErrInvalidLogFileMinLifeSpan, timeSpanInSeconds)
	}

	sls := &secondsLifeSpanner{
		timeSpanInSeconds: timeSpanInSeconds,
		baseLifeSpanner:   newBaseLifeSpanner(),
	}

	log.Info("the secondsLifeSpanner", "timespan", timeSpanInSeconds)

	ctx, cancelFunc := context.WithCancel(context.Background())
	sls.cancelFunc = cancelFunc

	go sls.startTicker(ctx)

	return sls, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sls *secondsLifeSpanner) IsInterfaceNil() bool {
	return sls == nil
}

func (sls *secondsLifeSpanner) startTicker(ctx context.Context) {
	ct := 0
	for {
		select {
		case <-time.After(sls.timeSpanInSeconds):
			log.Info("Ticked once", "timespan", sls.timeSpanInSeconds, "ct", ct)
			sls.tickChannel <- ""
		case <-ctx.Done():
			log.Debug("closing secondsLifeSpanner go routine")
			return
		}
	}
}
