package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

func (e *epochStartMetaSyncer) SetEpochStartMetaBlockInterceptorProcessor(proc EpochStartMetaBlockInterceptorProcessor) {
	e.metaBlockProcessor = proc
}

func (e *epochStartMetaBlockProcessor) GetMapMetaBlock() map[string]data.MetaHeaderHandler {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	return e.mapReceivedMetaBlocks
}
