package bootstrap

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/pendingMb"
	"github.com/multiversx/mx-chain-go/storage/factory"
)

type metaStorageHandler struct {
	*baseStorageHandler
}

// NewMetaStorageHandler will return a new instance of metaStorageHandler
func NewMetaStorageHandler(args StorageHandlerArgs) (*metaStorageHandler, error) {
	err := checkNilArgs(args)
	if err != nil {
		return nil, err
	}

	epochStartNotifier := &disabled.EpochStartNotifier{}
	storageFactory, err := factory.NewStorageServiceFactory(
		factory.StorageServiceFactoryArgs{
			Config:                        args.GeneralConfig,
			PrefsConfig:                   args.PreferencesConfig,
			ShardCoordinator:              args.ShardCoordinator,
			PathManager:                   args.PathManagerHandler,
			EpochStartNotifier:            epochStartNotifier,
			NodeTypeProvider:              args.NodeTypeProvider,
			StorageType:                   factory.BootstrapStorageService,
			ManagedPeersHolder:            args.ManagedPeersHolder,
			CurrentEpoch:                  args.CurrentEpoch,
			CreateTrieEpochRootHashStorer: false,
			NodeProcessingMode:            args.NodeProcessingMode,
			RepopulateTokensSupplies:      false,
			StateStatsHandler:             args.StateStatsHandler,
		},
	)
	if err != nil {
		return nil, err
	}

	storageService, err := storageFactory.CreateForMeta()
	if err != nil {
		return nil, err
	}

	base := &baseStorageHandler{
		storageService:                  storageService,
		shardCoordinator:                args.ShardCoordinator,
		marshalizer:                     args.Marshaller,
		hasher:                          args.Hasher,
		currentEpoch:                    args.CurrentEpoch,
		uint64Converter:                 args.Uint64Converter,
		nodesCoordinatorRegistryFactory: args.NodesCoordinatorRegistryFactory,
		proofsPool:                      args.ProofsPool,
		enableEpochsHandler:             args.EnableEpochsHandler,
	}

	return &metaStorageHandler{baseStorageHandler: base}, nil
}

// CloseStorageService closes the containing storage service
func (msh *metaStorageHandler) CloseStorageService() {
	err := msh.storageService.CloseAll()
	if err != nil {
		log.Warn("error while closing storers", "error", err)
	}
}

// SaveDataToStorage will save the fetched data to storage, so it will be used by the storage bootstrap component
func (msh *metaStorageHandler) SaveDataToStorage(components *ComponentsNeededForBootstrap) error {
	bootStorer, err := msh.storageService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return err
	}

	lastHeader, err := msh.saveLastHeader(components.EpochStartMetaBlock)
	if err != nil {
		return err
	}

	err = msh.saveMetaHdrForEpochTrigger(components.EpochStartMetaBlock)
	if err != nil {
		return err
	}

	err = msh.saveMetaHdrForEpochTrigger(components.PreviousEpochStart)
	if err != nil {
		return err
	}

	_, err = msh.saveMetaHdrToStorage(components.PreviousEpochStart)
	if err != nil {
		return err
	}

	msh.saveMiniblocksFromComponents(components)

	miniBlocks, err := computePendingMiniBlocks(components)
	if err != nil {
		return err
	}

	triggerConfigKey, err := msh.saveTriggerRegistry(components)
	if err != nil {
		return err
	}

	nodesCoordinatorConfigKey, err := msh.saveNodesCoordinatorRegistry(components.EpochStartMetaBlock, components.NodesConfig)
	if err != nil {
		return err
	}

	lastCrossNotarizedHeader, err := msh.saveLastCrossNotarizedHeaders(components.EpochStartMetaBlock, components.Headers)
	if err != nil {
		return err
	}

	lastSelfNotarizedHeaders, err := msh.getLastSelfNotarizedHeaders(components.EpochStartMetaBlock, lastHeader, components.Headers)
	if err != nil {
		return err
	}

	err = msh.saveIntermediateMetaBlocksToStorage(components.EpochStartMetaBlock, components.Headers)
	if err != nil {
		return err
	}

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                 lastHeader,
		LastCrossNotarizedHeaders:  lastCrossNotarizedHeader,
		LastSelfNotarizedHeaders:   lastSelfNotarizedHeaders,
		ProcessedMiniBlocks:        []bootstrapStorage.MiniBlocksInMeta{},
		PendingMiniBlocks:          miniBlocks,
		NodesCoordinatorConfigKey:  nodesCoordinatorConfigKey,
		EpochStartTriggerConfigKey: triggerConfigKey,
		HighestFinalBlockNonce:     lastHeader.Nonce,
		LastRound:                  0,
	}
	bootStrapDataBytes, err := msh.marshalizer.Marshal(&bootStrapData)
	if err != nil {
		return err
	}

	roundToUseAsKey := int64(components.EpochStartMetaBlock.GetRound())
	roundNum := bootstrapStorage.RoundNum{Num: roundToUseAsKey}
	roundNumBytes, err := msh.marshalizer.Marshal(&roundNum)
	if err != nil {
		return err
	}

	err = bootStorer.Put([]byte(common.HighestRoundFromBootStorage), roundNumBytes)
	if err != nil {
		return err
	}
	key := []byte(strconv.FormatInt(roundToUseAsKey, 10))
	err = bootStorer.Put(key, bootStrapDataBytes)
	if err != nil {
		return err
	}

	log.Debug("saved bootstrap data to storage", "round", roundToUseAsKey)
	return nil
}

