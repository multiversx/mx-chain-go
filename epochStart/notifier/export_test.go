package notifier

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/epochStart"
)

func (essh *epochStartSubscriptionHandler) RegisteredHandlers() ([]epochStart.ActionHandler, *sync.RWMutex) {
	return essh.epochStartHandlers, &essh.mutEpochStartHandler
}
