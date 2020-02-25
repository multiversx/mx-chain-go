package sharding

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	keyFormat = "%s_%v_%v_%v"
)

// TODO: move this to config parameters
const nodeCoordinatorStoredEpochs = 2

// TODO: add a parameter for shardId  when acting as observer
type epochNodesConfig struct {
	nbShards     uint32
	shardId      uint32
	eligibleMap  map[uint32][]Validator
	waitingMap   map[uint32][]Validator
	mutNodesMaps sync.RWMutex
}

// EpochStartSubscriber provides Register and Unregister functionality for the end of epoch events
type EpochStartSubscriber interface {
	RegisterHandler(handler epochStart.EpochStartHandler)
	UnregisterHandler(handler epochStart.EpochStartHandler)
}

type indexHashedNodesCoordinator struct {
	doExpandEligibleList    func(validators []Validator, mut *sync.RWMutex) []Validator
	hasher                  hashing.Hasher
	shuffler                NodesShuffler
	epochStartSubscriber    EpochStartSubscriber
	bootStorer              storage.Storer
	selfPubKey              []byte
	nodesConfig             map[uint32]*epochNodesConfig
	mutNodesConfig          sync.RWMutex
	currentEpoch            uint32
	savedStateKey           []byte
	mutSavedStateKey        sync.RWMutex
	numTotalEligible        uint64
	shardConsensusGroupSize int
	metaConsensusGroupSize  int
	consensusGroupCacher    Cacher
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	nodesConfig := make(map[uint32]*epochNodesConfig, nodeCoordinatorStoredEpochs)

	nodesConfig[arguments.Epoch] = &epochNodesConfig{
		nbShards:     arguments.NbShards,
		shardId:      arguments.ShardId,
		eligibleMap:  make(map[uint32][]Validator),
		waitingMap:   make(map[uint32][]Validator),
		mutNodesMaps: sync.RWMutex{},
	}

	savedKey := arguments.Hasher.Compute(string(arguments.SelfPublicKey))

	ihgs := &indexHashedNodesCoordinator{
		hasher:                  arguments.Hasher,
		shuffler:                arguments.Shuffler,
		epochStartSubscriber:    arguments.EpochStartSubscriber,
		bootStorer:              arguments.BootStorer,
		selfPubKey:              arguments.SelfPublicKey,
		nodesConfig:             nodesConfig,
		currentEpoch:            arguments.Epoch,
		savedStateKey:           savedKey,
		shardConsensusGroupSize: arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:  arguments.MetaConsensusGroupSize,
		consensusGroupCacher:    arguments.ConsensusGroupCache,
	}

	ihgs.doExpandEligibleList = ihgs.expandEligibleList

	err = ihgs.SetNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, arguments.Epoch)
	if err != nil {
		return nil, err
	}

	err = ihgs.saveState(ihgs.savedStateKey)
	if err != nil {
		log.Error("saving initial nodes coordinator config failed",
			"error", err.Error())
	}

	ihgs.epochStartSubscriber.RegisterHandler(ihgs)

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinator) error {
	if arguments.ShardConsensusGroupSize < 1 || arguments.MetaConsensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}
	if arguments.NbShards < 1 {
		return ErrInvalidNumberOfShards
	}
	if arguments.ShardId >= arguments.NbShards && arguments.ShardId != core.MetachainShardId {
		return ErrInvalidShardId
	}
	if arguments.Hasher == nil {
		return ErrNilHasher
	}
	if arguments.SelfPublicKey == nil {
		return ErrNilPubKey
	}
	if arguments.Shuffler == nil {
		return ErrNilShuffler
	}
	if check.IfNil(arguments.BootStorer) {
		return ErrNilBootStorer
	}
	if arguments.ConsensusGroupCache == nil {
		return ErrNilCacher
	}

	return nil
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinator) SetNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	epoch uint32,
) error {

	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	nodesConfig, ok := ihgs.nodesConfig[epoch]

	if !ok {
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	if eligible == nil || waiting == nil {
		return ErrNilInputNodesMap
	}

	nodesList, ok := eligible[core.MetachainShardId]
	if !ok || len(nodesList) < ihgs.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	numTotalEligible := uint64(len(nodesList[MetachainShardId]))
	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := len(eligible[shardId])
		if nbNodesShard < ihgs.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
		numTotalEligible += uint64(nbNodesShard)
	}

	// nbShards holds number of shards without meta
	nodesConfig.nbShards = uint32(len(eligible) - 1)
	nodesConfig.eligibleMap = eligible
	nodesConfig.waitingMap = waiting
	nodesConfig.shardId = ihgs.computeShardForPublicKey(nodesConfig)
	ihgs.nodesConfig[epoch] = nodesConfig

	return nil
}

