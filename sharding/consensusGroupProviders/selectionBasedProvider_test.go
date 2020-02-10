package consensusGroupProviders

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
	"math/big"
	"math/rand"
	"testing"
)

func TestSelectionBasedProvider_AddToSortedSlice(t *testing.T) {
	sbp := NewSelectionBasedProvider()

	v1 := &validatorEntry{
		startIndex:     0,
		numAppearances: 2,
	}
	v2 := &validatorEntry{
		startIndex:     7,
		numAppearances: 2,
	}
	v3 := &validatorEntry{
		startIndex:     4,
		numAppearances: 3,
	}
	v4 := &validatorEntry{
		startIndex:     5,
		numAppearances: 3,
	}
	v5 := &validatorEntry{
		startIndex:     12,
		numAppearances: 3,
	}
	v6 := &validatorEntry{
		startIndex:     9,
		numAppearances: 3,
	}
	v7 := &validatorEntry{
		startIndex:     5,
		numAppearances: 3,
	}

	sbp.addToSlice(v1)
	sbp.addToSlice(v2)
	sbp.addToSlice(v3)
	sbp.addToSlice(v4)
	sbp.addToSlice(v5)
	sbp.addToSlice(v6)
	sbp.addToSlice(v7)

	lastIndex := sbp.sortedSlice[0].startIndex
	for i := 1; i < len(sbp.sortedSlice); i++ {
		if sbp.sortedSlice[i].startIndex < lastIndex {
			assert.Fail(t, "slice is not sorted.")
		}
	}
}

func TestSelectionBasedProvider_Get(t *testing.T) {
	sbp := NewSelectionBasedProvider()

	numVals := 7
	randomness := uint64(4567890)
	expElList := getExpandedEligibleList(17)
	res, err := sbp.Get(randomness, int64(numVals), expElList)
	assert.Nil(t, err)
	assert.Equal(t, numVals, len(res))
	displayVals(res)
}
