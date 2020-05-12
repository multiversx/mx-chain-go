package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSenderAsBucketSortedMapItem_ComputeScore(t *testing.T) {
	list := newUnconstrainedListToTest()

	list.AddTx(createTxWithParams([]byte("a"), ".", 1, 1000, 200000, 100*oneTrilion))
	list.AddTx(createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneTrilion))
	list.AddTx(createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneTrilion))

	require.Equal(t, uint64(3), list.countTx())
	require.Equal(t, int64(2000), list.totalBytes.Get())
	require.Equal(t, int64(400000), list.totalGas.Get())
	require.Equal(t, int64(40*oneMilion), list.totalFee.Get())

	require.InDelta(t, float64(5.795382396), list.computeRawScore(), delta)
}

func TestSenderAsBucketSortedMapItem_ScoreFluctuatesDeterministicallyWhenTransactionsAreAddedOrRemoved(t *testing.T) {
	list := newUnconstrainedListToTest()

	A := createTxWithParams([]byte("A"), ".", 1, 1000, 200000, 100*oneTrilion)
	B := createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneTrilion)
	C := createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneTrilion)

	scoreNone := int(list.ComputeScore())
	list.AddTx(A)
	scoreA := int(list.ComputeScore())
	list.AddTx(B)
	scoreAB := int(list.ComputeScore())
	list.AddTx(C)
	scoreABC := int(list.ComputeScore())

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 17, scoreA)
	require.Equal(t, 8, scoreAB)
	require.Equal(t, 5, scoreABC)

	list.RemoveTx(C)
	scoreAB = int(list.ComputeScore())
	list.RemoveTx(B)
	scoreA = int(list.ComputeScore())
	list.RemoveTx(A)
	scoreNone = int(list.ComputeScore())

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 17, scoreA)
	require.Equal(t, 8, scoreAB)
	require.Equal(t, 5, scoreABC)
}

func Test_computeSenderScore(t *testing.T) {
	score := computeSenderScore(senderScoreParams{count: 14000, size: kBToBytes(100000), fee: toMicroERD(300000), gas: 2500000000, minGasPrice: 100})
	require.InDelta(t, float64(0.1789683371), score, delta)

	score = computeSenderScore(senderScoreParams{count: 19000, size: kBToBytes(3000), fee: toMicroERD(2300000), gas: 19000000000, minGasPrice: 100})
	require.InDelta(t, float64(0.2517997181), score, delta)

	score = computeSenderScore(senderScoreParams{count: 3, size: kBToBytes(2), fee: toMicroERD(40), gas: 400000, minGasPrice: 100})
	require.InDelta(t, float64(5.795382396), score, delta)

	score = computeSenderScore(senderScoreParams{count: 1, size: kBToBytes(0.3), fee: toMicroERD(50), gas: 100000, minGasPrice: 100})
	require.InDelta(t, float64(100), score, delta)
}

func Benchmark_computeSenderScore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 10000000; j++ {
			computeSenderScore(senderScoreParams{count: j, size: (j + 1) * 500, fee: toMicroERD(11 * j), gas: 100000 * j})
		}
	}
}
