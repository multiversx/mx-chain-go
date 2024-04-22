package nodesCoordinator

import (
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-go/common"
)

// LoadState loads the nodes coordinator state from the used boot storage
func (ihnc *indexHashedNodesCoordinator) LoadState(key []byte) error {
	return ihnc.baseLoadState(key)
}

func (ihnc *indexHashedNodesCoordinator) baseLoadState(key []byte) error {
	ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)

	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

	ihnc.loadingFromDisk.Store(true)
	defer ihnc.loadingFromDisk.Store(false)

	data, err := ihnc.bootStorer.Get(ncInternalkey)
	if err != nil {
		return err
	}

	config, err := ihnc.nodesCoordinatorRegistryFactory.CreateNodesCoordinatorRegistry(data)
	if err != nil {
		return err
	}

	ihnc.mutSavedStateKey.Lock()
	ihnc.savedStateKey = key
	ihnc.mutSavedStateKey.Unlock()

	ihnc.currentEpoch = config.GetCurrentEpoch()
	log.Debug("loaded nodes config", "current epoch", config.GetCurrentEpoch())

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

func (ihnc *indexHashedNodesCoordinator) saveState(key []byte, epoch uint32) error {
	registry := ihnc.NodesCoordinatorToRegistry(epoch)
	data, err := ihnc.nodesCoordinatorRegistryFactory.GetRegistryData(registry, ihnc.currentEpoch)
	if err != nil {
		return err
	}

	ncInternalKey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), key...)
	log.Debug("saving nodes coordinator config", "key", ncInternalKey, "epoch", epoch)

	return ihnc.bootStorer.Put(ncInternalKey, data)
}

// NodesCoordinatorToRegistry will export the nodesCoordinator data to the registry
func (ihnc *indexHashedNodesCoordinator) NodesCoordinatorToRegistry(epoch uint32) NodesCoordinatorRegistryHandler {
	if epoch >= ihnc.enableEpochsHandler.GetActivationEpoch(common.StakingV4Step2Flag) {
		log.Debug("indexHashedNodesCoordinator.NodesCoordinatorToRegistry called with auction registry", "epoch", epoch)
		return ihnc.nodesCoordinatorToRegistryWithAuction()
	}

	log.Debug("indexHashedNodesCoordinator.NodesCoordinatorToRegistry called with old registry", "epoch", epoch)
	return ihnc.nodesCoordinatorToOldRegistry()
}

func (ihnc *indexHashedNodesCoordinator) nodesCoordinatorToOldRegistry() NodesCoordinatorRegistryHandler {
	ihnc.mutNodesConfig.RLock()
	defer ihnc.mutNodesConfig.RUnlock()

	registry := &NodesCoordinatorRegistry{
		CurrentEpoch: ihnc.currentEpoch,
		EpochsConfig: make(map[string]*EpochValidators),
	}

	minEpoch, lastEpoch := ihnc.getMinAndLastEpoch()
	for epoch := minEpoch; epoch <= lastEpoch; epoch++ {
		epochNodesData, ok := ihnc.nodesConfig[epoch]
		if !ok {
			continue
		}

		registry.EpochsConfig[fmt.Sprint(epoch)] = epochNodesConfigToEpochValidators(epochNodesData)
	}

	return registry
}

func (ihnc *indexHashedNodesCoordinator) getMinAndLastEpoch() (uint32, uint32) {
	minEpoch := 0
	lastEpoch := ihnc.getLastEpochConfig()
	if lastEpoch >= nodesCoordinatorStoredEpochs {
		minEpoch = int(lastEpoch) - nodesCoordinatorStoredEpochs + 1
	}

	return uint32(minEpoch), lastEpoch
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
