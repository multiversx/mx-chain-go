package sharding

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/common"
)

// LoadState loads the nodes coordinator state from the used boot storage
func (ihgs *indexHashedNodesCoordinator) LoadState(key []byte) error {
	return ihgs.baseLoadState(key)
}

func (ihgs *indexHashedNodesCoordinator) baseLoadState(key []byte) error {
	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

	ihgs.loadingFromDisk.Store(true)
	defer ihgs.loadingFromDisk.Store(false)

	data, err := ihgs.bootStorer.Get(ncInternalkey)
	if err != nil {
		return err
	}

	config, err := CreateNodesCoordinatorRegistry(ihgs.marshalizer, data)
	if err != nil {
		return err
	}

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = key
	ihgs.mutSavedStateKey.Unlock()

	ihgs.currentEpoch = config.GetCurrentEpoch()
	log.Debug("loaded nodes config", "current epoch", config.GetCurrentEpoch())

	nodesConfig, err := ihgs.registryToNodesCoordinator(config)
	if err != nil {
		return err
	}

	displayNodesConfigInfo(nodesConfig)
	ihgs.mutNodesConfig.Lock()
	ihgs.nodesConfig = nodesConfig
	ihgs.mutNodesConfig.Unlock()

	return nil
}

func displayNodesConfigInfo(config map[uint32]*epochNodesConfig) {
	for epoch, cfg := range config {
		log.Debug("restored config for",
			"epoch", epoch,
			"computed shard ID", cfg.shardID,
		)
	}
}

func (ihgs *indexHashedNodesCoordinator) saveState(key []byte) error {
	data, err := ihgs.getRegistryData()
	if err != nil {
		return err
	}

	ncInternalKey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)
	log.Debug("saving nodes coordinator config", "key", ncInternalKey)

	return ihgs.bootStorer.Put(ncInternalKey, data)
}

func (ihgs *indexHashedNodesCoordinator) getRegistryData() ([]byte, error) {
	var err error
	var data []byte

	if ihgs.flagStakingV4.IsSet() {
		registry := ihgs.nodesCoordinatorToRegistryWithAuction()
		data, err = ihgs.marshalizer.Marshal(registry)
	} else {
		registry := ihgs.nodesCoordinatorToOldRegistry()
		data, err = json.Marshal(registry)
	}

	return data, err
}

// NodesCoordinatorToRegistry will export the nodesCoordinator data to the registry
func (ihgs *indexHashedNodesCoordinator) NodesCoordinatorToRegistry() NodesCoordinatorRegistryHandler {
	if ihgs.flagStakingV4.IsSet() {
		return ihgs.nodesCoordinatorToRegistryWithAuction()
	}

	return ihgs.nodesCoordinatorToOldRegistry()
}

func (ihgs *indexHashedNodesCoordinator) nodesCoordinatorToOldRegistry() NodesCoordinatorRegistryHandler {
	ihgs.mutNodesConfig.RLock()
	defer ihgs.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistry{
		CurrentEpoch: ihgs.currentEpoch,
		EpochsConfig: make(map[string]*EpochValidators),
	}

	minEpoch, lastEpoch := ihgs.getMinAndLastEpoch()
	for epoch := minEpoch; epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihgs.nodesConfig[epoch]
		if !ok {
			continue
		}

		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return registry
}

func (ihgs *indexHashedNodesCoordinator) getMinAndLastEpoch() (uint32, uint32) {
	minEpoch := 0
	lastEpoch := ihgs.getLastEpochConfig()
	if lastEpoch >= nodesCoordinatorStoredEpochs {
		minEpoch = int(lastEpoch) - nodesCoordinatorStoredEpochs + 1
	}

	return uint32(minEpoch), lastEpoch
}

func (ihgs *indexHashedNodesCoordinator) getLastEpochConfig() uint32 {
	lastEpoch := uint32(0)
	for epoch := range ihgs.nodesConfig {
		if lastEpoch < epoch {
			lastEpoch = epoch
		}
	}

	return lastEpoch
}

