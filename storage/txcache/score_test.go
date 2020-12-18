package txcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultScoreComputer_computeRawScore(t *testing.T) {
	_, txFeeHelper := dummyParamsWithGasPrice(100*oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)

	// 50k moveGas -> fee 1100
	score := computer.computeRawScore(senderScoreParams{count: 14000, feeScore: 30000000, gas: 2500000000})
	assert.InDelta(t, float64(2.3415883017), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 19000, feeScore: 200000000, gas: 19000000000})
	assert.InDelta(t, float64(1.5359207057), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 2, feeScore: 5000, gas: 400000})
	assert.InDelta(t, float64(21.2260779521), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 1, feeScore: 2300, gas: 100000})
	assert.InDelta(t, float64(96.7883275543), score, delta)
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
	require.Equal(t, int64(5748), list.totalFee.Get())

	scoreParams := list.getScoreParams()
	rawScore := computer.computeRawScore(scoreParams)
	require.InDelta(t, float64(24.97322800), rawScore, delta)
}

func TestDefaultScoreComputer_scoreFluctuatesDeterministicallyWhileTxListForSenderMutates(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(100*oneBillion)
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
	require.Equal(t, 33, scoreA)
	require.Equal(t, 28, scoreAB)
	require.Equal(t, 24, scoreABC)
	require.Equal(t, 26, scoreABCD)

	list.RemoveTx(D)
	scoreABC = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(C)
	scoreAB = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(B)
	scoreA = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(A)
	scoreNone = int(computer.computeScore(list.getScoreParams()))

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 33, scoreA)
	require.Equal(t, 28, scoreAB)
	require.Equal(t, 24, scoreABC)
}
