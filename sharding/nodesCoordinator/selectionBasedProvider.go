package nodesCoordinator

import (
	"encoding/binary"
	"sync"

	"github.com/multiversx/mx-chain-core-go/hashing"
)

type validatorEntry struct {
	startIndex     int64
	numAppearances int64
}

// SelectionBasedProvider will handle the returning of the consensus group by simulating a reslicing of the expanded
// eligible list. A comparison between a real reslicing and this can be found in common_test.go
type SelectionBasedProvider struct {
	hasher       hashing.Hasher
	sortedSlice  []*validatorEntry
	size         int64
	mutSelection sync.Mutex
}

// NewSelectionBasedProvider will return a new instance of SelectionBasedProvider
func NewSelectionBasedProvider(hasher hashing.Hasher, maxSize uint32) *SelectionBasedProvider {
	return &SelectionBasedProvider{
		hasher:      hasher,
		sortedSlice: make([]*validatorEntry, 0, maxSize),
		size:        0,
	}
}

func (sbp *SelectionBasedProvider) addToSortedSlice(ve *validatorEntry) {
	sbp.size += ve.numAppearances

	for i := 0; i < len(sbp.sortedSlice); i++ {
		if sbp.sortedSlice[i].startIndex >= ve.startIndex {
			sbp.sortedSlice = append(sbp.sortedSlice[:i], append([]*validatorEntry{ve}, sbp.sortedSlice[i:]...)...)
			return
		}
	}

	// add to last position
	sbp.sortedSlice = append(sbp.sortedSlice, ve)
}

func (sbp *SelectionBasedProvider) add(expElList []uint32, index int64) {
	startIdx, numAppearances := computeStartIndexAndNumAppearancesForValidator(expElList, index)
	ve := &validatorEntry{
		startIndex:     startIdx,
		numAppearances: numAppearances,
	}
	sbp.addToSortedSlice(ve)
}

// Get will return the consensus group based on the randomness.
// After an item is chosen, it is added to the slice, so it won't be selected again so next time a new item
// is needed, the index is recalculated until the validator doesn't exist in that slice
func (sbp *SelectionBasedProvider) Get(randomness []byte, numValidators int64, expandedEligibleList []uint32) ([]uint32, error) {
	if len(randomness) == 0 {
		return nil, ErrNilRandomness
	}
	if numValidators > int64(len(expandedEligibleList)) {
		return nil, ErrInvalidSampleSize
	}

	validators := make([]uint32, 0, numValidators)
	var index uint64
	lenExpandedList := int64(len(expandedEligibleList))

	sbp.mutSelection.Lock()
	defer sbp.mutSelection.Unlock()
	defer sbp.clean()

	for i := int64(0); i < numValidators; i++ {
		newRandomness := sbp.computeRandomnessAsUint64(randomness, int(i))
		if sbp.size >= lenExpandedList {
			return nil, ErrInvalidSampleSize
		}
		index = newRandomness % uint64(lenExpandedList-sbp.size)
		index = sbp.adjustIndex(index)
		validators = append(validators, expandedEligibleList[index])
		sbp.add(expandedEligibleList, int64(index))
	}

	return validators, nil
}

func (sbp *SelectionBasedProvider) clean() {
	sbp.size = 0
	sbp.sortedSlice = make([]*validatorEntry, 0, len(sbp.sortedSlice))
}

func (sbp *SelectionBasedProvider) computeRandomnessAsUint64(randomness []byte, index int) uint64 {
	buffCurrentIndex := make([]byte, 8)
	binary.BigEndian.PutUint64(buffCurrentIndex, uint64(index))

	indexHash := sbp.hasher.Compute(string(buffCurrentIndex) + string(randomness))

	randomnessAsUint64 := binary.BigEndian.Uint64(indexHash)

	return randomnessAsUint64
}

func (sbp *SelectionBasedProvider) adjustIndex(index uint64) uint64 {
	for _, entry := range sbp.sortedSlice {
		if uint64(entry.startIndex) > index {
			break
		}
		index += uint64(entry.numAppearances)
	}

	return index
}
