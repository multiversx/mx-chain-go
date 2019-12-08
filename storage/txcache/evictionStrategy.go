package txcache

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

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

// DoEvictionIfNecessary does cache eviction
func (model *EvictionStrategy) DoEvictionIfNecessary(incomingTx *transaction.Transaction) {
	if model.Cache.txByHash.Counter.Get() < int64(model.Config.CountThreshold) {
		return
	}

	// First pass
	// If senders capacity is close to be reached reached, arbitrarily evict senders
	// Senders capacity is close to be reached when there are a lot of senders with little or one transaction
	model.DoSendersEvictionIfNecessary()

	// Second pass
	// If still too many transactions
	// For senders with many transactions (> "ManyTransactionsForASender") evict "PartOfManyTransactionsOfASender" transactions
	model.DoHighNonceTransactionsEviction()
}

// DoSendersEvictionIfNecessary removes senders (along with their transactions) from the cache
// Only if capacity is close to be reached
func (model *EvictionStrategy) DoSendersEvictionIfNecessary() {
	sendersEvictionNecessary := model.Cache.txListBySender.Counter.Get() > int64(model.Config.CountThreshold)

	if sendersEvictionNecessary {
		_, _ = model.DoArbitrarySendersEviction()
	}
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

	return len(txsToEvict), len(sendersToEvict)
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

	return len(txsToEvict), len(sendersToEvict)
}
