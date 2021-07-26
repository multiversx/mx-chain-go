package forking

import vmcommon "github.com/ElrondNetwork/elrond-vm-common"

// Handlers -
func (gen *genericEpochNotifier) Handlers() []vmcommon.EpochSubscriberHandler {
	gen.mutHandler.RLock()
	defer gen.mutHandler.RUnlock()

	return gen.handlers
}

// CurrentTimestamp -
func (gen *genericEpochNotifier) CurrentTimestamp() uint64 {
	_, timestamp := gen.getEpochTimestamp()

	return timestamp
}
