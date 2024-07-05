package txcache

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

var _ scoreComputer = (*defaultScoreComputer)(nil)

type senderScoreParams struct {
	avgPpuNumerator   float64
	avgPpuDenominator uint64

	accountNonce        uint64
	accountNonceIsKnown bool
	maxTransactionNonce uint64
	minTransactionNonce uint64

	numOfTransactions           uint64
	hasSpotlessSequenceOfNonces bool
}

type defaultScoreComputer struct {
	worstPpu float64
}

func newDefaultScoreComputer(txGasHandler TxGasHandler) *defaultScoreComputer {
	worstPpu := computeWorstPpu(txGasHandler)

	return &defaultScoreComputer{
		worstPpu: worstPpu,
	}
}

func computeWorstPpu(txGasHandler TxGasHandler) float64 {
	minGasPrice := txGasHandler.MinGasPrice()
	maxGasLimitPerTx := txGasHandler.MaxGasLimitPerTx()
	worstPpuTx := &WrappedTransaction{
		Tx: &transaction.Transaction{
			GasLimit: maxGasLimitPerTx,
			GasPrice: minGasPrice,
		},
	}

	return worstPpuTx.computeFee(txGasHandler) / float64(maxGasLimitPerTx)
}

// computeScore computes the score of the sender, as an integer in [0, numberOfScoreChunks]
func (computer *defaultScoreComputer) computeScore(scoreParams senderScoreParams) uint32 {
	rawScore := computer.computeRawScore(scoreParams)
	truncatedScore := uint32(rawScore)
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

	// https://www.wolframalpha.com, with input "((1 / (1 + exp(-x)) - 1/2) * 2) * 100, where x is from 0 to 10"
	avgPpuNormalizedSubunitary := (1.0/(1+math.Exp(-avgPpuNormalizedLog)) - 0.5) * 2
	score := avgPpuNormalizedSubunitary * float64(numberOfScoreChunks)

	return score
}
