package notifier

import (
	"sync"
)

func (essh *epochStartSubscriptionHandler) RegisteredHandlers() ([]SubscribeFunctionHandler, *sync.RWMutex) {
	return essh.epochStartHandlers, &essh.mutEpochStartHandler
}
