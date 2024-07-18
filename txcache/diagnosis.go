package txcache

import (
	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// Diagnose checks the state of the cache for inconsistencies and displays a summary (senders and transactions).
func (cache *TxCache) Diagnose(_ bool) {
	sw := core.NewStopWatch()
	sw.Start("diagnose")

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

	cache.displaySendersAsDiagnostics()
	cache.displayTransactionsAsDiagnostics()

	sw.Stop("diagnose")
	duration := sw.GetMeasurement("diagnose")

	log.Debug("TxCache.Diagnose()",
		"duration", duration,
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

func (cache *TxCache) displaySendersAsDiagnostics() {
	if log.GetLevel() > logger.LogTrace {
		return
	}

	senders := cache.txListBySender.getSnapshotAscending()

	if len(senders) == 0 {
		return
	}

	numToDisplay := core.MinInt(diagnosisMaxSendersToDisplay, len(senders))
	logDiagnoseSenders.Trace("Senders (as newline-separated JSON)", "numSenders", len(senders), "numToDisplay", numToDisplay)
	logDiagnoseSenders.Trace(marshalSendersToNewlineDelimitedJson(senders[:numToDisplay]))
}

func (cache *TxCache) displayTransactionsAsDiagnostics() {
	if log.GetLevel() > logger.LogTrace {
		return
	}

	transactions := cache.getAllTransactions()

	if len(transactions) == 0 {
		return
	}

	numToDisplay := core.MinInt(diagnosisMaxTransactionsToDisplay, len(transactions))
	logDiagnoseTransactions.Trace("Transactions (as newline-separated JSON)", "numTransactions", len(transactions), "numToDisplay", numToDisplay)
	logDiagnoseTransactions.Trace(marshalTransactionsToNewlineDelimitedJson(transactions[:numToDisplay]))
}
