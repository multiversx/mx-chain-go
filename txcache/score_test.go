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
	require.Equal(t, float64(16.112421018189185), computer.scoreScalingFactor)
}

func TestComputeWorstPpu(t *testing.T) {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	require.Equal(t, float64(10082500), computeWorstPpu(gasHandler))
}

func TestDefaultScoreComputer_computeRawScore(t *testing.T) {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	computer := newDefaultScoreComputer(gasHandler)

	require.Equal(t, 74.06805875222626, computer.computeRawScore(senderScoreParams{
		avgPpuNumerator:             57500000000000,
		avgPpuDenominator:           57500,
		isAccountNonceKnown:         false,
		hasSpotlessSequenceOfNonces: true,
	}))

	require.Equal(t, 135.40260746155397, computer.computeRawScore(senderScoreParams{
		avgPpuNumerator:             57500000000000 * 45,
		avgPpuDenominator:           57500,
		isAccountNonceKnown:         false,
		hasSpotlessSequenceOfNonces: true,
	}))
}

func TestDefaultScoreComputer_computeScore(t *testing.T) {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	worstPpu := computeWorstPpu(gasHandler)
	excellentPpu := float64(gasHandler.MinGasPrice()) * excellentGasPriceFactor

	require.Equal(t, 0, computeScoreGivenAvgPpu(worstPpu))
	require.Equal(t, 11, computeScoreGivenAvgPpu(worstPpu*2))
	require.Equal(t, 31, computeScoreGivenAvgPpu(worstPpu*7))
	require.Equal(t, 74, computeScoreGivenAvgPpu(worstPpu*100))
	require.Equal(t, 90, computeScoreGivenAvgPpu(worstPpu*270))
	require.Equal(t, 99, computeScoreGivenAvgPpu(worstPpu*495))
	require.Equal(t, 100, computeScoreGivenAvgPpu(worstPpu*500))

	require.Equal(t, 55, computeScoreGivenAvgPpu(excellentPpu/16))
	require.Equal(t, 66, computeScoreGivenAvgPpu(excellentPpu/8))
	require.Equal(t, 77, computeScoreGivenAvgPpu(excellentPpu/4))
	require.Equal(t, 88, computeScoreGivenAvgPpu(excellentPpu/2))
	require.Equal(t, 99, computeScoreGivenAvgPpu(excellentPpu))
	require.Equal(t, 100, computeScoreGivenAvgPpu(excellentPpu+1))
}

func computeScoreGivenAvgPpu(avgPpu float64) int {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	computer := newDefaultScoreComputer(gasHandler)

	return computer.computeScore(senderScoreParams{
		avgPpuNumerator:             avgPpu,
		avgPpuDenominator:           1,
		isAccountNonceKnown:         true,
		hasSpotlessSequenceOfNonces: true,
	})
}

func TestDefaultScoreComputer_computeScore_consideringOneTransaction(t *testing.T) {
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

func BenchmarkScoreComputer_computeScore(b *testing.B) {
	gasHandler := txcachemocks.NewTxGasHandlerMock()
	computer := newDefaultScoreComputer(gasHandler)

	tx := &WrappedTransaction{
		Tx: &transaction.Transaction{
			Data:     make([]byte, 42),
			GasLimit: 50000000,
			GasPrice: 1000000000,
		},
	}

	for i := 0; i < b.N; i++ {
		txFee := tx.computeFee(gasHandler)

		for j := uint64(0); j < 1_000_000; j++ {
			computer.computeScore(senderScoreParams{
				avgPpuNumerator:             txFee,
				avgPpuDenominator:           tx.Tx.GetGasLimit(),
				hasSpotlessSequenceOfNonces: true,
			})
		}
	}

	// Results:
	//
	// (a) 12 ms to compute the score 1 million times:
	// 		cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
	// 		BenchmarkScoreComputer_computeScore-8   	     100	  11895452 ns/op	     297 B/op	      12 allocs/op
}
