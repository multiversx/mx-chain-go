package bootstrap

import (
	"context"
	"fmt"
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
	bootStrapFactory "github.com/multiversx/mx-chain-go/epochStart/bootstrap/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/heartbeat/validator"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/trie/factory"
	updateSync "github.com/multiversx/mx-chain-go/update/sync"
)

type bootStrapShardProcessor struct {
	*epochStartBootstrap
}

func (bp *bootStrapShardProcessor) requestAndProcessForShard(peerMiniBlocks []*block.MiniBlock) error {
	epochStartData, err := bp.findSelfShardEpochStartData()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = bp.miniBlocksSyncer.SyncPendingMiniBlocks(epochStartData.GetPendingMiniBlockHeaderHandlers(), ctx)
	cancel()
	if err != nil {
		return err
	}

	pendingMiniBlocks, err := bp.miniBlocksSyncer.GetMiniBlocks()
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

	bp.headersSyncer.ClearFields()
	ctx, cancel = context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err = bp.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return err
	}

	neededHeaders, err := bp.headersSyncer.GetHeaders()
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: SyncMissingHeadersByHash")

	for hash, hdr := range neededHeaders {
		bp.syncedHeaders[hash] = hdr
	}

	shardNotarizedHeader, ok := bp.syncedHeaders[string(epochStartData.GetHeaderHash())].(data.ShardHeaderHandler)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	dts, err := bp.getDataToSync(
		epochStartData,
		shardNotarizedHeader,
	)
	if err != nil {
		return err
	}

	for hash, hdr := range dts.additionalHeaders {
		bp.syncedHeaders[hash] = hdr
	}

	argsStorageHandler := StorageHandlerArgs{
		GeneralConfig:                   bp.generalConfig,
		PreferencesConfig:               bp.prefsConfig,
		ShardCoordinator:                bp.shardCoordinator,
		PathManagerHandler:              bp.coreComponentsHolder.PathHandler(),
		Marshaller:                      bp.coreComponentsHolder.InternalMarshalizer(),
		Hasher:                          bp.coreComponentsHolder.Hasher(),
		CurrentEpoch:                    bp.baseData.lastEpoch,
		Uint64Converter:                 bp.coreComponentsHolder.Uint64ByteSliceConverter(),
		NodeTypeProvider:                bp.coreComponentsHolder.NodeTypeProvider(),
		NodesCoordinatorRegistryFactory: bp.nodesCoordinatorRegistryFactory,
		ManagedPeersHolder:              bp.cryptoComponentsHolder.ManagedPeersHolder(),
		NodeProcessingMode:              bp.nodeProcessingMode,
		StateStatsHandler:               bp.stateStatsHandler,
		AdditionalStorageServiceCreator: bp.runTypeComponents.AdditionalStorageServiceCreator(),
	}
	storageHandlerComponent, err := NewShardStorageHandler(argsStorageHandler)
	if err != nil {
		return err
	}

	defer storageHandlerComponent.CloseStorageService()

	bp.closeTrieComponents()
	triesContainer, trieStorageManagers, err := factory.CreateTriesComponentsForShardId(
		bp.generalConfig,
		bp.coreComponentsHolder,
		storageHandlerComponent.storageService,
		bp.stateStatsHandler,
	)
	if err != nil {
		return err
	}

	bp.trieContainer = triesContainer
	bp.trieStorageManagers = trieStorageManagers

	log.Debug("start in epoch bootstrap: started syncUserAccountsState", "rootHash", dts.rootHashToSync)
	err = bp.syncUserAccountsState(dts.rootHashToSync)
	if err != nil {
		return err
	}
	log.Debug("start in epoch bootstrap: syncUserAccountsState")

	components := &ComponentsNeededForBootstrap{
		EpochStartMetaBlock: bp.epochStartMeta,
		PreviousEpochStart:  bp.prevEpochStartMeta,
		ShardHeader:         dts.ownShardHdr,
		NodesConfig:         bp.nodesConfig,
		Headers:             bp.syncedHeaders,
		ShardCoordinator:    bp.shardCoordinator,
		PendingMiniBlocks:   pendingMiniBlocks,
		PeerMiniBlocks:      peerMiniBlocks,
	}

	errSavingToStorage := storageHandlerComponent.SaveDataToStorage(components, shardNotarizedHeader, dts.withScheduled, dts.miniBlocks)
	if errSavingToStorage != nil {
		return errSavingToStorage
	}

	return nil
}

func (bp *bootStrapShardProcessor) computeNumShards(epochStartMeta data.MetaHeaderHandler) uint32 {
	return uint32(len(epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()))
}