// GetNodesPerShard returns the eligible nodes per shard map
func (ihgs *indexHashedNodesCoordinator) GetNodesPerShard(epoch uint32) (map[uint32][]Validator, error) {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	return nodesConfig.eligibleMap, nil
}

// ComputeConsensusGroup will generate a list of validators based on the the eligible list,
// consensus group size and a randomness source
// Steps:
// 1. generate expanded eligible list by multiplying entries from shards' eligible list according to stake and rating -> TODO
// 2. for each value in [0, consensusGroupSize), compute proposedindex = Hash( [index as string] CONCAT randomness) % len(eligible list)
// 3. if proposed index is already in the temp validator list, then proposedIndex++ (and then % len(eligible list) as to not
//    exceed the maximum index value permitted by the validator list), and then recheck against temp validator list until
//    the item at the new proposed index is not found in the list. This new proposed index will be called checked index
// 4. the item at the checked index is appended in the temp validator list
func (ihgs *indexHashedNodesCoordinator) ComputeConsensusGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) (validatorsGroup []Validator, err error) {
	var eligibleShardList []Validator
	var mut *sync.RWMutex

	log.Trace("computing consensus group for",
		"epoch", epoch,
		"shardId", shardId,
		"randomness", randomness,
		"round", round)

	if randomness == nil {
		return nil, ErrNilRandomness
	}
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if ok {
		eligibleShardList = nodesConfig.eligibleMap[shardId]
		mut = &nodesConfig.mutNodesMaps
	}
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDesNotExist
	}
	if shardId >= nodesConfig.nbShards && shardId != core.MetachainShardId {
		return nil, ErrInvalidShardId
	}

	if ihgs == nil {
		return nil, ErrNilRandomness
	}

	key := []byte(fmt.Sprintf(keyFormat, string(randomness), round, shardId, epoch))
	validators := ihgs.searchConsensusForKey(key)
	if validators != nil {
		return validators, nil
	}

	consensusSize := ihgs.ConsensusGroupSize(shardId)
	randomness = []byte(fmt.Sprintf("%d-%s", round, randomness))

	// TODO: pre-compute eligible list and update only on rating change.
	expandedList := ihgs.doExpandEligibleList(eligibleShardList, mut)

	log.Debug("ComputeValidatorsGroup",
		"randomness", randomness,
		"consensus size", consensusSize,
		"eligible list length", len(expandedList))

	consensusGroupProvider := NewSelectionBasedProvider(ihgs.hasher, uint32(consensusSize))
	tempList, err := consensusGroupProvider.Get(randomness, int64(consensusSize), expandedList)
	if err != nil {
		return nil, err
	}

	ihgs.consensusGroupCacher.Put(key, tempList)

	return tempList, nil
}

func (ihgs *indexHashedNodesCoordinator) searchConsensusForKey(key []byte) []Validator {
	value, ok := ihgs.consensusGroupCacher.Get(key)
	if ok {
		consensusGroup, typeOk := value.([]Validator)
		if typeOk {
			return consensusGroup
		}

	}
	return nil
}

// GetValidatorWithPublicKey gets the validator with the given public key
func (ihgs *indexHashedNodesCoordinator) GetValidatorWithPublicKey(
	publicKey []byte,
	epoch uint32,
) (Validator, uint32, error) {
	if publicKey == nil {
		return nil, 0, ErrNilPubKey
	}
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, 0, ErrEpochNodesConfigDesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardId, shardEligible := range nodesConfig.eligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			if bytes.Equal(publicKey, shardEligible[i].PubKey()) {
				return shardEligible[i], shardId, nil
			}
		}
	}

	return nil, 0, ErrValidatorNotFound
}

// GetConsensusValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihgs *indexHashedNodesCoordinator) GetConsensusValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeConsensusGroup(randomness, round, shardId, epoch)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range consensusNodes {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetConsensusValidatorsRewardsAddresses calculates the validator consensus group for a specific shard, randomness and round
// number, returning their staking/rewards addresses
func (ihgs *indexHashedNodesCoordinator) GetConsensusValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeConsensusGroup(randomness, round, shardId, epoch)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, len(consensusNodes))
	for i, v := range consensusNodes {
		addresses[i] = string(v.Address())
	}

	return addresses, nil
}

