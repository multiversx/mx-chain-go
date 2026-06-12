package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

func (e *epochStartMetaSyncer) SetEpochStartMetaBlockInterceptorProcessor(proc EpochStartMetaBlockInterceptorProcessor) {
	e.metaBlockProcessor = proc
}

func (e *epochStartMetaBlockProcessor) GetMapMetaBlock() map[string]data.MetaHeaderHandler {
	e.mutReceivedMetaBlocks.RLock()
	defer e.mutReceivedMetaBlocks.RUnlock()

	return e.mapReceivedMetaBlocks
}

func (e *epochStartBootstrap) RebuildNetworkComponentsForShard() error {
	return e.rebuildNetworkComponentsForShard()
}

func (e *epochStartBootstrap) ResolversContainer() dataRetriever.ResolversContainer {
	return e.resolversContainer
}

func (e *epochStartBootstrap) MainInterceptorContainer() process.InterceptorsContainer {
	return e.mainInterceptorContainer
}

func (e *epochStartBootstrap) FullArchiveInterceptorContainer() process.InterceptorsContainer {
	return e.fullArchiveInterceptorContainer
}

func (e *epochStartBootstrap) RequestHandler() process.RequestHandler {
	return e.requestHandler
}
