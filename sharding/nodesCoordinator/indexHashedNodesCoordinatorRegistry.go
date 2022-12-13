package nodesCoordinator

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
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
func (ihnc *indexHashedNodesCoordinator) LoadState(key []byte) error {
	return ihnc.baseLoadState(key)
}

// GetNodesCoordinatorRegistry will get the nodes coordinator registry from boot storage
func GetNodesCoordinatorRegistry(
	key []byte, // old key
	storer storage.Storer,
	lastEpoch uint32,
	numStoredEpochs uint32,
) (*NodesCoordinatorRegistry, error) {
	if check.IfNil(storer) {
		return nil, ErrNilBootStorer
	}

	minEpoch := 0
	if lastEpoch >= numStoredEpochs {
		minEpoch = int(lastEpoch) - int(numStoredEpochs) + 1
	}

	epochsConfig := make(map[string]*EpochValidators)
	for epoch := int(lastEpoch); epoch >= minEpoch; epoch-- {
		ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(epoch))...)

		log.Debug("getting nodes coordinator config", "key", ncInternalkey)

		epochConfigBytes, err := storer.Get(ncInternalkey)
		if err != nil {
			return getNodesCoordinatorRegistryByRandomnessKey(key, storer)
		}

		epochConfig := &NodesCoordinatorRegistry{}
		err = json.Unmarshal(epochConfigBytes, epochConfig)
		if err != nil {
			return nil, err
		}

		for epoch, config := range epochConfig.EpochsConfig {
			epochsConfig[epoch] = config
		}
	}

	return &NodesCoordinatorRegistry{
		EpochsConfig: epochsConfig,
		CurrentEpoch: lastEpoch,
	}, nil
}

func getNodesCoordinatorRegistryByRandomnessKey(
	key []byte,
	storer storage.Storer,
) (*NodesCoordinatorRegistry, error) {
	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)
	epochConfigBytes, err := storer.SearchFirst(ncInternalkey)
	if err != nil {
		return nil, err
	}
	epochConfig := &NodesCoordinatorRegistry{}
	err = json.Unmarshal(epochConfigBytes, epochConfig)
	if err != nil {
		return nil, err
	}

	return epochConfig, nil
}

func (ihnc *indexHashedNodesCoordinator) baseLoadState(key []byte) error {
	ihnc.loadingFromDisk.Store(true)
	defer ihnc.loadingFromDisk.Store(false)

	lastEpoch := ihnc.getLastEpochConfig()
	config, err := GetNodesCoordinatorRegistry(key, ihnc.bootStorer, lastEpoch, ihnc.numStoredEpochs)
	if err != nil {
		return err
	}

	ihnc.mutSavedStateKey.Lock()
	ihnc.savedStateKey = key
	ihnc.mutSavedStateKey.Unlock()

	ihnc.currentEpoch = config.CurrentEpoch
	log.Debug("loaded nodes config", "current epoch", config.CurrentEpoch)

	nodesConfig, err := ihnc.registryToNodesCoordinator(config)
	if err != nil {
		return err
	}

	displayNodesConfigInfo(nodesConfig)
	ihnc.mutNodesConfig.Lock()
	ihnc.nodesConfig = nodesConfig
	ihnc.mutNodesConfig.Unlock()

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

func (ihnc *indexHashedNodesCoordinator) saveState(key []byte) error {
	registry := ihnc.NodesCoordinatorToRegistry()

	err := SaveNodesCoordinatorRegistry(registry, ihnc.bootStorer)
	if err != nil {
		return err
	}

	return nil
}

// NodesCoordinatorToRegistry will export the nodesCoordinator data to the registry
func (ihnc *indexHashedNodesCoordinator) NodesCoordinatorToRegistry() *NodesCoordinatorRegistry {
	ihnc.mutNodesConfig.RLock()
	defer ihnc.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistry{
		CurrentEpoch: ihnc.currentEpoch,
		EpochsConfig: make(map[string]*EpochValidators),
	}

	minEpoch := 0
	lastEpoch := ihnc.getLastEpochConfig()
	if lastEpoch >= ihnc.numStoredEpochs {
		minEpoch = int(lastEpoch) - int(ihnc.numStoredEpochs) + 1
	}

	for epoch := uint32(minEpoch); epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihnc.nodesConfig[epoch]
		if !ok {
			continue
		}

		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return registry
}

// SaveNodesCoordinatorRegistry will save the nodes coordinator registry to storage
func SaveNodesCoordinatorRegistry(
	nodesConfig *NodesCoordinatorRegistry,
	storer storage.Storer,
) error {
	if check.IfNil(storer) {
		return ErrNilBootStorer
	}
	if nodesConfig == nil {
		return ErrNilNodesCoordinatorRegistry
	}

	for epoch, config := range nodesConfig.EpochsConfig {
		epochsConfig := make(map[string]*EpochValidators)
		epochsConfig[epoch] = config
		currentEpoch, err := strconv.ParseUint(epoch, 10, 32)
		if err != nil {
			return err
		}
		registry := &NodesCoordinatorRegistry{
			CurrentEpoch: uint32(currentEpoch),
			EpochsConfig: epochsConfig,
		}

		registryBytes, err := json.Marshal(registry)
		if err != nil {
			return err
		}

		ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(epoch)...)
		err = storer.Put(ncInternalkey, registryBytes)
		if err != nil {
			return err
		}

		log.Debug("saving nodes coordinator config", "key", ncInternalkey)
	}

	return nil
}

func (ihnc *indexHashedNodesCoordinator) getLastEpochConfig() uint32 {
	lastEpoch := uint32(0)
	for epoch := range ihnc.nodesConfig {
		if lastEpoch < epoch {
			lastEpoch = epoch
		}
	}

	return lastEpoch
}

func (ihnc *indexHashedNodesCoordinator) registryToNodesCoordinator(
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
		nodesConfig.shardID, _ = ihnc.computeShardForSelfPublicKey(nodesConfig)
		epoch32 := uint32(epoch)
		result[epoch32] = nodesConfig
		log.Debug("registry to nodes coordinator", "epoch", epoch32)
		result[epoch32].selectors, err = ihnc.createSelectors(nodesConfig)
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
