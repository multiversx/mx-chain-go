package bootstrap

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	bootStrapFactory "github.com/multiversx/mx-chain-go/epochStart/bootstrap/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/trie/factory"
)

type sovereignBootStrapShardProcessor struct {
	*sovereignChainEpochStartBootstrap
}

func (e *sovereignBootStrapShardProcessor) requestAndProcessForShard(peerMiniBlocks []*block.MiniBlock) error {
	argsStorageHandler := StorageHandlerArgs{
		GeneralConfig:                   e.generalConfig,
		PreferencesConfig:               e.prefsConfig,
		ShardCoordinator:                e.shardCoordinator,
		PathManagerHandler:              e.coreComponentsHolder.PathHandler(),
		Marshaller:                      e.coreComponentsHolder.InternalMarshalizer(),
		Hasher:                          e.coreComponentsHolder.Hasher(),
		CurrentEpoch:                    e.baseData.lastEpoch,
		Uint64Converter:                 e.coreComponentsHolder.Uint64ByteSliceConverter(),
		NodeTypeProvider:                e.coreComponentsHolder.NodeTypeProvider(),
		NodesCoordinatorRegistryFactory: e.nodesCoordinatorRegistryFactory,
		ManagedPeersHolder:              e.cryptoComponentsHolder.ManagedPeersHolder(),
		NodeProcessingMode:              e.nodeProcessingMode,
		StateStatsHandler:               e.stateStatsHandler,
		AdditionalStorageServiceCreator: e.runTypeComponents.AdditionalStorageServiceCreator(),
	}
	storageHandlerComponent, err := NewShardStorageHandler(argsStorageHandler)
	if err != nil {
		return err
	}

	sovStorageHandler := newSovereignShardStorageHandler(storageHandlerComponent)

	defer sovStorageHandler.CloseStorageService()

	e.closeTrieComponents()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		e.generalConfig,
		e.coreComponentsHolder,
		sovStorageHandler.storageService,
		e.stateStatsHandler,
	)
	if err != nil {
		return err
	}

	e.trieContainer = triesContainer
	e.trieStorageManagers = trieStorageManagers

	log.Debug("start in epoch bootstrap: started syncUserAccountsState", "rootHash", e.epochStartMeta.GetRootHash())
	err = e.syncUserAccountsState(e.epochStartMeta.GetRootHash())
	if err != nil {
		return err
	}

	log.Debug("start in epoch bootstrap: started syncValidatorAccountsState", "validatorRootHash", e.epochStartMeta.GetValidatorStatsRootHash())
	err = e.syncValidatorAccountsState(e.epochStartMeta.GetValidatorStatsRootHash())
	if err != nil {
		return err
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: e.epochStartMeta,
		PreviousEpochStart:  e.prevEpochStartMeta,
		ShardHeader:         e.epochStartMeta,
		NodesConfig:         e.nodesConfig,
		Headers:             e.syncedHeaders,
		ShardCoordinator:    e.shardCoordinator,
		PendingMiniBlocks:   make(map[string]*block.MiniBlock),
		PeerMiniBlocks:      peerMiniBlocks,
	}

	return sovStorageHandler.SaveDataToStorage(components, e.epochStartMeta, false, make(map[string]*block.MiniBlock))
}

func (e *sovereignBootStrapShardProcessor) computeNumShards(_ data.MetaHeaderHandler) uint32 {
	return 1
}

func (e *sovereignBootStrapShardProcessor) createRequestHandler() (process.RequestHandler, error) {
	requestersContainerArgs := requesterscontainer.FactoryArgs{
		RequesterConfig:                 e.generalConfig.Requesters,
		ShardCoordinator:                e.shardCoordinator,
		MainMessenger:                   e.mainMessenger,
		FullArchiveMessenger:            e.fullArchiveMessenger,
		Marshaller:                      e.coreComponentsHolder.InternalMarshalizer(),
		Uint64ByteSliceConverter:        uint64ByteSlice.NewBigEndianConverter(),
		OutputAntifloodHandler:          disabled.NewAntiFloodHandler(),
		CurrentNetworkEpochProvider:     disabled.NewCurrentNetworkEpochProviderHandler(),
		MainPreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		PeersRatingHandler:              disabled.NewDisabledPeersRatingHandler(),
		SizeCheckDelta:                  0,
	}
	requestersFactory, err := e.runTypeComponents.RequestersContainerFactoryCreator().CreateRequesterContainerFactory(requestersContainerArgs)
	if err != nil {
		return nil, err
	}

	container, err := requestersFactory.Create()
	if err != nil {
		return nil, err
	}

	finder, err := containers.NewRequestersFinder(container, e.shardCoordinator)
	if err != nil {
		return nil, err
	}

	return e.runTypeComponents.RequestHandlerCreator().CreateRequestHandler(
		requestHandlers.RequestHandlerArgs{
			RequestersFinder:      finder,
			RequestedItemsHandler: cache.NewTimeCache(timeBetweenRequests),
			WhiteListHandler:      e.whiteListHandler,
			MaxTxsToRequest:       maxToRequest,
			ShardID:               core.SovereignChainShardId,
			RequestInterval:       timeBetweenRequests,
		},
	)
}

