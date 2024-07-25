package txcache

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

func (cache *TxCache) monitorSelectionStart(contextualLogger logger.Logger) *core.StopWatch {
	contextualLogger.Debug("monitorSelectionStart()", "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders())
	sw := core.NewStopWatch()
	sw.Start("selection")
	return sw
}

func (cache *TxCache) monitorSelectionEnd(contextualLog logger.Logger, stopWatch *core.StopWatch, selection []*WrappedTransaction) {
	stopWatch.Stop("selection")
	duration := stopWatch.GetMeasurement("selection")

	numSendersSelected := cache.numSendersSelected.Reset()
	numSendersWithInitialGap := cache.numSendersWithInitialGap.Reset()
	numSendersWithMiddleGap := cache.numSendersWithMiddleGap.Reset()

	contextualLog.Debug("monitorSelectionEnd()", "duration", duration,
		"numTxSelected", len(selection),
		"numSendersSelected", numSendersSelected,
		"numSendersWithInitialGap", numSendersWithInitialGap,
		"numSendersWithMiddleGap", numSendersWithMiddleGap,
	)
}

func displaySelectionOutcome(contextualLogger logger.Logger, sortedSenders []*txListForSender, selection []*WrappedTransaction) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	if len(sortedSenders) > 0 {
		contextualLogger.Trace("Sorted senders (as newline-separated JSON):")
		contextualLogger.Trace(marshalSendersToNewlineDelimitedJson(sortedSenders))
	} else {
		contextualLogger.Trace("Sorted senders: none")
	}

	if len(selection) > 0 {
		contextualLogger.Trace("Selected transactions (as newline-separated JSON):")
		contextualLogger.Trace(marshalTransactionsToNewlineDelimitedJson(selection))
	} else {
		contextualLogger.Trace("Selected transactions: none")
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
