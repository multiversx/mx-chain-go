package txcache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// EvictionStrategyConfig is a cache eviction model
type EvictionStrategyConfig struct {
	CountThreshold                 uint32
	CountJustAFewSenders           uint32
	NoOldestSendersToEvict         uint32
	ALotOfTransactionsForASender   uint32
	NoTxsToEvictForASenderWithALot uint32
}

// EvictionStrategy is a cache eviction model
type EvictionStrategy struct {
	cache  *TxCache
	config EvictionStrategyConfig
	mutex  sync.Mutex
}

// NewEvictionStrategy creates a new EvictionModel
func NewEvictionStrategy(cache *TxCache, config EvictionStrategyConfig) *EvictionStrategy {
	model := &EvictionStrategy{
		cache:  cache,
		config: config,
	}

	return model
}

// DoEviction does cache eviction
// We do not allow more evictions to start concurrently
func (model *EvictionStrategy) DoEviction(incomingTx *transaction.Transaction) {
	if !model.areThereTooManyTxs() {
		return
	}

	model.mutex.Lock()

	if model.areThereTooManySenders() {
		countTxs, countSenders := model.evictOldestSenders()
		log.Debug("DoEviction, 1st pass:", "countTxs", countTxs, "countSenders", countSenders)
	}

	if model.areThereTooManyTxs() {
		countTxs, countSenders := model.evictHighNonceTransactions()
		log.Debug("DoEviction, 2nd pass:", "countTxs", countTxs, "countSenders", countSenders)
	}

	if model.areThereTooManyTxs() && !model.areThereJustAFewSenders() {
		steps, countTxs, countSenders := model.evictSendersWhileTooManyTxs()
		log.Debug("DoEviction, 3rd pass:", "steps", steps, "countTxs", countTxs, "countSenders", countSenders)
	}

	model.mutex.Unlock()
}

func (model *EvictionStrategy) areThereTooManySenders() bool {
	noSenders := model.cache.txListBySender.counter.Get()
	tooManySenders := noSenders > int64(model.config.CountThreshold)
	return tooManySenders
}

func (model *EvictionStrategy) areThereJustAFewSenders() bool {
	noSenders := model.cache.txListBySender.counter.Get()
	justAFewSenders := noSenders < int64(model.config.CountJustAFewSenders)
	return justAFewSenders
}

func (model *EvictionStrategy) areThereTooManyTxs() bool {
	noTxs := model.cache.txByHash.counter.Get()
	tooManyTxs := noTxs > int64(model.config.CountThreshold)
	return tooManyTxs
}

// evictOldestSenders removes transactions from the cache
func (model *EvictionStrategy) evictOldestSenders() (uint32, uint32) {
	listsOrdered := model.cache.txListBySender.GetListsSortedByOrderNumber()
	sliceEnd := core.MinUint32(model.config.NoOldestSendersToEvict, uint32(len(listsOrdered)))
	listsToEvict := listsOrdered[:sliceEnd]

	return model.evictSendersAndTheirTxs(listsToEvict)
}

func (model *EvictionStrategy) evictSendersAndTheirTxs(listsToEvict []*TxListForSender) (uint32, uint32) {
	sendersToEvict := make([]string, 0)
	txsToEvict := make([][]byte, 0)

	for _, txList := range listsToEvict {
		sendersToEvict = append(sendersToEvict, txList.sender)
		txsToEvict = append(txsToEvict, txList.GetTxHashes()...)
	}

	return model.doEvictItems(txsToEvict, sendersToEvict)
}

func (model *EvictionStrategy) doEvictItems(txsToEvict [][]byte, sendersToEvict []string) (countTxs uint32, countSenders uint32) {
	countTxs = model.cache.txByHash.RemoveTxsBulk(txsToEvict)
	countSenders = model.cache.txListBySender.RemoveSendersBulk(sendersToEvict)
	return
}

// evictHighNonceTransactions removes transactions from the cache
// For senders with many transactions (> "ALotOfTransactionsForASender"), evict "NoTxsToEvictForASenderWithALot" transactions
// Also makes sure that there's no sender with 0 transactions
func (model *EvictionStrategy) evictHighNonceTransactions() (uint32, uint32) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	model.cache.ForEachSender(func(key string, txList *TxListForSender) {
		aLot := model.config.ALotOfTransactionsForASender
		toEvictForSenderCount := model.config.NoTxsToEvictForASenderWithALot

		if txList.HasMoreThan(aLot) {
			txsToEvictForSender := txList.RemoveHighNonceTxs(toEvictForSenderCount)
			txsToEvict = append(txsToEvict, txsToEvictForSender...)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, key)
		}
	})

	return model.doEvictItems(txsToEvict, sendersToEvict)
}

// evictSendersWhileTooManyTxs removes transactions
// Eviction happens in ((transaction count) - CountThreshold) / NoOldestSendersToEvict + 1 steps
// One batch of senders is removed in each step
func (model *EvictionStrategy) evictSendersWhileTooManyTxs() (step uint32, countTxs uint32, countSenders uint32) {
	batchesSource := model.cache.txListBySender.GetListsSortedByOrderNumber()
	batchSize := model.config.NoOldestSendersToEvict
	batchStart := uint32(0)

	for step = 1; model.areThereTooManyTxs(); step++ {
		batchEnd := core.MinUint32(batchStart+batchSize, uint32(len(batchesSource)))
		batch := batchesSource[batchStart:batchEnd]

		stepCountTxs, stepCountSenders := model.evictSendersAndTheirTxs(batch)

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
