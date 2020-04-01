package sharding

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

var _ RandomSelector = (*selectorExpandedList)(nil)

type selectorExpandedList struct {
	validator         []Validator
	expandedList      []uint32
	hasher            hashing.Hasher
	uniqueItems       uint32
	consensusSelector *SelectionBasedProvider
}

const minWeight = 1

// NewSelectorExpandedList creates a new selector initializing selection set to the given lists of validators
// and expanding it according to each validator weight.
func NewSelectorExpandedList(validators []Validator, weightList []uint32, hasher hashing.Hasher) (*selectorExpandedList, error) {
	if len(validators) == 0 {
		return nil, ErrNilValidators
	}
	if len(weightList) == 0 {
		return nil, ErrNilWeights
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	uniqueItems := uint32(len(weightList))
	selectorConsensus := NewSelectionBasedProvider(hasher, uniqueItems)
	selector := &selectorExpandedList{
		validator:         validators,
		hasher:            hasher,
		uniqueItems:       uniqueItems,
		consensusSelector: selectorConsensus,
	}

	var err error
	selector.expandedList, err = selector.expandList(validators, weightList)
	if err != nil {
		return nil, err
	}

	return selector, nil
}

// Select selects from the preconfigured selection set sampleSize elements, returning the selected indexes from the set
func (s *selectorExpandedList) Select(randSeed []byte, sampleSize uint32) ([]uint32, error) {
	if len(randSeed) == 0 {
		return nil, ErrNilRandomness
	}
	if sampleSize == 0 || sampleSize > s.uniqueItems {
		return nil, ErrInvalidSampleSize
	}

	return s.consensusSelector.Get(randSeed, int64(sampleSize), s.expandedList)
}

func (s *selectorExpandedList) expandList(validators []Validator, weightList []uint32) ([]uint32, error) {
	minSize := len(validators)
	validatorList := make([]uint32, 0, minSize)

	for i := 0; i < len(validators); i++ {
		if weightList[i] < minWeight {
			return nil, ErrInvalidWeight
		}

		for j := uint32(0); j < weightList[i]; j++ {
			validatorList = append(validatorList, uint32(i))
		}
	}

	s.expandedList = validatorList

	return validatorList, nil
}

// IsInterfaceNil returns true if the receiver is nil
func (s *selectorExpandedList) IsInterfaceNil() bool {
	return s == nil
}
