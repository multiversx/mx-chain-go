package bootstrap

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/partitioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/containers"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/resolverscontainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/heartbeat/validator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/trie/factory"
)

type bootStrapShardRequester struct {
	*epochStartBootstrap
}

func (e *bootStrapShardRequester) requestAndProcessForShard(peerMiniBlocks []*block.MiniBlock) error {
	epochStartData, err := e.findSelfShardEpochStartData()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = e.miniBlocksSyncer.SyncPendingMiniBlocks(epochStartData.GetPendingMiniBlockHeaderHandlers(), ctx)
	cancel()
	if err != nil {
		return err
	}

	pendingMiniBlocks, err := e.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: GetMiniBlocks", "num synced", len(pendingMiniBlocks))

	shardIds := []uint32{
		core.MetachainShardId,
		core.MetachainShardId,
	}
	lastFinishedMeta := epochStartData.GetLastFinishedMetaBlock()
	firstPendingMetaBlock := epochStartData.GetFirstPendingMetaBlock()
	hashesToRequest := [][]byte{
		lastFinishedMeta,
		firstPendingMetaBlock,
	}

	e.headersSyncer.ClearFields()
	ctx, cancel = context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return err
	}

	neededHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: SyncMissingHeadersByHash")

	for hash, hdr := range neededHeaders {
		e.syncedHeaders[hash] = hdr
	}

	shardNotarizedHeader, ok := e.syncedHeaders[string(epochStartData.GetHeaderHash())].(data.ShardHeaderHandler)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	dts, err := e.getDataToSync(
		epochStartData,
		shardNotarizedHeader,
	)
	if err != nil {
		return err
	}

	for hash, hdr := range dts.additionalHeaders {
		e.syncedHeaders[hash] = hdr
	}

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

	defer storageHandlerComponent.CloseStorageService()

	e.closeTrieComponents()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		e.generalConfig,
		e.coreComponentsHolder,
		storageHandlerComponent.storageService,
		e.stateStatsHandler,
	)
	if err != nil {
		return err
	}

	e.trieContainer = triesContainer
	e.trieStorageManagers = trieStorageManagers

	log.Debug("start in epoch bootstrap: started syncUserAccountsState", "rootHash", dts.rootHashToSync)
	err = e.syncUserAccountsState(dts.rootHashToSync)
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: syncUserAccountsState")

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: e.epochStartMeta,
		PreviousEpochStart:  e.prevEpochStartMeta,
		ShardHeader:         dts.ownShardHdr,
		NodesConfig:         e.nodesConfig,
		Headers:             e.syncedHeaders,
		ShardCoordinator:    e.shardCoordinator,
		PendingMiniBlocks:   pendingMiniBlocks,
		PeerMiniBlocks:      peerMiniBlocks,
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(components, shardNotarizedHeader, dts.withScheduled, dts.miniBlocks)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (e *bootStrapShardRequester) computeNumShards(epochStartMeta data.MetaHeaderHandler) uint32 {
	return uint32(len(epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()))
}

func (e *bootStrapShardRequester) createRequestHandler() (process.RequestHandler, error) {
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
	requestersFactory, err := requesterscontainer.NewMetaRequestersContainerFactory(requestersContainerArgs)
	if err != nil {
		return nil, err
	}

	container, err := requestersFactory.Create()
	if err != nil {
		return nil, err
	}

	err = requestersFactory.AddShardTrieNodeRequesters(container)
	if err != nil {
		return nil, err
	}

	finder, err := containers.NewRequestersFinder(container, e.shardCoordinator)
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := cache.NewTimeCache(timeBetweenRequests)
	return requestHandlers.NewResolverRequestHandler(
		finder,
		requestedItemsHandler,
		e.whiteListHandler,
		maxToRequest,
		core.MetachainShardId,
		timeBetweenRequests,
	)
}

func (e *bootStrapShardRequester) createResolversContainer() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(e.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	storageService := disabled.NewChainStorer()

	payloadValidator, err := validator.NewPeerAuthenticationPayloadValidator(e.generalConfig.HeartbeatV2.HeartbeatExpiryTimespanInSec)
	if err != nil {
		return err
	}

	// TODO - create a dedicated request handler to be used when fetching required data with the correct shard coordinator
	//  this one should only be used before determining the correct shard where the node should reside
	log.Debug("epochStartBootstrap.createRequestHandler", "shard", e.shardCoordinator.SelfId())
	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:                    e.shardCoordinator,
		MainMessenger:                       e.mainMessenger,
		FullArchiveMessenger:                e.fullArchiveMessenger,
		Store:                               storageService,
		Marshalizer:                         e.coreComponentsHolder.InternalMarshalizer(),
		DataPools:                           e.dataPool,
		Uint64ByteSliceConverter:            uint64ByteSlice.NewBigEndianConverter(),
		NumConcurrentResolvingJobs:          10,
		NumConcurrentResolvingTrieNodesJobs: 3,
		DataPacker:                          dataPacker,
		TriesContainer:                      e.trieContainer,
		SizeCheckDelta:                      0,
		InputAntifloodHandler:               disabled.NewAntiFloodHandler(),
		OutputAntifloodHandler:              disabled.NewAntiFloodHandler(),
		MainPreferredPeersHolder:            disabled.NewPreferredPeersHolder(),
		FullArchivePreferredPeersHolder:     disabled.NewPreferredPeersHolder(),
		PayloadValidator:                    payloadValidator,
	}
	resolverFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerArgs)
	if err != nil {
		return err
	}

	container, err := resolverFactory.Create()
	if err != nil {
		return err
	}

	return resolverFactory.AddShardTrieNodeResolvers(container)
}

func (e *bootStrapShardRequester) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)
	shardIds := make([]uint32, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)

	for _, epochStartData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		hashesToRequest = append(hashesToRequest, epochStartData.GetHeaderHash())
		shardIds = append(shardIds, epochStartData.GetShardID())
	}

	if meta.GetEpoch() > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.MetachainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
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
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}

func (e *bootStrapShardRequester) syncHeadersFromStorage(
	meta data.MetaHeaderHandler,
	syncingShardID uint32,
	importDBTargetShardID uint32,
	timeToWaitForRequestedData time.Duration,
) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)
	shardIds := make([]uint32, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)

	for _, epochStartData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		shouldSkipHeaderFetch := epochStartData.GetShardID() != syncingShardID &&
			importDBTargetShardID != core.MetachainShardId
		if shouldSkipHeaderFetch {
			continue
		}

		hashesToRequest = append(hashesToRequest, epochStartData.GetHeaderHash())
		shardIds = append(shardIds, epochStartData.GetShardID())
	}

	if meta.GetEpoch() > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.MetachainShardId)
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
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}
