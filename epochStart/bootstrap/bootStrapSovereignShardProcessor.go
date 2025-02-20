package bootstrap

import (
	"context"
	"fmt"
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
	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/trie/factory"
	updateSync "github.com/multiversx/mx-chain-go/update/sync"
)

type sovereignBootStrapShardProcessor struct {
	*sovereignChainEpochStartBootstrap
}

func (sbp *sovereignBootStrapShardProcessor) requestAndProcessForShard(peerMiniBlocks []*block.MiniBlock) error {
	argsStorageHandler := StorageHandlerArgs{
		GeneralConfig:                   sbp.generalConfig,
		PreferencesConfig:               sbp.prefsConfig,
		ShardCoordinator:                sbp.shardCoordinator,
		PathManagerHandler:              sbp.coreComponentsHolder.PathHandler(),
		Marshaller:                      sbp.coreComponentsHolder.InternalMarshalizer(),
		Hasher:                          sbp.coreComponentsHolder.Hasher(),
		CurrentEpoch:                    sbp.baseData.lastEpoch,
		Uint64Converter:                 sbp.coreComponentsHolder.Uint64ByteSliceConverter(),
		NodeTypeProvider:                sbp.coreComponentsHolder.NodeTypeProvider(),
		NodesCoordinatorRegistryFactory: sbp.nodesCoordinatorRegistryFactory,
		ManagedPeersHolder:              sbp.cryptoComponentsHolder.ManagedPeersHolder(),
		NodeProcessingMode:              sbp.nodeProcessingMode,
		StateStatsHandler:               sbp.stateStatsHandler,
		AdditionalStorageServiceCreator: sbp.runTypeComponents.AdditionalStorageServiceCreator(),
	}
	storageHandlerComponent, err := NewShardStorageHandler(argsStorageHandler)
	if err != nil {
		return err
	}

	sovStorageHandler := newSovereignShardStorageHandler(storageHandlerComponent)

	defer sovStorageHandler.CloseStorageService()

	sbp.closeTrieComponents()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		sbp.generalConfig,
		sbp.coreComponentsHolder,
		sovStorageHandler.storageService,
		sbp.stateStatsHandler,
	)
	if err != nil {
		return err
	}

	sbp.trieContainer = triesContainer
	sbp.trieStorageManagers = trieStorageManagers

	log.Debug("start in epoch bootstrap: started syncUserAccountsState", "rootHash", sbp.epochStartMeta.GetRootHash())
	err = sbp.syncUserAccountsState(sbp.epochStartMeta.GetRootHash())
	if err != nil {
		return err
	}

	log.Debug("start in epoch bootstrap: started syncValidatorAccountsState", "validatorRootHash", sbp.epochStartMeta.GetValidatorStatsRootHash())
	err = sbp.syncValidatorAccountsState(sbp.epochStartMeta.GetValidatorStatsRootHash())
	if err != nil {
		return err
	}

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: sbp.epochStartMeta,
		PreviousEpochStart:  sbp.prevEpochStartMeta,
		ShardHeader:         sbp.epochStartMeta,
		NodesConfig:         sbp.nodesConfig,
		Headers:             sbp.syncedHeaders,
		ShardCoordinator:    sbp.shardCoordinator,
		PendingMiniBlocks:   make(map[string]*block.MiniBlock),
		PeerMiniBlocks:      peerMiniBlocks,
	}

	return sovStorageHandler.SaveDataToStorage(components, sbp.epochStartMeta, false, make(map[string]*block.MiniBlock))
}

func (sbp *sovereignBootStrapShardProcessor) computeNumShards(_ data.MetaHeaderHandler) uint32 {
	return 1
}

