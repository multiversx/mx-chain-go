package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSenderAsBucketSortedMapItem_ComputeScore(t *testing.T) {
	list := newTxListForSender(".")

	list.AddTx([]byte("a"), createTxWithParams(".", 1, 1000, 200000, 100*oneTrilion))
	list.AddTx([]byte("b"), createTxWithParams(".", 1, 500, 100000, 100*oneTrilion))
	list.AddTx([]byte("c"), createTxWithParams(".", 1, 500, 100000, 100*oneTrilion))

	require.Equal(t, int64(3), list.countTx())
	require.Equal(t, int64(2000), list.totalBytes.Get())
	require.Equal(t, int64(400000), list.totalGas.Get())
	require.Equal(t, int64(40*oneMilion), list.totalFee.Get())

	require.InDelta(t, float64(5.795382396), list.computeRawScore(), delta)
}

func Test_computeSenderScore(t *testing.T) {

	require.InDelta(t, float64(0.1789683371), computeSenderScore(senderScoreParams{count: 14000, size: kBToBytes(100000), fee: toMicroERD(300000), gas: 2500000000}), delta)
	require.InDelta(t, float64(0.2517997181), computeSenderScore(senderScoreParams{count: 19000, size: kBToBytes(3000), fee: toMicroERD(2300000), gas: 19000000000}), delta)
	require.InDelta(t, float64(5.795382396), computeSenderScore(senderScoreParams{count: 3, size: kBToBytes(2), fee: toMicroERD(40), gas: 400000}), delta)
	require.InDelta(t, float64(100), computeSenderScore(senderScoreParams{count: 1, size: kBToBytes(0.3), fee: toMicroERD(50), gas: 100000}), delta)
}

func Benchmark_computeSenderScore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < 15000; j++ {
			computeSenderScore(senderScoreParams{count: int64(j), size: int64((j + 1) * 500), fee: toMicroERD(int64(11 * j)), gas: int64(100000 * j)})
		}
	}
}
