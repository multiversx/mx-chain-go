package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultScoreComputer_computeRawScore(t *testing.T) {
	_, txFeeHelper := dummyParamsWithGasPrice(100*oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)

	score := computer.computeRawScore(senderScoreParams{count: 14000, feeScore: toNanoERD(300), gas: 2500000000})
	require.InDelta(t, float64(0.1789683371), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 19000, feeScore: toNanoERD(2300), gas: 19000000000})
	require.InDelta(t, float64(0.2517997181), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 3, feeScore: toNanoERD(0.04), gas: 400000})
	require.InDelta(t, float64(5.795382396), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 1, feeScore: toNanoERD(0.05), gas: 100000})
	require.InDelta(t, float64(100), score, delta)
}

func BenchmarkScoreComputer_computeRawScore(b *testing.B) {
	_, txFeeHelper := dummyParams()
	computer := newDefaultScoreComputer(txFeeHelper)

	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 10000000; j++ {
			computer.computeRawScore(senderScoreParams{count: j, feeScore: toNanoERD(float64(0.011) * float64(j)), gas: 100000 * j})
		}
	}
}

func TestDefaultScoreComputer_computeRawScoreOfTxListForSender(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(100*oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)
	list := newUnconstrainedListToTest()

	list.AddTx(createTxWithParams([]byte("a"), ".", 1, 1000, 200000, 100*oneBillion), txGasHandler, txFeeHelper)
	list.AddTx(createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneBillion), txGasHandler, txFeeHelper)
	list.AddTx(createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneBillion), txGasHandler, txFeeHelper)

	require.Equal(t, uint64(3), list.countTx())
	require.Equal(t, int64(2000), list.totalBytes.Get())
	require.Equal(t, int64(400000), list.totalGas.Get())
	require.Equal(t, int64(40*oneMilion), list.totalFee.Get())

	scoreParams := list.getScoreParams()
	rawScore := computer.computeRawScore(scoreParams)
	require.InDelta(t, float64(5.795382396), rawScore, delta)
}

func TestDefaultScoreComputer_scoreFluctuatesDeterministicallyWhileTxListForSenderMutates(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParams()
	computer := newDefaultScoreComputer(txFeeHelper)
	list := newUnconstrainedListToTest()

	A := createTxWithParams([]byte("A"), ".", 1, 1000, 200000, 100*oneBillion)
	B := createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneBillion)
	C := createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneBillion)
	D := createTxWithParams([]byte("d"), ".", 1, 128, 50000, 100*oneBillion)

	scoreNone := int(computer.computeScore(list.getScoreParams()))
	list.AddTx(A, txGasHandler, txFeeHelper)
	scoreA := int(computer.computeScore(list.getScoreParams()))
	list.AddTx(B, txGasHandler, txFeeHelper)
	scoreAB := int(computer.computeScore(list.getScoreParams()))
	list.AddTx(C, txGasHandler, txFeeHelper)
	scoreABC := int(computer.computeScore(list.getScoreParams()))
	list.AddTx(D, txGasHandler, txFeeHelper)
	scoreABCD := int(computer.computeScore(list.getScoreParams()))

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 17, scoreA)
	require.Equal(t, 8, scoreAB)
	require.Equal(t, 5, scoreABC)
	require.Equal(t, 10, scoreABCD)

	list.RemoveTx(D)
	scoreABC = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(C)
	scoreAB = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(B)
	scoreA = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(A)
	scoreNone = int(computer.computeScore(list.getScoreParams()))

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 17, scoreA)
	require.Equal(t, 8, scoreAB)
	require.Equal(t, 5, scoreABC)
}
