package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
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
	log.Trace("TxCache: eviction started")
	cache.displayState()
	sw := core.NewStopWatch()
	sw.Start("eviction")
	return sw
}

func (cache *TxCache) monitorEvictionEnd(stopWatch *core.StopWatch) {
	duration := stopWatch.GetMeasurement("selection")
	numTx := cache.numTxAddedDuringEviction.Reset()
	log.Trace("TxCache: eviction ended", "duration", duration, "numTxAddedDuringEviction", numTx)
	cache.evictionJournal.display()
	cache.displayState()
}

func (cache *TxCache) monitorSelectionStart() *core.StopWatch {
	log.Trace("TxCache: selection started")
	sw := core.NewStopWatch()
	sw.Start("selection")
	return sw
}

func (cache *TxCache) monitorSelectionEnd(stopWatch *core.StopWatch) {
	duration := stopWatch.GetMeasurement("selection")
	numTx := cache.numTxAddedBetweenSelections.Reset()
	log.Trace("TxCache: selection ended", "duration", duration, "numTxAddedBetweenSelections", numTx)
}

func (cache *TxCache) displayState() {
	txListBySenderMap := cache.txListBySender.backingMap
	chunksCount := txListBySenderMap.Count()
	scoreChunksCount := txListBySenderMap.CountSorted()
	sendersCount := uint32(cache.CountSenders())

	log.Trace("TxCache:", "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", sendersCount)

	if chunksCount != sendersCount {
		log.Error("TxCache.CountSenders() inconsistency:", "counter", sendersCount, "in-map", chunksCount)
	}
	if chunksCount != scoreChunksCount {
		log.Error("TxCache.txListBySender.backingMap counts inconsistency:", "chunks", chunksCount, "scoreChunks", scoreChunksCount)
	}

	log.Trace("TxCache.txListBySender.histogram:", "chunks", txListBySenderMap.ChunksCounts(), "scoreChunks", txListBySenderMap.ScoreChunksCounts())
}

func (cache *TxCache) onRemoveTxInconsistency(txHash []byte) {
	// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
	log.Error("TxCache.onRemoveTxInconsistency(): detected maps sync inconsistency", "tx", txHash)
}

func (txMap *txListBySenderMap) onRemoveTxInconsistency(sender string) {
	log.Error("txListBySenderMap.removeTx() detected inconsistency: sender of tx not in cache", "sender", sender)
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
	log.Trace("Eviction journal:")
	log.Trace("Pass 1:", "txs", journal.passOneNumTxs, "senders", journal.passOneNumSenders)
	log.Trace("Pass 2:", "txs", journal.passTwoNumTxs, "senders", journal.passTwoNumSenders, "steps", journal.passTwoNumSteps)
}
