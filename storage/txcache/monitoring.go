package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("txcache")

func (cache *TxCache) monitorTxAddition() {
	cache.numTxAddedBetweenSelections.Increment()

	if cache.isEvictionInProgress.IsSet() {
		cache.numTxAddedDuringEviction.Increment()
	}
}

func (cache *TxCache) monitorEvictionStart() *core.StopWatch {
	log.Trace("TxCache: eviction started", "name", cache.name, "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders())
	cache.displaySendersHistogram()
	sw := core.NewStopWatch()
	sw.Start("eviction")
	return sw
}

func (cache *TxCache) monitorEvictionEnd(stopWatch *core.StopWatch) {
	stopWatch.Stop("eviction")
	duration := stopWatch.GetMeasurement("eviction")
	numTx := cache.numTxAddedDuringEviction.Reset()
	log.Trace("TxCache: eviction ended", "name", cache.name, "duration", duration, "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders(), "numTxAddedDuringEviction", numTx)
	cache.evictionJournal.display()
	cache.displaySendersHistogram()
}

func (cache *TxCache) monitorSelectionStart() *core.StopWatch {
	log.Trace("TxCache: selection started", "name", cache.name, "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders())
	cache.displaySendersHistogram()
	sw := core.NewStopWatch()
	sw.Start("selection")
	return sw
}

func (cache *TxCache) monitorSelectionEnd(selection []data.TransactionHandler, stopWatch *core.StopWatch) {
	stopWatch.Stop("selection")
	duration := stopWatch.GetMeasurement("selection")
	numTx := cache.numTxAddedBetweenSelections.Reset()
	log.Trace("TxCache: selection ended", "name", cache.name, "duration", duration, "numTxSelected", len(selection), "numTxAddedBetweenSelections", numTx)
	cache.displaySendersHistogram()
}

func (cache *TxCache) displaySendersHistogram() {
	txListBySenderMap := cache.txListBySender.backingMap
	log.Trace("TxCache.sendersHistogram:", "chunks", txListBySenderMap.ChunksCounts(), "scoreChunks", txListBySenderMap.ScoreChunksCounts())
}

func (cache *TxCache) onRemoveTxInconsistency(txHash []byte) {
	// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
	log.Error("TxCache.onRemoveTxInconsistency(): detected maps sync inconsistency", "name", cache.name, "tx", txHash)
}

func (txMap *txListBySenderMap) onRemoveTxInconsistency(sender string) {
	log.Error("txListBySenderMap.removeTx() detected inconsistency: sender of tx not in cache", "sender", []byte(sender))
}

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	evictionPerformed bool
	passOneNumTxs     uint32
	passOneNumSenders uint32
	passTwoNumTxs     uint32
	passTwoNumSenders uint32
	passTwoNumSteps   uint32
}

func (journal *evictionJournal) display() {
	log.Trace("Eviction.pass1:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders)
	log.Trace("Eviction.pass2:", "txs", journal.passTwoNumTxs, "senders", journal.passTwoNumSenders, "steps", journal.passTwoNumSteps)
}
