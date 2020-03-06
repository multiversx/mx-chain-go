package notifier

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/epochStart"
)

func (essh *epochStartSubscriptionHandler) RegisteredHandlers() ([]epochStart.EpochStartHandler, *sync.RWMutex) {
	return essh.epochStartHandlers, &essh.mutEpochStartHandler
}
