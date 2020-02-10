package consensusGroupProviders

import (
	"bytes"

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

func NewSelectionBasedProvider() *selectionBasedProvider {
	return &selectionBasedProvider{
		sortedSlice: make([]*validatorEntry, 0),
		size:        0,
	}
}

func (sbp *selectionBasedProvider) addToSlice(ve *validatorEntry) {
	defer func() {
		sbp.size += ve.numAppearances
	}()

	if len(sbp.sortedSlice) == 0 {
		sbp.sortedSlice = append(sbp.sortedSlice, ve)
		return
	}

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
	startIdx, numAppearances := sbp.computeNumAppearancesForValidator(expElList, x)
	ve := &validatorEntry{
		startIndex:     startIdx,
		numAppearances: numAppearances,
	}
	sbp.addToSlice(ve)
}

func (sbp *selectionBasedProvider) Get(randomness uint64, numVal int64, originalExpEligibleList []sharding.Validator) ([]sharding.Validator, error) {
	expEligibleList := make([]sharding.Validator, len(originalExpEligibleList))
	copy(expEligibleList, originalExpEligibleList)
	valSlice := make([]sharding.Validator, 0)

	x := randomness % uint64(len(expEligibleList))
	valSlice = append(valSlice, expEligibleList[x])
	sbp.add(expEligibleList, int64(x))
	for i := int64(1); i < numVal; i++ {
		x = randomness % uint64(int64(len(expEligibleList))-sbp.size)
		sliceIdx := 0
		for {
			if sliceIdx == len(sbp.sortedSlice) {
				sbp.add(expEligibleList, int64(x))
				valSlice = append(valSlice, expEligibleList[x])
				break
			}
			if uint64(sbp.sortedSlice[sliceIdx].startIndex) <= x {
				x += uint64(sbp.sortedSlice[sliceIdx].numAppearances)
				sliceIdx++
			} else {
				sbp.add(expEligibleList, int64(x))
				valSlice = append(valSlice, expEligibleList[x])
				break
			}
		}
	}

	return valSlice, nil
}

func (sbp *selectionBasedProvider) computeNumAppearancesForValidator(expEligibleList []sharding.Validator, idx int64) (int64, int64) {
	val := expEligibleList[idx].Address()
	originalIdx := idx
	startIdx := int64(0)
	if idx == 0 {
		startIdx = 0
	} else {
		for {
			idx--
			if idx == 0 || !bytes.Equal(expEligibleList[idx].Address(), val) {
				startIdx = idx + 1
				break
			}
		}
	}

	idx = originalIdx
	var endIdx int64
	if idx == int64(len(expEligibleList)) {
		endIdx = idx
	} else {
		for {
			idx++
			if idx == int64(len(expEligibleList)) || !bytes.Equal(expEligibleList[idx].Address(), val) {
				endIdx = idx
				break
			}
		}
	}

	return startIdx, endIdx - startIdx
}
