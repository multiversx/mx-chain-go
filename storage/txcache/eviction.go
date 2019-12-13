package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// EvictionConfig is a cache eviction model
type EvictionConfig struct {
	Enabled                        bool
	CountThreshold                 uint32
	CountJustAFewSenders           uint32
	NoOldestSendersToEvict         uint32
	ALotOfTransactionsForASender   uint32
	NoTxsToEvictForASenderWithALot uint32
}

// doEviction does cache eviction
// We do not allow more evictions to start concurrently
func (cache *TxCache) doEviction(incomingTx data.TransactionHandler) {
	if !cache.areThereTooManyTxs() {
		return
	}

	cache.evictionMutex.Lock()

	if cache.areThereTooManySenders() {
		countTxs, countSenders := cache.evictOldestSenders()
		log.Debug("DoEviction, 1st pass:", "countTxs", countTxs, "countSenders", countSenders)
	}

	if cache.areThereTooManyTxs() {
		countTxs, countSenders := cache.evictHighNonceTransactions()
		log.Debug("DoEviction, 2nd pass:", "countTxs", countTxs, "countSenders", countSenders)
	}

	if cache.areThereTooManyTxs() && !cache.areThereJustAFewSenders() {
		steps, countTxs, countSenders := cache.evictSendersWhileTooManyTxs()
		log.Debug("DoEviction, 3rd pass:", "steps", steps, "countTxs", countTxs, "countSenders", countSenders)
	}

	cache.evictionMutex.Unlock()
}

func (cache *TxCache) areThereTooManySenders() bool {
	noSenders := cache.CountSenders()
	tooManySenders := noSenders > int64(cache.evictionConfig.CountThreshold)
	return tooManySenders
}

func (cache *TxCache) areThereJustAFewSenders() bool {
	noSenders := cache.CountSenders()
	justAFewSenders := noSenders < int64(cache.evictionConfig.CountJustAFewSenders)
	return justAFewSenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	noTxs := cache.CountTx()
	tooManyTxs := noTxs > int64(cache.evictionConfig.CountThreshold)
	return tooManyTxs
}

// evictOldestSenders removes transactions from the cache
func (cache *TxCache) evictOldestSenders() (uint32, uint32) {
	listsOrdered := cache.txListBySender.GetListsSortedByOrderNumber()
	sliceEnd := core.MinUint32(cache.evictionConfig.NoOldestSendersToEvict, uint32(len(listsOrdered)))
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
// For senders with many transactions (> "ALotOfTransactionsForASender"), evict "NoTxsToEvictForASenderWithALot" transactions
// Also makes sure that there's no sender with 0 transactions
func (cache *TxCache) evictHighNonceTransactions() (uint32, uint32) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	cache.forEachSender(func(key string, txList *txListForSender) {
		aLot := cache.evictionConfig.ALotOfTransactionsForASender
		toEvictForSenderCount := cache.evictionConfig.NoTxsToEvictForASenderWithALot

		if txList.HasMoreThan(aLot) {
			txsToEvictForSender := txList.RemoveHighNonceTxs(toEvictForSenderCount)
			txsToEvict = append(txsToEvict, txsToEvictForSender...)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, key)
		}
	})

	return cache.doEvictItems(txsToEvict, sendersToEvict)
}

// evictSendersWhileTooManyTxs removes transactions
// Eviction happens in ((transaction count) - CountThreshold) / NoOldestSendersToEvict + 1 steps
// One batch of senders is removed in each step
func (cache *TxCache) evictSendersWhileTooManyTxs() (step uint32, countTxs uint32, countSenders uint32) {
	batchesSource := cache.txListBySender.GetListsSortedByOrderNumber()
	batchSize := cache.evictionConfig.NoOldestSendersToEvict
	batchStart := uint32(0)

	for step = 1; cache.areThereTooManyTxs(); step++ {
		batchEnd := core.MinUint32(batchStart+batchSize, uint32(len(batchesSource)))
		batch := batchesSource[batchStart:batchEnd]

		stepCountTxs, stepCountSenders := cache.evictSendersAndTheirTxs(batch)

		countTxs += stepCountTxs
		countSenders += stepCountSenders
		batchStart += batchSize

		// Infinite loop otherwise
		if stepCountTxs == 0 {
			break
		}

		if stepCountSenders < batchSize {
			break
		}
	}

	return
}
