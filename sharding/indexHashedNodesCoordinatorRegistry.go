package sharding

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/core"
)

// SerializableValidator holds the minimal data required for marshalling and un-marshalling a validator
type SerializableValidator struct {
	PubKey  []byte `json:"pubKey"`
	Chances uint32 `json:"chances"`
	Index   uint32 `json:"index"`
}

// EpochValidators holds one epoch configuration for a nodes coordinator
type EpochValidators struct {
	EligibleValidators map[string][]*SerializableValidator `json:"eligibleValidators"`
	WaitingValidators  map[string][]*SerializableValidator `json:"waitingValidators"`
	LeavingValidators  map[string][]*SerializableValidator `json:"leavingValidators"`
}

// NodesCoordinatorRegistry holds the data that can be used to initialize a nodes coordinator
type NodesCoordinatorRegistry struct {
	EpochsConfig map[string]*EpochValidators `json:"epochConfigs"`
	CurrentEpoch uint32                      `json:"currentEpoch"`
}

// TODO: add proto marshalizer for these package - replace all json marshalizers
// LoadState loads the nodes coordinator state from the used boot storage
func (ihgs *indexHashedNodesCoordinator) LoadState(key []byte) error {
	return ihgs.baseLoadState(key)
}

// LoadState loads the nodes coordinator state from the used boot storage
func (ihgs *indexHashedNodesCoordinatorWithRater) LoadState(key []byte) error {
	return ihgs.baseLoadState(key)
}

func (ihgs *indexHashedNodesCoordinator) baseLoadState(key []byte) error {
	ncInternalkey := append([]byte(core.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

	ihgs.loadingFromDisk.Store(true)
	defer ihgs.loadingFromDisk.Store(false)

	data, err := ihgs.bootStorer.Get(ncInternalkey)
	if err != nil {
		return err
	}

	config := &NodesCoordinatorRegistry{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = key
	ihgs.mutSavedStateKey.Unlock()

	ihgs.currentEpoch = config.CurrentEpoch
	log.Debug("loaded nodes config", "current epoch", config.CurrentEpoch)

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
	registry := ihgs.NodesCoordinatorToRegistry()
	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	ncInternalkey := append([]byte(core.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("saving nodes coordinator config", "key", ncInternalkey)

	return ihgs.bootStorer.Put(ncInternalkey, data)
}

// NodesCoordinatorToRegistry will export the nodesCoordinator data to the registry
func (ihgs *indexHashedNodesCoordinator) NodesCoordinatorToRegistry() *NodesCoordinatorRegistry {
	ihgs.mutNodesConfig.RLock()
	defer ihgs.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistry{
		CurrentEpoch: ihgs.currentEpoch,
		EpochsConfig: make(map[string]*EpochValidators, len(ihgs.nodesConfig)),
	}

	for epoch, epochNodesData := range ihgs.nodesConfig {
		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return registry
}

func (ihgs *indexHashedNodesCoordinator) registryToNodesCoordinator(
	config *NodesCoordinatorRegistry,
) (map[uint32]*epochNodesConfig, error) {
	var err error
	var epoch int64
	result := make(map[uint32]*epochNodesConfig)

	for epochStr, epochValidators := range config.EpochsConfig {
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
		nodesConfig.shardID = ihgs.computeShardForSelfPublicKey(nodesConfig)
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

func epochValidatorsToEpochNodesConfig(config *EpochValidators) (*epochNodesConfig, error) {
	result := &epochNodesConfig{}
	var err error

	result.eligibleMap, err = serializableValidatorsMapToValidatorsMap(config.EligibleValidators)
	if err != nil {
		return nil, err
	}

	result.waitingMap, err = serializableValidatorsMapToValidatorsMap(config.WaitingValidators)
	if err != nil {
		return nil, err
	}

	result.leavingMap, err = serializableValidatorsMapToValidatorsMap(config.LeavingValidators)
	if err != nil {
		return nil, err
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
