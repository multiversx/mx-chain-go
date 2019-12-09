package txcache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("txcache/eviction")

// EvictionStrategyConfig is a cache eviction model
type EvictionStrategyConfig struct {
	CountThreshold                  int
	NoOldestSendersToEvict          int
	ManyTransactionsForASender      int
	PartOfManyTransactionsOfASender int
}

// EvictionStrategy is a cache eviction model
type EvictionStrategy struct {
	Cache  *TxCache
	Config EvictionStrategyConfig
	mutex  sync.Mutex
}

// NewEvictionStrategy creates a new EvictionModel
func NewEvictionStrategy(cache *TxCache, config EvictionStrategyConfig) *EvictionStrategy {
	model := &EvictionStrategy{
		Cache:  cache,
		Config: config,
	}

	return model
}

// DoEviction does cache eviction
func (model *EvictionStrategy) DoEviction(incomingTx *transaction.Transaction) {
	if !model.areThereTooManyTxs() {
		return
	}

	// We do not allow more evictions to start concurrently
	model.mutex.Lock()

	// First pass
	// Senders capacity is close to be reached first (before txs capacity) when there are a lot of senders with little or one transaction
	if model.areThereTooManySenders() {
		log.Debug("DoEviction: 1st pass")
		countTxs, countSenders := model.EvictOldestSenders()
		log.Debug("Evicted:", "countTxs", countTxs, "countSenders", countSenders)
	}

	// Second pass
	// For senders with many transactions (> "ManyTransactionsForASender"), evict "PartOfManyTransactionsOfASender" transactions
	if model.areThereTooManyTxs() {
		log.Debug("DoEviction: 2nd pass")
		countTxs, countSenders := model.EvictHighNonceTransactions()
		log.Debug("Evicted:", "countTxs", countTxs, "countSenders", countSenders)
	}

	// Third pass
	if model.areThereTooManyTxs() {
		log.Debug("DoEviction: 3nd pass")
		countTxs, countSenders := model.EvictSendersWhileTooManyTxs()
		log.Debug("Evicted:", "countTxs", countTxs, "countSenders", countSenders)
	}

	model.mutex.Unlock()
}

func (model *EvictionStrategy) areThereTooManySenders() bool {
	tooManySenders := model.Cache.txListBySender.Counter.Get() > int64(model.Config.CountThreshold)
	return tooManySenders
}

func (model *EvictionStrategy) areThereTooManyTxs() bool {
	tooManyTxs := model.Cache.txByHash.Counter.Get() > int64(model.Config.CountThreshold)
	return tooManyTxs
}

// EvictOldestSenders removes transactions from the cache
func (model *EvictionStrategy) EvictOldestSenders() (int, int) {
	listsOrdered := model.Cache.txListBySender.GetListsSortedByOrderNumber()
	sliceEnd := core.MinInt(model.Config.NoOldestSendersToEvict, len(listsOrdered))
	listsToEvict := listsOrdered[:sliceEnd]

	sendersToEvict := make([]string, 0)
	txsToEvict := make([][]byte, 0)

	for _, txList := range listsToEvict {
		sendersToEvict = append(sendersToEvict, txList.sender)
		txHashes := txList.GetTxHashes()
		txsToEvict = append(txsToEvict, txHashes...)
	}

	return model.evictItems(txsToEvict, sendersToEvict)
}

func (model *EvictionStrategy) evictItems(txsToEvict [][]byte, sendersToEvict []string) (int, int) {
	model.Cache.txByHash.RemoveTransactionsBulk(txsToEvict)
	model.Cache.txListBySender.removeSenders(sendersToEvict)

	return len(txsToEvict), len(sendersToEvict)
}

// EvictHighNonceTransactions removes transactions from the cache
func (model *EvictionStrategy) EvictHighNonceTransactions() (int, int) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	model.Cache.txListBySender.ForEach(func(key string, txList *TxListForSender) {
		if txList.HasMoreThan(model.Config.ManyTransactionsForASender) {
			txHashes := txList.RemoveHighNonceTxs(model.Config.PartOfManyTransactionsOfASender)
			txsToEvict = append(txsToEvict, txHashes...)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, key)
		}
	})

	return model.evictItems(txsToEvict, sendersToEvict)
}

// EvictSendersWhileTooManyTxs removes transactions
func (model *EvictionStrategy) EvictSendersWhileTooManyTxs() (int, int) {
	listsOrdered := model.Cache.txListBySender.GetListsSortedByOrderNumber()

	countTxs := 0
	countSenders := 0

	sliceStart := 0
	for step := 1; model.areThereTooManyTxs(); step++ {
		log.Debug("EvictSendersWhileTooManyTxs", "step", step)

		batchSize := model.Config.NoOldestSendersToEvict
		sliceEnd := core.MinInt(sliceStart+batchSize, len(listsOrdered))
		listsToEvict := listsOrdered[sliceStart:sliceEnd]
		sendersToEvict := make([]string, 0)
		txsToEvict := make([][]byte, 0)

		for _, txList := range listsToEvict {
			sendersToEvict = append(sendersToEvict, txList.sender)
			txHashes := txList.GetTxHashes()
			txsToEvict = append(txsToEvict, txHashes...)
		}

		model.evictItems(txsToEvict, sendersToEvict)

		countTxs += len(txsToEvict)
		countSenders += len(listsToEvict)

		sliceStart += batchSize
	}

	return countTxs, countSenders
}
