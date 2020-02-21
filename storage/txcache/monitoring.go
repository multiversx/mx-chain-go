package txcache

import (
	"bytes"

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

func (cache *TxCache) monitorTxRemoval() {
	cache.numTxRemovedBetweenSelections.Increment()

	if cache.isEvictionInProgress.IsSet() {
		cache.numTxRemovedDuringEviction.Increment()
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
	numTxAdded := cache.numTxAddedDuringEviction.Reset()
	numTxRemoved := cache.numTxRemovedDuringEviction.Reset()
	log.Trace("TxCache: eviction ended", "name", cache.name, "duration", duration, "numBytes", cache.NumBytes(), "txs", cache.CountTx(), "senders", cache.CountSenders(), "numTxAddedDuringEviction", numTxAdded, "numTxRemovedDuringEviction", numTxRemoved)
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

func (cache *TxCache) monitorSelectionEnd(selection []data.TransactionHandler, selectionHashes [][]byte, stopWatch *core.StopWatch) {
	stopWatch.Stop("selection")
	duration := stopWatch.GetMeasurement("selection")
	numTxAdded := cache.numTxAddedBetweenSelections.Reset()
	numTxRemoved := cache.numTxRemovedBetweenSelections.Reset()
	log.Trace("TxCache: selection ended", "name", cache.name, "duration", duration, "numTxSelected", len(selection), "numTxAddedBetweenSelections", numTxAdded, "numTxRemovedBetweenSelections", numTxRemoved)
	cache.displaySendersHistogram()

	go cache.diagnose()
	//go cache.dumpSelection(selection, selectionHashes)
}

func (cache *TxCache) dumpSelection(selection []data.TransactionHandler, selectionHashes [][]byte) {
	if len(selection) == 0 {
		return
	}

	log.Trace("dumpSelection")

	sender := selection[0].GetSndAddress()
	senderTxs := make([]uint64, 0)
	senderNonce := cache.getNonceOfSender(sender)

	for i := 0; i < len(selection); i++ {
		tx := selection[i]
		currentSender := tx.GetSndAddress()
		currentNonce := tx.GetNonce()

		if !bytes.Equal(sender, currentSender) {
			log.Trace("sender", "addr", sender, "nonce", senderNonce, "txs", senderTxs)

			sender = currentSender
			senderTxs = make([]uint64, 0)
			senderNonce = cache.getNonceOfSender(sender)
		}

		senderTxs = append(senderTxs, currentNonce)
	}

	log.Trace("sender", "addr", sender, "txs", senderTxs)
}

func (cache *TxCache) getNonceOfSender(sender []byte) int64 {
	txList, ok := cache.txListBySender.getListForSender(string(sender))
	if !ok {
		return -2
	}
	if !txList.accountNonceKnown.IsSet() {
		return -1
	}

	return int64(txList.accountNonce.Get())
}

func (cache *TxCache) displaySendersHistogram() {
	txListBySenderMap := cache.txListBySender.backingMap
	log.Trace("TxCache.sendersHistogram:", "chunks", txListBySenderMap.ChunksCounts(), "scoreChunks", txListBySenderMap.ScoreChunksCounts())
}

func (cache *TxCache) onRemoveTxInconsistency(txHash []byte) {
	// This happens when one transaction is processed and it has to be removed from the cache, but it has already been evicted soon after its selection.
	log.Trace("TxCache.onRemoveTxInconsistency(): detected maps sync inconsistency", "name", cache.name, "tx", txHash)
}

func (txMap *txListBySenderMap) onRemoveTxInconsistency(sender string) {
	// This happens when a sender whose transactions were selected for processing is evicted from cache.
	// When it comes to remove one if its transactions due to processing, they don't exist in cache anymore.
	log.Trace("txListBySenderMap.removeTx() detected inconsistency: sender of tx not in cache", "sender", []byte(sender))
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

func (cache *TxCache) diagnose() {
	sw := core.NewStopWatch()
	sw.Start("diagnose")

	sizeInBytes := cache.NumBytes()
	numTxsEstimate := int(cache.CountTx())
	numTxsInChunks := cache.txByHash.backingMap.Count()
	txsKeys := cache.txByHash.backingMap.Keys()
	numSendersEstimate := uint32(cache.CountSenders())
	numSendersInChunks := cache.txListBySender.backingMap.Count()
	numSendersInScoreChunks := cache.txListBySender.backingMap.CountSorted()
	sendersKeys := cache.txListBySender.backingMap.Keys()
	sendersKeysSorted := cache.txListBySender.backingMap.KeysSorted()
	sendersSnapshot := cache.txListBySender.getSnapshotAscending()

	sw.Stop("diagnose")
	duration := sw.GetMeasurement("diagnose")

	fine := (numSendersEstimate == numSendersInChunks && numSendersEstimate == numSendersInScoreChunks)
	fine = fine && (len(sendersKeys) == len(sendersKeysSorted) && len(sendersKeys) == len(sendersSnapshot))
	fine = fine && (int(numSendersEstimate) == len(sendersKeys))
	fine = fine && (numTxsEstimate == numTxsInChunks && numTxsEstimate == len(txsKeys))

	logFunc := log.Trace
	if !fine {
		logFunc = log.Warn
	}

	logFunc("Diagnose", "name", cache.name, "duration", duration, "fine", fine)
	logFunc("Size:", "current", sizeInBytes, "max", cache.config.NumBytesThreshold)
	logFunc("NumSenders:", "estimate", numSendersEstimate, "inChunks", numSendersInChunks, "inScoreChunks", numSendersInScoreChunks)
	logFunc("NumSenders (continued):", "keys", len(sendersKeys), "keysSorted", len(sendersKeysSorted), "snapshot", len(sendersSnapshot))
	logFunc("NumTxs:", "estimate", numTxsEstimate, "inChunks", numTxsInChunks, "keys", len(txsKeys))

	//cache.displayNonceGaps()
}

type gapJournal struct {
	accountNonce        uint64
	previousNonce       uint64
	numFailedSelections int64
}

func (cache *TxCache) displayNonceGaps() {
	gaps := make(map[string]gapJournal)

	cache.forEachSenderDescending(func(key string, txList *txListForSender) {
		if !txList.copyDetectedGap {
			return
		}

		gaps[key] = gapJournal{
			accountNonce:        txList.accountNonce.Get(),
			previousNonce:       txList.copyPreviousNonce,
			numFailedSelections: txList.numFailedSelections.Get(),
		}
	})

	if len(gaps) > 0 {
		log.Trace("Detected gaps", "count", len(gaps))

		dumped := 0
		for key, value := range gaps {
			log.Trace("Gap", "sender", []byte(key), "info", value)

			dumped++
			if dumped > 100 {
				break
			}
		}
	}
}
