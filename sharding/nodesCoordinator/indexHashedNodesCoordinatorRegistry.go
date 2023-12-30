package nodesCoordinator

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
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

func (ihnc *indexHashedNodesCoordinator) nodesConfigFromStaticStorer(
	epoch uint32,
) (*epochNodesConfig, error) {
	metaBlock, err := ihnc.metaBlockFromStaticStorer(epoch)
	if err != nil {
		return nil, err
	}

	validatorsInfo, err := ihnc.createValidatorsInfoFromStatic(metaBlock, epoch)
	if err != nil {
		return nil, err
	}

	return ihnc.nodesConfigFromValidatorsInfo(metaBlock, validatorsInfo)
}

func (ihnc *indexHashedNodesCoordinator) createValidatorsInfoFromStatic(
	metaBlock data.HeaderHandler,
	epoch uint32,
) ([]*state.ShardValidatorInfo, error) {
	mbHeaderHandlers := metaBlock.GetMiniBlockHeaderHandlers()

	allValidatorInfo := make([]*state.ShardValidatorInfo, 0)

	for _, mbHeader := range mbHeaderHandlers {
		if mbHeader.GetTypeInt32() != int32(block.PeerBlock) {
			continue
		}

		mbBytes, err := ihnc.epochStartStaticStorer.Get(mbHeader.GetHash())
		if err != nil {
			log.Error("createValidatorsInfoFromStatic:", "mbHeaderHash", mbHeader.GetHash(), "error", err)
			return nil, err
		}

		mb := &block.MiniBlock{}
		err = ihnc.marshalizer.Unmarshal(mb, mbBytes)
		if err != nil {
			log.Error("createValidatorsInfoFromStatic: Unmarshal", "mbHeaderHash", mbHeader.GetHash(), "error", err)
			return nil, err
		}

		for _, txHash := range mb.TxHashes {
			shardValidatorInfo, err := ihnc.getShardValidatorInfoFromStatic(txHash, epoch)
			if err != nil {
				return nil, err
			}

			allValidatorInfo = append(allValidatorInfo, shardValidatorInfo)
		}
	}

	return allValidatorInfo, nil
}

func (ihnc *indexHashedNodesCoordinator) getShardValidatorInfoFromStatic(
	txHash []byte,
	epoch uint32,
) (*state.ShardValidatorInfo, error) {
	marshalledShardValidatorInfo := txHash
	if epoch >= ihnc.enableEpochsHandler.GetActivationEpoch(common.RefactorPeersMiniBlocksFlag) {
		shardValidatorInfoBytes, err := ihnc.epochStartStaticStorer.Get(txHash)
		if err != nil {
			return nil, err
		}

		marshalledShardValidatorInfo = shardValidatorInfoBytes
	}

	shardValidatorInfo := &state.ShardValidatorInfo{}
	err := ihnc.marshalizer.Unmarshal(shardValidatorInfo, marshalledShardValidatorInfo)
	if err != nil {
		return nil, err
	}

	return shardValidatorInfo, nil
}

func (ihnc *indexHashedNodesCoordinator) nodesConfigFromValidatorsInfo(
	metaBlock data.HeaderHandler,
	validatorsInfo []*state.ShardValidatorInfo,
) (*epochNodesConfig, error) {
	epoch := metaBlock.GetEpoch()
	randomness := metaBlock.GetRandSeed()

	newNodesConfig, err := ihnc.computeNodesConfigFromList(&epochNodesConfig{}, validatorsInfo)
	if err != nil {
		return nil, err
	}

	additionalLeavingMap, err := ihnc.nodesCoordinatorHelper.ComputeAdditionalLeaving(validatorsInfo)
	if err != nil {
		return nil, err
	}

	unStakeLeavingList := ihnc.createSortedListFromMap(newNodesConfig.leavingMap)
	additionalLeavingList := ihnc.createSortedListFromMap(additionalLeavingMap)

	shufflerArgs := ArgsUpdateNodes{
		Eligible:          newNodesConfig.eligibleMap,
		Waiting:           newNodesConfig.waitingMap,
		NewNodes:          newNodesConfig.newList,
		UnStakeLeaving:    unStakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              randomness,
		NbShards:          newNodesConfig.nbShards,
		Epoch:             epoch,
	}

	resUpdateNodes, err := ihnc.shuffler.UpdateNodeLists(shufflerArgs)
	if err != nil {
		return nil, err
	}

	leavingNodesMap, _ := createActuallyLeavingPerShards(
		newNodesConfig.leavingMap,
		additionalLeavingMap,
		resUpdateNodes.Leaving,
	)

	nodesConfig := &epochNodesConfig{}
	nodesConfig.nbShards = uint32(len(resUpdateNodes.Eligible) - 1)
	nodesConfig.eligibleMap = resUpdateNodes.Eligible
	nodesConfig.waitingMap = resUpdateNodes.Waiting
	nodesConfig.leavingMap = leavingNodesMap
	nodesConfig.shardID, _ = ihnc.computeShardForSelfPublicKey(nodesConfig)
	nodesConfig.selectors, err = ihnc.createSelectors(nodesConfig)
	if err != nil {
		return nil, err
	}

	return nodesConfig, nil
}

