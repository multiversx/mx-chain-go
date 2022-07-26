package transactionsfee

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"math/big"
)

const (
	writeLogOperation    = "writeLog"
	signalErrorOperation = "signalError"
)

type ArgTransactionsFeeProcessor struct {
	TxFeeCalculator FeesProcessorHandler
}

type transactionsFeeProcessor struct {
	txFeeCalculator FeesProcessorHandler
}

func NewTransactionFeeProcessor(arg ArgTransactionsFeeProcessor) (*transactionsFeeProcessor, error) {
	if check.IfNil(arg.TxFeeCalculator) {
		return nil, ErrNilTransactionFeeCalculator
	}

	return &transactionsFeeProcessor{
		txFeeCalculator: arg.TxFeeCalculator,
	}, nil
}

func (tep *transactionsFeeProcessor) PutFeeAndGasUsed(pool *indexer.Pool) {
	tep.prepareInvalidTxs(pool)

	txsWithResultsMap := groupTransactionsWithResults(pool)
	tep.prepareNormalTxs(txsWithResultsMap)

}

func (tep *transactionsFeeProcessor) prepareInvalidTxs(pool *indexer.Pool) {
	for _, invalidTx := range pool.Invalid {
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(invalidTx, invalidTx.GetGasLimit())
		invalidTx.SetGasUsed(invalidTx.GetGasLimit())
		invalidTx.SetFee(fee)
	}
}

func (tep *transactionsFeeProcessor) prepareNormalTxs(groupedTxs *groupedTransactionsAndScrs) {
	for txHash, txWithResult := range groupedTxs.txsWithResults {
		gasUsed := tep.txFeeCalculator.ComputeGasLimit(txWithResult)
		fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResult, gasUsed)

		txWithResult.SetGasUsed(gasUsed)
		txWithResult.SetFee(fee)

		tep.prepareTxWithResults([]byte(txHash), txWithResult)
	}
}

func (tep *transactionsFeeProcessor) prepareTxWithResults(txHash []byte, txWithResults *transactionWithResults) {
	txHashHasRefund := make(map[string]struct{})
	for _, scrHandler := range txWithResults.scrs {
		scr, ok := scrHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		if isSCRForSenderWithRefund(scr, txHash, txWithResults) || isRefundForRelayed(scr, txWithResults) {
			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults, scr.Value)

			txWithResults.SetGasUsed(gasUsed)
			txWithResults.SetFee(fee)
			txHashHasRefund[string(txHash)] = struct{}{}
		}
	}

	if check.IfNil(txWithResults.logs) {
		return
	}

	for _, event := range txWithResults.logs.GetLogEvents() {
		identifier := string(event.GetIdentifier())
		switch identifier {
		case signalErrorOperation:
			fee := tep.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(txWithResults, txWithResults.GetGasLimit())
			txWithResults.SetGasUsed(txWithResults.GetGasLimit())
			txWithResults.SetFee(fee)
			return
		case writeLogOperation:
			_, found := txHashHasRefund[string(txHash)]
			if !found {
				return
			}

			gasUsed, fee := tep.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(txWithResults, big.NewInt(0))
			txWithResults.SetGasUsed(gasUsed)
			txWithResults.SetFee(fee)
		default:
			continue
		}
	}
}
