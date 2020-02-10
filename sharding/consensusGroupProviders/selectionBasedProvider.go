package consensusGroupProviders

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type validatorEntry struct {
	startIndex     int64
	numAppearances int64
}

type selectionBasedProvider struct {
	sortedSlice []*validatorEntry
	size        int64
}

func NewSelectionBasedProvider(maxSize uint32) *selectionBasedProvider {
	return &selectionBasedProvider{
		sortedSlice: make([]*validatorEntry, 0, maxSize),
		size:        0,
	}
}

func (sbp *selectionBasedProvider) addToSortedSlice(ve *validatorEntry) {
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

func (sbp *selectionBasedProvider) add(expElList []sharding.Validator, x int64) {
	startIdx, numAppearances := computeNumAppearancesForValidator(expElList, x)
	ve := &validatorEntry{
		startIndex:     startIdx,
		numAppearances: numAppearances,
	}
	sbp.addToSortedSlice(ve)
}

func (sbp *selectionBasedProvider) Get(randomness uint64, numVal int64, originalExpEligibleList []sharding.Validator) ([]sharding.Validator, error) {
	valSlice := make([]sharding.Validator, 0, numVal)
	var x uint64
	lenExpandedList := int64(len(originalExpEligibleList))

	for i := int64(0); i < numVal; i++ {
		x = randomness % uint64(lenExpandedList-sbp.size)
		x = sbp.adjustIndex(x)
		valSlice = append(valSlice, originalExpEligibleList[x])
		sbp.add(originalExpEligibleList, int64(x))
	}

	return valSlice, nil
}

func (sbp *selectionBasedProvider) adjustIndex(index uint64) uint64 {
	for sliceIdx := 0; sliceIdx < len(sbp.sortedSlice); sliceIdx++ {
		if uint64(sbp.sortedSlice[sliceIdx].startIndex) > index {
			break
		}
		index += uint64(sbp.sortedSlice[sliceIdx].numAppearances)
	}

	return index
}
