package logging

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-logger/check"
)

const minEpochsLifeSpan = 1

type epochsLifeSpanner struct {
	*baseLifeSpanner
	spanInEpochs uint32
	cancelFunc   context.CancelFunc
}

func newEpochsLifeSpanner(es EpochStartNotifierWithConfirm, epochsLifeSpan uint32) (*epochsLifeSpanner, error) {
	log.Info("newEpochsLifeSpanner entered", "timespan", epochsLifeSpan)
	if check.IfNil(es) {
		return nil, fmt.Errorf("%w, epoch start notifier is nil", core.ErrInvalidLogFileMinLifeSpan)
	}
	if epochsLifeSpan < minEpochsLifeSpan {
		return nil, fmt.Errorf("%w, min: %v, provided %v", core.ErrInvalidLogFileMinLifeSpan, minEpochsLifeSpan, epochsLifeSpan)
	}

	els := &epochsLifeSpanner{
		spanInEpochs:    epochsLifeSpan,
		baseLifeSpanner: newBaseLifeSpanner(),
	}

	es.RegisterForEpochChangeConfirmed(
		func(epoch uint32) {
			if epoch%els.spanInEpochs == 0 {
				els.lifeSpanChannel <- fmt.Sprintf("%v", epoch)
			}
		},
	)

	return els, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sls *epochsLifeSpanner) IsInterfaceNil() bool {
	return sls == nil
}
