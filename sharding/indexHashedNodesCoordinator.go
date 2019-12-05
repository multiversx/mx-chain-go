package sharding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

type indexHashedNodesCoordinator struct {
	doExpandEligibleList    func(uint32) []Validator
	nbShards                uint32
	shardId                 uint32
	hasher                  hashing.Hasher
	nodesMap                map[uint32][]Validator
	shardConsensusGroupSize int
	metaConsensusGroupSize  int
	selfPubKey              []byte
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(arguments ArgNodesCoordinator) (*indexHashedNodesCoordinator, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	ihgs := &indexHashedNodesCoordinator{
		nbShards:                arguments.NbShards,
		shardId:                 arguments.ShardId,
		hasher:                  arguments.Hasher,
		nodesMap:                make(map[uint32][]Validator),
		shardConsensusGroupSize: arguments.ShardConsensusGroupSize,
		metaConsensusGroupSize:  arguments.MetaConsensusGroupSize,
		selfPubKey:              arguments.SelfPublicKey,
	}

	ihgs.doExpandEligibleList = ihgs.expandEligibleList

	err = ihgs.SetNodesPerShards(arguments.Nodes)
	if err != nil {
		return nil, err
	}

	return ihgs, nil
}

func checkArguments(arguments ArgNodesCoordinator) error {
	if arguments.ShardConsensusGroupSize < 1 || arguments.MetaConsensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}
	if arguments.NbShards < 1 {
		return ErrInvalidNumberOfShards
	}
	if arguments.ShardId >= arguments.NbShards && arguments.ShardId != MetachainShardId {
		return ErrInvalidShardId
	}
	if arguments.Hasher == nil {
		return ErrNilHasher
	}
	if arguments.SelfPublicKey == nil {
		return ErrNilPubKey
	}

	return nil
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinator) SetNodesPerShards(nodes map[uint32][]Validator) error {
	if nodes == nil {
		return ErrNilInputNodesMap
	}

	nodesList, ok := nodes[MetachainShardId]
	if ok && len(nodesList) < ihgs.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	for shardId := uint32(0); shardId < ihgs.nbShards; shardId++ {
		nbNodesShard := len(nodes[shardId])
		if nbNodesShard < ihgs.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
	}

	ihgs.nodesMap = nodes

	return nil
}

// GetNodesPerShard returns the nodes per shard map
func (ihgs *indexHashedNodesCoordinator) GetNodesPerShard() map[uint32][]Validator {
	return ihgs.nodesMap
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
) (validatorsGroup []Validator, err error) {
	if randomness == nil {
		return nil, ErrNilRandomness
	}

	if shardId >= ihgs.nbShards && shardId != MetachainShardId {
		return nil, ErrInvalidShardId
	}

	if ihgs == nil {
		return nil, ErrNilRandomness
	}

	tempList := make([]Validator, 0)
	consensusSize := ihgs.consensusGroupSize(shardId)
	randomness = []byte(fmt.Sprintf("%d-%s", round, core.ToB64(randomness)))

	// TODO: pre-compute eligible list and update only on rating change.
	expandedList := ihgs.doExpandEligibleList(shardId)
	lenExpandedList := len(expandedList)

	for startIdx := 0; startIdx < consensusSize; startIdx++ {
		proposedIndex := ihgs.computeListIndex(startIdx, lenExpandedList, string(randomness))
		checkedIndex := ihgs.checkIndex(proposedIndex, expandedList, tempList)
		tempList = append(tempList, expandedList[checkedIndex])
	}

	return tempList, nil
}

// GetValidatorWithPublicKey gets the validator with the given public key
func (ihgs *indexHashedNodesCoordinator) GetValidatorWithPublicKey(publicKey []byte) (Validator, uint32, error) {
	if publicKey == nil {
		return nil, 0, ErrNilPubKey
	}

	for shardId, shardEligible := range ihgs.nodesMap {
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
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness, round, shardId)
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
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness, round, shardId)
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
func (ihgs *indexHashedNodesCoordinator) GetSelectedPublicKeys(selection []byte, shardId uint32) (publicKeys []string, err error) {
	if shardId >= ihgs.nbShards && shardId != MetachainShardId {
		return nil, ErrInvalidShardId
	}

	selectionLen := uint16(len(selection) * 8) // 8 selection bits in each byte
	shardEligibleLen := uint16(len(ihgs.nodesMap[shardId]))
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

		publicKeys[cnt] = string(ihgs.nodesMap[shardId][i].PubKey())
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
func (ihgs *indexHashedNodesCoordinator) GetAllValidatorsPublicKeys() map[uint32][][]byte {
	validatorsPubKeys := make(map[uint32][][]byte)

	for shardId, shardEligible := range ihgs.nodesMap {
		for i := 0; i < len(shardEligible); i++ {
			validatorsPubKeys[shardId] = append(validatorsPubKeys[shardId], ihgs.nodesMap[shardId][i].PubKey())
		}
	}

	return validatorsPubKeys
}

// GetValidatorsIndexes will return validators indexes for a block
func (ihgs *indexHashedNodesCoordinator) GetValidatorsIndexes(publicKeys []string) []uint64 {
	validatorsPubKeys := ihgs.GetAllValidatorsPublicKeys()
	signersIndexes := make([]uint64, 0)

	for _, pubKey := range publicKeys {
		for index, value := range validatorsPubKeys[ihgs.shardId] {
			if bytes.Equal([]byte(pubKey), value) {
				signersIndexes = append(signersIndexes, uint64(index))
			}
		}
	}

	return signersIndexes
}

func (ihgs *indexHashedNodesCoordinator) expandEligibleList(shardId uint32) []Validator {
	//TODO implement an expand eligible list variant
	return ihgs.nodesMap[shardId]
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

func (ihgs *indexHashedNodesCoordinator) consensusGroupSize(shardId uint32) int {
	if shardId == MetachainShardId {
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
