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
	"github.com/multiversx/mx-chain-go/process"
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

	epochStartMeta := components.EpochStartMetaBlock
	var lastSelfNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo
	if !epochStartMeta.IsHeaderV3() {
		lastSelfNotarizedHeaders = []bootstrapStorage.BootstrapHeaderInfo{lastHeader}
	} else {
		for _, epochStartData := range epochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
			lastFinishedMetaBlockHash := epochStartData.GetLastFinishedMetaBlock()
			lastFinishedMetaBlock, ok := components.Headers[string(lastFinishedMetaBlockHash)]
			if !ok {
				log.Error("should be able to find header",
					"hash", lastFinishedMetaBlockHash,
					"shardID", epochStartData.GetShardID(),
				)
				return epochStart.ErrMissingHeader
			}

			shardHeaderHash := epochStartData.GetHeaderHash()
			shardHeader, ok := components.Headers[string(shardHeaderHash)]
			if !ok {
				log.Error("should be able to find header",
					"hash", lastFinishedMetaBlockHash,
					"shardID", epochStartData.GetShardID(),
				)
				return epochStart.ErrMissingHeader
			}

			shardHeaderHandler, ok := shardHeader.(data.ShardHeaderHandler)
			if !ok {
				log.Error("epoch start data shard header",
					"error", process.ErrWrongTypeAssertion,
				)
				return epochStart.ErrWrongTypeAssertion
			}

			if len(shardHeaderHandler.GetMetaBlockHashes()) > 0 {
				metaHash := shardHeaderHandler.GetMetaBlockHashes()[0]
				metaHeader0, ok := components.Headers[string(metaHash)]
				if !ok {
					log.Error("should be able to find header",
						"hash", metaHash,
						"shardID", epochStartData.GetShardID(),
					)
					return epochStart.ErrMissingHeader
				}

				prevHash := metaHeader0.GetPrevHash()

				metaHeader1, ok := components.Headers[string(prevHash)]
				if !ok {
					log.Error("should be able to find header",
						"hash", prevHash,
						"shardID", epochStartData.GetShardID(),
					)
					return epochStart.ErrMissingHeader
				}

				lastFinishedMetaBlockHash = prevHash
				lastFinishedMetaBlock = metaHeader1
			}

			bootstrapHdrInfo := bootstrapStorage.BootstrapHeaderInfo{
				ShardId: epochStartData.GetShardID(),
				Epoch:   lastFinishedMetaBlock.GetEpoch(),
				Nonce:   lastFinishedMetaBlock.GetNonce(),
				Hash:    lastFinishedMetaBlockHash,
			}

			log.Debug("SaveDataToStorage",
				"GetLastFinishedMetaBlock: hash", lastFinishedMetaBlockHash,
				"shard", epochStartData.GetShardID(),
			)

			lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, bootstrapHdrInfo)
		}

		lastExecRes, err := common.GetLastBaseExecutionResultHandler(epochStartMeta)
		if err != nil {
			return err
		}

		lastExecMetaHeader, ok := components.Headers[string(lastExecRes.GetHeaderHash())]
		if !ok {
			log.Error("should be able to find last exec meta header",
				"hash", lastExecRes.GetHeaderHash(),
			)
			return epochStart.ErrMissingHeader
		}

		bootstrapHdrInfoMeta := bootstrapStorage.BootstrapHeaderInfo{
			ShardId: core.MetachainShardId,
			Epoch:   lastExecMetaHeader.GetEpoch(),
			Nonce:   lastExecMetaHeader.GetNonce(),
			Hash:    lastExecRes.GetHeaderHash(),
		}

		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, bootstrapHdrInfoMeta)
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

func (msh *metaStorageHandler) saveEpochStartMetaHdrs(components *ComponentsNeededForBootstrap) error {
	for hash, hdr := range components.Headers {
		isForCurrentShard := hdr.GetShardID() == msh.shardCoordinator.SelfId()
		if !isForCurrentShard {
			log.Debug("not for curr shard",
				"hash", hash,
				"shard", hdr.GetShardID(),
				"nonce", hdr.GetNonce(),
			)
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

		log.Debug("saved header",
			"hash", hash,
			"shard", hdr.GetShardID(),
			"nonce", hdr.GetNonce(),
		)
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
