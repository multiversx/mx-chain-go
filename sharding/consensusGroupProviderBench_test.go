package sharding

import (
	"testing"
)

func BenchmarkReslicingBasedProvider_Get(b *testing.B) {
	numVals := 63
	expElList := getExpandedEligibleList(400)
	randomness := make([]byte, 8)
	for i := 0; i < 8; i++ {
		randomness[i] = 5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testWithReslicing(randomness, numVals, expElList)
	}
}

func BenchmarkSelectionBasedProvider_Get(b *testing.B) {
	numVals := 63
	expElList := getExpandedEligibleList(400)
	randomness := make([]byte, 8)
	for i := 0; i < 8; i++ {
		randomness[i] = 5
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testWithSelection(randomness, numVals, expElList)
	}
}
