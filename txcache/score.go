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
	worstPpuLog        float64
	scoreScalingFactor float64
}

func newDefaultScoreComputer(txGasHandler TxGasHandler) *defaultScoreComputer {
	worstPpu := computeWorstPpu(txGasHandler)
	worstPpuLog := math.Log(worstPpu)
	excellentPpu := float64(txGasHandler.MinGasPrice()) * excellentGasPriceFactor
	excellentPpuNormalized := excellentPpu / worstPpu
	excellentPpuNormalizedLog := math.Log(excellentPpuNormalized)
	scoreScalingFactor := float64(numberOfScoreChunks) / excellentPpuNormalizedLog

	return &defaultScoreComputer{
		worstPpuLog:        worstPpuLog,
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
func (computer *defaultScoreComputer) computeScore(scoreParams senderScoreParams) int {
	rawScore := computer.computeRawScore(scoreParams)
	truncatedScore := int(rawScore)

	if truncatedScore > numberOfScoreChunks {
		return numberOfScoreChunks
	}

	return truncatedScore
}

// computeRawScore computes the score of a sender, as follows:
// score = log(sender's average price per unit / worst price per unit) * scoreScalingFactor,
// where scoreScalingFactor = highest score / log(excellent price per unit / worst price per unit)
func (computer *defaultScoreComputer) computeRawScore(params senderScoreParams) float64 {
	if !params.hasSpotlessSequenceOfNonces {
		return 0
	}

	avgPpu := params.avgPpuNumerator / float64(params.avgPpuDenominator)

	// We use the worst possible price per unit for normalization.
	// The expression below is same as log(avgPpu / worstPpu), but we precompute "worstPpuLog" in the constructor.
	avgPpuNormalizedLog := math.Log(avgPpu) - computer.worstPpuLog

	score := avgPpuNormalizedLog * computer.scoreScalingFactor
	return score
}