// GetSelectedPublicKeys returns the stringified public keys of the marked validators in the selection bitmap
// TODO: This function needs to be revised when the requirements are clarified
func (ihgs *indexHashedNodesCoordinator) GetSelectedPublicKeys(
	selection []byte,
	shardId uint32,
	epoch uint32,
) (publicKeys []string, err error) {

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDesNotExist
	}

	if shardId >= nodesConfig.nbShards && shardId != core.MetachainShardId {
		return nil, ErrInvalidShardId
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	selectionLen := uint16(len(selection) * 8) // 8 selection bits in each byte
	shardEligibleLen := uint16(len(nodesConfig.eligibleMap[shardId]))
	invalidSelection := selectionLen < shardEligibleLen

	if invalidSelection {
		return nil, ErrEligibleSelectionMismatch
	}

	consensusSize := ihgs.ConsensusGroupSize(shardId)
	publicKeys = make([]string, consensusSize)
	cnt := 0

	for i := uint16(0); i < shardEligibleLen; i++ {
		isSelected := (selection[i/8] & (1 << (i % 8))) != 0

		if !isSelected {
			continue
		}

		publicKeys[cnt] = string(nodesConfig.eligibleMap[shardId][i].PubKey())
		cnt++

		if cnt > consensusSize {
			return nil, ErrEligibleTooManySelections
		}
	}

	if cnt < consensusSize {
		return nil, ErrEligibleTooFewSelections
	}

	return publicKeys, nil
}

// GetAllValidatorsPublicKeys will return all validators public keys for all shards
func (ihgs *indexHashedNodesCoordinator) GetAllValidatorsPublicKeys(epoch uint32) (map[uint32][][]byte, error) {
	validatorsPubKeys := make(map[uint32][][]byte)

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDesNotExist
	}

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	for shardId, shardEligible := range nodesConfig.eligibleMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardId] = append(validatorsPubKeys[shardId], nodesConfig.eligibleMap[shardId][i].PubKey())
		}
	}

	return validatorsPubKeys, nil
}

// GetValidatorsIndexes will return validators indexes for a block
func (ihgs *indexHashedNodesCoordinator) GetValidatorsIndexes(
	publicKeys []string,
	epoch uint32,
) ([]uint64, error) {
	signersIndexes := make([]uint64, 0)

	validatorsPubKeys, err := ihgs.GetAllValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	ihgs.mutNodesConfig.RLock()
	nodesConfig := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	for _, pubKey := range publicKeys {
		for index, value := range validatorsPubKeys[nodesConfig.shardId] {
			if bytes.Equal([]byte(pubKey), value) {
				signersIndexes = append(signersIndexes, uint64(index))
			}
		}
	}

	if len(publicKeys) != len(signersIndexes) {
		log.Error("public keys not found", "len pubKeys", len(publicKeys), "len signers", len(signersIndexes))
		return nil, ErrNotInvalidNumberPubKeys
	}

	return signersIndexes, nil
}

// EpochStartPrepare wis called when an epoch start event is observed, but not yet confirmed/committed.
// Some components may need to do some initialisation on this event
func (ihgs *indexHashedNodesCoordinator) EpochStartPrepare(metaHeader data.HeaderHandler) {
	randomness := metaHeader.GetPrevRandSeed()
	newEpoch := metaHeader.GetEpoch()

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[newEpoch-1]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		log.Error("no configured epoch found")
		return
	}

	// TODO: update the new nodes and leaving nodes as well
	shufflerArgs := ArgsUpdateNodes{
		Eligible: nodesConfig.eligibleMap,
		Waiting:  nodesConfig.waitingMap,
		NewNodes: make([]Validator, 0),
		Leaving:  make([]Validator, 0),
		Rand:     randomness,
		NbShards: nodesConfig.nbShards,
	}

	eligibleMap, waitingMap, _ := ihgs.shuffler.UpdateNodeLists(shufflerArgs)

	_ = ihgs.SetNodesPerShards(eligibleMap, waitingMap, newEpoch)
	err := ihgs.saveState(randomness)
	if err != nil {
		log.Error("saving nodes coordinator config failed", "error", err.Error())
	}
	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = randomness
	ihgs.mutSavedStateKey.Unlock()
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihgs *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - nodeCoordinatorStoredEpochs
	needToRemove := epochToRemove >= 0
	ihgs.currentEpoch = newEpoch

	ihgs.mutNodesConfig.Lock()
	if needToRemove {
		for epoch := range ihgs.nodesConfig {
			if epoch <= uint32(epochToRemove) {
				delete(ihgs.nodesConfig, epoch)
			}
		}
	}
	ihgs.mutNodesConfig.Unlock()
}

