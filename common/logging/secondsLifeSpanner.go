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
	log.Debug("newSecondsLifeSpanner entered", "timespan", timeSpanInSeconds)
	if timeSpanInSeconds < minFileLifeSpan {
		return nil, fmt.Errorf("%w, provided %v", core.ErrInvalidLogFileMinLifeSpan, timeSpanInSeconds)
	}

	sls := &secondsLifeSpanner{
		timeSpanInSeconds: timeSpanInSeconds,
		baseLifeSpanner:   newBaseLifeSpanner(),
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	sls.cancelFunc = cancelFunc

	go sls.startTicker(ctx)

	return sls, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sls *secondsLifeSpanner) IsInterfaceNil() bool {
	return sls == nil
}

// Close closes all internal components
func (sls *secondsLifeSpanner) Close() error {
	if sls.cancelFunc != nil {
		sls.cancelFunc()
	}
	return nil
}

func (sls *secondsLifeSpanner) startTicker(ctx context.Context) {
	for {
		select {
		case <-time.After(sls.timeSpanInSeconds):
			sls.lifeSpanChannel <- ""
		case <-ctx.Done():
			log.Debug("closing secondsLifeSpanner go routine")
			sls.baseLifeSpanner.Close()
			return
		}
	}
}
