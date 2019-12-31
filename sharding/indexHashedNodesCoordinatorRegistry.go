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

func (ihgs *indexHashedNodesCoordinator) saveState(key []byte) error {
	data, err := ihgs.nodesCoordinatorToRegistry()
	if err != nil {
		return err
	}

	key = append([]byte(keyPrefix), key...)

	return ihgs.bootStorer.Put(key, data)
}

func (ihgs *indexHashedNodesCoordinator) loadState(key []byte) error {
	key = append([]byte(keyPrefix), key...)

	data, err := ihgs.bootStorer.Get(key)
	if err != nil {
		return err
	}

	config := &NodesCoordinatorRegistry{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return ihgs.registryToNodesCoordinator(config)
}

func (ihgs *indexHashedNodesCoordinator) nodesCoordinatorToRegistry() ([]byte, error) {
	ihgs.mutNodesConfig.RLock()
	defer ihgs.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistry{
		CurrentEpoch: ihgs.currentEpoch,
		EpochsConfig: make(map[string]*EpochValidators, len(ihgs.nodesConfig)),
	}

	for epoch, epochNodesData := range ihgs.nodesConfig {
		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return json.Marshal(registry)
}

func (ihgs *indexHashedNodesCoordinator) registryToNodesCoordinator(config *NodesCoordinatorRegistry) error {
	var err error
	var epoch int64

	for epochStr, epochValidators := range config.EpochsConfig {
		epoch, err = strconv.ParseInt(epochStr, 10, 32)
		if err != nil {
			return err
		}

		var nodesConfig *epochNodesConfig

		nodesConfig, err = epochValidatorsToEpochNodesConfig(epochValidators)
		if err != nil {
			return err
		}

		nbShards := uint32(len(nodesConfig.eligibleMap))
		if nbShards <= 1 {
			return ErrInvalidNumberOfShards
		}

		// shards without metachain shard
		nodesConfig.nbShards = nbShards - 1
		nodesConfig.shardId = ihgs.computeShardForPublicKey(nodesConfig)
		epoch32 := uint32(epoch)
		ihgs.nodesConfig[epoch32] = nodesConfig
	}

	ihgs.currentEpoch = config.CurrentEpoch

	return nil
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
		key, err := strconv.ParseInt(k, 10, 32)
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
