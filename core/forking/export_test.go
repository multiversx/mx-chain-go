package forking

import "github.com/ElrondNetwork/elrond-go/core"

// Handlers -
func (gen *genericEpochNotifier) Handlers() []core.EpochSubscriberHandler {
	gen.mutHandler.RLock()
	defer gen.mutHandler.RUnlock()

	return gen.handlers
}

// CurrentTimestamp -
func (gen *genericEpochNotifier) CurrentTimestamp() uint64 {
	_, timestamp := gen.getEpochTimestamp()

	return timestamp
}
