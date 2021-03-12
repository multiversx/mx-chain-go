package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

func (e *epochStartMetaSyncer) SetEpochStartMetaBlockInterceptorProcessor(proc EpochStartMetaBlockInterceptorProcessor) {
	e.metaBlockProcessor = proc
}

func (e *epochStartMetaBlockProcessor) GetMapMetaBlock() map[string]data.HeaderHandler {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	return e.mapReceivedMetaBlocks
}