func (sbp *sovereignBootStrapShardProcessor) createRequestHandler() (process.RequestHandler, error) {
	requestersContainerArgs := requesterscontainer.FactoryArgs{
		RequesterConfig:                 sbp.generalConfig.Requesters,
		ShardCoordinator:                sbp.shardCoordinator,
		MainMessenger:                   sbp.mainMessenger,
		FullArchiveMessenger:            sbp.fullArchiveMessenger,
		Marshaller:                      sbp.coreComponentsHolder.InternalMarshalizer(),
		Uint64ByteSliceConverter:        uint64ByteSlice.NewBigEndianConverter(),
		OutputAntifloodHandler:          disabled.NewAntiFloodHandler(),
		CurrentNetworkEpochProvider:     disabled.NewCurrentNetworkEpochProviderHandler(),
		MainPreferredPeersHolder:        disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		PeersRatingHandler:              disabled.NewDisabledPeersRatingHandler(),
		SizeCheckDelta:                  0,
	}
	requestersFactory, err := sbp.runTypeComponents.RequestersContainerFactoryCreator().CreateRequesterContainerFactory(requestersContainerArgs)
	if err != nil {
		return nil, err
	}

	container, err := requestersFactory.Create()
	if err != nil {
		return nil, err
	}

	finder, err := containers.NewRequestersFinder(container, sbp.shardCoordinator)
	if err != nil {
		return nil, err
	}

	return sbp.runTypeComponents.RequestHandlerCreator().CreateRequestHandler(
		requestHandlers.RequestHandlerArgs{
			RequestersFinder:      finder,
			RequestedItemsHandler: cache.NewTimeCache(timeBetweenRequests),
			WhiteListHandler:      sbp.whiteListHandler,
			MaxTxsToRequest:       maxToRequest,
			ShardID:               core.SovereignChainShardId,
			RequestInterval:       timeBetweenRequests,
		},
	)
}

func (sbp *sovereignBootStrapShardProcessor) createResolversContainer() error {
	return nil
}

func (sbp *sovereignBootStrapShardProcessor) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	return sbp.baseSyncHeaders(meta, DefaultTimeToWaitForRequestedData)
}

func (sbp *sovereignBootStrapShardProcessor) syncHeadersFromStorage(
	meta data.MetaHeaderHandler,
	_ uint32,
	_ uint32,
	timeToWaitForRequestedData time.Duration,
) (map[string]data.HeaderHandler, error) {
	return sbp.baseSyncHeaders(meta, timeToWaitForRequestedData)
}

func (sbp *sovereignBootStrapShardProcessor) baseSyncHeaders(
	meta data.MetaHeaderHandler,
	timeToWaitForRequestedData time.Duration,
) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, 2)
	shardIds := make([]uint32, 0, 2)

	for _, epochStartData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		hashesToRequest = append(hashesToRequest, epochStartData.GetHeaderHash())
		shardIds = append(shardIds, epochStartData.GetShardID())
	}

	if meta.GetEpoch() > sbp.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.SovereignChainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToWaitForRequestedData)
	err := sbp.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := sbp.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == sbp.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.SovereignChainHeader{}
	}

	return syncedHeaders, nil
}

func (sbp *sovereignBootStrapShardProcessor) processNodesConfigFromStorage(pubKey []byte, _ uint32) (nodesCoordinator.NodesCoordinatorRegistryHandler, uint32, error) {
	var err error
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:                         sbp.dataPool,
		Marshalizer:                      sbp.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler:                   sbp.requestHandler,
		ChanceComputer:                   sbp.rater,
		GenesisNodesConfig:               sbp.genesisNodesConfig,
		NodeShuffler:                     sbp.nodeShuffler,
		Hasher:                           sbp.coreComponentsHolder.Hasher(),
		PubKey:                           pubKey,
		ShardIdAsObserver:                core.SovereignChainShardId,
		ChanNodeStop:                     sbp.coreComponentsHolder.ChanStopNodeProcess(),
		NodeTypeProvider:                 sbp.coreComponentsHolder.NodeTypeProvider(),
		IsFullArchive:                    sbp.prefsConfig.FullArchive,
		EnableEpochsHandler:              sbp.coreComponentsHolder.EnableEpochsHandler(),
		NodesCoordinatorRegistryFactory:  sbp.nodesCoordinatorRegistryFactory,
		NodesCoordinatorWithRaterFactory: sbp.runTypeComponents.NodesCoordinatorWithRaterCreator(),
	}
	sbp.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return nil, 0, err
	}

	// no need to save the peers miniblocks here as they were already fetched from the DB
	nodesConfig, _, _, err := sbp.nodesConfigHandler.NodesConfigFromMetaBlock(sbp.epochStartMeta, sbp.prevEpochStartMeta)
	return nodesConfig, core.SovereignChainShardId, err
}