// GetSavedStateKey returns the key for the last nodes coordinator saved state
func (ihgs *indexHashedNodesCoordinator) GetSavedStateKey() []byte {
	ihgs.mutSavedStateKey.RLock()
	key := ihgs.savedStateKey
	ihgs.mutSavedStateKey.RUnlock()

	return key
}

// ShardIdForEpoch returns the nodesCoordinator configured ShardId for specified epoch if epoch configuration exists,
// otherwise error
func (ihgs *indexHashedNodesCoordinator) ShardIdForEpoch(epoch uint32) (uint32, error) {
	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return 0, ErrEpochNodesConfigDesNotExist
	}

	return nodesConfig.shardId, nil
}

// GetConsensusWhitelistedNodes return the whitelisted nodes allowed to send consensus messages, for each of the shards
func (ihgs *indexHashedNodesCoordinator) GetConsensusWhitelistedNodes(
	epoch uint32,
) (map[string]struct{}, error) {
	publicKeysPrevEpoch := make(map[uint32][][]byte)
	var err error

	if epoch > 0 {
		publicKeysPrevEpoch, err = ihgs.GetAllValidatorsPublicKeys(epoch - 1)
		if err != nil {
			return nil, err
		}
	}

	publicKeysNewEpoch, err := ihgs.GetAllValidatorsPublicKeys(epoch)
	if err != nil {
		return nil, err
	}

	estimatedMapSize := len(publicKeysNewEpoch) * len(publicKeysNewEpoch[0])
	shardEligible := make(map[string]struct{}, estimatedMapSize)

	prevEpochShardId, err := ihgs.ShardIdForEpoch(epoch - 1)
	if err == nil {
		for _, pubKey := range publicKeysPrevEpoch[prevEpochShardId] {
			shardEligible[string(pubKey)] = struct{}{}
		}
	} else {
		log.Debug("error getting shardId for epoch", "epoch", epoch-1, "error", err)
	}

	epochShardId, err := ihgs.ShardIdForEpoch(epoch)
	if err != nil {
		return nil, err
	}

	for _, pubKey := range publicKeysNewEpoch[epochShardId] {
		shardEligible[string(pubKey)] = struct{}{}
	}

	return shardEligible, nil
}

func (ihgs *indexHashedNodesCoordinator) expandEligibleList(validators []Validator, mut *sync.RWMutex) []Validator {
	//TODO implement an expand eligible list variant
	return validators
}

func (ihgs *indexHashedNodesCoordinator) computeShardForPublicKey(nodesConfig *epochNodesConfig) uint32 {
	pubKey := ihgs.selfPubKey
	selfShard := uint32(0)
	epochNodesConfig, ok := ihgs.nodesConfig[ihgs.currentEpoch]
	if ok {
		selfShard = epochNodesConfig.shardId
	}

	for shard, validators := range nodesConfig.eligibleMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				return shard
			}
		}
	}

	for shard, validators := range nodesConfig.waitingMap {
		for _, v := range validators {
			if bytes.Equal(v.PubKey(), pubKey) {
				return shard
			}
		}
	}

	return selfShard
}

// validatorIsInList returns true if a validator has been found in provided list
func (ihgs *indexHashedNodesCoordinator) validatorIsInList(v Validator, list []Validator) bool {
	for i := 0; i < len(list); i++ {
		if bytes.Equal(v.PubKey(), list[i].PubKey()) {
			return true
		}
	}

	return false
}

// ConsensusGroupSize returns the consensus group size for a specific shard
func (ihgs *indexHashedNodesCoordinator) ConsensusGroupSize(
	shardId uint32,
) int {
	if shardId == core.MetachainShardId {
		return ihgs.metaConsensusGroupSize
	}

	return ihgs.shardConsensusGroupSize
}

// GetNumTotalEligible returns the number of total eligible accross all shards from current setup
func (ihgs *indexHashedNodesCoordinator) GetNumTotalEligible() uint64 {
	return ihgs.numTotalEligible
}

// GetOwnPublicKey will return current node public key  for block sign
func (ihgs *indexHashedNodesCoordinator) GetOwnPublicKey() []byte {
	return ihgs.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	return ihgs == nil
}
