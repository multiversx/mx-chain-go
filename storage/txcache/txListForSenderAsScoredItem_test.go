package txcache

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_computeSenderScore(t *testing.T) {
	delta := 0.00000001
	require.InDelta(t, float64(0.1789683371), computeSenderScore(senderScoreParams{count: 14000, size: 100000, fee: toMicroERD(300000), gas: 2500000000}), delta)
	require.InDelta(t, float64(0.2517997181), computeSenderScore(senderScoreParams{count: 19000, size: 3000, fee: toMicroERD(2300000), gas: 19000000000}), delta)
	require.InDelta(t, float64(5.795382396), computeSenderScore(senderScoreParams{count: 3, size: 2, fee: toMicroERD(40), gas: 400000}), delta)
	require.InDelta(t, float64(100), computeSenderScore(senderScoreParams{count: 1, size: 0.3, fee: toMicroERD(50), gas: 100000}), delta)
}

func toMicroERD(erd int64) int64 {
	return erd * 1000000
}
