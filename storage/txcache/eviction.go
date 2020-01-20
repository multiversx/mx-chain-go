package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// EvictionConfig is a cache eviction model
type EvictionConfig struct {
	Enabled                         bool
	NumBytesThreshold               uint32
	CountThreshold                  uint32
	NumSendersToEvictInOneStep      uint32
	ALotOfTransactionsForASender    uint32
	NumTxsToEvictForASenderWithALot uint32
}

// doEviction does cache eviction
// We do not allow more evictions to start concurrently
func (cache *TxCache) doEviction() {
	tooManyBytes := cache.areThereTooManyBytes()
	tooManyTxs := cache.areThereTooManyTxs()
	tooManySenders := cache.areThereTooManySenders()
	journal := evictionJournal{}

	isCapacityExceeded := tooManyBytes || tooManyTxs || tooManySenders
	if !isCapacityExceeded {
		return
	}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	cache.maybeEvictionInProgress.Set()
	defer cache.maybeEvictionInProgress.Unset()

	cache.onEvictionStarted()

	if tooManyTxs {
		journal.passOneNumTxs, journal.passOneNumSenders = cache.evictHighNonceTransactions()
		journal.evictionPerformed = true
	}

	if cache.shouldContinueEvictingSenders() {
		journal.passTwoNumSteps, journal.passTwoNumTxs, journal.passTwoNumSenders = cache.evictSendersInLoop()
		journal.evictionPerformed = true
	}

	cache.evictionJournal = journal
	cache.onEvictionEnded()
}

func (cache *TxCache) areThereTooManyBytes() bool {
	numBytes := cache.NumBytes()
	tooManyBytes := numBytes > int64(cache.evictionConfig.NumBytesThreshold)
	return tooManyBytes
}

func (cache *TxCache) areThereTooManySenders() bool {
	numSenders := cache.CountSenders()
	tooManySenders := numSenders > int64(cache.evictionConfig.CountThreshold)
	return tooManySenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	numTxs := cache.CountTx()
	tooManyTxs := numTxs > int64(cache.evictionConfig.CountThreshold)
	return tooManyTxs
}

func (cache *TxCache) shouldContinueEvictingSenders() bool {
	return cache.areThereTooManyTxs() || cache.areThereTooManySenders() || cache.areThereTooManyBytes()
}

// evictHighNonceTransactions removes transactions from the cache
// For senders with many transactions (> "ALotOfTransactionsForASender"), evict "NumTxsToEvictForASenderWithALot" transactions
// Also makes sure that there's no sender with 0 transactions
func (cache *TxCache) evictHighNonceTransactions() (uint32, uint32) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	cache.forEachSenderAscending(func(key string, txList *txListForSender) {
		aLot := cache.evictionConfig.ALotOfTransactionsForASender
		numTxsToEvict := cache.evictionConfig.NumTxsToEvictForASenderWithALot

		if txList.HasMoreThan(aLot) {
			txsToEvictForSender := txList.RemoveHighNonceTxs(numTxsToEvict)
			txsToEvict = append(txsToEvict, txsToEvictForSender...)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, key)
		}
	})

	if len(txsToEvict) == 0 {
		return 0, 0
	}

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

	batchesSource := cache.txListBySender.getSnapshotAscending()
	batchSize := cache.evictionConfig.NumSendersToEvictInOneStep
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
	sendersToEvict := make([]string, 0)
	txsToEvict := make([][]byte, 0)

	for _, txList := range listsToEvict {
		sendersToEvict = append(sendersToEvict, txList.sender)
		txsToEvict = append(txsToEvict, txList.getTxHashes()...)
	}

	return cache.doEvictItems(txsToEvict, sendersToEvict)
}
