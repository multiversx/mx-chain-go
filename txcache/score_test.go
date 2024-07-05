package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultScoreComputer(t *testing.T) {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	computer := newDefaultScoreComputer(gasHandler)

	require.NotNil(t, computer)
	require.Equal(t, float64(16.12631180572966), computer.worstPpuLog)
}

func TestComputeWorstPpu(t *testing.T) {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	require.Equal(t, float64(10082500), computeWorstPpu(gasHandler))
}

func TestDefaultScoreComputer_computeScore(t *testing.T) {
	// Simple transfers:
	require.Equal(t, 74, computeScoreOfTransaction(0, 50000, oneBillion))
	require.Equal(t, 80, computeScoreOfTransaction(0, 50000, 1.5*oneBillion))
	require.Equal(t, 85, computeScoreOfTransaction(0, 50000, 2*oneBillion))
	require.Equal(t, 99, computeScoreOfTransaction(0, 50000, 5*oneBillion))
	require.Equal(t, 100, computeScoreOfTransaction(0, 50000, 10*oneBillion))

	// Simple transfers, with some data (same scores as above):
	require.Equal(t, 74, computeScoreOfTransaction(100, 50000+1500*100, oneBillion))
	require.Equal(t, 80, computeScoreOfTransaction(100, 50000+1500*100, 1.5*oneBillion))
	require.Equal(t, 85, computeScoreOfTransaction(100, 50000+1500*100, 2*oneBillion))
	require.Equal(t, 99, computeScoreOfTransaction(100, 50000+1500*100, 5*oneBillion))
	require.Equal(t, 100, computeScoreOfTransaction(100, 50000+1500*100, 10*oneBillion))

	// Smart contract calls:
	require.Equal(t, 28, computeScoreOfTransaction(1, 1000000, oneBillion))
	require.Equal(t, 40, computeScoreOfTransaction(42, 1000000, oneBillion))
	// Even though the gas price is high, it does not compensate the network's contract execution subsidies (thus, score is not excellent).
	require.Equal(t, 46, computeScoreOfTransaction(42, 1000000, 1.5*oneBillion))
	require.Equal(t, 51, computeScoreOfTransaction(42, 1000000, 2*oneBillion))
	require.Equal(t, 66, computeScoreOfTransaction(42, 1000000, 5*oneBillion))
	require.Equal(t, 77, computeScoreOfTransaction(42, 1000000, 10*oneBillion))
	require.Equal(t, 88, computeScoreOfTransaction(42, 1000000, 20*oneBillion))
	require.Equal(t, 94, computeScoreOfTransaction(42, 1000000, 30*oneBillion))
	require.Equal(t, 99, computeScoreOfTransaction(42, 1000000, 40*oneBillion))
	require.Equal(t, 100, computeScoreOfTransaction(42, 1000000, 50*oneBillion))

	// Smart contract calls with extremely large gas limit:
	require.Equal(t, 0, computeScoreOfTransaction(3, 150000000, oneBillion))
	require.Equal(t, 0, computeScoreOfTransaction(3, 300000000, oneBillion))
	require.Equal(t, 6, computeScoreOfTransaction(3, 150000000, 1.5*oneBillion))
	require.Equal(t, 11, computeScoreOfTransaction(3, 150000000, 2*oneBillion))
	require.Equal(t, 26, computeScoreOfTransaction(3, 150000000, 5*oneBillion))
	require.Equal(t, 37, computeScoreOfTransaction(3, 150000000, 10*oneBillion))
	require.Equal(t, 48, computeScoreOfTransaction(3, 150000000, 20*oneBillion))
	require.Equal(t, 55, computeScoreOfTransaction(3, 150000000, 30*oneBillion))
	// With a very high gas price, the transaction reaches the score of a simple transfer:
	require.Equal(t, 74, computeScoreOfTransaction(3, 150000000, 100*oneBillion))

	// Smart contract calls with max gas limit:
	require.Equal(t, 0, computeScoreOfTransaction(3, 600000000, oneBillion))
	require.Equal(t, 37, computeScoreOfTransaction(3, 600000000, 10*oneBillion))
	require.Equal(t, 63, computeScoreOfTransaction(3, 600000000, 50*oneBillion))
	// With a very high gas price, the transaction reaches the score of a simple transfer:
	require.Equal(t, 74, computeScoreOfTransaction(3, 600000000, 100*oneBillion))
	require.Equal(t, 85, computeScoreOfTransaction(3, 600000000, 200*oneBillion))
}

