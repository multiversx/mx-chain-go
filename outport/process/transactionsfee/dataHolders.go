package transactionsfee

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
)

type txHandlerWithFeeInfo interface {
	GetTxHandler() data.TransactionHandler
	GetFeeInfo() *outportcore.FeeInfo
}

type transactionWithResults struct {
	txHandlerWithFeeInfo
	scrs []txHandlerWithFeeInfo
	log  *data.LogData
}

type transactionsAndScrsHolder struct {
	txsWithResults map[string]*transactionWithResults
	scrsNoTx       map[string]txHandlerWithFeeInfo
}

func newTransactionsAndScrsHolder(nrTxs, nrScrs int) *transactionsAndScrsHolder {
	return &transactionsAndScrsHolder{
		txsWithResults: make(map[string]*transactionWithResults, nrTxs),
		scrsNoTx:       make(map[string]txHandlerWithFeeInfo, nrScrs),
	}
}

func prepareTransactionsAndScrs(txPool *outportcore.TransactionPool) *transactionsAndScrsHolder {
	totalTxs := len(txPool.Transactions) + len(txPool.InvalidTxs) + len(txPool.Rewards)
	if totalTxs == 0 && len(txPool.SmartContractResults) == 0 {
		return newTransactionsAndScrsHolder(0, 0)
	}

	transactionsAndScrs := newTransactionsAndScrsHolder(totalTxs, len(txPool.SmartContractResults))
	for txHash, tx := range txPool.Transactions {
		transactionsAndScrs.txsWithResults[txHash] = &transactionWithResults{
			txHandlerWithFeeInfo: tx,
		}
	}

	for _, txLog := range txPool.Logs {
		txHash := txLog.TxHash
		txWithResults, ok := transactionsAndScrs.txsWithResults[txHash]
		if !ok {
			continue
		}

		txWithResults.log = &data.LogData{
			LogHandler: txLog.Log,
			TxHash:     txHash,
		}
	}

	for scrHash, scrHandler := range txPool.SmartContractResults {
		scr := scrHandler.SmartContractResult

		txWithResults, ok := transactionsAndScrs.txsWithResults[hex.EncodeToString(scr.OriginalTxHash)]
		if !ok {
			transactionsAndScrs.scrsNoTx[scrHash] = scrHandler
			continue
		}

		txWithResults.scrs = append(txWithResults.scrs, scrHandler)
	}

	return transactionsAndScrs
}
