package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

func (e *epochStartMetaSyncer) SetEpochStartMetaBlockInterceptorProcessor(proc EpochStartMetaBlockInterceptorProcessor) {
	e.metaBlockProcessor = proc
}

func (e *epochStartMetaBlockProcessor) GetMapMetaBlock() map[string]data.MetaHeaderHandler {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	return e.mapReceivedMetaBlocks
}
