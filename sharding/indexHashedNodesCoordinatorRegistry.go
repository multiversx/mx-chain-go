package sharding

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
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

func (ihgs *indexHashedNodesCoordinator) baseLoadState(key []byte) error {
	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

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

	ihgs.SetCurrentEpoch(config.CurrentEpoch)
	log.Debug("loaded nodes config", "current epoch", config.CurrentEpoch)

	nodesConfig, err := ihgs.registryToNodesCoordinator(config)
	if err != nil {
		return err
	}

	displayNodesConfigInfo(nodesConfig)
	ihgs.SetNodesConfig(nodesConfig)

	return nil
}

func displayNodesConfigInfo(config map[uint32]*nodesCoordinator.EpochNodesConfig) {
	for epoch, cfg := range config {
		log.Debug("restored config for",
			"epoch", epoch,
			"computed shard ID", cfg.ShardID,
		)
	}
}

func (ihgs *indexHashedNodesCoordinator) saveState(key []byte) error {
	registry := ihgs.NodesCoordinatorToRegistry()
	data, err := json.Marshal(registry)
	if err != nil {
		return err
	}

	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("saving nodes coordinator config", "key", ncInternalkey)

	return ihgs.bootStorer.Put(ncInternalkey, data)
}

// NodesCoordinatorToRegistry will export the nodesCoordinator data to the registry
func (ihgs *indexHashedNodesCoordinator) NodesCoordinatorToRegistry() *NodesCoordinatorRegistry {
	ihgs.mutNodesConfig.RLock()
	defer ihgs.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistry{
		CurrentEpoch: ihgs.GetCurrentEpoch(),
		EpochsConfig: make(map[string]*EpochValidators),
	}

	minEpoch := 0
	lastEpoch := ihgs.GetLastEpochConfig()
	if lastEpoch >= nodesCoordinatorStoredEpochs {
		minEpoch = int(lastEpoch) - nodesCoordinatorStoredEpochs + 1
	}

	for epoch := uint32(minEpoch); epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihgs.GetNodesConfigPerEpoch(epoch)
		if !ok {
			continue
		}

		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return registry
}

func (ihgs *indexHashedNodesCoordinator) registryToNodesCoordinator(
	config *NodesCoordinatorRegistry,
) (map[uint32]*nodesCoordinator.EpochNodesConfig, error) {
	var err error
	var epoch int64
	result := make(map[uint32]*nodesCoordinator.EpochNodesConfig)

	for epochStr, epochValidators := range config.EpochsConfig {
		epoch, err = strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			return nil, err
		}

		var nodesConfig *nodesCoordinator.EpochNodesConfig
		nodesConfig, err = epochValidatorsToEpochNodesConfig(epochValidators)
		if err != nil {
			return nil, err
		}

		nbShards := uint32(len(nodesConfig.EligibleMap))
		if nbShards < 2 {
			return nil, ErrInvalidNumberOfShards
		}

		// shards without metachain shard
		nodesConfig.NbShards = nbShards - 1
		nodesConfig.ShardID, _ = ihgs.computeShardForSelfPublicKey(nodesConfig)
		epoch32 := uint32(epoch)
		result[epoch32] = nodesConfig
		log.Debug("registry to nodes coordinator", "epoch", epoch32)
		result[epoch32].Selectors, err = ihgs.CreateSelectors(nodesConfig)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func epochNodesConfigToEpochValidators(config *nodesCoordinator.EpochNodesConfig) *EpochValidators {
	result := &EpochValidators{
		EligibleValidators: make(map[string][]*SerializableValidator, len(config.EligibleMap)),
		WaitingValidators:  make(map[string][]*SerializableValidator, len(config.WaitingMap)),
		LeavingValidators:  make(map[string][]*SerializableValidator, len(config.LeavingMap)),
	}

	for k, v := range config.EligibleMap {
		result.EligibleValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.WaitingMap {
		result.WaitingValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	for k, v := range config.LeavingMap {
		result.LeavingValidators[fmt.Sprint(k)] = ValidatorArrayToSerializableValidatorArray(v)
	}

	return result
}

func epochValidatorsToEpochNodesConfig(config *EpochValidators) (*nodesCoordinator.EpochNodesConfig, error) {
	result := &nodesCoordinator.EpochNodesConfig{}
	var err error

	result.EligibleMap, err = serializableValidatorsMapToValidatorsMap(config.EligibleValidators)
	if err != nil {
		return nil, err
	}

	result.WaitingMap, err = serializableValidatorsMapToValidatorsMap(config.WaitingValidators)
	if err != nil {
		return nil, err
	}

	result.LeavingMap, err = serializableValidatorsMapToValidatorsMap(config.LeavingValidators)
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
		result[i], err = nodesCoordinator.NewValidator(v.PubKey, v.Chances, v.Index)
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
			validatorObj, err := nodesCoordinator.NewValidator(nodeInfo.PubKeyBytes(), nodesCoordinator.DefaultSelectionChances, uint32(index))
			if err != nil {
				return nil, err
			}

			validators = append(validators, validatorObj)
		}
		validatorsMap[shId] = validators
	}

	return validatorsMap, nil
}
