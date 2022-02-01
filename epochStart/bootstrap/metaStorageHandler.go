package bootstrap

import (
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
)

type metaStorageHandler struct {
	*baseStorageHandler
}

// NewMetaStorageHandler will return a new instance of metaStorageHandler
func NewMetaStorageHandler(
	generalConfig config.Config,
	prefsConfig config.PreferencesConfig,
	shardCoordinator sharding.Coordinator,
	pathManagerHandler storage.PathManagerHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	currentEpoch uint32,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
	nodeTypeProvider NodeTypeProviderHandler,
) (*metaStorageHandler, error) {
	epochStartNotifier := &disabled.EpochStartNotifier{}
	storageFactory, err := factory.NewStorageServiceFactory(
		&generalConfig,
		&prefsConfig,
		shardCoordinator,
		pathManagerHandler,
		epochStartNotifier,
		nodeTypeProvider,
		currentEpoch,
		false,
	)
	if err != nil {
		return nil, err
	}

	storageService, err := storageFactory.CreateForMeta()
	if err != nil {
		return nil, err
	}

	base := &baseStorageHandler{
		storageService:   storageService,
		shardCoordinator: shardCoordinator,
		marshalizer:      marshalizer,
		hasher:           hasher,
		currentEpoch:     currentEpoch,
		uint64Converter:  uint64Converter,
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

// SaveDataToStorage will save the fetched data to storage so it will be used by the storage bootstrap component
func (msh *metaStorageHandler) SaveDataToStorage(components *ComponentsNeededForBootstrap) error {
	bootStorer := msh.storageService.GetStorer(dataRetriever.BootstrapUnit)

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

	lastCrossNotarizedHeader, err := msh.saveLastCrossNotarizedHeaders(components.EpochStartMetaBlock, components.Headers)
	if err != nil {
		return err
	}

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                 lastHeader,
		LastCrossNotarizedHeaders:  lastCrossNotarizedHeader,
		LastSelfNotarizedHeaders:   []bootstrapStorage.BootstrapHeaderInfo{lastHeader},
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

	bootstrapStorageUnit := msh.storageService.GetStorer(dataRetriever.BootstrapUnit)
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
