package nodesCoordinator

import (
	"encoding/binary"
	"math"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
)

const maxUint64 = 0xFFFFFFFFFFFFFFFF

var _ RandomSelector = (*selectorWRS)(nil)

type keyIndex struct {
	key   float64
	index uint32
}

// Weighted random selection selector
type selectorWRS struct {
	weights []uint32
	hasher  hashing.Hasher
}

// NewSelectorWRS creates a new selector initializing selection set to the given lists of weights
func NewSelectorWRS(weightList []uint32, hasher hashing.Hasher) (*selectorWRS, error) {
	if len(weightList) == 0 {
		return nil, ErrNilWeights
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	selector := &selectorWRS{
		weights: weightList,
		hasher:  hasher,
	}

	return selector, nil
}

// Select selects from the preconfigured selection set sampleSize elements, returning the selected indexes from the set
func (s *selectorWRS) Select(randSeed []byte, sampleSize uint32) ([]uint32, error) {
	if len(randSeed) == 0 {
		return nil, ErrNilRandomness
	}
	if sampleSize > uint32(len(s.weights)) {
		return nil, ErrInvalidSampleSize
	}

	keyIndexList := make([]*keyIndex, len(s.weights))
	buffIndex := make([]byte, 8)
	for i, w := range s.weights {
		if w < minWeight {
			return nil, ErrInvalidWeight
		}
		binary.BigEndian.PutUint64(buffIndex, uint64(i))
		indexHash := s.hasher.Compute(string(buffIndex) + string(randSeed))
		randomUint64 := binary.BigEndian.Uint64(indexHash)
		pow := 1.0 / float64(w)
		key := math.Pow(float64(randomUint64)/maxUint64, pow)
		keyIndexList[i] = &keyIndex{
			key:   key,
			index: uint32(i),
		}
	}
	sort.SliceStable(keyIndexList, func(i, j int) bool {
		return keyIndexList[i].key > keyIndexList[j].key
	})

	ret := make([]uint32, sampleSize)
	for i := uint32(0); i < sampleSize; i++ {
		ret[i] = keyIndexList[i].index
	}

	return ret, nil
}

// IsInterfaceNil returns true if the receiver is nil
func (s *selectorWRS) IsInterfaceNil() bool {
	return s == nil
}
