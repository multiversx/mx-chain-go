package txcache

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxCache is
type TxCache struct {
	txListBySender   TxListBySenderMap
	txByHash         TxByHashMap
	evictionStrategy *EvictionStrategy
}

// NewTxCache creates a new transaction cache
// "size" dictates the maximum number of transactions to hold in this cache at a given time
// "shardsHint" is used to configure the internal concurrent maps on which the implementation relies
func NewTxCache(size int, shardsHint int) *TxCache {
	// Note: for simplicity, we use the same "shardsHint" for both internal concurrent maps

	// We'll hold at most "size" lists of 1 transaction
	txBySender := NewConcurrentMap(size, shardsHint)
	// We'll hold at most "size" transactions
	txByHash := NewConcurrentMap(size, shardsHint)

	txCache := &TxCache{
		txListBySender: TxListBySenderMap{
			Map:     txBySender,
			Counter: 0,
		},
		txByHash: TxByHashMap{
			Map:     txByHash,
			Counter: 0,
		},
	}

	return txCache
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(txHash []byte, tx *transaction.Transaction) {
	if cache.evictionStrategy != nil {
		cache.evictionStrategy.DoEvictionIfNecessary(tx)
	}

	cache.txByHash.addTransaction(txHash, tx)
	cache.txListBySender.addTransaction(txHash, tx)
}

// GetByTxHash gets the transaction by hash
func (cache *TxCache) GetByTxHash(txHash []byte) (*transaction.Transaction, bool) {
	tx, ok := cache.txByHash.getTransaction(string(txHash))
	return tx, ok
}

// GetSorted gets
func (cache *TxCache) GetSorted(noRequested int, batchSizePerSender int) []*transaction.Transaction {
	result := make([]*transaction.Transaction, noRequested)
	resultFillIndex := 0
	resultIsFull := false

	for pass := 0; ; pass++ {
		copiedInThisPass := 0

		cache.txListBySender.Map.IterCb(func(key string, txListUntyped interface{}) {
			if resultIsFull {
				return
			}

			txList := txListUntyped.(*TxListForSender)

			// Do this on first pass only
			if pass == 0 {
				txList.RestartBatchCopying(batchSizePerSender)
			}

			copied := txList.CopyBatchTo(result[resultFillIndex:])

			resultFillIndex += copied
			copiedInThisPass += copied
			resultIsFull = resultFillIndex == noRequested
		})

		nothingCopiedThisPass := copiedInThisPass == 0

		// No more passes needed
		if nothingCopiedThisPass || resultIsFull {
			break
		}
	}

	return result[:resultFillIndex]
}

// RemoveByTxHash removes
func (cache *TxCache) RemoveByTxHash(txHash []byte) {
	tx, ok := cache.txByHash.removeTransaction(string(txHash))
	if ok {
		cache.txListBySender.removeTransaction(tx)
	}
}

// CountTx gets the number of transactions in the cache
func (cache *TxCache) CountTx() int64 {
	return cache.txByHash.Counter.Get()
}

func (cache *TxCache) selfCheck() {
	// todo check sync between the two maps
}
