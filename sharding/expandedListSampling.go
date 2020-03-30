package sharding

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

type selectorExpandedList struct {
	validator         []Validator
	expandedList      []uint32
	hasher            hashing.Hasher
	uniqueItems       uint32
	selectorConsensus *SelectionBasedProvider
}

// NewSelectorExpandedList creates a new selector initializing selection set to the given lists of validators
// and expanding it according to each validator weight.
func NewSelectorExpandedList(validators []Validator, weightList []uint32, hasher hashing.Hasher) (RandomSelector, error) {
	if len(validators) == 0 {
		return nil, ErrNilValidators
	}
	if len(weightList) == 0 {
		return nil, ErrNilParam
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
		selectorConsensus: selectorConsensus,
	}

	var err error
	selector.expandedList, err = selector.expandList(validators, weightList)

	return selector, err
}

// Select selects from the preconfigured selection set sampleSize elements, returning the selected indexes from the set
func (s *selectorExpandedList) Select(randSeed []byte, sampleSize uint32) ([]uint32, error) {
	if len(randSeed) == 0 {
		return nil, ErrNilRandomness
	}
	if sampleSize > s.uniqueItems {
		return nil, ErrInvalidSampleSize
	}

	return s.selectorConsensus.Get(randSeed, int64(sampleSize), s.expandedList)
}

func (s *selectorExpandedList) expandList(validators []Validator, weightList []uint32) ([]uint32, error) {
	minSize := len(validators)
	validatorList := make([]uint32, 0, minSize)

	for i := 0; i < len(validators); i++ {
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
