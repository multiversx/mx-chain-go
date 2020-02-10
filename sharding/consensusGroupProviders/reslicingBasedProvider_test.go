package consensusGroupProviders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReslicingBasedProvider_Get(t *testing.T) {
	rbp := NewReslicingBasedProvider()

	numVals := 5
	randomness := uint64(4567890020202020)
	expElList := getExpandedEligibleList(8)
	res, err := rbp.Get(randomness, int64(numVals), expElList)
	assert.Nil(t, err)
	assert.Equal(t, numVals, len(res))
}
