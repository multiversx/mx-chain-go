package logging

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

const minEpochsLifeSpan = 1

type epochsLifeSpanner struct {
	*baseLifeSpanner
	spanInEpochs uint32
	cancelFunc   context.CancelFunc
}

func newEpochsLifeSpanner(es EpochStartNotifierWithConfirm, epochsLifeSpan uint32) (*epochsLifeSpanner, error) {
	log.Info("newEpochsLifeSpanner entered", "timespan", epochsLifeSpan)
	if epochsLifeSpan < minEpochsLifeSpan {
		return nil, fmt.Errorf("NewEpochsLifeSpanner %w, provided %v", core.ErrInvalidLogFileMinLifeSpan, epochsLifeSpan)
	}

	sls := &epochsLifeSpanner{
		spanInEpochs:    epochsLifeSpan,
		baseLifeSpanner: newBaseLifeSpanner(),
	}

	log.Info("the epochsLifeSpanner", "timespan", epochsLifeSpan)

	es.RegisterForEpochChangeConfirmed(
		func(epoch uint32) {
			if epoch%sls.spanInEpochs == 0 {
				log.Info("Ticked once", "timespan", sls.spanInEpochs, "epoch", epoch)
				sls.tickChannel <- fmt.Sprintf("%v", epoch)
			}
		},
	)

	return sls, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sls *epochsLifeSpanner) IsInterfaceNil() bool {
	return sls == nil
}