func (sbp *sovereignBootStrapShardProcessor) createEpochStartMetaSyncer() (epochStart.StartOfEpochMetaSyncer, error) {
	epochStartConfig := sbp.generalConfig.EpochStartConfig
	metaBlockProcessor, err := NewEpochStartMetaBlockProcessor(
		sbp.mainMessenger,
		sbp.requestHandler,
		sbp.coreComponentsHolder.InternalMarshalizer(),
		sbp.coreComponentsHolder.Hasher(),
		thresholdForConsideringMetaBlockCorrect,
		epochStartConfig.MinNumConnectedPeersToStart,
		epochStartConfig.MinNumOfPeersToConsiderBlockValid,
	)
	if err != nil {
		return nil, err
	}

	argsEpochStartSyncer := ArgsNewEpochStartMetaSyncer{
		CoreComponentsHolder:    sbp.coreComponentsHolder,
		CryptoComponentsHolder:  sbp.cryptoComponentsHolder,
		RequestHandler:          sbp.requestHandler,
		Messenger:               sbp.mainMessenger,
		ShardCoordinator:        sbp.shardCoordinator,
		EconomicsData:           sbp.economicsData,
		WhitelistHandler:        sbp.whiteListHandler,
		StartInEpochConfig:      epochStartConfig,
		HeaderIntegrityVerifier: sbp.headerIntegrityVerifier,
		MetaBlockProcessor:      newEpochStartSovereignBlockProcessor(metaBlockProcessor),
	}

	return newEpochStartSovereignSyncer(argsEpochStartSyncer)
}

func (sbp *sovereignBootStrapShardProcessor) createStorageEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (epochStart.StartOfEpochMetaSyncer, error) {
	return newEpochStartSovereignSyncer(args)
}

func (sbp *sovereignBootStrapShardProcessor) createEpochStartInterceptorsContainers(args bootStrapFactory.ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	containerFactoryArgs, err := bootStrapFactory.CreateEpochStartContainerFactoryArgs(args)
	if err != nil {
		return nil, nil, err
	}

	sp, err := interceptorscontainer.NewShardInterceptorsContainerFactory(*containerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	incomingHeaderProcessor, err := incomingHeader.CreateIncomingHeaderProcessor(
		sbp.generalConfig.SovereignConfig.NotifierConfig.WebSocketConfig,
		sbp.dataPool,
		sbp.generalConfig.SovereignConfig.MainChainNotarization.MainChainNotarizationStartRound,
		sbp.runTypeComponents,
	)
	if err != nil {
		return nil, nil, err
	}

	interceptorsContainerFactory, err := interceptorscontainer.NewSovereignShardInterceptorsContainerFactory(interceptorscontainer.ArgsSovereignShardInterceptorsContainerFactory{
		ShardContainer:           sp,
		IncomingHeaderSubscriber: incomingHeaderProcessor,
	})
	if err != nil {
		return nil, nil, err
	}

	return interceptorsContainerFactory.Create()
}

func (bp *sovereignBootStrapShardProcessor) createCrossHeaderRequester() (updateSync.CrossHeaderRequester, error) {
	extendedHeaderRequester, castOk := bp.requestHandler.(updateSync.ExtendedShardHeaderRequestHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignBootStrapShardProcessor.createCrossHeaderRequester for extendedHeaderRequester", process.ErrWrongTypeAssertion)
	}

	return updateSync.NewExtendedHeaderRequester(extendedHeaderRequester)
}
