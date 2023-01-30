package transactionAPI

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type gasUsedAndFeeProcessor struct {
	txFeeCalculator feeComputer
}

func newGasUsedAndFeeProcessor(txFeeCalculator feeComputer) *gasUsedAndFeeProcessor {
	return &gasUsedAndFeeProcessor{
		txFeeCalculator: txFeeCalculator,
	}
}

func (gfp *gasUsedAndFeeProcessor) computeAndAttachGasUsedAndFee(tx *transaction.ApiTransactionResult) {
	gasUsed := gfp.txFeeCalculator.ComputeGasLimit(tx)
	fee := gfp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)

	tx.GasUsed = gasUsed
	tx.Fee = fee.String()

	if tx.IsRelayed {
		tx.GasUsed = tx.GasLimit
		tx.Fee = tx.InitiallyPaidFee
	}

	hasRefund := false
	for _, scr := range tx.SmartContractResults {
		if scr.IsRefund {
			gasUsed, fee = gfp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, scr.Value)

			tx.GasUsed = gasUsed
			tx.Fee = fee.String()
			hasRefund = true
			break
		}
	}

	gfp.prepareTxWithResultsBasedOnLogs(tx, hasRefund)
}

func (gfp *gasUsedAndFeeProcessor) prepareTxWithResultsBasedOnLogs(
	tx *transaction.ApiTransactionResult,
	hasRefund bool,
) {
	if tx.Logs == nil {
		return
	}

	for _, event := range tx.Logs.Events {
		if core.WriteLogIdentifier == event.Identifier && !hasRefund {
			gasUsed, fee := gfp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, big.NewInt(0))
			tx.GasUsed = gasUsed
			tx.Fee = fee.String()

			continue
		}
		if core.SignalErrorOperation == event.Identifier {
			fee := gfp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
			tx.GasUsed = tx.GasLimit
			tx.Fee = fee.String()
		}
	}
}
