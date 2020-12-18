package txcache

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultScoreComputer_computeRawScore(t *testing.T) {
	_, txFeeHelper := dummyParamsWithGasPrice(100 * oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)

	// 50k moveGas, 100Bil minPrice -> normalizedFee 8940
	score := computer.computeRawScore(senderScoreParams{count: 1, feeScore: 18000, gas: 100000})
	assert.InDelta(t, float64(34.78839226910904), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 1, feeScore: 1500000, gas: 10000000})
	assert.InDelta(t, float64(19.61086955543323), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 1, feeScore: 5000000, gas: 30000000})
	assert.InDelta(t, float64(26.60960295569976), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 2, feeScore: 36000, gas: 200000})
	assert.InDelta(t, float64(23.12944684325884), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 1000, feeScore: 18000000, gas: 100000000})
	assert.InDelta(t, float64(3.9327419187495716), score, delta)

	score = computer.computeRawScore(senderScoreParams{count: 10000, feeScore: 180000000, gas: 1000000000})
	assert.InDelta(t, float64(3.000829279334849), score, delta)
}

func BenchmarkScoreComputer_computeRawScore(b *testing.B) {
	_, txFeeHelper := dummyParams()
	computer := newDefaultScoreComputer(txFeeHelper)

	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < 10000000; j++ {
			computer.computeRawScore(senderScoreParams{count: j, feeScore: uint64(float64(8000) * float64(j)), gas: 100000 * j})
		}
	}
}

func TestDefaultScoreComputer_computeRawScoreOfTxListForSender(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(100 * oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)
	list := newUnconstrainedListToTest()

	list.AddTx(createTxWithParams([]byte("a"), ".", 1, 1000, 50000, 100*oneBillion), txGasHandler, txFeeHelper)
	list.AddTx(createTxWithParams([]byte("b"), ".", 1, 500, 100000, 100*oneBillion), txGasHandler, txFeeHelper)
	list.AddTx(createTxWithParams([]byte("c"), ".", 1, 500, 100000, 100*oneBillion), txGasHandler, txFeeHelper)

	require.Equal(t, uint64(3), list.countTx())
	require.Equal(t, int64(2000), list.totalBytes.Get())
	require.Equal(t, int64(250000), list.totalGas.Get())
	require.Equal(t, int64(40260), list.totalFee.Get())

	scoreParams := list.getScoreParams()
	rawScore := computer.computeRawScore(scoreParams)
	require.InDelta(t, float64(12.6078942666), rawScore, delta)
}

func TestDefaultScoreComputer_scoreFluctuatesDeterministicallyWhileTxListForSenderMutates(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(100 * oneBillion)
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
	require.Equal(t, 18, scoreA)
	require.Equal(t, 12, scoreAB)
	require.Equal(t, 10, scoreABC)
	require.Equal(t, 9, scoreABCD)

	list.RemoveTx(D)
	scoreABC = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(C)
	scoreAB = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(B)
	scoreA = int(computer.computeScore(list.getScoreParams()))
	list.RemoveTx(A)
	scoreNone = int(computer.computeScore(list.getScoreParams()))

	require.Equal(t, 0, scoreNone)
	require.Equal(t, 18, scoreA)
	require.Equal(t, 12, scoreAB)
	require.Equal(t, 10, scoreABC)
}

func TestDefaultScoreComputer_DifferentSenders(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(100 * oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)

	A := createTxWithParams([]byte("a"), "a", 1, 128, 50000, 100*oneBillion)    // min value normal tx
	B := createTxWithParams([]byte("b"), "b", 1, 128, 50000, 150*oneBillion)    // 50% higher value normal tx
	C := createTxWithParams([]byte("c"), "c", 1, 128, 10000000, 100*oneBillion) // min value SC call
	D := createTxWithParams([]byte("d"), "d", 1, 128, 10000000, 150*oneBillion) // 50% higher value SC call

	listA := newUnconstrainedListToTest()
	listA.AddTx(A, txGasHandler, txFeeHelper)
	scoreA := int(computer.computeScore(listA.getScoreParams()))

	listB := newUnconstrainedListToTest()
	listB.AddTx(B, txGasHandler, txFeeHelper)
	scoreB := int(computer.computeScore(listB.getScoreParams()))

	listC := newUnconstrainedListToTest()
	listC.AddTx(C, txGasHandler, txFeeHelper)
	scoreC := int(computer.computeScore(listC.getScoreParams()))

	listD := newUnconstrainedListToTest()
	listD.AddTx(D, txGasHandler, txFeeHelper)
	scoreD := int(computer.computeScore(listD.getScoreParams()))

	require.Equal(t, 34, scoreA)
	require.Equal(t, 83, scoreB)
	require.Equal(t, 15, scoreC)
	require.Equal(t, 17, scoreD)

	// adding same type of transactions for each sender decreases the score
	for i := 2; i < 1000; i++ {
		A = createTxWithParams([]byte("a"+strconv.Itoa(i)), "a", uint64(i), 128, 50000, 100*oneBillion) // min value normal tx
		listA.AddTx(A, txGasHandler, txFeeHelper)
		B = createTxWithParams([]byte("b"+strconv.Itoa(i)), "b", uint64(i), 128, 50000, 150*oneBillion) // 50% higher value normal tx
		listB.AddTx(B, txGasHandler, txFeeHelper)
		C = createTxWithParams([]byte("c"+strconv.Itoa(i)), "c", uint64(i), 128, 10000000, 100*oneBillion) // min value SC call
		listC.AddTx(C, txGasHandler, txFeeHelper)
		D = createTxWithParams([]byte("d"+strconv.Itoa(i)), "d", uint64(i), 128, 10000000, 150*oneBillion) // 50% higher value SC call
		listD.AddTx(D, txGasHandler, txFeeHelper)
	}

	scoreA = int(computer.computeScore(listA.getScoreParams()))
	scoreB = int(computer.computeScore(listB.getScoreParams()))
	scoreC = int(computer.computeScore(listC.getScoreParams()))
	scoreD = int(computer.computeScore(listD.getScoreParams()))

	require.Equal(t, 3, scoreA)
	require.Equal(t, 12, scoreB)
	require.Equal(t, 1, scoreC)
	require.Equal(t, 1, scoreD)
}
