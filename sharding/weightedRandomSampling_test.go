package sharding

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/stretchr/testify/require"
)

type benchParams struct {
	name            string
	selectionSample uint32
	weights         []uint32
}

func createDummyWeights(
	setSize uint32,
	startingWeight uint32,
	changeStep float32,
	getNextWeight func(prevValue uint32, change float32) uint32,
) []uint32 {
	result := make([]uint32, setSize)

	result[0] = startingWeight
	for i := uint32(1); i < setSize; i++ {
		result[i] = getNextWeight(result[i-1], changeStep)
	}

	return result
}

func nextWeightGeometricProgression(prevValue uint32, ratio float32) uint32 {
	return uint32(float32(prevValue) * ratio)
}

func nextWeightArithmeticProgression(prevValue uint32, difference float32) uint32 {
	return uint32(float32(prevValue) + difference)
}

func displayStatistics(statistics []uint32, timesFirst []uint32, weights []uint32) {
	for i := 0; i < len(statistics); i++ {
		_, _ = fmt.Println("weight", weights[i], "selected", statistics[i], "times.", "Selected first", timesFirst[i], "times")
	}
}

func benchmarkSelection(
	b *testing.B,
	selectionSample uint32,
	weights []uint32,
	rands [][]byte,
	hasher hashing.Hasher,
	statistics []uint32,
	timesFirst []uint32,
) {
	nbPrecomputedRands := len(rands)
	selector, err := NewSelectorWRS(weights, hasher)
	require.Nil(b, err)
	require.NotNil(b, selector)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StartTimer()

		index := n % nbPrecomputedRands
		selectedIndexes, errSelect := selector.Select(rands[index], selectionSample)

		b.StopTimer()
		require.Nil(b, errSelect)
		require.Equal(b, selectionSample, uint32(len(selectedIndexes)))
		timesFirst[selectedIndexes[0]]++
		for i := uint32(0); i < selectionSample; i++ {
			statistics[selectedIndexes[i]]++
		}
	}
}

func BenchmarkNewSelectorWRS_SelectConsensusShard(b *testing.B) {
	benchData := []benchParams{
		{
			name:            "63 of 400 arithmetic progression",
			selectionSample: 63,
			weights:         createDummyWeights(400, 100, 100, nextWeightArithmeticProgression),
		},
		{
			name:            "400 of 400 arithmetic progression",
			selectionSample: 400,
			weights:         createDummyWeights(400, 100, 100, nextWeightArithmeticProgression),
		},
		{
			name:            "63 of 400 same weights",
			selectionSample: 63,
			weights:         createDummyWeights(400, 100, 1, nextWeightGeometricProgression),
		},
		{
			name:            "400 of 400 same weights",
			selectionSample: 400,
			weights:         createDummyWeights(400, 100, 1, nextWeightGeometricProgression),
		},
		{
			name:            "5 of 10 geometric progression",
			selectionSample: 5,
			weights:         createDummyWeights(10, 32, 2, nextWeightGeometricProgression),
		},
	}

	nbPrecomputedRands := 100000
	hasher := sha256.NewSha256()

	rands := make([][]byte, nbPrecomputedRands)
	for i := 0; i < nbPrecomputedRands; i++ {
		rands[i] = hasher.Compute(strconv.Itoa(i))
	}

	for _, bd := range benchData {
		statistics := make([]uint32, len(bd.weights))
		timesFirst := make([]uint32, len(bd.weights))
		b.Run(bd.name, func(b *testing.B) {
			benchmarkSelection(b, bd.selectionSample, bd.weights, rands, hasher, statistics, timesFirst)
		})
		displayStatistics(statistics, timesFirst, bd.weights)
	}
}

func TestNewSelectorWRSNilWeightsShouldErr(t *testing.T) {
	t.Parallel()

	hasher := sha256.NewSha256()
	selector, err := NewSelectorWRS(nil, hasher)
	require.Equal(t, ErrNilWeights, err)
	require.Nil(t, selector)
}

func TestNewSelectorWRSNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	selector, err := NewSelectorWRS(weights, nil)
	require.Equal(t, ErrNilHasher, err)
	require.Nil(t, selector)
}

func TestNewSelectorWRSOK(t *testing.T) {
	t.Parallel()

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()
	selector, err := NewSelectorWRS(weights, hasher)
	require.Nil(t, err)
	require.NotNil(t, selector)
}

func TestSelectorWRS_SelectNilRandSeedShouldErr(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	sampleSize := uint32(5)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()
	selector, _ := NewSelectorWRS(weights, hasher)
	selectedIndexes, err := selector.Select(nil, sampleSize)
	require.Equal(t, ErrNilRandomness, err)
	require.Nil(t, selectedIndexes)
}

func TestSelectorWRS_SelectInvalidSampleSizeShouldErr(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()
	selector, _ := NewSelectorWRS(weights, hasher)
	r := hasher.Compute("seed")
	selectedIndexes, err := selector.Select(r, setSize+1)
	require.Equal(t, ErrInvalidSampleSize, err)
	require.Nil(t, selectedIndexes)
}

func TestSelectorWRS_SelectInvalidWeightShouldErr(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	sampleSize := uint32(5)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	weights[1] = 0
	hasher := sha256.NewSha256()
	selector, _ := NewSelectorWRS(weights, hasher)
	r := hasher.Compute("seed")
	selectedIndexes, err := selector.Select(r, sampleSize)
	require.Equal(t, ErrInvalidWeight, err)
	require.Nil(t, selectedIndexes)
}

func TestSelectorWRS_SelectOK(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	sampleSize := uint32(5)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()
	selector, _ := NewSelectorWRS(weights, hasher)
	r := hasher.Compute("seed")
	selectedIndexes, err := selector.Select(r, sampleSize)
	require.Equal(t, sampleSize, uint32(len(selectedIndexes)))
	require.Nil(t, err)
	for i := 0; i < len(selectedIndexes)-1; i++ {
		for j := i + 1; j < len(selectedIndexes); j++ {
			require.NotEqual(t, selectedIndexes[i], selectedIndexes[j])
		}
	}
}

func TestSelectorWRS_SelectOKSameParametersSameResult(t *testing.T) {
	t.Parallel()

	setSize := uint32(10)
	sampleSize := uint32(5)
	weights := createDummyWeights(setSize, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()
	selector, _ := NewSelectorWRS(weights, hasher)
	r := hasher.Compute("seed")

	selectedIndexes, err := selector.Select(r, sampleSize)
	require.Equal(t, sampleSize, uint32(len(selectedIndexes)))
	require.Nil(t, err)

	selectedIndexes2, err2 := selector.Select(r, sampleSize)
	require.Equal(t, sampleSize, uint32(len(selectedIndexes2)))
	require.Nil(t, err2)

	for i := 0; i < len(selectedIndexes)-1; i++ {
		for j := i + 1; j < len(selectedIndexes); j++ {
			require.NotEqual(t, selectedIndexes[i], selectedIndexes[j])

		}
		require.Equal(t, selectedIndexes[i], selectedIndexes2[i])
	}
	require.Equal(t, selectedIndexes[sampleSize-1], selectedIndexes2[sampleSize-1])
}

func TestSelectorWRS_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var selector RandomSelector
	var err error

	weights := createDummyWeights(10, 32, 2, nextWeightGeometricProgression)
	hasher := sha256.NewSha256()

	require.True(t, check.IfNil(selector))
	selector, err = NewSelectorWRS(weights, hasher)
	require.Nil(t, err)
	require.False(t, check.IfNil(selector))
}
