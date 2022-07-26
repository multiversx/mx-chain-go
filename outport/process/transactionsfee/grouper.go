package transactionsfee

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
)

type transactionWithResults struct {
	data.TransactionHandlerWithGasUsedAndFee
	scrs    []data.TransactionHandler
	receipt []*receipt.Receipt
	logs    *data.LogData
}

type groupedTransactionsAndScrs struct {
	txsWithResults map[string]*transactionWithResults
	scrsNoTx       map[string]data.TransactionHandlerWithGasUsedAndFee
}

func newGroupedTransactionsAndScrs(nrTxs, nrScrs int) *groupedTransactionsAndScrs {
	return &groupedTransactionsAndScrs{
		txsWithResults: make(map[string]*transactionWithResults, nrTxs),
		scrsNoTx:       make(map[string]data.TransactionHandlerWithGasUsedAndFee, nrScrs),
	}
}

func groupTransactionsWithResults(txPool *indexer.Pool) *groupedTransactionsAndScrs {
	totalTxs := len(txPool.Txs) + len(txPool.Invalid) + len(txPool.Rewards)
	if totalTxs == 0 && len(txPool.Scrs) == 0 {
		return newGroupedTransactionsAndScrs(0, 0)
	}

	groupedTxsAndScrs := newGroupedTransactionsAndScrs(totalTxs, len(txPool.Scrs))
	for txHash, tx := range txPool.Txs {
		groupedTxsAndScrs.txsWithResults[txHash] = &transactionWithResults{
			TransactionHandlerWithGasUsedAndFee: tx,
		}
	}

	for _, txLog := range txPool.Logs {
		txWithResults, ok := groupedTxsAndScrs.txsWithResults[txLog.TxHash]
		if !ok {
			continue
		}

		txWithResults.logs = txLog
	}

	for scrHash, scr := range txPool.Scrs {
		txWithResults, ok := groupedTxsAndScrs.txsWithResults[scrHash]
		if !ok {
			groupedTxsAndScrs.scrsNoTx[scrHash] = scr
		}

		txWithResults.scrs = append(txWithResults.scrs, scr)
	}

	return groupedTxsAndScrs
}