func (bp *bootStrapShardProcessor) createRequestHandler() (process.RequestHandler, error) {
	requestersContainerArgs := requesterscontainer.FactoryArgs{
		RequesterConfig:                 bp.generalConfig.Requesters,
		ShardCoordinator:                bp.shardCoordinator,
		MainMessenger:                   bp.mainMessenger,
		FullArchiveMessenger:            bp.fullArchiveMessenger,
		Marshaller:                      bp.coreComponentsHolder.InternalMarshalizer(),
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

	finder, err := containers.NewRequestersFinder(container, bp.shardCoordinator)
	if err != nil {
		return nil, err
	}

	requestedItemsHandler := cache.NewTimeCache(timeBetweenRequests)
	return requestHandlers.NewResolverRequestHandler(
		finder,
		requestedItemsHandler,
		bp.whiteListHandler,
		maxToRequest,
		core.MetachainShardId,
		timeBetweenRequests,
	)
}

func (bp *bootStrapShardProcessor) createResolversContainer() error {
	dataPacker, err := partitioning.NewSimpleDataPacker(bp.coreComponentsHolder.InternalMarshalizer())
	if err != nil {
		return err
	}

	storageService := disabled.NewChainStorer()

	payloadValidator, err := validator.NewPeerAuthenticationPayloadValidator(bp.generalConfig.HeartbeatV2.HeartbeatExpiryTimespanInSec)
	if err != nil {
		return err
	}

	// TODO - create a dedicated request handler to be used when fetching required data with the correct shard coordinator
	//  this one should only be used before determining the correct shard where the node should reside
	log.Debug("epochStartBootstrap.createRequestHandler", "shard", bp.shardCoordinator.SelfId())
	resolversContainerArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:                    bp.shardCoordinator,
		MainMessenger:                       bp.mainMessenger,
		FullArchiveMessenger:                bp.fullArchiveMessenger,
		Store:                               storageService,
		Marshalizer:                         bp.coreComponentsHolder.InternalMarshalizer(),
		DataPools:                           bp.dataPool,
		Uint64ByteSliceConverter:            uint64ByteSlice.NewBigEndianConverter(),
		NumConcurrentResolvingJobs:          10, // TODO: We need to take this from config
		NumConcurrentResolvingTrieNodesJobs: 3,
		DataPacker:                          dataPacker,
		TriesContainer:                      bp.trieContainer,
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

func (bp *bootStrapShardProcessor) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)
	shardIds := make([]uint32, 0, len(meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers())+1)

	for _, epochStartData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		hashesToRequest = append(hashesToRequest, epochStartData.GetHeaderHash())
		shardIds = append(shardIds, epochStartData.GetShardID())
	}

	if meta.GetEpoch() > bp.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.MetachainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := bp.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := bp.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == bp.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}

func (bp *bootStrapShardProcessor) syncHeadersFromStorage(
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

	if meta.GetEpoch() > bp.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.MetachainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToWaitForRequestedData)
	err := bp.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := bp.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == bp.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.MetaBlock{}
	}

	return syncedHeaders, nil
}

func (bp *bootStrapShardProcessor) processNodesConfigFromStorage(pubKey []byte, importDBTargetShardID uint32) (nodesCoordinator.NodesCoordinatorRegistryHandler, uint32, error) {
	var err error
	shardId := bp.destinationShardAsObserver
	if shardId > bp.baseData.numberOfShards && shardId != core.MetachainShardId {
		shardId = bp.genesisShardCoordinator.SelfId()
	}
	argsNewValidatorStatusSyncers := ArgsNewSyncValidatorStatus{
		DataPool:                         bp.dataPool,
		Marshalizer:                      bp.coreComponentsHolder.InternalMarshalizer(),
		RequestHandler:                   bp.requestHandler,
		ChanceComputer:                   bp.rater,
		GenesisNodesConfig:               bp.genesisNodesConfig,
		NodeShuffler:                     bp.nodeShuffler,
		Hasher:                           bp.coreComponentsHolder.Hasher(),
		PubKey:                           pubKey,
		ShardIdAsObserver:                shardId,
		ChanNodeStop:                     bp.coreComponentsHolder.ChanStopNodeProcess(),
		NodeTypeProvider:                 bp.coreComponentsHolder.NodeTypeProvider(),
		IsFullArchive:                    bp.prefsConfig.FullArchive,
		EnableEpochsHandler:              bp.coreComponentsHolder.EnableEpochsHandler(),
		NodesCoordinatorRegistryFactory:  bp.nodesCoordinatorRegistryFactory,
		NodesCoordinatorWithRaterFactory: bp.runTypeComponents.NodesCoordinatorWithRaterCreator(),
	}
	bp.nodesConfigHandler, err = NewSyncValidatorStatus(argsNewValidatorStatusSyncers)
	if err != nil {
		return nil, 0, err
	}

	clonedHeader := bp.epochStartMeta.ShallowClone()
	clonedEpochStartMeta, ok := clonedHeader.(*block.MetaBlock)
	if !ok {
		return nil, 0, fmt.Errorf("%w while trying to assert clonedHeader to *block.MetaBlock", epochStart.ErrWrongTypeAssertion)
	}
	err = bp.applyCurrentShardIDOnMiniblocksCopy(clonedEpochStartMeta, importDBTargetShardID)
	if err != nil {
		return nil, 0, err
	}

	clonedHeader = bp.prevEpochStartMeta.ShallowClone()
	clonedPrevEpochStartMeta, ok := clonedHeader.(*block.MetaBlock)
	if !ok {
		return nil, 0, fmt.Errorf("%w while trying to assert prevClonedHeader to *block.MetaBlock", epochStart.ErrWrongTypeAssertion)
	}

	err = bp.applyCurrentShardIDOnMiniblocksCopy(clonedPrevEpochStartMeta, importDBTargetShardID)
	if err != nil {
		return nil, 0, err
	}

	// no need to save the peers miniblocks here as they were already fetched from the DB
	nodesConfig, shardId, _, err := bp.nodesConfigHandler.NodesConfigFromMetaBlock(clonedEpochStartMeta, clonedPrevEpochStartMeta)
	return nodesConfig, bp.applyShardIDAsObserverIfNeeded(shardId), err
}

