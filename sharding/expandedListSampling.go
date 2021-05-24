package sharding

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

var _ RandomSelector = (*selectorExpandedList)(nil)

type selectorExpandedList struct {
	expandedList []uint32
	hasher       hashing.Hasher
	uniqueItems  uint32
}

const minWeight = 1

// NewSelectorExpandedList creates a new selector initializing selection set to the given lists of validators
// and expanding it according to each validator weight.
func NewSelectorExpandedList(weightList []uint32, hasher hashing.Hasher) (*selectorExpandedList, error) {
	if len(weightList) == 0 {
		return nil, ErrNilWeights
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	uniqueItems := uint32(len(weightList))
	selector := &selectorExpandedList{
		hasher:      hasher,
		uniqueItems: uniqueItems,
	}

	var err error
	selector.expandedList, err = selector.expandList(weightList)
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
	selectorConsensus := NewSelectionBasedProvider(s.hasher, s.uniqueItems)
	return selectorConsensus.Get(randSeed, int64(sampleSize), s.expandedList)
}

func (s *selectorExpandedList) expandList(weightList []uint32) ([]uint32, error) {
	minSize := len(weightList)
	expandedValidatorList := make([]uint32, 0, minSize)

	for i := 0; i < len(weightList); i++ {
		if weightList[i] < minWeight {
			return nil, ErrInvalidWeight
		}

		for j := uint32(0); j < weightList[i]; j++ {
			expandedValidatorList = append(expandedValidatorList, uint32(i))
		}
	}

	return expandedValidatorList, nil
}

// IsInterfaceNil returns true if the receiver is nil
func (s *selectorExpandedList) IsInterfaceNil() bool {
	return s == nil
}
