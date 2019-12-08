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
	txCache := &TxCache{
		txListBySender: NewTxListBySenderMap(size, shardsHint),
		txByHash:       NewTxByHashMap(size, shardsHint),
	}

	return txCache
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(txHash []byte, tx *transaction.Transaction) {
	if cache.evictionStrategy != nil {
		cache.evictionStrategy.DoEvictionIfNecessary(tx)
	}

	cache.txByHash.AddTx(txHash, tx)
	cache.txListBySender.AddTx(txHash, tx)
}

// GetByTxHash gets the transaction by hash
func (cache *TxCache) GetByTxHash(txHash []byte) (*transaction.Transaction, bool) {
	tx, ok := cache.txByHash.GetTx(string(txHash))
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
				txList.StartBatchCopying(batchSizePerSender)
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

// RemoveTxByHash removes
func (cache *TxCache) RemoveTxByHash(txHash []byte) {
	tx, ok := cache.txByHash.RemoveTx(string(txHash))
	if ok {
		cache.txListBySender.RemoveTx(tx)
	}
}

// CountTx gets the number of transactions in the cache
func (cache *TxCache) CountTx() int64 {
	return cache.txByHash.Counter.Get()
}
