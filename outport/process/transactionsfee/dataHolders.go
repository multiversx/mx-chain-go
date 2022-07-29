package transactionsfee

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
)

type transactionWithResults struct {
	data.TransactionHandlerWithGasUsedAndFee
	scrs []data.TransactionHandler
	logs *data.LogData
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

		txWithResults.logs = txLog
	}

	for scrHash, scr := range txPool.Scrs {
		txWithResults, ok := transactionsAndScrs.txsWithResults[scrHash]
		if !ok {
			transactionsAndScrs.scrsNoTx[scrHash] = scr
		}

		txWithResults.scrs = append(txWithResults.scrs, scr)
	}

	return transactionsAndScrs
}
