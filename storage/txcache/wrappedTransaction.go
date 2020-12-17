package txcache

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/data"
)

const processFeeFactor = 80

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
	normalizedGasPriceProcess, remainingProcessingPriceShift := normalizeGasPriceProcessing(tx, txGasHandler, txFeeHelper)

	normalizedFeeMove := normalizedMoveGas * normalizedGasPriceMove
	normalizedFeeProcess := normalizedProcessGas >> remainingProcessingPriceShift * normalizedGasPriceProcess

	tx.TxFeeScoreNormalized = normalizedFeeMove + normalizedFeeProcess*processFeeFactor

	return tx.TxFeeScoreNormalized
}

func normalizeGasPriceProcessing(tx *WrappedTransaction, txGasHandler TxGasHandler, txFeeHelper feeHelper) (uint64, uint64) {
	normalizedGasPriceProcess := txGasHandler.GasPriceForProcessing(tx.Tx) >> txFeeHelper.gasPriceShift()
	remainingProcessingPriceShift := uint64(0)

	if normalizedGasPriceProcess == 0 {
		var p uint64
		remainingProcessingPriceShift = txFeeHelper.gasPriceShift()

		for p = txGasHandler.GasPriceForProcessing(tx.Tx); p > 0; p >>= 1 {
			remainingProcessingPriceShift--
		}
	}

	return normalizedGasPriceProcess, remainingProcessingPriceShift
}