func (msh *metaStorageHandler) getLastSelfNotarizedHeaders(
	epochStartMeta data.MetaHeaderHandler,
	epochStartMetaBootstrapInfo bootstrapStorage.BootstrapHeaderInfo,
	syncedHeaders map[string]data.HeaderHandler,
) ([]bootstrapStorage.BootstrapHeaderInfo, error) {
	lastSelfNotarizedHeaders := []bootstrapStorage.BootstrapHeaderInfo{epochStartMetaBootstrapInfo}

	for _, epochStartData := range epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		selfNotarizedMetaHash, selfNotarizedMetaBlock, found := msh.getSelfNotarizedMetaForShard(epochStartData, syncedHeaders)
		if !found {
			continue
		}

		_, err := msh.saveMetaHdrToStorage(selfNotarizedMetaBlock)
		if err != nil {
			return nil, err
		}

		bootstrapHdrInfo := bootstrapStorage.BootstrapHeaderInfo{
			ShardId: epochStartData.GetShardID(),
			Epoch:   selfNotarizedMetaBlock.GetEpoch(),
			Nonce:   selfNotarizedMetaBlock.GetNonce(),
			Hash:    selfNotarizedMetaHash,
		}

		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, bootstrapHdrInfo)
	}

	return lastSelfNotarizedHeaders, nil
}

func (msh *metaStorageHandler) saveIntermediateMetaBlocksToStorage(
	epochStartMeta data.MetaHeaderHandler,
	syncedHeaders map[string]data.HeaderHandler,
) error {
	prevHash := epochStartMeta.GetPrevHash()

	for len(prevHash) > 0 {
		metaBlock, ok := syncedHeaders[string(prevHash)]
		if !ok {
			break
		}

		_, err := msh.saveMetaHdrToStorage(metaBlock)
		if err != nil {
			return err
		}

		if metaBlock.GetNonce() == 0 {
			break
		}

		prevHash = metaBlock.GetPrevHash()
	}

	return nil
}

// computePendingMiniBlocks replays the intermediate metablocks on a fresh handler
// to derive the correct pending state, matching the AddProcessedHeader sequence
// that a continuously-running node executes.
func computePendingMiniBlocks(
	components *ComponentsNeededForBootstrap,
) ([]bootstrapStorage.PendingMiniBlocksInfo, error) {
	handler, err := pendingMb.NewPendingMiniBlocks()
	if err != nil {
		return nil, err
	}

	if !check.IfNil(components.PreviousEpochStart) {
		err = handler.AddProcessedHeader(components.PreviousEpochStart)
		if err != nil {
			log.Debug("computePendingMiniBlocks: could not process previous epoch start", "error", err)
		}
	}

	intermediateBlocks := collectIntermediateMetaBlocks(
		components.EpochStartMetaBlock,
		components.PreviousEpochStart,
		components.Headers,
	)
	sort.Slice(intermediateBlocks, func(i, j int) bool {
		return intermediateBlocks[i].GetNonce() < intermediateBlocks[j].GetNonce()
	})

	for _, metaBlock := range intermediateBlocks {
		errProcess := handler.AddProcessedHeader(metaBlock)
		if errProcess != nil {
			log.Debug("computePendingMiniBlocks: could not process intermediate block",
				"nonce", metaBlock.GetNonce(), "error", errProcess)
		}
	}

	err = handler.AddProcessedHeader(components.EpochStartMetaBlock)
	if err != nil {
		return nil, fmt.Errorf("could not process epoch start meta block: %w", err)
	}

	numShards := uint32(len(components.EpochStartMetaBlock.GetEpochStartHandler().GetLastFinalizedHeaderHandlers()))
	pendingMbs := make([]bootstrapStorage.PendingMiniBlocksInfo, 0, numShards)
	for shardID := uint32(0); shardID < numShards; shardID++ {
		hashes := handler.GetPendingMiniBlocks(shardID)
		if len(hashes) == 0 {
			continue
		}
		pendingMbs = append(pendingMbs, bootstrapStorage.PendingMiniBlocksInfo{
			ShardID:          shardID,
			MiniBlocksHashes: hashes,
		})
	}

	log.Debug("computePendingMiniBlocks",
		"numIntermediateBlocks", len(intermediateBlocks),
		"numShardsWithPending", len(pendingMbs))

	return pendingMbs, nil
}

