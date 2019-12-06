package txcache

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// EvictionStrategy is a cache eviction model
type EvictionStrategy struct {
	CountThreshold         int
	EachAndEverySender     int
	ManyTransactions       int
	PartOfManyTransactions int
	Cache                  *TxCache
}

// NewEvictionStrategy creates a new EvictionModel
func NewEvictionStrategy(capacity int, cache *TxCache) *EvictionStrategy {
	model := &EvictionStrategy{
		CountThreshold:         capacity * 99 / 100,
		EachAndEverySender:     capacity/100 + 1,
		ManyTransactions:       capacity * 1 / 100,
		PartOfManyTransactions: capacity * (1 / 4) / 100,
		Cache:                  cache,
	}

	return model
}

// DoEvictionIfNecessary does cache eviction
func (model *EvictionStrategy) DoEvictionIfNecessary(incomingTx *transaction.Transaction) {
	if model.Cache.txCount.Get() < int64(model.CountThreshold) {
		return
	}

	// First pass
	// If senders capacity is close to be reached reached, arbitrarily evict ~1/256 senders
	// Senders capacity is close to be reached when there are a lot of senders with little or one transaction
	model.DoSendersEvictionIfNecessary()

	// Second pass
	// If still too many transactions
	// For senders with many transactions (> "evictionManyTransactions") evict "evictionPartOfManyTransactions" transactions
	model.DoHighNonceTransactionsEviction()
}

// DoSendersEvictionIfNecessary removes senders (along with their transactions) from the cache
// Removes "each and every" sender from the cache
func (model *EvictionStrategy) DoSendersEvictionIfNecessary() {
	sendersEvictionNecessary := model.Cache.sendersCount.Get() > int64(model.CountThreshold)

	if sendersEvictionNecessary {
		model.doArbitrarySendersEviction()
	}
}

func (model *EvictionStrategy) doArbitrarySendersEviction() {
	sendersToEvict := make([]string, 0)

	index := 0
	model.Cache.txListBySender.IterCb(func(key string, txListUntyped interface{}) {
		if index%model.EachAndEverySender == 0 {
			sendersToEvict = append(sendersToEvict, key)
		}

		index++
	})

	// to do. delete from txByHash also!
	model.Cache.removeSenders(sendersToEvict)
}

// DoHighNonceTransactionsEviction removes transactions from the cache
func (model *EvictionStrategy) DoHighNonceTransactionsEviction() {
	sendersToEvict := make([]string, 0)

	model.Cache.txListBySender.IterCb(func(key string, txListUntyped interface{}) {
		txList := txListUntyped.(*TxListForSender)

		if txList.HasMoreThan(model.ManyTransactions) {
			txList.RemoveHighNonceTransactions(model.PartOfManyTransactions)
		}

		if txList.IsEmpty() {
			sendersToEvict = append(sendersToEvict, key)
		}
	})

	// to do. delete from txByHash also!
	model.Cache.removeSenders(sendersToEvict)
}
