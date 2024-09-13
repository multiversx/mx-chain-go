package bootstrap

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
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

	// TODO: MX-15748 Analyse this
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

	dts, err := e.getDataToSyncMethod(
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
