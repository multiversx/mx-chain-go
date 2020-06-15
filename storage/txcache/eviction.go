package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// doEviction does cache eviction
// We do not allow more evictions to start concurrently
func (cache *TxCache) doEviction() {
	if cache.isEvictionInProgress.IsSet() {
		return
	}

	if !cache.isCapacityExceeded() {
		return
	}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	cache.isEvictionInProgress.Set()
	defer cache.isEvictionInProgress.Unset()

	if !cache.isCapacityExceeded() {
		return
	}

	stopWatch := cache.monitorEvictionStart()
	cache.makeSnapshotOfSenders()

	journal := evictionJournal{}
	journal.passOneNumSteps, journal.passOneNumTxs, journal.passOneNumSenders = cache.evictSendersInLoop()
	journal.evictionPerformed = true
	cache.evictionJournal = journal

	cache.monitorEvictionEnd(stopWatch)
	cache.destroySnapshotOfSenders()
}

func (cache *TxCache) makeSnapshotOfSenders() {
	cache.evictionSnapshotOfSenders = cache.txListBySender.getSnapshotAscending()
}

func (cache *TxCache) destroySnapshotOfSenders() {
	cache.evictionSnapshotOfSenders = nil
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

// This is called concurrently by two goroutines: the eviction one and the sweeping one
func (cache *TxCache) doEvictItems(txsToEvict [][]byte, sendersToEvict []string) (countTxs uint32, countSenders uint32) {
	countTxs = cache.txByHash.RemoveTxsBulk(txsToEvict)
	countSenders = cache.txListBySender.RemoveSendersBulk(sendersToEvict)
	return
}

func (cache *TxCache) evictSendersInLoop() (uint32, uint32, uint32) {
	return cache.evictSendersWhile(cache.isCapacityExceeded)
}

// evictSendersWhileTooManyTxs removes transactions in a loop, as long as "shouldContinue" is true
// One batch of senders is removed in each step
func (cache *TxCache) evictSendersWhile(shouldContinue func() bool) (step uint32, numTxs uint32, numSenders uint32) {
	if !shouldContinue() {
		return
	}

	snapshot := cache.evictionSnapshotOfSenders
	snapshotLength := uint32(len(snapshot))
	batchSize := cache.config.NumSendersToPreemptivelyEvict
	batchStart := uint32(0)

	for step = 0; shouldContinue(); step++ {
		batchEnd := batchStart + batchSize
		batchEndBounded := core.MinUint32(batchEnd, snapshotLength)
		batch := snapshot[batchStart:batchEndBounded]

		numTxsEvictedInStep, numSendersEvictedInStep := cache.evictSendersAndTheirTxs(batch)

		numTxs += numTxsEvictedInStep
		numSenders += numSendersEvictedInStep
		batchStart += batchSize

		reachedEnd := batchStart >= snapshotLength
		noTxsEvicted := numTxsEvictedInStep == 0
		incompleteBatch := numSendersEvictedInStep < batchSize

		shouldBreak := noTxsEvicted || incompleteBatch || reachedEnd
		if shouldBreak {
			break
		}
	}

	return
}

// This is called concurrently by two goroutines: the eviction one and the sweeping one
func (cache *TxCache) evictSendersAndTheirTxs(listsToEvict []*txListForSender) (uint32, uint32) {
	sendersToEvict := make([]string, 0, len(listsToEvict))
	txsToEvict := make([][]byte, 0, approximatelyCountTxInLists(listsToEvict))

	for _, txList := range listsToEvict {
		sendersToEvict = append(sendersToEvict, txList.sender)
		txsToEvict = append(txsToEvict, txList.getTxHashes()...)
	}

	return cache.doEvictItems(txsToEvict, sendersToEvict)
}
