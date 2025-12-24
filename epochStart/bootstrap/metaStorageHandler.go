package bootstrap

import (
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
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

	err = msh.saveEpochStartMetaHdrs(components)
	if err != nil {
		return err
	}

	msh.saveMiniblocksFromComponents(components)

	miniBlocks, err := msh.groupMiniBlocksByShard(components.PendingMiniBlocks)
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

	lastCrossNotarizedHeaders, err := msh.saveLastCrossNotarizedHeaders(components.EpochStartMetaBlock, components.Headers)
	if err != nil {
		return err
	}

	lastSelfNotarizedHeaders, err := msh.getLastSelfNotarizedHeaders(
		components.EpochStartMetaBlock,
		lastHeader,
		components.Headers,
	)
	if err != nil {
		return err
	}

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                 lastHeader,
		LastCrossNotarizedHeaders:  lastCrossNotarizedHeaders,
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
	var lastSelfNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo
	if !epochStartMeta.IsHeaderV3() {
		return []bootstrapStorage.BootstrapHeaderInfo{
			epochStartMetaBootstrapInfo,
		}, nil
	}

	for _, epochStartData := range epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {

		bootstrapHdrInfo, err := msh.getLastNotarizedBootstrapInfoForEpochStartData(epochStartData, syncedHeaders)
		if err != nil {
			return nil, err
		}

		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, bootstrapHdrInfo)
	}

	bootstrapHdrInfoMeta, err := msh.getLastMetaBootstrapInfo(epochStartMeta, syncedHeaders)
	if err != nil {
		return nil, err
	}

	lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, bootstrapHdrInfoMeta)

	return lastSelfNotarizedHeaders, nil
}

func (msh *metaStorageHandler) getLastMetaBootstrapInfo(
	epochStartMeta data.MetaHeaderHandler,
	syncedHeaders map[string]data.HeaderHandler,
) (bootstrapStorage.BootstrapHeaderInfo, error) {
	lastExecRes, err := common.GetLastBaseExecutionResultHandler(epochStartMeta)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	lastExecMetaHeader, ok := syncedHeaders[string(lastExecRes.GetHeaderHash())]
	if !ok {
		return bootstrapStorage.BootstrapHeaderInfo{}, epochStart.ErrMissingHeader
	}

	bootstrapHdrInfoMeta := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MetachainShardId,
		Epoch:   lastExecMetaHeader.GetEpoch(),
		Nonce:   lastExecMetaHeader.GetNonce(),
		Hash:    lastExecRes.GetHeaderHash(),
	}

	return bootstrapHdrInfoMeta, nil
}

func (msh *metaStorageHandler) getLastNotarizedBootstrapInfoForEpochStartData(
	epochStartData data.EpochStartShardDataHandler,
	syncedHeaders map[string]data.HeaderHandler,
) (bootstrapStorage.BootstrapHeaderInfo, error) {
	lastFinishedMetaBlockHash := epochStartData.GetLastFinishedMetaBlock()
	lastFinishedMetaBlock, ok := syncedHeaders[string(lastFinishedMetaBlockHash)]
	if !ok {
		return bootstrapStorage.BootstrapHeaderInfo{}, epochStart.ErrMissingHeader
	}

	shardHeaderHash := epochStartData.GetHeaderHash()
	shardHeader, ok := syncedHeaders[string(shardHeaderHash)]
	if !ok {
		return bootstrapStorage.BootstrapHeaderInfo{}, epochStart.ErrMissingHeader
	}

	shardHeaderHandler, ok := shardHeader.(data.ShardHeaderHandler)
	if !ok {
		return bootstrapStorage.BootstrapHeaderInfo{}, epochStart.ErrWrongTypeAssertion
	}

	if len(shardHeaderHandler.GetMetaBlockHashes()) <= 0 {
		return bootstrapStorage.BootstrapHeaderInfo{
			ShardId: epochStartData.GetShardID(),
			Epoch:   lastFinishedMetaBlock.GetEpoch(),
			Nonce:   lastFinishedMetaBlock.GetNonce(),
			Hash:    lastFinishedMetaBlockHash,
		}, nil
	}

	metaHash := shardHeaderHandler.GetMetaBlockHashes()[0]
	metaHeader, ok := syncedHeaders[string(metaHash)]
	if !ok {
		return bootstrapStorage.BootstrapHeaderInfo{}, epochStart.ErrMissingHeader
	}

	prevHash := metaHeader.GetPrevHash()

	prevMetaHeader, ok := syncedHeaders[string(prevHash)]
	if !ok {
		return bootstrapStorage.BootstrapHeaderInfo{}, epochStart.ErrMissingHeader
	}

	lastFinishedMetaBlockHash = prevHash
	lastFinishedMetaBlock = prevMetaHeader

	bootstrapHdrInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: epochStartData.GetShardID(),
		Epoch:   lastFinishedMetaBlock.GetEpoch(),
		Nonce:   lastFinishedMetaBlock.GetNonce(),
		Hash:    lastFinishedMetaBlockHash,
	}

	return bootstrapHdrInfo, nil
}

func (msh *metaStorageHandler) saveEpochStartMetaHdrs(components *ComponentsNeededForBootstrap) error {
	for _, hdr := range components.Headers {
		isForCurrentShard := hdr.GetShardID() == msh.shardCoordinator.SelfId()
		if !isForCurrentShard {
			_, err := msh.saveShardHdrToStorage(hdr)
			if err != nil {
				return err
			}

			continue
		}

		_, err := msh.saveMetaHdrToStorage(hdr)
		if err != nil {
			return err
		}
	}

	return nil
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
	metaBlock := components.EpochStartMetaBlock
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilMetaBlock
	}

	hash, err := core.CalculateHash(msh.marshalizer, msh.hasher, metaBlock)
	if err != nil {
		return nil, err
	}

	triggerReg := epochStart.CreateMetaRegistryHandler(metaBlock)
	_ = triggerReg.SetEpochStartMetaHeaderHandler(metaBlock)
	_ = triggerReg.SetEpoch(metaBlock.GetEpoch())
	_ = triggerReg.SetEpochStartMetaHash(hash)
	_ = triggerReg.SetCurrEpochStartRound(metaBlock.GetRound())
	_ = triggerReg.SetPrevEpochStartRound(components.PreviousEpochStart.GetRound())
	_ = triggerReg.SetEpochFinalityAttestingRound(metaBlock.GetRound())
	_ = triggerReg.SetEpochChangeProposed(false)

	bootstrapKey := []byte(fmt.Sprint(metaBlock.GetRound()))
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), bootstrapKey...)

	triggerRegBytes, err := msh.marshalizer.Marshal(triggerReg)
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