func (ihgs *indexHashedNodesCoordinator) registryToNodesCoordinator(
	config NodesCoordinatorRegistryHandler,
) (map[uint32]*epochNodesConfig, error) {
	var err error
	var epoch int64
	result := make(map[uint32]*epochNodesConfig)

	for epochStr, epochValidators := range config.GetEpochsConfig() {
		epoch, err = strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			return nil, err
		}

		var nodesConfig *epochNodesConfig
		nodesConfig, err = epochValidatorsToEpochNodesConfig(epochValidators)
		if err != nil {
			return nil, err
		}

		nbShards := uint32(len(nodesConfig.eligibleMap))
		if nbShards < 2 {
			return nil, ErrInvalidNumberOfShards
		}

		// shards without metachain shard
		nodesConfig.nbShards = nbShards - 1
		nodesConfig.shardID, _ = ihgs.computeShardForSelfPublicKey(nodesConfig)
		epoch32 := uint32(epoch)
		result[epoch32] = nodesConfig
		log.Debug("registry to nodes coordinator", "epoch", epoch32)
		result[epoch32].selectors, err = ihgs.createSelectors(nodesConfig)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func epochNodesConfigToEpochValidators(config *epochNodesConfig) *EpochValidators {
	result := &EpochValidators{
		EligibleValidators: make(map[string][]*SerializableValidator, len(config.eligibleMap)),
		WaitingValidators:  make(map[string][]*SerializableValidator, len(config.waitingMap)),
		LeavingValidators:  make(map[string][]*SerializableValidator, len(config.leavingMap)),
	}

	for k, v := range config.eligibleMap {
		result.EligibleValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.waitingMap {
		result.WaitingValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.leavingMap {
		result.LeavingValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	return result
}

func epochValidatorsToEpochNodesConfig(config EpochValidatorsHandler) (*epochNodesConfig, error) {
	result := &epochNodesConfig{}
	var err error

	result.eligibleMap, err = serializableValidatorsMapToValidatorsMap(config.GetEligibleValidators())
	if err != nil {
		return nil, err
	}

	result.waitingMap, err = serializableValidatorsMapToValidatorsMap(config.GetWaitingValidators())
	if err != nil {
		return nil, err
	}

	result.leavingMap, err = serializableValidatorsMapToValidatorsMap(config.GetLeavingValidators())
	if err != nil {
		return nil, err
	}

	configWithAuction, castOk := config.(EpochValidatorsHandlerWithAuction)
	if castOk {
		result.shuffledOutMap, err = serializableValidatorsMapToValidatorsMap(configWithAuction.GetShuffledOutValidators())
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func serializableValidatorsMapToValidatorsMap(
	sValidators map[string][]*SerializableValidator,
) (map[uint32][]Validator, error) {

	result := make(map[uint32][]Validator, len(sValidators))

	for k, v := range sValidators {
		key, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return nil, err
		}

		result[uint32(key)], err = serializableValidatorArrayToValidatorArray(v)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ValidatorArrayToSerializableValidatorArray -
func ValidatorArrayToSerializableValidatorArray(validators []Validator) []*SerializableValidator {
	result := make([]*SerializableValidator, len(validators))

	for i, v := range validators {
		result[i] = &SerializableValidator{
			PubKey:  v.PubKey(),
			Chances: v.Chances(),
			Index:   v.Index(),
		}
	}

	return result
}

func serializableValidatorArrayToValidatorArray(sValidators []*SerializableValidator) ([]Validator, error) {
	result := make([]Validator, len(sValidators))
	var err error

	for i, v := range sValidators {
		result[i], err = NewValidator(v.PubKey, v.Chances, v.Index)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// NodesInfoToValidators maps nodeInfo to validator interface
func NodesInfoToValidators(nodesInfo map[uint32][]GenesisNodeInfoHandler) (map[uint32][]Validator, error) {
	validatorsMap := make(map[uint32][]Validator)

	for shId, nodeInfoList := range nodesInfo {
		validators := make([]Validator, 0, len(nodeInfoList))
		for index, nodeInfo := range nodeInfoList {
			validatorObj, err := NewValidator(nodeInfo.PubKeyBytes(), defaultSelectionChances, uint32(index))
			if err != nil {
				return nil, err
			}

			validators = append(validators, validatorObj)
		}
		validatorsMap[shId] = validators
	}

	return validatorsMap, nil
}
