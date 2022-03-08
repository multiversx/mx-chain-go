package sharding

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/common"
)

// EpochValidators holds one epoch configuration for a nodes coordinator
type EpochValidators struct {
	EligibleValidators map[string][]*SerializableValidator `json:"eligibleValidators"`
	WaitingValidators  map[string][]*SerializableValidator `json:"waitingValidators"`
	LeavingValidators  map[string][]*SerializableValidator `json:"leavingValidators"`
}

func (ev *EpochValidators) GetEligibleValidators() map[string][]*SerializableValidator {
	return ev.EligibleValidators
}

func (ev *EpochValidators) GetWaitingValidators() map[string][]*SerializableValidator {
	return ev.WaitingValidators
}

func (ev *EpochValidators) GetLeavingValidators() map[string][]*SerializableValidator {
	return ev.LeavingValidators
}

// NodesCoordinatorRegistry holds the data that can be used to initialize a nodes coordinator
type NodesCoordinatorRegistry struct {
	EpochsConfig map[string]*EpochValidators `json:"epochConfigs"`
	CurrentEpoch uint32                      `json:"currentEpoch"`
}

func (ncr *NodesCoordinatorRegistry) GetCurrentEpoch() uint32 {
	return ncr.CurrentEpoch
}

func (ncr *NodesCoordinatorRegistry) GetEpochsConfig() map[string]EpochValidatorsHandler {
	ret := make(map[string]EpochValidatorsHandler)
	for epoch, config := range ncr.EpochsConfig {
		ret[epoch] = config
	}

	return ret
}

func (ncr *NodesCoordinatorRegistry) SetCurrentEpoch(epoch uint32) {
	ncr.CurrentEpoch = epoch
}

func (ncr *NodesCoordinatorRegistry) SetEpochsConfig(epochsConfig map[string]EpochValidatorsHandler) {
	ncr.EpochsConfig = make(map[string]*EpochValidators)

	for epoch, config := range epochsConfig {
		ncr.EpochsConfig[epoch] = &EpochValidators{
			EligibleValidators: config.GetEligibleValidators(),
			WaitingValidators:  config.GetWaitingValidators(),
			LeavingValidators:  config.GetLeavingValidators(),
		}
	}
}

// EpochValidatorsHandler defines what one epoch configuration for a nodes coordinator should hold
type EpochValidatorsHandler interface {
	GetEligibleValidators() map[string][]*SerializableValidator
	GetWaitingValidators() map[string][]*SerializableValidator
	GetLeavingValidators() map[string][]*SerializableValidator
}

type EpochValidatorsHandlerWithAuction interface {
	EpochValidatorsHandler
	GetShuffledOutValidators() map[string][]*SerializableValidator
}

// NodesCoordinatorRegistryHandler defines that used to initialize nodes coordinator
type NodesCoordinatorRegistryHandler interface {
	GetEpochsConfig() map[string]EpochValidatorsHandler
	GetCurrentEpoch() uint32

	SetCurrentEpoch(epoch uint32)
	SetEpochsConfig(epochsConfig map[string]EpochValidatorsHandler)
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

	var config NodesCoordinatorRegistryHandler
	if ihgs.flagStakingV4.IsSet() {
		config = &NodesCoordinatorRegistryWithAuction{}
		err = json.Unmarshal(data, config)
		if err != nil {
			return err
		}
	} else {
		config = &NodesCoordinatorRegistry{}
		err = json.Unmarshal(data, config)
		if err != nil {
			return err
		}
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
	registry := ihgs.NodesCoordinatorToRegistry()
	data, err := json.Marshal(registry) // TODO: Choose different marshaller depending on registry
	if err != nil {
		return err
	}

	ncInternalKey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("saving nodes coordinator config", "key", ncInternalKey)

	return ihgs.bootStorer.Put(ncInternalKey, data)
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

	minEpoch := 0
	lastEpoch := ihgs.getLastEpochConfig()
	if lastEpoch >= nodesCoordinatorStoredEpochs {
		minEpoch = int(lastEpoch) - nodesCoordinatorStoredEpochs + 1
	}

	for epoch := uint32(minEpoch); epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihgs.nodesConfig[epoch]
		if !ok {
			continue
		}

		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return registry
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
