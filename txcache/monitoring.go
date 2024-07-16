package txcache

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("txcache/main")
var logAdd = logger.GetOrCreate("txcache/add")
var logRemove = logger.GetOrCreate("txcache/remove")
var logSelect = logger.GetOrCreate("txcache/select")

func (cache *TxCache) monitorEvictionWrtSenderLimit(sender []byte, evicted [][]byte) {
	logRemove.Debug("monitorEvictionWrtSenderLimit()", "sender", sender, "num", len(evicted))
}

func (cache *TxCache) monitorEvictionWrtSenderNonce(sender []byte, senderNonce uint64, evicted [][]byte) {
	logRemove.Trace("monitorEvictionWrtSenderNonce()", "sender", sender, "nonce", senderNonce, "num", len(evicted))
}

func (cache *TxCache) monitorEvictionStart() *core.StopWatch {
	logRemove.Debug("monitorEvictionStart()", "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders())
	sw := core.NewStopWatch()
	sw.Start("eviction")
	return sw
}

func (cache *TxCache) monitorEvictionEnd(stopWatch *core.StopWatch) {
	stopWatch.Stop("eviction")
	duration := stopWatch.GetMeasurement("eviction")
	logRemove.Debug("monitorEvictionEnd()", "duration", duration, "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders())
	cache.evictionJournal.display()
}

func (cache *TxCache) monitorSelectionStart() *core.StopWatch {
	logSelect.Debug("monitorSelectionStart()", "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders())
	sw := core.NewStopWatch()
	sw.Start("selection")
	return sw
}

func (cache *TxCache) monitorSelectionEnd(stopWatch *core.StopWatch, selection []*WrappedTransaction) {
	stopWatch.Stop("selection")
	duration := stopWatch.GetMeasurement("selection")

	numSendersSelected := cache.numSendersSelected.Reset()
	numSendersWithInitialGap := cache.numSendersWithInitialGap.Reset()
	numSendersWithMiddleGap := cache.numSendersWithMiddleGap.Reset()

	logSelect.Debug("monitorSelectionEnd()", "duration", duration,
		"numTxSelected", len(selection),
		"numSendersSelected", numSendersSelected,
		"numSendersWithInitialGap", numSendersWithInitialGap,
		"numSendersWithMiddleGap", numSendersWithMiddleGap,
	)
}

func displaySelectionOutcome(sortedSenders []*txListForSender, selection []*WrappedTransaction) {
	if logSelect.GetLevel() > logger.LogTrace {
		return
	}

	if len(sortedSenders) > 0 {
		logSelect.Trace("Sorted senders (as newline-separated JSON):")
		logSelect.Trace(marshalSendersToNewlineDelimitedJson(sortedSenders))
	}

	if len(selection) > 0 {
		logSelect.Trace("Selected transactions (as newline-separated JSON):")
		logSelect.Trace(marshalTransactionsToNewlineDelimitedJson(selection))
	}
}

type batchSelectionJournal struct {
	selectedNum   int
	selectedGas   uint64
	isFirstBatch  bool
	hasInitialGap bool
	hasMiddleGap  bool
}

func (cache *TxCache) monitorBatchSelectionEnd(journal batchSelectionJournal) {
	if !journal.isFirstBatch {
		return
	}

	if journal.hasInitialGap {
		cache.numSendersWithInitialGap.Increment()
	} else if journal.hasMiddleGap {
		// Currently, we only count middle gaps on first batch (for simplicity)
		cache.numSendersWithMiddleGap.Increment()
	}

	if journal.selectedNum > 0 {
		cache.numSendersSelected.Increment()
	}
}

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	evictionPerformed bool
	passOneNumTxs     uint32
	passOneNumSenders uint32
	passOneNumSteps   uint32
}

func (journal *evictionJournal) display() {
	logRemove.Debug("Eviction.pass1:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders, "steps", journal.passOneNumSteps)
}

// Diagnose checks the state of the cache for inconsistencies and displays a summary
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
	senders := cache.txListBySender.getSnapshotAscending()

	cache.displaySendersSummary(senders)

	sw.Stop("diagnose")
	duration := sw.GetMeasurement("diagnose")

	fine := numSendersEstimate == numSendersInChunks
	fine = fine && len(sendersKeys) == len(senders)
	fine = fine && (int(numSendersEstimate) == len(sendersKeys))
	fine = fine && (numTxsEstimate == numTxsInChunks && numTxsEstimate == len(txsKeys))

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
		"len(senders)", len(senders),
	)
}

func (cache *TxCache) displaySendersSummary(senders []*txListForSender) {
	if log.GetLevel() > logger.LogTrace {
		return
	}

	if len(senders) == 0 {
		return
	}

	log.Trace("displaySendersSummary(), as newline-separated JSON:")
	log.Trace(marshalSendersToNewlineDelimitedJson(senders))
}

func monitorSendersScoreHistogram(scoreGroups [][]*txListForSender) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	stringBuilder := strings.Builder{}

	for i, group := range scoreGroups {
		if len(group) == 0 {
			continue
		}

		stringBuilder.WriteString(fmt.Sprintf("#%d: %d; ", i, len(group)))
	}

	log.Debug("monitorSendersScoreHistogram()", "histogram", stringBuilder.String())
}