// Generally speaking, the score is computed for a sender, not for a single transaction.
// However, for the sake of testing, we consider a sender with a single transaction.
func computeScoreOfTransaction(dataLength int, gasLimit uint64, gasPrice uint64) int {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	computer := newDefaultScoreComputer(gasHandler)

	tx := &WrappedTransaction{
		Tx: &transaction.Transaction{
			Data:     make([]byte, dataLength),
			GasLimit: gasLimit,
			GasPrice: gasPrice,
		},
	}

	txFee := tx.computeFee(gasHandler)

	scoreParams := senderScoreParams{
		avgPpuNumerator:             txFee,
		avgPpuDenominator:           gasLimit,
		hasSpotlessSequenceOfNonces: true,
	}

	return int(computer.computeScore(scoreParams))
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
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)
	list := newUnconstrainedListToTest()

	list.AddTx(createTxWithParams([]byte("a"), ".", 1, 1000, 50000, oneBillion), txGasHandler, txFeeHelper)
	list.AddTx(createTxWithParams([]byte("b"), ".", 1, 500, 100000, oneBillion), txGasHandler, txFeeHelper)
	list.AddTx(createTxWithParams([]byte("c"), ".", 1, 500, 100000, oneBillion), txGasHandler, txFeeHelper)

	require.Equal(t, uint64(3), list.countTx())
	require.Equal(t, int64(2000), list.totalBytes.Get())
	require.Equal(t, int64(250000), list.totalGas.Get())
	require.Equal(t, int64(51588), list.totalFeeScore.Get())

	scoreParams := list.getScoreParams()
	rawScore := computer.computeRawScore(scoreParams)
	require.InDelta(t, float64(12.4595615805), rawScore, delta)
}

func TestDefaultScoreComputer_scoreFluctuatesDeterministicallyWhileTxListForSenderMutates(t *testing.T) {
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)
	list := newUnconstrainedListToTest()

	A := createTxWithParams([]byte("A"), ".", 1, 1000, 200000, oneBillion)
	B := createTxWithParams([]byte("b"), ".", 1, 500, 100000, oneBillion)
	C := createTxWithParams([]byte("c"), ".", 1, 500, 100000, oneBillion)
	D := createTxWithParams([]byte("d"), ".", 1, 128, 50000, oneBillion)

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
	txGasHandler, txFeeHelper := dummyParamsWithGasPrice(oneBillion)
	computer := newDefaultScoreComputer(txFeeHelper)

	A := createTxWithParams([]byte("a"), "a", 1, 128, 50000, oneBillion)                // min value normal tx
	B := createTxWithParams([]byte("b"), "b", 1, 128, 50000, uint64(1.5*oneBillion))    // 50% higher value normal tx
	C := createTxWithParams([]byte("c"), "c", 1, 128, 10000000, oneBillion)             // min value SC call
	D := createTxWithParams([]byte("d"), "d", 1, 128, 10000000, uint64(1.5*oneBillion)) // 50% higher value SC call

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

	require.Equal(t, 33, scoreA)
	require.Equal(t, 82, scoreB)
	require.Equal(t, 15, scoreC)
	require.Equal(t, 16, scoreD)

	// adding same type of transactions for each sender decreases the score
	for i := 2; i < 1000; i++ {
		A = createTxWithParams([]byte("a"+strconv.Itoa(i)), "a", uint64(i), 128, 50000, oneBillion) // min value normal tx
		listA.AddTx(A, txGasHandler, txFeeHelper)
		B = createTxWithParams([]byte("b"+strconv.Itoa(i)), "b", uint64(i), 128, 50000, uint64(1.5*oneBillion)) // 50% higher value normal tx
		listB.AddTx(B, txGasHandler, txFeeHelper)
		C = createTxWithParams([]byte("c"+strconv.Itoa(i)), "c", uint64(i), 128, 10000000, oneBillion) // min value SC call
		listC.AddTx(C, txGasHandler, txFeeHelper)
		D = createTxWithParams([]byte("d"+strconv.Itoa(i)), "d", uint64(i), 128, 10000000, uint64(1.5*oneBillion)) // 50% higher value SC call
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
