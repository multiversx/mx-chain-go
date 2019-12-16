package sharding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

// TODO: move this to config parameters
const nodeCoordinatorStoredEpochs = 2

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
	doExpandEligibleList    func(uint32) []Validator
	hasher                  hashing.Hasher
	shuffler                NodesShuffler
	epochStartSubscriber    EpochStartSubscriber
	selfPubKey              []byte
	nodesConfig             map[uint32]*epochNodesConfig
	mutNodesConfig          sync.RWMutex
	currentEpoch            uint32
	shardConsensusGroupSize int
	metaConsensusGroupSize  int
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

	ihgs := &indexHashedNodesCoordinator{
		hasher:                  arguments.Hasher,
		shuffler:                arguments.Shuffler,
		epochStartSubscriber:    arguments.EpochStartSubscriber,
		selfPubKey:              arguments.SelfPublicKey,
		nodesConfig:             nodesConfig,
		currentEpoch:            arguments.Epoch,
		shardConsensusGroupSize: arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:  arguments.MetaConsensusGroupSize,
	}

	ihgs.doExpandEligibleList = ihgs.expandEligibleList

	err = ihgs.SetNodesPerShards(arguments.EligibleNodes, arguments.WaitingNodes, arguments.Epoch)
	if err != nil {
		return nil, err
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
	if ok && len(nodesList) < ihgs.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	for shardId := uint32(0); shardId < uint32(len(eligible)-1); shardId++ {
		nbNodesShard := len(eligible[shardId])
		if nbNodesShard < ihgs.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
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

// ComputeValidatorsGroup will generate a list of validators based on the the eligible list,
// consensus group size and a randomness source
// Steps:
// 1. generate expanded eligible list by multiplying entries from shards' eligible list according to stake and rating -> TODO
// 2. for each value in [0, consensusGroupSize), compute proposedindex = Hash( [index as string] CONCAT randomness) % len(eligible list)
// 3. if proposed index is already in the temp validator list, then proposedIndex++ (and then % len(eligible list) as to not
//    exceed the maximum index value permitted by the validator list), and then recheck against temp validator list until
//    the item at the new proposed index is not found in the list. This new proposed index will be called checked index
// 4. the item at the checked index is appended in the temp validator list
func (ihgs *indexHashedNodesCoordinator) ComputeValidatorsGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) (validatorsGroup []Validator, err error) {
	if randomness == nil {
		return nil, ErrNilRandomness
	}

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[epoch]
	ihgs.mutNodesConfig.RUnlock()

	if !ok {
		return nil, ErrEpochNodesConfigDesNotExist
	}
	if shardId >= nodesConfig.nbShards && shardId != core.MetachainShardId {
		return nil, ErrInvalidShardId
	}

	tempList := make([]Validator, 0)
	consensusSize := ihgs.consensusGroupSize(shardId)
	randomness = []byte(fmt.Sprintf("%d-%s", round, core.ToB64(randomness)))

	nodesConfig.mutNodesMaps.RLock()
	defer nodesConfig.mutNodesMaps.RUnlock()

	// TODO: pre-compute eligible list and update only on rating change.
	expandedList, _ := ihgs.doExpandEligibleList(shardId, epoch)
	lenExpandedList := len(expandedList)

	for startIdx := 0; startIdx < consensusSize; startIdx++ {
		proposedIndex := ihgs.computeListIndex(startIdx, lenExpandedList, string(randomness))
		checkedIndex := ihgs.checkIndex(proposedIndex, expandedList, tempList)
		tempList = append(tempList, expandedList[checkedIndex])
	}

	return tempList, nil
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

// GetValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihgs *indexHashedNodesCoordinator) GetValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness, round, shardId, epoch)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range consensusNodes {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetValidatorsRewardsAddresses calculates the validator consensus group for a specific shard, randomness and round
// number, returning their staking/rewards addresses
func (ihgs *indexHashedNodesCoordinator) GetValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
	epoch uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness, round, shardId, epoch)
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

	consensusSize := ihgs.consensusGroupSize(shardId)
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

	return signersIndexes, nil
}

// EpochStartAction is called upon a start of epoch event.
// NodeCoordinator has to get the nodes assignment to shards using the shuffler.
func (ihgs *indexHashedNodesCoordinator) EpochStartAction(hdr data.HeaderHandler) {
	randomness := hdr.GetRandSeed()
	newEpoch := hdr.GetEpoch()
	epochToRemove := int32(newEpoch) - nodeCoordinatorStoredEpochs
	needToRemove := epochToRemove >= 0

	ihgs.mutNodesConfig.RLock()
	nodesConfig, ok := ihgs.nodesConfig[ihgs.currentEpoch]
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

	err := ihgs.SetNodesPerShards(eligibleMap, waitingMap, newEpoch)
	if err != nil {
		return
	}

	ihgs.currentEpoch = newEpoch

	ihgs.mutNodesConfig.Lock()
	if needToRemove {
		delete(ihgs.nodesConfig, uint32(epochToRemove))
	}
	ihgs.mutNodesConfig.Unlock()
}

func (ihgs *indexHashedNodesCoordinator) expandEligibleList(shardId uint32, epoch uint32) ([]Validator, error) {
	ihgs.mutNodesConfig.RLock()
	defer ihgs.mutNodesConfig.RUnlock()

	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if !ok {
		return nil, ErrEpochNodesConfigDesNotExist
	}

	//TODO implement an expand eligible list variant
	return nodesConfig.eligibleMap[shardId], nil
}

func (ihgs *indexHashedNodesCoordinator) computeShardForPublicKey(nodesConfig *epochNodesConfig) uint32 {
	pubKey := ihgs.selfPubKey
	selfShard := nodesConfig.shardId

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

// computeListIndex computes a proposed index from expanded eligible list
func (ihgs *indexHashedNodesCoordinator) computeListIndex(currentIndex int, lenList int, randomSource string) int {
	buffCurrentIndex := make([]byte, 8)
	binary.BigEndian.PutUint64(buffCurrentIndex, uint64(currentIndex))

	indexHash := ihgs.hasher.Compute(string(buffCurrentIndex) + randomSource)

	computedLargeIndex := big.NewInt(0)
	computedLargeIndex.SetBytes(indexHash)
	lenExpandedEligibleList := big.NewInt(int64(lenList))

	// computedListIndex = computedLargeIndex % len(expandedEligibleList)
	computedListIndex := big.NewInt(0).Mod(computedLargeIndex, lenExpandedEligibleList).Int64()

	return int(computedListIndex)
}

// checkIndex returns a checked index starting from a proposed index
func (ihgs *indexHashedNodesCoordinator) checkIndex(
	proposedIndex int,
	eligibleList []Validator,
	selectedList []Validator,
) int {
	for {
		v := eligibleList[proposedIndex]

		if ihgs.validatorIsInList(v, selectedList) {
			proposedIndex++
			proposedIndex = proposedIndex % len(eligibleList)
			continue
		}

		return proposedIndex
	}
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

func (ihgs *indexHashedNodesCoordinator) consensusGroupSize(
	shardId uint32,
) int {
	if shardId == core.MetachainShardId {
		return ihgs.metaConsensusGroupSize
	}

	return ihgs.shardConsensusGroupSize
}

// GetOwnPublicKey will return current node public key  for block sign
func (ihgs *indexHashedNodesCoordinator) GetOwnPublicKey() []byte {
	return ihgs.selfPubKey
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihgs *indexHashedNodesCoordinator) IsInterfaceNil() bool {
	if ihgs == nil {
		return true
	}
	return false
}
