package txcache

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("txcache")

// TxCache is
type TxCache struct {
	txListBySender   TxListBySenderMap
	txByHash         TxByHashMap
	EvictionStrategy *EvictionStrategy
}

// NewTxCache creates a new transaction cache
// "size" dictates the maximum number of transactions to hold in this cache at a given time
// "noChunksHint" is used to configure the internal concurrent maps on which the implementation relies
func NewTxCache(size uint32, noChunksHint uint32) *TxCache {
	// Note: for simplicity, we use the same "noChunksHint" for both internal concurrent maps
	txCache := &TxCache{
		txListBySender: NewTxListBySenderMap(size, noChunksHint),
		txByHash:       NewTxByHashMap(size, noChunksHint),
	}

	return txCache
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(txHash []byte, tx data.TransactionHandler) {
	if cache.EvictionStrategy != nil {
		cache.EvictionStrategy.DoEviction(tx)
	}

	cache.txByHash.AddTx(txHash, tx)
	cache.txListBySender.AddTx(txHash, tx)
}

// GetByTxHash gets the transaction by hash
func (cache *TxCache) GetByTxHash(txHash []byte) (data.TransactionHandler, bool) {
	tx, ok := cache.txByHash.GetTx(string(txHash))
	return tx, ok
}

// GetTransactions gets a reasonably fair list of transactions to be included in the next miniblock
// It returns at most "noRequested" transactions
// Each sender gets the chance to give at least "batchSizePerSender" transactions, unless "noRequested" limit is reached before iterating over all senders
func (cache *TxCache) GetTransactions(noRequested int, batchSizePerSender int) []data.TransactionHandler {
	result := make([]data.TransactionHandler, noRequested)
	resultFillIndex := 0
	resultIsFull := false

	for pass := 0; !resultIsFull; pass++ {
		copiedInThisPass := 0

		cache.ForEachSender(func(key string, txList *TxListForSender) {
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
		if nothingCopiedThisPass {
			break
		}
	}

	return result[:resultFillIndex]
}

// RemoveTxByHash removes
func (cache *TxCache) RemoveTxByHash(txHash []byte) error {
	tx, ok := cache.txByHash.RemoveTx(string(txHash))
	if !ok {
		return errorTxNotFound
	}

	found := cache.txListBySender.RemoveTx(tx)
	if !found {
		// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
		log.Error("RemoveTxByHash detected maps sync inconsistency", "tx", txHash)
		return errorMapsSyncInconsistency
	}

	return nil
}

// CountTx gets the number of transactions in the cache
func (cache *TxCache) CountTx() int64 {
	return cache.txByHash.counter.Get()
}

// ForEachSender iterates over the senders
func (cache *TxCache) ForEachSender(function ForEachSender) {
	cache.txListBySender.ForEach(function)
}
