package groupSelectors

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

type indexHashedGroupSelector struct {
	hasher               hashing.Hasher
	eligibleList         []consensus.Validator
	expandedEligibleList []consensus.Validator
	consensusGroupSize   int
}

// NewIndexHashedGroupSelector creates a new index hashed group selector
func NewIndexHashedGroupSelector(consensusGroupSize int, hasher hashing.Hasher) (*indexHashedGroupSelector, error) {
	if hasher == nil {
		return nil, consensus.ErrNilHasher
	}

	ihgs := &indexHashedGroupSelector{
		hasher:               hasher,
		eligibleList:         make([]consensus.Validator, 0),
		expandedEligibleList: make([]consensus.Validator, 0),
	}

	err := ihgs.SetConsensusGroupSize(consensusGroupSize)
	if err != nil {
		return nil, err
	}

	return ihgs, nil
}

// LoadEligibleList loads the eligible list
func (ihgs *indexHashedGroupSelector) LoadEligibleList(eligibleList []consensus.Validator) error {
	if eligibleList == nil {
		return consensus.ErrNilInputSlice
	}

	ihgs.eligibleList = make([]consensus.Validator, len(eligibleList))
	copy(ihgs.eligibleList, eligibleList)
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
func (ihgs *indexHashedGroupSelector) ComputeValidatorsGroup(randomness []byte) (validatorsGroup []consensus.Validator, err error) {
	if len(ihgs.eligibleList) < ihgs.consensusGroupSize {
		return nil, consensus.ErrSmallEligibleListSize
	}

	if randomness == nil {
		return nil, consensus.ErrNilRandomness
	}

	ihgs.expandedEligibleList = ihgs.expandEligibleList()

	tempList := make([]consensus.Validator, 0)

	for startIdx := 0; startIdx < ihgs.consensusGroupSize; startIdx++ {
		proposedIndex := ihgs.computeListIndex(startIdx, string(randomness))

		checkedIndex := ihgs.checkIndex(proposedIndex, tempList)
		tempList = append(tempList, ihgs.expandedEligibleList[checkedIndex])
	}

	return tempList, nil
}

// GetSelectedPublicKeys returns the stringified public keys of the marked validators in the selection bitmap
// TODO: This function needs to be revised when the requirements are clarified
func (ihgs *indexHashedGroupSelector) GetSelectedPublicKeys(selection []byte) (publicKeys []string, err error) {
	selectionLen := uint16(len(selection) * 8) // 8 selection bits in each byte
	shardEligibleLen := uint16(len(ihgs.eligibleList))
	invalidSelection := selectionLen < shardEligibleLen

	if invalidSelection {
		return nil, consensus.ErrEligibleSelectionMismatch
	}

	publicKeys = make([]string, ihgs.consensusGroupSize)
	cnt := 0

	for i := uint16(0); i < shardEligibleLen; i++ {
		isSelected := (selection[i/8] & (1 << (i % 8))) != 0

		if !isSelected {
			continue
		}

		publicKeys[cnt] = string(ihgs.eligibleList[i].PubKey())
		cnt++

		if cnt > ihgs.consensusGroupSize {
			return nil, consensus.ErrEligibleTooManySelections
		}
	}

	if cnt < ihgs.consensusGroupSize {
		return nil, consensus.ErrEligibleTooFewSelections
	}

	return publicKeys, nil
}

func (ihgs *indexHashedGroupSelector) expandEligibleList() []consensus.Validator {
	//TODO implement an expand eligible list variant
	return ihgs.eligibleList
}

// computeListIndex computes a proposed index from expanded eligible list
func (ihgs *indexHashedGroupSelector) computeListIndex(currentIndex int, randomSource string) int {
	buffCurrentIndex := make([]byte, 8)
	binary.BigEndian.PutUint64(buffCurrentIndex, uint64(currentIndex))

	indexHash := ihgs.hasher.Compute(string(buffCurrentIndex) + randomSource)

	computedLargeIndex := big.NewInt(0)
	computedLargeIndex.SetBytes(indexHash)

	// computedListIndex = computedLargeIndex % len(expandedEligibleList)
	computedListIndex := big.NewInt(0).Mod(computedLargeIndex, big.NewInt(int64(len(ihgs.expandedEligibleList)))).Int64()
	return int(computedListIndex)
}

// checkIndex returns a checked index starting from a proposed index
func (ihgs *indexHashedGroupSelector) checkIndex(proposedIndex int, selectedList []consensus.Validator) int {

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
func (ihgs *indexHashedGroupSelector) validatorIsInList(v consensus.Validator, list []consensus.Validator) bool {
	for i := 0; i < len(list); i++ {
		if bytes.Equal(v.PubKey(), list[i].PubKey()) {
			return true
		}
	}

	return false
}

// ConsensusGroupSize returns the consensus group size
func (ihgs *indexHashedGroupSelector) ConsensusGroupSize() int {
	return ihgs.consensusGroupSize
}

// SetConsensusGroupSize sets the consensus group size
func (ihgs *indexHashedGroupSelector) SetConsensusGroupSize(consensusGroupSize int) error {
	if consensusGroupSize < 1 {
		return consensus.ErrInvalidConsensusGroupSize
	}

	ihgs.consensusGroupSize = consensusGroupSize
	return nil
}
