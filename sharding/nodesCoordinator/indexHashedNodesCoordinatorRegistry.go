package nodesCoordinator

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
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
func (ihnc *indexHashedNodesCoordinator) LoadState(key []byte, epoch uint32) error {
	return ihnc.baseLoadState(key, epoch)
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

	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(lastEpoch))...)
	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

	epochConfigBytes, err := storer.SearchFirst(ncInternalkey)
	if err != nil {
		log.Debug("failed to get nodes coordinator config", "key", ncInternalkey)
		return handleNodesCoordinatorRegistryByOldKey(key, storer)
	}

	err = updateEpochsConfig(epochsConfig, epochConfigBytes)
	if err != nil {
		return nil, err
	}

	for epoch := int(lastEpoch) - 1; epoch >= minEpoch; epoch-- {
		err := setEpochConfigPerEpoch(epoch, epochsConfig, storer)
		if err != nil {
			log.Debug("failed to get nodes coordinator config for epoch", "epoch", epoch)
		}
	}

	ncr := &NodesCoordinatorRegistry{
		EpochsConfig: epochsConfig,
		CurrentEpoch: lastEpoch,
	}

	nc, err := registryToNodesCoordinator(ncr)
	if err != nil {
		return nil, err
	}

	log.Debug("nodes configuration on get")
	displayNodesConfiguration(
		nc[lastEpoch].eligibleMap,
		nc[lastEpoch].waitingMap,
		nc[lastEpoch].leavingMap,
		make(map[uint32][]Validator),
		uint32(len(nc[lastEpoch].eligibleMap)-1),
	)

	return ncr, nil
}

func setEpochConfigPerEpoch(
	epoch int,
	epochsConfig map[string]*EpochValidators,
	storer storage.Storer,
) error {
	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(epoch))...)
	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

	epochConfigBytes, err := storer.SearchFirst(ncInternalkey)
	if err != nil {
		return err
	}

	err = updateEpochsConfig(epochsConfig, epochConfigBytes)
	if err != nil {
		return err
	}

	return nil
}

func updateEpochsConfig(epochsConfig map[string]*EpochValidators, epochConfig []byte) error {
	registry := &NodesCoordinatorRegistry{}
	err := json.Unmarshal(epochConfig, registry)
	if err != nil {
		return err
	}

	for epoch, config := range registry.EpochsConfig {
		epochsConfig[epoch] = config
	}

	return nil
}

func handleNodesCoordinatorRegistryByOldKey(
	key []byte,
	storer storage.Storer,
) (*NodesCoordinatorRegistry, error) {
	registry, err := getNodesCoordinatorRegistryByRandomnessKey(key, storer)
	if err != nil {
		return nil, err
	}

	err = SaveNodesCoordinatorRegistry(registry, storer)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func getNodesCoordinatorRegistryByRandomnessKey(
	key []byte,
	storer storage.Storer,
) (*NodesCoordinatorRegistry, error) {
	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("getting nodes coordinator config by randomness", "key", ncInternalkey)

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

func (ihnc *indexHashedNodesCoordinator) baseLoadState(key []byte, lastEpoch uint32) error {
	ihnc.loadingFromDisk.Store(true)
	defer ihnc.loadingFromDisk.Store(false)

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

	maxEpoch := uint32(0)
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
		debug.PrintStack()

		log.Debug("saving nodes coordinator config", "key", ncInternalkey)

		if currentEpoch > uint64(maxEpoch) {
			maxEpoch = uint32(currentEpoch)
		}
	}

	nc, err := registryToNodesCoordinator(nodesConfig)
	if err != nil {
		return err
	}

	log.Debug("nodes configuration on save")
	displayNodesConfiguration(
		nc[maxEpoch].eligibleMap,
		nc[maxEpoch].waitingMap,
		nc[maxEpoch].leavingMap,
		make(map[uint32][]Validator),
		uint32(len(nodesConfig.EpochsConfig[fmt.Sprint(maxEpoch)].EligibleValidators)-1),
	)

	return nil
}

func registryToNodesCoordinator(
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
		epoch32 := uint32(epoch)
		result[epoch32] = nodesConfig
		log.Debug("TEST: registry to nodes coordinator", "epoch", epoch32)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
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
