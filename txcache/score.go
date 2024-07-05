package txcache

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

var _ scoreComputer = (*defaultScoreComputer)(nil)

type senderScoreParams struct {
	avgPpuNumerator             float64
	avgPpuDenominator           uint64
	hasSpotlessSequenceOfNonces bool
}

type defaultScoreComputer struct {
	worstPpu           float64
	scoreScalingFactor float64
}

func newDefaultScoreComputer(txGasHandler TxGasHandler) *defaultScoreComputer {
	worstPpu := computeWorstPpu(txGasHandler)
	excellentPpu := float64(txGasHandler.MinGasPrice()) * excellentGasPriceFactor
	excellentPpuNormalized := excellentPpu / worstPpu
	excellentPpuNormalizedLog := math.Log(excellentPpuNormalized)
	scoreScalingFactor := float64(numberOfScoreChunks) / excellentPpuNormalizedLog

	return &defaultScoreComputer{
		worstPpu:           worstPpu,
		scoreScalingFactor: scoreScalingFactor,
	}
}

func computeWorstPpu(txGasHandler TxGasHandler) float64 {
	gasLimit := txGasHandler.MaxGasLimitPerTx()
	gasPrice := txGasHandler.MinGasPrice()

	worstPpuTx := &WrappedTransaction{
		Tx: &transaction.Transaction{
			GasLimit: gasLimit,
			GasPrice: gasPrice,
		},
	}

	return worstPpuTx.computeFee(txGasHandler) / float64(gasLimit)
}

// computeScore computes the score of the sender, as an integer in [0, numberOfScoreChunks]
func (computer *defaultScoreComputer) computeScore(scoreParams senderScoreParams) uint32 {
	rawScore := computer.computeRawScore(scoreParams)
	truncatedScore := uint32(rawScore)

	if truncatedScore > numberOfScoreChunks {
		return numberOfScoreChunks
	}

	return truncatedScore
}

func (computer *defaultScoreComputer) computeRawScore(params senderScoreParams) float64 {
	if !params.hasSpotlessSequenceOfNonces {
		return 0
	}

	avgPpu := params.avgPpuNumerator / float64(params.avgPpuDenominator)

	// We use the worst possible price per unit for normalization.
	avgPpuNormalized := avgPpu / computer.worstPpu
	avgPpuNormalizedLog := math.Log(avgPpuNormalized)

	score := avgPpuNormalizedLog * computer.scoreScalingFactor
	return score
}
