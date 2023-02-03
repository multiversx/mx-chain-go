package transactionsfee

import (
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
)

type transactionWithResults struct {
	data.TransactionHandlerWithGasUsedAndFee
	scrs []data.TransactionHandlerWithGasUsedAndFee
	log  *data.LogData
}

type transactionsAndScrsHolder struct {
	txsWithResults map[string]*transactionWithResults
	scrsNoTx       map[string]data.TransactionHandlerWithGasUsedAndFee
}

func newTransactionsAndScrsHolder(nrTxs, nrScrs int) *transactionsAndScrsHolder {
	return &transactionsAndScrsHolder{
		txsWithResults: make(map[string]*transactionWithResults, nrTxs),
		scrsNoTx:       make(map[string]data.TransactionHandlerWithGasUsedAndFee, nrScrs),
	}
}

func prepareTransactionsAndScrs(txPool *outportcore.Pool) *transactionsAndScrsHolder {
	totalTxs := len(txPool.Txs) + len(txPool.Invalid) + len(txPool.Rewards)
	if totalTxs == 0 && len(txPool.Scrs) == 0 {
		return newTransactionsAndScrsHolder(0, 0)
	}

	transactionsAndScrs := newTransactionsAndScrsHolder(totalTxs, len(txPool.Scrs))
	for txHash, tx := range txPool.Txs {
		transactionsAndScrs.txsWithResults[txHash] = &transactionWithResults{
			TransactionHandlerWithGasUsedAndFee: tx,
		}
	}

	for _, txLog := range txPool.Logs {
		txWithResults, ok := transactionsAndScrs.txsWithResults[txLog.TxHash]
		if !ok {
			continue
		}

		txWithResults.log = txLog
	}

	for scrHash, scrHandler := range txPool.Scrs {
		scr, ok := scrHandler.GetTxHandler().(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		txWithResults, ok := transactionsAndScrs.txsWithResults[string(scr.OriginalTxHash)]
		if !ok {
			transactionsAndScrs.scrsNoTx[scrHash] = scrHandler
			continue
		}

		txWithResults.scrs = append(txWithResults.scrs, scrHandler)
	}

	return transactionsAndScrs
}
