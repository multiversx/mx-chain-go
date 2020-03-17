package storagehandler

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/structs"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
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
	shardCoordinator sharding.Coordinator,
	pathManagerHandler storage.PathManagerHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	currentEpoch uint32,
) (*metaStorageHandler, error) {
	epochStartNotifier := &disabled.EpochStartNotifier{}
	storageFactory, err := factory.NewStorageServiceFactory(
		&generalConfig,
		shardCoordinator,
		pathManagerHandler,
		epochStartNotifier,
		currentEpoch,
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
	}

	return &metaStorageHandler{baseStorageHandler: base}, nil
}

// SaveDataToStorage will save the fetched data to storage so it will be used by the storage bootstrap component
func (msh *metaStorageHandler) SaveDataToStorage(components structs.ComponentsNeededForBootstrap) error {
	// TODO: here we should save all needed data

	defer func() {
		err := msh.storageService.CloseAll()
		if err != nil {
			log.Debug("error while closing storers", "error", err)
		}
	}()

	bootStorer := msh.storageService.GetStorer(dataRetriever.BootstrapUnit)

	lastHeader, err := msh.getAndSaveLastHeader(components.EpochStartMetaBlock)
	if err != nil {
		return err
	}

	miniBlocks, err := msh.getAndSavePendingMiniBlocks(components.PendingMiniBlocks)
	if err != nil {
		return err
	}

	triggerConfigKey, err := msh.getAndSaveTriggerRegistry(components)
	if err != nil {
		return err
	}

	nodesCoordinatorConfigKey, err := msh.getAndSaveNodesCoordinatorKey(components.EpochStartMetaBlock)
	if err != nil {
		return err
	}

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                lastHeader,                                         // meta - epoch start metablock ; shard - shard header
		LastCrossNotarizedHeaders: nil,                                                // lastFinalizedMetaBlock + firstPendingMetaBlock
		LastSelfNotarizedHeaders:  []bootstrapStorage.BootstrapHeaderInfo{lastHeader}, // meta - epoch start metablock , shard: shard header
		ProcessedMiniBlocks:       nil,                                                // first pending metablock and pending miniblocks - difference between them
		// (shard - only shard ; meta - possible not to fill at all)
		PendingMiniBlocks:          miniBlocks,                // pending miniblocks
		NodesCoordinatorConfigKey:  nodesCoordinatorConfigKey, // wait for radu's component
		EpochStartTriggerConfigKey: triggerConfigKey,          // metachain/shard trigger registery
		HighestFinalBlockNonce:     lastHeader.Nonce,          //
		LastRound:                  int64(components.EpochStartMetaBlock.Round),
	}
	bootStrapDataBytes, err := msh.marshalizer.Marshal(&bootStrapData)
	if err != nil {
		return err
	}
	err = bootStorer.Put([]byte(highestRoundFromBootStorage), bootStrapDataBytes)
	if err != nil {
		return err
	}
	log.Info("saved bootstrap data to storage")
	return nil
}

func (msh *metaStorageHandler) getAndSaveLastHeader(metaBlock *block.MetaBlock) (bootstrapStorage.BootstrapHeaderInfo, error) {
	lastHeaderHash, err := core.CalculateHash(msh.marshalizer, msh.hasher, metaBlock)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	//metaBlock.

	lastHeaderBytes, err := msh.marshalizer.Marshal(metaBlock)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	err = msh.storageService.GetStorer(dataRetriever.MetaBlockUnit).Put(lastHeaderHash, lastHeaderBytes)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	bootstrapHdrInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MetachainShardId,
		Nonce:   metaBlock.Nonce,
		Hash:    lastHeaderHash,
	}

	return bootstrapHdrInfo, nil
}

func (msh *metaStorageHandler) getAndSaveTriggerRegistry(components structs.ComponentsNeededForBootstrap) ([]byte, error) {
	metaBlock := components.EpochStartMetaBlock
	hash, err := core.CalculateHash(msh.marshalizer, msh.hasher, metaBlock)
	if err != nil {
		return nil, err
	}

	triggerReg := metachain.TriggerRegistry{
		Epoch:                       metaBlock.Epoch,
		CurrentRound:                metaBlock.Round,
		EpochFinalityAttestingRound: metaBlock.Round,
		CurrEpochStartRound:         metaBlock.Round,
		PrevEpochStartRound:         components.PreviousEpochStartMetaBlock.Round,
		EpochStartMetaHash:          hash,
		EpochStartMeta:              metaBlock,
	}

	trigStateKey := fmt.Sprintf("initial_value_epoch%d", metaBlock.Epoch)
	key := []byte(triggerRegistrykeyPrefix + trigStateKey)

	triggerRegBytes, err := json.Marshal(&triggerReg)
	if err != nil {
		return nil, err
	}

	errPut := msh.storageService.GetStorer(dataRetriever.BootstrapUnit).Put(key, triggerRegBytes)
	if errPut != nil {
		return nil, errPut
	}

	return key, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (msh *metaStorageHandler) IsInterfaceNil() bool {
	return msh == nil
}
