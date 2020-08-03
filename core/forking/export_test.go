package forking

import "github.com/ElrondNetwork/elrond-go/core"

func (gen *genericEpochNotifier) Handlers() []core.EpochNotifiedHandler {
	gen.mutHandler.RLock()
	defer gen.mutHandler.RUnlock()

	return gen.handlers
}
