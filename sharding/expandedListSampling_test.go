package sharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/stretchr/testify/require"
)

func TestNewSelectorExpandedListNilWeightsShouldErr(t *testing.T) {
	t.Parallel()

	hasher := sha256.NewSha256()

	selector, err := NewSelectorExpandedList(nil, hasher)
	require.Equal(t, ErrNilWeights, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)

	selector, err := NewSelectorExpandedList(weights, nil)
	require.Equal(t, ErrNilHasher, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListInvalidWeightShouldErr(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	weights[1] = 0
	hasher := sha256.NewSha256()

	selector, err := NewSelectorExpandedList(weights, hasher)
	require.Equal(t, ErrInvalidWeight, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListOK(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightArithmeticProgression)
	hasher := sha256.NewSha256()

	selector, err := NewSelectorExpandedList(weights, hasher)
	require.Nil(t, err)
	require.NotNil(t, selector)

	for i := uint32(0); i < uint32(len(weights)); i++ {
		offset := uint32(0)
		if i > 0 {
			offset = weights[i-1]
		}
		for j := i + offset; j < weights[i]; j++ {
			require.Equal(t, selector.expandedList[j], i)
		}
	}
}

func TestSelectorExpandedList_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var selector RandomSelector
	require.True(t, check.IfNil(selector))

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()

	selector, _ = NewSelectorExpandedList(weights, hasher)
	require.False(t, check.IfNil(selector))
}

func TestSelectorExpandedList_SelectNilRandomessShouldErr(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()

	selector, _ := NewSelectorExpandedList(weights, hasher)
	indexes, err := selector.Select(nil, 5)

	require.Equal(t, ErrNilRandomness, err)
	require.Nil(t, indexes)
}

func TestSelectorExpandedList_Select0SampleSizeShouldErr(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()

	selector, _ := NewSelectorExpandedList(weights, hasher)
	indexes, err := selector.Select([]byte("random"), 0)

	require.Equal(t, ErrInvalidSampleSize, err)
	require.Nil(t, indexes)
}

func TestSelectorExpandedList_SelectSampleSizeGreaterThanSetShouldErr(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()

	selector, _ := NewSelectorExpandedList(weights, hasher)
	indexes, err := selector.Select([]byte("random"), setSize+1)

	require.Equal(t, ErrInvalidSampleSize, err)
	require.Nil(t, indexes)
}

func TestSelectorExpandedList_SelectOK(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()
	sampleSize := setSize / 2

	selector, _ := NewSelectorExpandedList(weights, hasher)
	indexes, err := selector.Select([]byte("random"), sampleSize)
	require.Nil(t, err)
	require.NotNil(t, indexes)
	require.Equal(t, sampleSize, uint32(len(indexes)))
}
