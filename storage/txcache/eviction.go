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

	tooManyBytes := cache.areThereTooManyBytes()
	tooManyTxs := cache.areThereTooManyTxs()
	tooManySenders := cache.areThereTooManySenders()

	isCapacityExceeded := tooManyBytes || tooManyTxs || tooManySenders
	if !isCapacityExceeded {
		return
	}

	journal := evictionJournal{}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	cache.isEvictionInProgress.Set()
	defer cache.isEvictionInProgress.Unset()

	stopWatch := cache.monitorEvictionStart()

	if tooManyTxs {
		cache.makeSnapshotOfSenders()
		journal.passOneNumTxs, journal.passOneNumSenders = cache.evictHighNonceTransactions()
		journal.evictionPerformed = true
	}

	if cache.shouldContinueEvictingSenders() {
		cache.makeSnapshotOfSenders()
		journal.passTwoNumSteps, journal.passTwoNumTxs, journal.passTwoNumSenders = cache.evictSendersInLoop()
		journal.evictionPerformed = true
	}

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

func (cache *TxCache) areThereTooManyBytes() bool {
	numBytes := cache.NumBytes()
	tooManyBytes := numBytes > int64(cache.config.NumBytesThreshold)
	return tooManyBytes
}

func (cache *TxCache) areThereTooManySenders() bool {
	numSenders := cache.CountSenders()
	tooManySenders := numSenders > int64(cache.config.CountThreshold)
	return tooManySenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	numTxs := cache.CountTx()
	tooManyTxs := numTxs > int64(cache.config.CountThreshold)
	return tooManyTxs
}

func (cache *TxCache) shouldContinueEvictingSenders() bool {
	return cache.areThereTooManyTxs() || cache.areThereTooManySenders() || cache.areThereTooManyBytes()
}

// evictHighNonceTransactions removes transactions from the cache
// For senders with many transactions (> "LargeNumOfTxsForASender"), evict "NumTxsToEvictFromASender" transactions
// Also makes sure that there's no sender with 0 transactions
func (cache *TxCache) evictHighNonceTransactions() (uint32, uint32) {
	threshold := cache.config.LargeNumOfTxsForASender
	numTxsToEvict := cache.config.NumTxsToEvictFromASender

	// Heuristics: estimate that ~10% of senders have more transactions than the threshold
	sendersToEvictInitialCapacity := len(cache.evictionSnapshotOfSenders)/10 + 1
	txsToEvictInitialCapacity := sendersToEvictInitialCapacity * int(numTxsToEvict)

	sendersToEvict := make([]string, 0, sendersToEvictInitialCapacity)
	txsToEvict := make([][]byte, 0, txsToEvictInitialCapacity)

	for _, txList := range cache.evictionSnapshotOfSenders {
		if txList.HasMoreThan(threshold) {
			txsToEvictForSender := txList.RemoveHighNonceTxs(numTxsToEvict)
			cache.txListBySender.notifyScoreChange(txList)
			txsToEvict = append(txsToEvict, txsToEvictForSender...)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, txList.sender)
		}
	}

	// Note that, at this very moment, high nonce transactions have been evicted from senders' lists,
	// but not yet from the map of transactions.
	//
	// This may cause slight inconsistencies, such as:
	// - if a tx previously (recently) removed from the sender's list ("RemoveHighNonceTxs") arrives again at the pool,
	// before the execution of "doEvictItems", the tx will be ignored as it still exists (for a short time) in the map of transactions.
	return cache.doEvictItems(txsToEvict, sendersToEvict)
}

func (cache *TxCache) doEvictItems(txsToEvict [][]byte, sendersToEvict []string) (countTxs uint32, countSenders uint32) {
	countTxs = cache.txByHash.RemoveTxsBulk(txsToEvict)
	countSenders = cache.txListBySender.RemoveSendersBulk(sendersToEvict)
	return
}

func (cache *TxCache) evictSendersInLoop() (uint32, uint32, uint32) {
	return cache.evictSendersWhile(cache.shouldContinueEvictingSenders)
}

// evictSendersWhileTooManyTxs removes transactions in a loop, as long as "shouldContinue" is true
// One batch of senders is removed in each step
// Before starting the loop, the senders are sorted as specified by "sendersSortKind"
func (cache *TxCache) evictSendersWhile(shouldContinue func() bool) (step uint32, countTxs uint32, countSenders uint32) {
	if !shouldContinue() {
		return
	}

	batchesSource := cache.evictionSnapshotOfSenders
	batchSize := cache.config.NumSendersToEvictInOneStep
	batchStart := uint32(0)

	for step = 0; shouldContinue(); step++ {
		batchEnd := core.MinUint32(batchStart+batchSize, uint32(len(batchesSource)))
		batch := batchesSource[batchStart:batchEnd]

		countTxsEvictedInStep, countSendersEvictedInStep := cache.evictSendersAndTheirTxs(batch)

		countTxs += countTxsEvictedInStep
		countSenders += countSendersEvictedInStep
		batchStart += batchSize

		shouldBreak := countTxsEvictedInStep == 0 || countSendersEvictedInStep < batchSize
		if shouldBreak {
			break
		}
	}

	return
}

func (cache *TxCache) evictSendersAndTheirTxs(listsToEvict []*txListForSender) (uint32, uint32) {
	sendersToEvict := make([]string, 0, len(listsToEvict))
	txsToEvict := make([][]byte, 0, approximatelyCountTxInLists(listsToEvict))

	for _, txList := range listsToEvict {
		sendersToEvict = append(sendersToEvict, txList.sender)
		txsToEvict = append(txsToEvict, txList.getTxHashes()...)
	}

	return cache.doEvictItems(txsToEvict, sendersToEvict)
}