func (ihnc *indexHashedNodesCoordinator) metaBlockFromStaticStorer(
	epoch uint32,
) (data.HeaderHandler, error) {
	epochStartBootstrapKey := append([]byte(common.EpochStartStaticBlockKeyPrefix), []byte(fmt.Sprint(epoch))...)
	metaBlockBytes, err := ihnc.epochStartStaticStorer.Get(epochStartBootstrapKey)
	if err != nil {
		return nil, err
	}

	metaBlock := &block.MetaBlock{}
	err = ihnc.marshalizer.Unmarshal(metaBlock, metaBlockBytes)
	if err != nil {
		return nil, err
	}

	if metaBlock.GetNonce() > 1 && !metaBlock.IsStartOfEpochBlock() {
		return nil, epochStart.ErrNotEpochStartBlock
	}

	if metaBlock.GetEpoch() != epoch {
		return nil, ErrInvalidEpochStartEpoch
	}

	return metaBlock, nil
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

	log.Debug(string(debug.Stack()))

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

	nc := &NodesCoordinatorRegistry{
		EpochsConfig: epochsConfig,
		CurrentEpoch: lastEpoch,
	}

	DisplayNodesCoordinatorRegistry(nc)

	return nc, nil
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
	for epoch, nConfig := range nodesConfig {
		ihnc.nodesConfigCacher.Put([]byte(fmt.Sprint(epoch)), nConfig, 0)
		log.Debug("baseLoadState: put nodes config in cache", "epoch", epoch, "shard ID", ihnc.shuffledOutHandler.CurrentShardID())

		displayNodesConfiguration(
			nConfig.eligibleMap,
			nConfig.waitingMap,
			nConfig.leavingMap,
			make(map[uint32][]Validator),
			nConfig.nbShards)
	}
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

func (ihnc *indexHashedNodesCoordinator) checkInitialSaveState(key []byte) error {
	registry := ihnc.NodesCoordinatorToRegistry()

	for epoch := range registry.EpochsConfig {
		ncInternalkey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(epoch)...)
		_, err := ihnc.bootStorer.Get(ncInternalkey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ihnc *indexHashedNodesCoordinator) saveState(key []byte) error {
	err := ihnc.checkInitialSaveState(key)
	if err == nil {
		log.Debug("initial saveState: no need to rewrite nodes coordinator keys")
		return nil
	}

	registry := ihnc.NodesCoordinatorToRegistry()

	err = SaveNodesCoordinatorRegistry(registry, ihnc.bootStorer)
	if err != nil {
		return err
	}

	nodesConfig, err := ihnc.registryToNodesCoordinator(registry)
	if err != nil {
		return err
	}

	for epoch, nConfig := range nodesConfig {
		log.Debug("saveState: nodes config", "epoch", epoch)
		displayNodesConfiguration(
			nConfig.eligibleMap,
			nConfig.waitingMap,
			nConfig.leavingMap,
			make(map[uint32][]Validator),
			nConfig.nbShards)
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
		epochNodesData, ok := ihnc.getNodesConfig(epoch)
		if !ok {
			log.Debug("NodesCoordinatorToRegistry: did not find nodes config", "epoch", epoch)
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

	log.Debug(string(debug.Stack()))

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

	DisplayNodesCoordinatorRegistry(nodesConfig)

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
