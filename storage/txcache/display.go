package txcache

import (
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("txcache")

func (cache *TxCache) onEvictionStarted() {
	log.Trace("TxCache.onEvictionStarted()")
	cache.displayState()
}

func (cache *TxCache) onEvictionEnded() {
	log.Trace("TxCache.onEvictionEnded()")
	cache.evictionJournal.display()
	cache.displayState()
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
