package nodesCoordinator

import (
	"testing"
)

const numValidators = 63
const numValidatorsInEligibleList = 400

func getRandomness() []byte {
	randomness := make([]byte, 8)
	for i := 0; i < 8; i++ {
		randomness[i] = 5
	}

	return randomness
}

func BenchmarkReslicingBasedProvider_Get(b *testing.B) {
	numVals := numValidators
	expElList := getExpandedEligibleList(numValidatorsInEligibleList)
	randomness := getRandomness()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testWithReslicing(randomness, numVals, expElList)
	}
}

func BenchmarkSelectionBasedProvider_Get(b *testing.B) {
	numVals := numValidators
	expElList := getExpandedEligibleList(numValidatorsInEligibleList)
	randomness := getRandomness()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testWithSelection(randomness, numVals, expElList)
	}
}