// applyCurrentShardIDOnMiniblocksCopy will alter the fetched metablocks making the sender shard ID for each miniblock
// header to  be exactly the shard ID used in the import-db process. This is necessary as to allow the miniblocks to be requested
// on the available resolver and should be called only from this storage-base bootstrap instance.
// This method also copies the MiniBlockHeaders slice pointer. Otherwise, the node will end up stating
// "start of epoch metablock mismatch"
func (bp *bootStrapShardProcessor) applyCurrentShardIDOnMiniblocksCopy(metablock data.HeaderHandler, importDBTargetShardID uint32) error {
	originalMiniblocksHeaders := metablock.GetMiniBlockHeaderHandlers()
	mbsHeaderHandlersToSet := make([]data.MiniBlockHeaderHandler, 0, len(originalMiniblocksHeaders))
	var err error

	for i := range originalMiniblocksHeaders {
		mb := originalMiniblocksHeaders[i].ShallowClone()
		err = mb.SetSenderShardID(importDBTargetShardID) // it is safe to modify here as mb is a shallow clone
		if err != nil {
			return err
		}

		mbsHeaderHandlersToSet = append(mbsHeaderHandlersToSet, mb)
	}

	err = metablock.SetMiniBlockHeaderHandlers(mbsHeaderHandlersToSet)
	return err
}

func (bp *bootStrapShardProcessor) createEpochStartMetaSyncer() (epochStart.StartOfEpochMetaSyncer, error) {
	epochStartConfig := bp.generalConfig.EpochStartConfig
	metaBlockProcessor, err := NewEpochStartMetaBlockProcessor(
		bp.mainMessenger,
		bp.requestHandler,
		bp.coreComponentsHolder.InternalMarshalizer(),
		bp.coreComponentsHolder.Hasher(),
		thresholdForConsideringMetaBlockCorrect,
		epochStartConfig.MinNumConnectedPeersToStart,
		epochStartConfig.MinNumOfPeersToConsiderBlockValid,
	)
	if err != nil {
		return nil, err
	}

	argsEpochStartSyncer := ArgsNewEpochStartMetaSyncer{
		CoreComponentsHolder:    bp.coreComponentsHolder,
		CryptoComponentsHolder:  bp.cryptoComponentsHolder,
		RequestHandler:          bp.requestHandler,
		Messenger:               bp.mainMessenger,
		ShardCoordinator:        bp.shardCoordinator,
		EconomicsData:           bp.economicsData,
		WhitelistHandler:        bp.whiteListHandler,
		StartInEpochConfig:      epochStartConfig,
		HeaderIntegrityVerifier: bp.headerIntegrityVerifier,
		MetaBlockProcessor:      metaBlockProcessor,
	}

	return NewEpochStartMetaSyncer(argsEpochStartSyncer)
}

func (bp *bootStrapShardProcessor) createStorageEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (epochStart.StartOfEpochMetaSyncer, error) {
	return NewEpochStartMetaSyncer(args)
}

func (bp *bootStrapShardProcessor) createEpochStartInterceptorsContainers(args bootStrapFactory.ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	return bootStrapFactory.NewEpochStartInterceptorsContainer(args)
}

func (bp *bootStrapShardProcessor) createCrossHeaderRequester() (updateSync.CrossHeaderRequester, error) {
	return updateSync.NewMetaHeaderRequester(bp.requestHandler)
}
