package notifier

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

func (essh *epochStartSubscriptionHandler) RegisteredHandlers() ([]core.EpochStartActionHandler, *sync.RWMutex) {
	return essh.epochStartHandlers, &essh.mutEpochStartHandler
}

func (mesn *manualEpochStartNotifier) Handlers() []core.EpochStartActionHandler {
	mesn.mutHandlers.RLock()
	defer mesn.mutHandlers.RUnlock()

	handlers := make([]core.EpochStartActionHandler, len(mesn.handlers))
	copy(handlers, mesn.handlers)

	return handlers
}
