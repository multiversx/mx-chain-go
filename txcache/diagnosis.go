package txcache

import (
	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// Diagnose checks the state of the cache for inconsistencies and displays a summary, senders and transactions.
func (cache *TxCache) Diagnose(_ bool) {
	cache.diagnoseCounters()
	cache.diagnoseTransactions()
	cache.diagnoseSelection()
}

func (cache *TxCache) diagnoseCounters() {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	sizeInBytes := cache.NumBytes()
	numTxsEstimate := int(cache.CountTx())
	numTxsInChunks := cache.txByHash.backingMap.Count()
	txsKeys := cache.txByHash.backingMap.Keys()
	numSendersEstimate := int(cache.CountSenders())
	numSendersInChunks := cache.txListBySender.backingMap.Count()
	sendersKeys := cache.txListBySender.backingMap.Keys()

	fine := numSendersEstimate == numSendersInChunks
	fine = fine && (int(numSendersEstimate) == len(sendersKeys))
	fine = fine && (numTxsEstimate == numTxsInChunks && numTxsEstimate == len(txsKeys))

	log.Debug("diagnoseCounters",
		"fine", fine,
		"numTxsEstimate", numTxsEstimate,
		"numTxsInChunks", numTxsInChunks,
		"len(txsKeys)", len(txsKeys),
		"sizeInBytes", sizeInBytes,
		"numBytesThreshold", cache.config.NumBytesThreshold,
		"numSendersEstimate", numSendersEstimate,
		"numSendersInChunks", numSendersInChunks,
		"len(sendersKeys)", len(sendersKeys),
	)
}

func (cache *TxCache) diagnoseTransactions() {
	if logDiagnoseTransactions.GetLevel() > logger.LogTrace {
		return
	}

	transactions := cache.getAllTransactions()

	if len(transactions) == 0 {
		return
	}

	numToDisplay := core.MinInt(diagnosisMaxTransactionsToDisplay, len(transactions))
	logDiagnoseTransactions.Trace("diagnoseTransactions", "numTransactions", len(transactions), "numToDisplay", numToDisplay)
	logDiagnoseTransactions.Trace(marshalTransactionsToNewlineDelimitedJson(transactions[:numToDisplay]))
}

func (cache *TxCache) diagnoseSelection() {
	if logDiagnoseSelection.GetLevel() > logger.LogDebug {
		return
	}

	transactions := cache.doSelectTransactions(
		logDiagnoseSelection,
		diagnosisSelectionGasRequested,
	)

	displaySelectionOutcome(logDiagnoseSelection, transactions)
}

func displaySelectionOutcome(contextualLogger logger.Logger, selection []*WrappedTransaction) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	if len(selection) > 0 {
		contextualLogger.Trace("displaySelectionOutcome - transactions (as newline-separated JSON):")
		contextualLogger.Trace(marshalTransactionsToNewlineDelimitedJson(selection))
	} else {
		contextualLogger.Trace("displaySelectionOutcome - transactions: none")
	}
}
