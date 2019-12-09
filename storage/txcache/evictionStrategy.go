package txcache

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("txcache/eviction")

// EvictionStrategyConfig is a cache eviction model
type EvictionStrategyConfig struct {
	CountThreshold                  int
	EachAndEverySender              int
	ManyTransactionsForASender      int
	PartOfManyTransactionsOfASender int
}

// EvictionStrategy is a cache eviction model
type EvictionStrategy struct {
	Cache  *TxCache
	Config EvictionStrategyConfig
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

	// First pass
	// If senders capacity is close to be reached reached, arbitrarily evict senders
	// Senders capacity is close to be reached first (before txs capacity) when there are a lot of senders with little or one transaction
	if model.areThereTooManySenders() {
		log.Debug("DoEviction: 1st pass")
		model.DoArbitrarySendersEviction()
	}

	// Second pass
	// For senders with many transactions (> "ManyTransactionsForASender"), evict "PartOfManyTransactionsOfASender" transactions
	if model.areThereTooManyTxs() {
		log.Debug("DoEviction: 2nd pass")
		model.DoHighNonceTransactionsEviction()
	}

	// Third pass (in a loop)
	// While tx capacity is still close to be reached, arbitrarily evict senders
	for step := 1; model.areThereTooManyTxs(); step++ {
		log.Debug("DoEviction: 3nd pass", "step", step)
		model.DoArbitrarySendersEviction()
	}
}

func (model *EvictionStrategy) areThereTooManySenders() bool {
	tooManySenders := model.Cache.txListBySender.Counter.Get() > int64(model.Config.CountThreshold)
	return tooManySenders
}

func (model *EvictionStrategy) areThereTooManyTxs() bool {
	tooManyTxs := model.Cache.txByHash.Counter.Get() > int64(model.Config.CountThreshold)
	return tooManyTxs
}

// DoArbitrarySendersEviction removes senders (along with their transactions) from the cache
func (model *EvictionStrategy) DoArbitrarySendersEviction() (int, int) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	index := 0
	model.Cache.txListBySender.Map.IterCb(func(key string, txListUntyped interface{}) {
		txList := txListUntyped.(*TxListForSender)

		if index%model.Config.EachAndEverySender == 0 {
			txHashes := txList.GetTxHashes()
			txsToEvict = append(txsToEvict, txHashes...)
			sendersToEvict = append(sendersToEvict, key)
		}

		index++
	})

	model.Cache.txByHash.RemoveTransactionsBulk(txsToEvict)
	model.Cache.txListBySender.removeSenders(sendersToEvict)

	countTxs := len(txsToEvict)
	countSenders := len(sendersToEvict)

	log.Debug("DoArbitrarySendersEviction", "countTxs", countTxs, "countSenders", countSenders)
	return countTxs, countSenders
}

// DoHighNonceTransactionsEviction removes transactions from the cache
func (model *EvictionStrategy) DoHighNonceTransactionsEviction() (int, int) {
	txsToEvict := make([][]byte, 0)
	sendersToEvict := make([]string, 0)

	model.Cache.txListBySender.Map.IterCb(func(key string, txListUntyped interface{}) {
		txList := txListUntyped.(*TxListForSender)

		if txList.HasMoreThan(model.Config.ManyTransactionsForASender) {
			txHashes := txList.RemoveHighNonceTxs(model.Config.PartOfManyTransactionsOfASender)
			txsToEvict = append(txsToEvict, txHashes...)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, key)
		}
	})

	model.Cache.txByHash.RemoveTransactionsBulk(txsToEvict)
	model.Cache.txListBySender.removeSenders(sendersToEvict)

	countTxs := len(txsToEvict)
	countSenders := len(sendersToEvict)

	log.Debug("DoHighNonceTransactionsEviction", "countTxs", countTxs, "countSenders", countSenders)
	return countTxs, countSenders
}