func collectIntermediateMetaBlocks(
	epochStartMeta data.MetaHeaderHandler,
	previousEpochStart data.MetaHeaderHandler,
	syncedHeaders map[string]data.HeaderHandler,
) []data.HeaderHandler {
	prevNonce := uint64(0)
	if !check.IfNil(previousEpochStart) {
		prevNonce = previousEpochStart.GetNonce()
	}

	blocks := make([]data.HeaderHandler, 0)
	prevHash := epochStartMeta.GetPrevHash()
	for len(prevHash) > 0 {
		meta, ok := syncedHeaders[string(prevHash)]
		if !ok {
			break
		}
		if meta.GetNonce() <= prevNonce {
			break
		}
		blocks = append(blocks, meta)
		prevHash = meta.GetPrevHash()
	}

	return blocks
}

func (msh *metaStorageHandler) getSelfNotarizedMetaForShard(
	epochStartData data.EpochStartShardDataHandler,
	syncedHeaders map[string]data.HeaderHandler,
) ([]byte, data.HeaderHandler, bool) {
	metaHash := epochStartData.GetFirstPendingMetaBlock()
	if len(metaHash) == 0 {
		return nil, nil, false
	}

	metaBlock, found := syncedHeaders[string(metaHash)]
	if !found {
		return nil, nil, false
	}

	return metaHash, metaBlock, true
}

func (msh *metaStorageHandler) saveLastCrossNotarizedHeaders(
	meta data.MetaHeaderHandler,
	mapHeaders map[string]data.HeaderHandler,
) ([]bootstrapStorage.BootstrapHeaderInfo, error) {
	crossNotarizedHdrs := make([]bootstrapStorage.BootstrapHeaderInfo, 0)
	for _, epochStartShardData := range meta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		crossNotarizedHdrs = append(crossNotarizedHdrs, bootstrapStorage.BootstrapHeaderInfo{
			ShardId: epochStartShardData.GetShardID(),
			Nonce:   epochStartShardData.GetNonce(),
			Hash:    epochStartShardData.GetHeaderHash(),
		})

		hdr, ok := mapHeaders[string(epochStartShardData.GetHeaderHash())]
		if !ok {
			return nil, epochStart.ErrMissingHeader
		}

		_, err := msh.saveShardHdrToStorage(hdr)
		if err != nil {
			return nil, err
		}
	}

	return crossNotarizedHdrs, nil
}

func (msh *metaStorageHandler) saveLastHeader(metaBlock data.HeaderHandler) (bootstrapStorage.BootstrapHeaderInfo, error) {
	lastHeaderHash, err := msh.saveMetaHdrToStorage(metaBlock)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	bootstrapHdrInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MetachainShardId,
		Epoch:   metaBlock.GetEpoch(),
		Nonce:   metaBlock.GetNonce(),
		Hash:    lastHeaderHash,
	}

	return bootstrapHdrInfo, nil
}

func (msh *metaStorageHandler) saveTriggerRegistry(components *ComponentsNeededForBootstrap) ([]byte, error) {
	metaBlock, ok := components.EpochStartMetaBlock.(*block.MetaBlock)
	if !ok {
		return nil, epochStart.ErrWrongTypeAssertion
	}

	hash, err := core.CalculateHash(msh.marshalizer, msh.hasher, metaBlock)
	if err != nil {
		return nil, err
	}

	triggerReg := block.MetaTriggerRegistry{
		Epoch:                       metaBlock.GetEpoch(),
		CurrentRound:                metaBlock.GetRound(),
		EpochFinalityAttestingRound: metaBlock.GetRound(),
		CurrEpochStartRound:         metaBlock.GetRound(),
		PrevEpochStartRound:         components.PreviousEpochStart.GetRound(),
		EpochStartMetaHash:          hash,
		EpochStartMeta:              metaBlock,
	}

	bootstrapKey := []byte(fmt.Sprint(metaBlock.GetRound()))
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), bootstrapKey...)

	triggerRegBytes, err := msh.marshalizer.Marshal(&triggerReg)
	if err != nil {
		return nil, err
	}

	bootstrapStorageUnit, err := msh.storageService.GetStorer(dataRetriever.BootstrapUnit)
	if err != nil {
		return nil, err
	}

	errPut := bootstrapStorageUnit.Put(trigInternalKey, triggerRegBytes)
	if errPut != nil {
		return nil, errPut
	}

	return bootstrapKey, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (msh *metaStorageHandler) IsInterfaceNil() bool {
	return msh == nil
}
