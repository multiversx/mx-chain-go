package txcache

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	numTxs     uint32
	numSenders uint32
	numSteps   uint32
}

// doEviction does cache eviction
// We do not allow more evictions to start concurrently
func (cache *TxCache) doEviction() *evictionJournal {
	if cache.isEvictionInProgress.IsSet() {
		return nil
	}

	if !cache.isCapacityExceeded() {
		return nil
	}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	_ = cache.isEvictionInProgress.SetReturningPrevious()
	defer cache.isEvictionInProgress.Reset()

	if !cache.isCapacityExceeded() {
		return nil
	}

	logRemove.Debug("doEviction(): before eviction",
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)

	stopWatch := core.NewStopWatch()
	stopWatch.Start("eviction")

	sendersSnapshot := cache.txListBySender.getSnapshotAscending()
	evictionJournal := cache.evictSendersInLoop(sendersSnapshot)

	stopWatch.Stop("eviction")

	logRemove.Debug(
		"doEviction(): after eviction",
		"num bytes", cache.NumBytes(),
		"num now", cache.CountTx(),
		"num senders", cache.CountSenders(),
		"duration", stopWatch.GetMeasurement("eviction"),
		"evicted txs", evictionJournal.numTxs,
		"evicted senders", evictionJournal.numSenders,
		"eviction steps", evictionJournal.numSteps,
	)

	return &evictionJournal
}

func (cache *TxCache) isCapacityExceeded() bool {
	return cache.areThereTooManyBytes() || cache.areThereTooManySenders() || cache.areThereTooManyTxs()
}

func (cache *TxCache) areThereTooManyBytes() bool {
	numBytes := cache.NumBytes()
	tooManyBytes := numBytes > int(cache.config.NumBytesThreshold)
	return tooManyBytes
}

func (cache *TxCache) areThereTooManySenders() bool {
	numSenders := cache.CountSenders()
	tooManySenders := numSenders > uint64(cache.config.CountThreshold)
	return tooManySenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	numTxs := cache.CountTx()
	tooManyTxs := numTxs > uint64(cache.config.CountThreshold)
	return tooManyTxs
}

func (cache *TxCache) doEvictItems(txsToEvict [][]byte, sendersToEvict []string) (countTxs uint32, countSenders uint32) {
	countTxs = cache.txByHash.RemoveTxsBulk(txsToEvict)
	countSenders = cache.txListBySender.RemoveSendersBulk(sendersToEvict)
	return
}

func (cache *TxCache) evictSendersInLoop(sendersSnapshot []*txListForSender) evictionJournal {
	return cache.evictSendersWhile(sendersSnapshot, cache.isCapacityExceeded)
}

// evictSendersWhileTooManyTxs removes transactions in a loop, as long as "shouldContinue" is true
// One batch of senders is removed in each step
func (cache *TxCache) evictSendersWhile(sendersSnapshot []*txListForSender, shouldContinue func() bool) evictionJournal {
	if !shouldContinue() {
		return evictionJournal{}
	}

	snapshotLength := uint32(len(sendersSnapshot))
	batchSize := cache.config.NumSendersToPreemptivelyEvict
	batchStart := uint32(0)

	journal := evictionJournal{}

	for ; shouldContinue(); journal.numSteps++ {
		batchEnd := batchStart + batchSize
		batchEndBounded := core.MinUint32(batchEnd, snapshotLength)
		batch := sendersSnapshot[batchStart:batchEndBounded]

		numTxsEvictedInStep, numSendersEvictedInStep := cache.evictSendersAndTheirTxs(batch)

		journal.numTxs += numTxsEvictedInStep
		journal.numSenders += numSendersEvictedInStep
		batchStart += batchSize

		reachedEnd := batchStart >= snapshotLength
		noTxsEvicted := numTxsEvictedInStep == 0
		incompleteBatch := numSendersEvictedInStep < batchSize

		shouldBreak := noTxsEvicted || incompleteBatch || reachedEnd
		if shouldBreak {
			break
		}
	}

	return journal
}

func (cache *TxCache) evictSendersAndTheirTxs(listsToEvict []*txListForSender) (uint32, uint32) {
	sendersToEvict := make([]string, 0, len(listsToEvict))
	txsToEvict := make([][]byte, 0, approximatelyCountTxInLists(listsToEvict))

	for _, txList := range listsToEvict {
		sendersToEvict = append(sendersToEvict, txList.sender)
		txsToEvict = append(txsToEvict, txList.getTxsHashes()...)
	}

	return cache.doEvictItems(txsToEvict, sendersToEvict)
}
