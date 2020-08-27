package notifier

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/epochStart"
)

func (essh *epochStartSubscriptionHandler) RegisteredHandlers() ([]epochStart.ActionHandler, *sync.RWMutex) {
	return essh.epochStartHandlers, &essh.mutEpochStartHandler
}

func (mesn *manualEpochStartNotifier) Handlers() []epochStart.ActionHandler {
	mesn.mutHandlers.RLock()
	defer mesn.mutHandlers.RUnlock()

	handlers := make([]epochStart.ActionHandler, len(mesn.handlers))
	copy(handlers, mesn.handlers)

	return handlers
}
