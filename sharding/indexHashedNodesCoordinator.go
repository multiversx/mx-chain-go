package sharding

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/hashing"
)

type indexHashedNodesCoordinator struct {
	nbShards             uint32
	shardId              uint32
	hasher               hashing.Hasher
	nodesMap             map[uint32][]Validator
	expandedEligibleList []Validator
	consensusGroupSize   int
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinator(
	consensusGroupSize int,
	hasher hashing.Hasher,
	shardId uint32,
	nbShards uint32,
) (*indexHashedNodesCoordinator, error) {
	if consensusGroupSize < 1 {
		return nil, ErrInvalidConsensusGroupSize
	}

	if nbShards < 1 {
		return nil, ErrInvalidNumberOfShards
	}

	if shardId >= nbShards && shardId != MetachainShardId {
		return nil, ErrInvalidShardId
	}

	if hasher == nil {
		return nil, ErrNilHasher
	}

	ihgs := &indexHashedNodesCoordinator{
		nbShards:             nbShards,
		shardId:              shardId,
		hasher:               hasher,
		nodesMap:             make(map[uint32][]Validator),
		expandedEligibleList: make([]Validator, 0),
	}

	err := ihgs.SetConsensusGroupSize(consensusGroupSize)
	if err != nil {
		return nil, err
	}

	return ihgs, nil
}

// LoadNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinator) LoadNodesPerShards(nodes map[uint32][]Validator) error {
	if nodes == nil {
		return ErrNilInputNodesMap
	}

	ihgs.nodesMap = nodes
	ihgs.expandedEligibleList = ihgs.expandEligibleList()

	return nil
}

// ComputeValidatorsGroup will generate a list of validators based on the the eligible list,
// consensus group size and a randomness source
// Steps:
// 1. generate expanded eligible list by multiplying entries from eligible list according to stake and rating -> TODO
// 2. for each value in [0, consensusGroupSize), compute proposedindex = Hash( [index as string] CONCAT randomness) % len(eligible list)
// 3. if proposed index is already in the temp validator list, then proposedIndex++ (and then % len(eligible list) as to not
//    exceed the maximum index value permitted by the validator list), and then recheck against temp validator list until
//    the item at the new proposed index is not found in the list. This new proposed index will be called checked index
// 4. the item at the checked index is appended in the temp validator list
func (ihgs *indexHashedNodesCoordinator) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []Validator, err error) {
	if len(ihgs.nodesMap[ihgs.shardId]) < ihgs.consensusGroupSize {
		return nil, ErrSmallEligibleListSize
	}

	if randomness == nil {
		return nil, ErrNilRandomness
	}

	tempList := make([]Validator, 0)

	for startIdx := 0; startIdx < ihgs.consensusGroupSize; startIdx++ {
		proposedIndex := ihgs.computeListIndex(startIdx, string(randomness))

		checkedIndex := ihgs.checkIndex(proposedIndex, tempList)
		tempList = append(tempList, ihgs.expandedEligibleList[checkedIndex])
	}

	return tempList, nil
}

// GetSelectedValidatorsPublicKeys calculates the validators group for a specific randomness,
// returning their public keys
func (ihgs *indexHashedNodesCoordinator) GetValidatorsPublicKeys(randomness []byte) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range consensusNodes {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetSelectedPublicKeys returns the stringified public keys of the marked validators in the selection bitmap
// TODO: This function needs to be revised when the requirements are clarified
func (ihgs *indexHashedNodesCoordinator) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	selectionLen := uint16(len(selection) * 8) // 8 selection bits in each byte
	shardEligibleLen := uint16(len(ihgs.nodesMap[ihgs.shardId]))
	invalidSelection := selectionLen < shardEligibleLen

	if invalidSelection {
		return nil, ErrEligibleSelectionMismatch
	}

	publicKeys = make([]string, ihgs.consensusGroupSize)
	cnt := 0

	for i := uint16(0); i < shardEligibleLen; i++ {
		isSelected := (selection[i/8] & (1 << (i % 8))) != 0

		if !isSelected {
			continue
		}

		publicKeys[cnt] = string(ihgs.nodesMap[ihgs.shardId][i].PubKey())
		cnt++

		if cnt > ihgs.consensusGroupSize {
			return nil, ErrEligibleTooManySelections
		}
	}

	if cnt < ihgs.consensusGroupSize {
		return nil, ErrEligibleTooFewSelections
	}

	return publicKeys, nil
}

func (ihgs *indexHashedNodesCoordinator) expandEligibleList() []Validator {
	//TODO implement an expand eligible list variant
	return ihgs.nodesMap[ihgs.shardId]
}

// computeListIndex computes a proposed index from expanded eligible list
func (ihgs *indexHashedNodesCoordinator) computeListIndex(currentIndex int, randomSource string) int {
	buffCurrentIndex := make([]byte, 8)
	binary.BigEndian.PutUint64(buffCurrentIndex, uint64(currentIndex))

	indexHash := ihgs.hasher.Compute(string(buffCurrentIndex) + randomSource)

	computedLargeIndex := big.NewInt(0)
	computedLargeIndex.SetBytes(indexHash)
	lenExpandedEligibleList := big.NewInt(int64(len(ihgs.expandedEligibleList)))

	// computedListIndex = computedLargeIndex % len(expandedEligibleList)
	computedListIndex := big.NewInt(0).Mod(computedLargeIndex, lenExpandedEligibleList).Int64()

	return int(computedListIndex)
}

// checkIndex returns a checked index starting from a proposed index
func (ihgs *indexHashedNodesCoordinator) checkIndex(proposedIndex int, selectedList []Validator) int {

	for {
		v := ihgs.expandedEligibleList[proposedIndex]

		if ihgs.validatorIsInList(v, selectedList) {
			proposedIndex++
			proposedIndex = proposedIndex % len(ihgs.expandedEligibleList)
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

// ConsensusGroupSize returns the consensus group size
func (ihgs *indexHashedNodesCoordinator) ConsensusGroupSize() int {
	return ihgs.consensusGroupSize
}

// SetConsensusGroupSize sets the consensus group size
func (ihgs *indexHashedNodesCoordinator) SetConsensusGroupSize(consensusGroupSize int) error {
	if consensusGroupSize < 1 {
		return ErrInvalidConsensusGroupSize
	}

	ihgs.consensusGroupSize = consensusGroupSize
	return nil
}
