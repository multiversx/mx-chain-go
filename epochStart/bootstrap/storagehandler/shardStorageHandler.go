package storagehandler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/structs"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
)

type shardStorageHandler struct {
	*baseStorageHandler
}

// NewShardStorageHandler will return a new instance of shardStorageHandler
func NewShardStorageHandler(
	generalConfig config.Config,
	shardCoordinator sharding.Coordinator,
	pathManagerHandler storage.PathManagerHandler,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	currentEpoch uint32,
) (*shardStorageHandler, error) {
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
	storageService, err := storageFactory.CreateForShard()
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

	return &shardStorageHandler{baseStorageHandler: base}, nil
}

// SaveDataToStorage will save the fetched data to storage so it will be used by the storage bootstrap component
func (ssh *shardStorageHandler) SaveDataToStorage(components structs.ComponentsNeededForBootstrap) error {
	// TODO: here we should save all needed data

	defer func() {
		err := ssh.storageService.CloseAll()
		if err != nil {
			log.Debug("error while closing storers", "error", err)
		}
	}()

	bootStorer := ssh.storageService.GetStorer(dataRetriever.BootstrapUnit)

	lastHeader, err := ssh.getAndSaveLastHeader(components.ShardHeader)
	if err != nil {
		return err
	}

	miniBlocks, err := ssh.getAndSavePendingMiniBlocks(components.PendingMiniBlocks)
	if err != nil {
		return err
	}

	triggerConfigKey, err := ssh.getAndSaveTriggerRegistry(components)
	if err != nil {
		return err
	}

	nodesCoordinatorConfigKey, err := ssh.getAndSaveNodesCoordinatorKey(components.EpochStartMetaBlock)
	if err != nil {
		return err
	}

	bootStrapData := bootstrapStorage.BootstrapData{
		LastHeader:                lastHeader,                                         // meta - epoch start metablock ; shard - shard header
		LastCrossNotarizedHeaders: nil,                                                // lastFinalizedMetaBlock + firstPendingMetaBlock
		LastSelfNotarizedHeaders:  []bootstrapStorage.BootstrapHeaderInfo{lastHeader}, // meta - epoch start metablock , shard: shard header
		ProcessedMiniBlocks:       nil,                                                // first pending metablock si pending miniblocks - difference between them
		// (shard - only shard ; meta - possible not to fill at all)
		PendingMiniBlocks:          miniBlocks,                // pending miniblocks
		NodesCoordinatorConfigKey:  nodesCoordinatorConfigKey, // wait for radu's component
		EpochStartTriggerConfigKey: triggerConfigKey,          // metachain/shard trigger registery
		HighestFinalBlockNonce:     0,                         //
		LastRound:                  int64(components.ShardHeader.Round),
	}
	bootStrapDataBytes, err := ssh.marshalizer.Marshal(&bootStrapData)
	if err != nil {
		return err
	}
	roundToUseAsKey := int64(components.ShardHeader.Round + 2) // TODO: change this. added 2 in order to skip
	// equality check between round and LastRound from bootstrap from storage component
	roundNum := bootstrapStorage.RoundNum{Num: roundToUseAsKey}
	roundNumBytes, err := ssh.marshalizer.Marshal(&roundNum)
	if err != nil {
		return err
	}

	err = bootStorer.Put([]byte(highestRoundFromBootStorage), roundNumBytes)
	if err != nil {
		return err
	}

	log.Info("saved bootstrap data to storage")
	key := []byte(strconv.FormatInt(roundToUseAsKey, 10))
	err = bootStorer.Put(key, bootStrapDataBytes)
	if err != nil {
		return err
	}

	return nil
}

func (ssh *shardStorageHandler) getAndSaveLastHeader(shardHeader *block.Header) (bootstrapStorage.BootstrapHeaderInfo, error) {
	lastHeaderHash, err := core.CalculateHash(ssh.marshalizer, ssh.hasher, shardHeader)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	lastHeaderBytes, err := ssh.marshalizer.Marshal(shardHeader)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	err = ssh.storageService.GetStorer(dataRetriever.BlockHeaderUnit).Put(lastHeaderHash, lastHeaderBytes)
	if err != nil {
		return bootstrapStorage.BootstrapHeaderInfo{}, err
	}

	bootstrapHdrInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: core.MetachainShardId,
		Nonce:   shardHeader.Nonce,
		Hash:    lastHeaderHash,
	}

	return bootstrapHdrInfo, nil
}

func (ssh *shardStorageHandler) getAndSaveTriggerRegistry(components structs.ComponentsNeededForBootstrap) ([]byte, error) {
	shardHeader := components.ShardHeader

	metaBlock := components.EpochStartMetaBlock
	metaBlockHash, err := core.CalculateHash(ssh.marshalizer, ssh.hasher, metaBlock)
	if err != nil {
		return nil, err
	}

	triggerReg := shardchain.TriggerRegistry{
		Epoch:                       shardHeader.Epoch,
		CurrentRoundIndex:           int64(shardHeader.Round),
		EpochStartRound:             shardHeader.Round,
		EpochMetaBlockHash:          metaBlockHash,
		IsEpochStart:                false,
		NewEpochHeaderReceived:      false,
		EpochFinalityAttestingRound: 0,
	}

	trigStateKey := fmt.Sprintf("initial_value_epoch%d", metaBlock.Epoch)
	key := []byte(triggerRegistrykeyPrefix + trigStateKey)

	triggerRegBytes, err := json.Marshal(&triggerReg)
	if err != nil {
		return nil, err
	}

	errPut := ssh.storageService.GetStorer(dataRetriever.BootstrapUnit).Put(key, triggerRegBytes)
	if errPut != nil {
		return nil, errPut
	}

	return key, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ssh *shardStorageHandler) IsInterfaceNil() bool {
	return ssh == nil
}
