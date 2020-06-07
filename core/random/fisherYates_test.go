package random

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/assert"
)

// ------- FisherYatesShuffle

func TestFisherYatesShuffle_EmptyShouldReturnEmpty(t *testing.T) {
	indexes := make([]int, 0)
	randomizer := &mock.IntRandomizerStub{}

	resultIndexes := FisherYatesShuffle(indexes, randomizer)

	assert.Empty(t, resultIndexes)
}

func TestFisherYatesShuffle_OneElementShouldReturnTheSame(t *testing.T) {
	indexes := []int{1}
	randomizer := &mock.IntRandomizerStub{
		IntnCalled: func(n int) int {
			return n - 1
		},
	}

	resultIndexes := FisherYatesShuffle(indexes, randomizer)

	assert.Equal(t, indexes, resultIndexes)
}

func TestFisherYatesShuffle_ShouldWork(t *testing.T) {
	indexes := []int{1, 2, 3, 4, 5}
	randomizer := &mock.IntRandomizerStub{
		IntnCalled: func(n int) int {
			return 0
		},
	}

	//this will cause a rotation of the first element:
	//i = 4: 5, 2, 3, 4, 1 (swap 1 <-> 5)
	//i = 3: 4, 2, 3, 5, 1 (swap 5 <-> 4)
	//i = 2: 3, 2, 4, 5, 1 (swap 3 <-> 4)
	//i = 1: 2, 3, 4, 5, 1 (swap 3 <-> 2)

	resultIndexes := FisherYatesShuffle(indexes, randomizer)
	expectedResult := []int{2, 3, 4, 5, 1}

	assert.Equal(t, expectedResult, resultIndexes)
}