func (e *sovereignBootStrapShardProcessor) createResolversContainer() error {
	return nil
}

func (e *sovereignBootStrapShardProcessor) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	return e.baseSyncHeaders(meta, DefaultTimeToWaitForRequestedData)
}

func (e *sovereignBootStrapShardProcessor) syncHeadersFromStorage(
	meta data.MetaHeaderHandler,
	_ uint32,
	_ uint32,
	timeToWaitForRequestedData time.Duration,
) (map[string]data.HeaderHandler, error) {
	return e.baseSyncHeaders(meta, timeToWaitForRequestedData)
}

func (e *sovereignBootStrapShardProcessor) baseSyncHeaders(
	meta data.MetaHeaderHandler,
	timeToWaitForRequestedData time.Duration,
) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, 1)
	shardIds := make([]uint32, 0, 1)
	if meta.GetEpoch() > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.SovereignChainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToWaitForRequestedData)
	err := e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == e.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.SovereignChainHeader{}
	}

	return syncedHeaders, nil
}

func (e *sovereignBootStrapShardProcessor) processNodesConfigFromStorage(pubKey []byte, _ uint32) (nodesCoordinator.NodesCoordinatorRegistryHandler, uint32, error) {
	var err error
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:                         e.dataPool,
		Marshalizer:                      e.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler:                   e.requestHandler,
		ChanceComputer:                   e.rater,
		GenesisNodesConfig:               e.genesisNodesConfig,
		NodeShuffler:                     e.nodeShuffler,
		Hasher:                           e.coreComponentsHolder.Hasher(),
		PubKey:                           pubKey,
		ShardIdAsObserver:                core.SovereignChainShardId,
		ChanNodeStop:                     e.coreComponentsHolder.ChanStopNodeProcess(),
		NodeTypeProvider:                 e.coreComponentsHolder.NodeTypeProvider(),
		IsFullArchive:                    e.prefsConfig.FullArchive,
		EnableEpochsHandler:              e.coreComponentsHolder.EnableEpochsHandler(),
		NodesCoordinatorRegistryFactory:  e.nodesCoordinatorRegistryFactory,
		NodesCoordinatorWithRaterFactory: e.runTypeComponents.NodesCoordinatorWithRaterCreator(),
	}
	e.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return nil, 0, err
	}

	// no need to save the peers miniblocks here as they were already fetched from the DB
	nodesConfig, _, _, err := e.nodesConfigHandler.NodesConfigFromMetaBlock(e.epochStartMeta, e.prevEpochStartMeta)
	return nodesConfig, core.SovereignChainShardId, err
}

func (e *sovereignBootStrapShardProcessor) createEpochStartMetaSyncer() (epochStart.StartOfEpochMetaSyncer, error) {
	epochStartConfig := e.generalConfig.EpochStartConfig
	metaBlockProcessor, err := NewEpochStartMetaBlockProcessor(
		e.mainMessenger,
		e.requestHandler,
		e.coreComponentsHolder.InternalMarshalizer(),
		e.coreComponentsHolder.Hasher(),
		thresholdForConsideringMetaBlockCorrect,
		epochStartConfig.MinNumConnectedPeersToStart,
		epochStartConfig.MinNumOfPeersToConsiderBlockValid,
	)
	if err != nil {
		return nil, err
	}

	argsEpochStartSyncer := ArgsNewEpochStartMetaSyncer{
		CoreComponentsHolder:    e.coreComponentsHolder,
		CryptoComponentsHolder:  e.cryptoComponentsHolder,
		RequestHandler:          e.requestHandler,
		Messenger:               e.mainMessenger,
		ShardCoordinator:        e.shardCoordinator,
		EconomicsData:           e.economicsData,
		WhitelistHandler:        e.whiteListHandler,
		StartInEpochConfig:      epochStartConfig,
		HeaderIntegrityVerifier: e.headerIntegrityVerifier,
		MetaBlockProcessor:      newEpochStartSovereignBlockProcessor(metaBlockProcessor),
	}

	return newEpochStartSovereignSyncer(argsEpochStartSyncer)
}

func (e *sovereignBootStrapShardProcessor) createStorageEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (epochStart.StartOfEpochMetaSyncer, error) {
	return newEpochStartSovereignSyncer(args)
}

func (e *sovereignBootStrapShardProcessor) createEpochStartInterceptorsContainers(args bootStrapFactory.ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	containerFactoryArgs, err := bootStrapFactory.CreateEpochStartContainerFactoryArgs(args)
	if err != nil {
		return nil, nil, err
	}

	sp, err := interceptorscontainer.NewShardInterceptorsContainerFactory(*containerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	interceptorsContainerFactory, err := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(interceptorscontainer.ArgsSovereignShardInterceptorsContainerFactory{
		ShardContainer:           sp,
		IncomingHeaderSubscriber: &sovereign.IncomingHeaderSubscriberStub{},
	})
	if err != nil {
		return nil, nil, err
	}

	return interceptorsContainerFactory.Create()
}
