package txcache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("txcache")

// TxCache represents a cache-like structure (it has a fixed capacity and implements an eviction mechanism) for holding transactions
type TxCache struct {
	txListBySender txListBySenderMap
	txByHash       txByHashMap
	evictionConfig EvictionConfig
	evictionMutex  sync.Mutex
}

// NewTxCache creates a new transaction cache
// "nChunksHint" is used to configure the internal concurrent maps on which the implementation relies
func NewTxCache(nChunksHint uint32) *TxCache {
	// Note: for simplicity, we use the same "nChunksHint" for both internal concurrent maps
	txCache := &TxCache{
		txListBySender: newTxListBySenderMap(nChunksHint),
		txByHash:       newTxByHashMap(nChunksHint),
		evictionConfig: EvictionConfig{Enabled: false},
	}

	return txCache
}

// NewTxCacheWithEviction creates a new transaction cache with eviction
func NewTxCacheWithEviction(nChunksHint uint32, evictionConfig EvictionConfig) *TxCache {
	txCache := NewTxCache(nChunksHint)
	txCache.evictionConfig = evictionConfig

	return txCache
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(txHash []byte, tx data.TransactionHandler) {
	if check.IfNil(tx) {
		return
	}

	if cache.evictionConfig.Enabled {
		cache.doEviction()
	}

	cache.txByHash.addTx(txHash, tx)
	cache.txListBySender.addTx(txHash, tx)
}

// GetByTxHash gets the transaction by hash
func (cache *TxCache) GetByTxHash(txHash []byte) (data.TransactionHandler, bool) {
	tx, ok := cache.txByHash.getTx(string(txHash))
	return tx, ok
}

// GetTransactions gets a reasonably fair list of transactions to be included in the next miniblock
// It returns at most "numRequested" transactions
// Each sender gets the chance to give at least "batchSizePerSender" transactions, unless "numRequested" limit is reached before iterating over all senders
func (cache *TxCache) GetTransactions(numRequested int, batchSizePerSender int) []data.TransactionHandler {
	result := make([]data.TransactionHandler, numRequested)
	resultFillIndex := 0
	resultIsFull := false

	for pass := 0; !resultIsFull; pass++ {
		copiedInThisPass := 0

		cache.forEachSender(func(key string, txList *txListForSender) {
			// Reset happens on first pass only
			shouldResetCopy := pass == 0
			copied := txList.copyBatchTo(shouldResetCopy, result[resultFillIndex:], batchSizePerSender)

			resultFillIndex += copied
			copiedInThisPass += copied
			resultIsFull = resultFillIndex == numRequested
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
	tx, ok := cache.txByHash.removeTx(string(txHash))
	if !ok {
		return ErrTxNotFound
	}

	found := cache.txListBySender.removeTx(tx)
	if !found {
		// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
		log.Error("RemoveTxByHash detected maps sync inconsistency", "tx", txHash)
		return ErrMapsSyncInconsistency
	}

	return nil
}

// CountTx gets the number of transactions in the cache
func (cache *TxCache) CountTx() int64 {
	return cache.txByHash.counter.Get()
}

// CountSenders gets the number of senders in the cache
func (cache *TxCache) CountSenders() int64 {
	return cache.txListBySender.counter.Get()
}

// forEachSender iterates over the senders
func (cache *TxCache) forEachSender(function ForEachSender) {
	cache.txListBySender.forEach(function)
}
