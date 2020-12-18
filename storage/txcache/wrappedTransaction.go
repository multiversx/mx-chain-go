package txcache

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/data"
)

const processFeeFactor = float64(0.8) // 80%

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx                   data.TransactionHandler
	TxHash               []byte
	SenderShardID        uint32
	ReceiverShardID      uint32
	Size                 int64
	TxFeeScoreNormalized uint64
}

func (wrappedTx *WrappedTransaction) sameAs(another *WrappedTransaction) bool {
	return bytes.Equal(wrappedTx.TxHash, another.TxHash)
}

// estimateTxGas returns an approximation for the necessary computation units (gas units)
func estimateTxGas(tx *WrappedTransaction) uint64 {
	gasLimit := tx.Tx.GetGasLimit()
	return gasLimit
}

// estimateTxFeeScore returns an approximation for the cost of a transaction, in nano ERD
func estimateTxFeeScore(tx *WrappedTransaction, txGasHandler TxGasHandler, txFeeHelper feeHelper) uint64 {
	moveGas, processGas := txGasHandler.SplitTxGasInCategories(tx.Tx)

	normalizedMoveGas := moveGas >> txFeeHelper.gasLimitShift()
	normalizedProcessGas := processGas >> txFeeHelper.gasLimitShift()

	normalizedGasPriceMove := txGasHandler.GasPriceForMove(tx.Tx) >> txFeeHelper.gasPriceShift()
	normalizedGasPriceProcess := normalizeGasPriceProcessing(tx, txGasHandler, txFeeHelper)

	normalizedFeeMove := normalizedMoveGas * normalizedGasPriceMove
	normalizedFeeProcess := normalizedProcessGas * normalizedGasPriceProcess

	adjustmentFactor := computeProcessingGasPriceAdjustment(tx, txGasHandler, txFeeHelper)

	tx.TxFeeScoreNormalized = normalizedFeeMove + normalizedFeeProcess*adjustmentFactor

	return tx.TxFeeScoreNormalized
}

func normalizeGasPriceProcessing(tx *WrappedTransaction, txGasHandler TxGasHandler, txFeeHelper feeHelper) uint64 {
	return txGasHandler.GasPriceForProcessing(tx.Tx) >> txFeeHelper.gasPriceShift()
}

func computeProcessingGasPriceAdjustment(
	tx *WrappedTransaction,
	txGasHandler TxGasHandler,
	txFeeHelper feeHelper,
) uint64 {
	minPriceFactor := txFeeHelper.minGasPriceFactor()

	if minPriceFactor <= 2 {
		return 1
	}

	actualPriceFactor := float64(1)
	if txGasHandler.MinGasPriceProcessing() != 0 {
		actualPriceFactor = float64(txGasHandler.GasPriceForProcessing(tx.Tx)) / float64(txGasHandler.MinGasPriceProcessing())
	}

	return uint64(float64(txFeeHelper.minGasPriceFactor()) * processFeeFactor / float64(actualPriceFactor))
}
