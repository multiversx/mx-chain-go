package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
)

func (e *epochStartBootstrap) initializeFromLocalStorage() {
	latestData, errNotCritical := storageFactory.FindLatestDataFromStorage(
		e.generalConfig,
		e.marshalizer,
		e.workingDir,
		e.genesisNodesConfig.GetChainId(),
		e.defaultDBPath,
		e.defaultEpochString,
		e.defaultShardString,
	)
	if errNotCritical != nil {
		e.baseData.storageExists = false
		log.Debug("no epoch db found in storage", "error", errNotCritical.Error())
	} else {
		e.baseData.storageExists = true
		e.baseData.lastEpoch = latestData.Epoch
		e.baseData.shardId = latestData.ShardID
		e.baseData.lastRound = latestData.LastRound
		e.baseData.epochStartRound = latestData.EpochStartRound
		log.Debug("got last data from storage",
			"epoch", e.baseData.lastEpoch,
			"last round", e.baseData.lastRound,
			"last shard ID", e.baseData.shardId,
			"epoch start Round", e.baseData.epochStartRound)
	}
}

func (e *epochStartBootstrap) prepareEpochFromStorage() (Parameters, error) {
	args := storageFactory.ArgsNewOpenStorageUnits{
		GeneralConfig:      e.generalConfig,
		Marshalizer:        e.marshalizer,
		WorkingDir:         e.workingDir,
		ChainID:            e.genesisNodesConfig.GetChainId(),
		DefaultDBPath:      e.defaultDBPath,
		DefaultEpochString: e.defaultEpochString,
		DefaultShardString: e.defaultShardString,
	}
	openStorageHandler, err := storageFactory.NewStorageUnitOpenHandler(args)
	if err != nil {
		return Parameters{}, err
	}

	unitsToOpen := []string{e.generalConfig.BootstrapStorage.DB.FilePath, e.generalConfig.MetaBlockStorage.DB.FilePath}

	storageUnits, err := openStorageHandler.OpenStorageUnits(unitsToOpen)
	defer func() {
		for _, storer := range storageUnits {
			errClose := storer.Close()
			log.LogIfError(errClose)
		}
	}()

	if err != nil || len(storageUnits) != len(unitsToOpen) {
		return Parameters{}, err
	}

	_, e.nodesConfig, err = e.getLastBootstrapData(storageUnits[0])
	if err != nil {
		return Parameters{}, err
	}

	pubKey, err := e.publicKey.ToByteArray()
	if err != nil {
		return Parameters{}, err
	}

	e.epochStartMeta, err = e.getEpochStartMetaFromStorage(storageUnits[1])
	if err != nil {
		return Parameters{}, err
	}
	e.baseData.numberOfShards = uint32(len(e.epochStartMeta.EpochStart.LastFinalizedHeaders))

	if !e.checkIfShuffledOut(pubKey, e.nodesConfig) {
		parameters := Parameters{
			Epoch:       e.baseData.lastEpoch,
			SelfShardId: e.baseData.shardId,
			NumOfShards: e.baseData.numberOfShards,
		}
		return parameters, nil
	}

	log.Debug("prepareEpochFromStorage for shuffled out")

	err = e.createSyncers()
	if err != nil {
		return Parameters{}, err
	}

	e.syncedHeaders, err = e.syncHeadersFrom(e.epochStartMeta)
	if err != nil {
		return Parameters{}, err
	}

	prevEpochStartMetaHash := e.epochStartMeta.EpochStart.Economics.PrevEpochStartHash
	prevEpochStartMeta, ok := e.syncedHeaders[string(prevEpochStartMetaHash)].(*block.MetaBlock)
	if !ok {
		return Parameters{}, epochStart.ErrWrongTypeAssertion
	}
	e.prevEpochStartMeta = prevEpochStartMeta

	e.shardCoordinator, err = sharding.NewMultiShardCoordinator(e.baseData.numberOfShards, e.baseData.shardId)
	if err != nil {
		return Parameters{}, err
	}

	if e.shardCoordinator.SelfId() != e.genesisShardCoordinator.SelfId() {
		err = e.createTriesForNewShardId(e.shardCoordinator.SelfId())
		if err != nil {
			return Parameters{}, err
		}
	}

	if e.shardCoordinator.SelfId() == core.MetachainShardId {
		err = e.requestAndProcessForMeta()
		if err != nil {
			return Parameters{}, err
		}
	} else {
		err = e.requestAndProcessForShard()
		if err != nil {
			return Parameters{}, err
		}
	}

	parameters := Parameters{
		Epoch:       e.baseData.lastEpoch,
		SelfShardId: e.shardCoordinator.SelfId(),
		NumOfShards: e.shardCoordinator.NumberOfShards(),
	}
	return parameters, nil
}

func (e *epochStartBootstrap) checkIfShuffledOut(
	pubKey []byte,
	nodesConfig *sharding.NodesCoordinatorRegistry,
) bool {
	epochConfig := nodesConfig.EpochsConfig[fmt.Sprint(e.baseData.lastEpoch)]
	shardIdForConfig := fmt.Sprint(e.baseData.shardId)

	for _, validator := range epochConfig.WaitingValidators[shardIdForConfig] {
		if bytes.Equal(pubKey, validator.PubKey) {
			return false
		}
	}

	for _, validator := range epochConfig.EligibleValidators[shardIdForConfig] {
		if bytes.Equal(pubKey, validator.PubKey) {
			return false
		}
	}

	return true
}

func (e *epochStartBootstrap) getLastBootstrapData(storer storage.Storer) (*bootstrapStorage.BootstrapData, *sharding.NodesCoordinatorRegistry, error) {
	bootStorer, err := bootstrapStorage.NewBootstrapStorer(e.marshalizer, storer)
	if err != nil {
		return nil, nil, err
	}

	highestRound := bootStorer.GetHighestRound()
	bootstrapData, err := bootStorer.Get(highestRound)
	if err != nil {
		return nil, nil, err
	}

	ncInternalkey := append([]byte(core.NodesCoordinatorRegistryKeyPrefix), bootstrapData.NodesCoordinatorConfigKey...)
	data, err := storer.SearchFirst(ncInternalkey)
	if err != nil {
		log.Debug("this should not error - bootstrapData - NodesCoordinatorRegistryKey - getLastBootstrapData", "error", err)
		return nil, nil, err
	}

	config := &sharding.NodesCoordinatorRegistry{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, nil, err
	}

	return &bootstrapData, config, nil
}

func (e *epochStartBootstrap) getEpochStartMetaFromStorage(storer storage.Storer) (*block.MetaBlock, error) {
	epochIdentifier := core.EpochStartIdentifier(e.baseData.lastEpoch)
	data, err := storer.SearchFirst([]byte(epochIdentifier))
	if err != nil {
		log.Debug("getEpochStartMetaFromStorage", "key", epochIdentifier, "error", err)
		return nil, err
	}

	metaBlock := &block.MetaBlock{}
	err = e.marshalizer.Unmarshal(metaBlock, data)
	if err != nil {
		return nil, err
	}

	return metaBlock, nil
}
