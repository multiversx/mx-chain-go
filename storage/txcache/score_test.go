package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScoreComputer_computeRawScore(t *testing.T) {
	computer := &scoreComputer{}

	score := computer.computeRawScore(senderScoreParams{count: 14000, size: kBToBytes(100000), fee: toNanoERD(300), gas: 2500000000, minGasPrice: 100})
	require.InDelta(t, float64(0.1789683371), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 19000, size: kBToBytes(3000), fee: toNanoERD(2300), gas: 19000000000, minGasPrice: 100})
	require.InDelta(t, float64(0.2517997181), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 3, size: kBToBytes(2), fee: toNanoERD(0.04), gas: 400000, minGasPrice: 100})
	require.InDelta(t, float64(5.795382396), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 1, size: kBToBytes(0.3), fee: toNanoERD(0.05), gas: 100000, minGasPrice: 100})
	require.InDelta(t, float64(100), score, delta)
}

func BenchmarkScoreComputer_computeRawScore(b *testing.B) {
	computer := &scoreComputer{}

	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 10000000; j++ {
			computer.computeRawScore(senderScoreParams{count: j, size: (j + 1) * 500, fee: toNanoERD(float64(0.011) * float64(j)), gas: 100000 * j})
		}
	}
}
