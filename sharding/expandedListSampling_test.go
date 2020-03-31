package sharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/stretchr/testify/require"
)

func TestNewSelectorExpandedListNilValidatorsShouldErr(t *testing.T) {
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}

	selector, err := NewSelectorExpandedList(nil, weights, hasher)
	require.Equal(t, ErrNilValidators, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListNilWeightsShouldErr(t *testing.T) {
	validators := createDummyNodesList(10, "shard_0")
	hasher := &sha256.Sha256{}

	selector, err := NewSelectorExpandedList(validators, nil, hasher)
	require.Equal(t, ErrNilWeights, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListNilHasherShouldErr(t *testing.T) {
	validators := createDummyNodesList(10, "shard_0")
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)

	selector, err := NewSelectorExpandedList(validators, weights, nil)
	require.Equal(t, ErrNilHasher, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListInvalidWeightShouldErr(t *testing.T) {
	validators := createDummyNodesList(10, "shard_0")
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	weights[1] = 0
	hasher := &sha256.Sha256{}

	selector, err := NewSelectorExpandedList(validators, weights, hasher)
	require.Equal(t, ErrInvalidWeight, err)
	require.Nil(t, selector)
}

func TestNewSelectorExpandedListOK(t *testing.T) {
	validators := createDummyNodesList(10, "shard_0")
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}

	selector, err := NewSelectorExpandedList(validators, weights, hasher)
	require.Nil(t, err)
	require.NotNil(t, selector)
}

func TestSelectorExpandedList_IsInterfaceNil(t *testing.T) {
	var selector RandomSelector

	require.True(t, check.IfNil(selector))

	validators := createDummyNodesList(10, "shard_0")
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}

	selector, _ = NewSelectorExpandedList(validators, weights, hasher)
	require.False(t, check.IfNil(selector))
}

func TestSelectorExpandedList_SelectNilRandomessShouldErr(t *testing.T) {
	validators := createDummyNodesList(10, "shard_0")
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}

	selector, _ := NewSelectorExpandedList(validators, weights, hasher)
	indexes, err := selector.Select(nil, 5)

	require.Equal(t, ErrNilRandomness, err)
	require.Nil(t, indexes)
}

func TestSelectorExpandedList_Select0SampleSizeShouldErr(t *testing.T) {
	validators := createDummyNodesList(10, "shard_0")
	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}

	selector, _ := NewSelectorExpandedList(validators, weights, hasher)
	indexes, err := selector.Select([]byte("random"), 0)

	require.Equal(t, ErrInvalidSampleSize, err)
	require.Nil(t, indexes)
}

func TestSelectorExpandedList_SelectSampleSizeGreaterThanSetShouldErr(t *testing.T) {
	setSize := uint32(10)
	validators := createDummyNodesList(setSize, "shard_0")
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}

	selector, _ := NewSelectorExpandedList(validators, weights, hasher)
	indexes, err := selector.Select([]byte("random"), setSize+1)

	require.Equal(t, ErrInvalidSampleSize, err)
	require.Nil(t, indexes)
}

func TestSelectorExpandedList_SelectOK(t *testing.T) {
	setSize := uint32(10)
	validators := createDummyNodesList(setSize, "shard_0")
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := &sha256.Sha256{}
	sampleSize := setSize / 2

	selector, _ := NewSelectorExpandedList(validators, weights, hasher)
	indexes, err := selector.Select([]byte("random"), sampleSize)
	require.Nil(t, err)
	require.NotNil(t, indexes)
	require.Equal(t, sampleSize, uint32(len(indexes)))
}
