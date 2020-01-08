package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// EvictionConfig is a cache eviction model
type EvictionConfig struct {
	Enabled                         bool
	NumBytesThreshold               uint32
	CountThreshold                  uint32
	ThresholdEvictSenders           uint32
	NumSendersToEvictInOneStep      uint32
	ALotOfTransactionsForASender    uint32
	NumTxsToEvictForASenderWithALot uint32
}

// doEviction does cache eviction
// We do not allow more evictions to start concurrently
func (cache *TxCache) doEviction() evictionJournal {
	if !cache.areThereTooManyBytes() && !cache.areThereTooManyTxs() {
		return evictionJournal{}
	}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	//log.Info("TxCache.doEviction()")

	journal := evictionJournal{}

	if cache.isSizeEvictionEnabled() && cache.areThereTooManyBytes() {
		journal.passZeroNumSteps, journal.passZeroNumTxs, journal.passZeroNumSenders = cache.evictSendersWhileTooManyBytes()
	}

	if cache.areThereTooManySenders() {
		journal.passOneNumTxs, journal.passOneNumSenders = cache.evictOldestSenders()
	}

	if cache.areThereTooManyTxs() {
		journal.passTwoNumTxs, journal.passTwoNumSenders = cache.evictHighNonceTransactions()
	}

	if cache.areThereTooManyTxs() && !cache.areThereJustAFewSenders() {
		journal.passThreeNumSteps, journal.passThreeNumTxs, journal.passThreeNumSenders = cache.evictSendersWhileTooManyTxs()
	}

	cache.evictionJournal = journal
	journal.display()
	return journal
}

func (cache *TxCache) isSizeEvictionEnabled() bool {
	return cache.evictionConfig.NumBytesThreshold > 0
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

func (cache *TxCache) areThereJustAFewSenders() bool {
	numSenders := cache.CountSenders()
	justAFewSenders := numSenders < int64(cache.evictionConfig.ThresholdEvictSenders)
	return justAFewSenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	numTxs := cache.CountTx()
	tooManyTxs := numTxs > int64(cache.evictionConfig.CountThreshold)
	return tooManyTxs
}

// evictOldestSenders removes transactions from the cache
func (cache *TxCache) evictOldestSenders() (uint32, uint32) {
	listsOrdered := cache.txListBySender.GetListsSortedByOrderNumber()
	sliceEnd := core.MinUint32(cache.evictionConfig.NumSendersToEvictInOneStep, uint32(len(listsOrdered)))
	listsToEvict := listsOrdered[:sliceEnd]

	return cache.evictSendersAndTheirTxs(listsToEvict)
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

func (cache *TxCache) doEvictItems(txsToEvict [][]byte, sendersToEvict []string) (countTxs uint32, countSenders uint32) {
	countTxs = cache.txByHash.RemoveTxsBulk(txsToEvict)
	countSenders = cache.txListBySender.RemoveSendersBulk(sendersToEvict)
	return
}

// evictHighNonceTransactions removes transactions from the cache
// For senders with many transactions (> "ALotOfTransactionsForASender"), evict "NumTxsToEvictForASenderWithALot" transactions
// Also makes sure that there's no sender with 0 transactions
func (cache *TxCache) evictHighNonceTransactions() (uint32, uint32) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	cache.forEachSender(func(key string, txList *txListForSender) {
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

	return cache.doEvictItems(txsToEvict, sendersToEvict)
}

func (cache *TxCache) evictSendersWhileTooManyTxs() (step uint32, countTxs uint32, countSenders uint32) {
	return cache.evictSendersWhile(cache.areThereTooManyTxs, SortByOrderNumberAsc)
}

func (cache *TxCache) evictSendersWhileTooManyBytes() (step uint32, countTxs uint32, countSenders uint32) {
	return cache.evictSendersWhile(cache.areThereTooManyBytes, SortByTotalGas)
}

// evictSendersWhileTooManyTxs removes transactions in a loop, as long as "stopCondition" is true
// One batch of senders is removed in each step
// Before starting the loop, the senders are sorted as specified by "sendersSortKind"
func (cache *TxCache) evictSendersWhile(stopCondition func() bool, sendersSortKind txListBySenderSortKind) (step uint32, countTxs uint32, countSenders uint32) {
	batchesSource := cache.txListBySender.GetListsSortedBy(sendersSortKind)
	batchSize := cache.evictionConfig.NumSendersToEvictInOneStep
	batchStart := uint32(0)

	for step = 1; stopCondition(); step++ {
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
