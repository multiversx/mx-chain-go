package sharding

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const keyPrefix = "indexHashed_"

// SerializableValidator holds the minimal data required for marshalling and un-marshalling a validator
type SerializableValidator struct {
	PubKey  []byte `json:"pubKey"`
	Address []byte `json:"address"`
}

// EpochValidators holds one epoch configuration for a nodes coordinator
type EpochValidators struct {
	EligibleValidators map[string][]*SerializableValidator `json:"eligibleValidators"`
	WaitingValidators  map[string][]*SerializableValidator `json:"waitingValidators"`
}

// NodesCoordinatorRegistry holds the data that can be used to initialize a nodes coordinator
type NodesCoordinatorRegistry struct {
	EpochsConfig map[string]*EpochValidators `json:"epochConfigs"`
	CurrentEpoch uint32                      `json:"currentEpoch"`
}

// LoadState loads the nodes coordinator state from the used boot storage
func (ihgs *indexHashedNodesCoordinator) LoadState(key []byte) error {
	ncInternalkey := append([]byte(keyPrefix), key...)

	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

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

	nodesConfig, err := ihgs.registryToNodesCoordinator(config)
	if err != nil {
		return err
	}

	ihgs.mutNodesConfig.Lock()
	ihgs.nodesConfig = nodesConfig
	ihgs.mutNodesConfig.Unlock()

	return nil
}

func (ihgs *indexHashedNodesCoordinator) saveState(key []byte) error {
	registry := ihgs.nodesCoordinatorToRegistry()
	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	ncInternalkey := append([]byte(keyPrefix), key...)

	log.Debug("saving nodes coordinator config", "key", ncInternalkey)

	return ihgs.bootStorer.Put(ncInternalkey, data)
}

func (ihgs *indexHashedNodesCoordinator) nodesCoordinatorToRegistry() *NodesCoordinatorRegistry {
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
		if nbShards <= 1 {
			return nil, ErrInvalidNumberOfShards
		}

		// shards without metachain shard
		nodesConfig.nbShards = nbShards - 1
		nodesConfig.shardId = ihgs.computeShardForPublicKey(nodesConfig)
		epoch32 := uint32(epoch)
		result[epoch32] = nodesConfig
	}

	return result, nil
}

func epochNodesConfigToEpochValidators(config *epochNodesConfig) *EpochValidators {
	result := &EpochValidators{
		EligibleValidators: make(map[string][]*SerializableValidator, len(config.eligibleMap)),
		WaitingValidators:  make(map[string][]*SerializableValidator, len(config.waitingMap)),
	}

	for k, v := range config.eligibleMap {
		result.EligibleValidators[fmt.Sprint(k)] = validatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.waitingMap {
		result.WaitingValidators[fmt.Sprint(k)] = validatorArrayToSerializableValidatorArray(v)
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

func validatorArrayToSerializableValidatorArray(validators []Validator) []*SerializableValidator {
	result := make([]*SerializableValidator, len(validators))

	for i, v := range validators {
		result[i] = &SerializableValidator{
			PubKey:  v.PubKey(),
			Address: v.Address(),
		}
	}

	return result
}

func serializableValidatorArrayToValidatorArray(sValidators []*SerializableValidator) ([]Validator, error) {
	result := make([]Validator, len(sValidators))
	var err error

	for i, v := range sValidators {
		result[i], err = NewValidator(v.PubKey, v.Address)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
