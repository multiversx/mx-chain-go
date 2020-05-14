package txcache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*TxCache)(nil)
var _ txCache = (*TxCache)(nil)

type txHashes = [][]byte

// TxCache represents a cache-like structure (it has a fixed capacity and implements an eviction mechanism) for holding transactions
type TxCache struct {
	name                          string
	txListBySender                txListBySenderMap
	txByHash                      txByHashMap
	config                        CacheConfig
	evictionMutex                 sync.Mutex
	evictionJournal               evictionJournal
	evictionSnapshotOfSenders     []*txListForSender
	isEvictionInProgress          atomic.Flag
	numTxAddedBetweenSelections   atomic.Counter
	numTxAddedDuringEviction      atomic.Counter
	numTxRemovedBetweenSelections atomic.Counter
	numTxRemovedDuringEviction    atomic.Counter
	sweepingMutex                 sync.Mutex
	sweepingListOfSenders         []*txListForSender
}

// NewTxCache creates a new transaction cache
func NewTxCache(config CacheConfig) (*TxCache, error) {
	log.Debug("NewTxCache", "config", config)

	err := config.verify()
	if err != nil {
		return nil, err
	}

	// Note: for simplicity, we use the same "numChunksHint" for both internal concurrent maps
	numChunksHint := config.NumChunksHint
	senderConstraints := config.getSenderConstraints()
	scoreComputer := newDefaultScoreComputer(config.MinGasPriceNanoErd)

	txCache := &TxCache{
		name:            config.Name,
		txListBySender:  newTxListBySenderMap(numChunksHint, senderConstraints, scoreComputer),
		txByHash:        newTxByHashMap(numChunksHint),
		config:          config,
		evictionJournal: evictionJournal{},
	}

	txCache.initSweepable()
	return txCache, nil
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	ok = false
	added = false

	if tx == nil || check.IfNil(tx.Tx) {
		return
	}

	if cache.config.EvictionEnabled {
		cache.doEviction()
	}

	ok = true
	added, evicted := cache.txListBySender.addTx(tx)
	if added {
		cache.txByHash.addTx(tx)
		cache.monitorTxAddition()
	}

	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}

	return
}

// GetByTxHash gets the transaction by hash
func (cache *TxCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	tx, ok := cache.txByHash.getTx(string(txHash))
	return tx, ok
}

// SelectTransactions selects a reasonably fair list of transactions to be included in the next miniblock
// It returns at most "numRequested" transactions
// Each sender gets the chance to give at least "batchSizePerSender" transactions, unless "numRequested" limit is reached before iterating over all senders
func (cache *TxCache) SelectTransactions(numRequested int, batchSizePerSender int) []*WrappedTransaction {
	result := cache.doSelectTransactions(numRequested, batchSizePerSender)
	go cache.doAfterSelection()
	return result
}

func (cache *TxCache) doSelectTransactions(numRequested int, batchSizePerSender int) []*WrappedTransaction {
	stopWatch := cache.monitorSelectionStart()

	result := make([]*WrappedTransaction, numRequested)
	resultFillIndex := 0
	resultIsFull := false

	snapshotOfSenders := cache.getSendersEligibleForSelection()

	for pass := 0; !resultIsFull; pass++ {
		copiedInThisPass := 0

		for _, txList := range snapshotOfSenders {
			batchSizeWithScoreCoefficient := batchSizePerSender * int(txList.getLastComputedScore()+1)
			// Reset happens on first pass only
			isFirstBatch := pass == 0
			copied := txList.selectBatchTo(isFirstBatch, result[resultFillIndex:], batchSizeWithScoreCoefficient)

			if isFirstBatch {
				cache.collectSweepable(txList)
			}

			resultFillIndex += copied
			copiedInThisPass += copied
			resultIsFull = resultFillIndex == numRequested
			if resultIsFull {
				break
			}
		}

		nothingCopiedThisPass := copiedInThisPass == 0

		// No more passes needed
		if nothingCopiedThisPass {
			break
		}
	}

	result = result[:resultFillIndex]
	cache.monitorSelectionEnd(result, stopWatch)
	return result
}

func (cache *TxCache) getSendersEligibleForSelection() []*txListForSender {
	return cache.txListBySender.getSnapshotDescending()
}

func (cache *TxCache) doAfterSelection() {
	cache.sweepSweepable()
	cache.diagnose()
}

// RemoveTxByHash removes tx by hash
func (cache *TxCache) RemoveTxByHash(txHash []byte) error {
	tx, ok := cache.txByHash.removeTx(string(txHash))
	if !ok {
		return errTxNotFound
	}

	cache.monitorTxRemoval()

	found := cache.txListBySender.removeTx(tx)
	if !found {
		cache.onRemoveTxInconsistency(txHash)
		return errMapsSyncInconsistency
	}

	return nil
}

// NumBytes gets the approximate number of bytes stored in the cache
func (cache *TxCache) NumBytes() int64 {
	return cache.txByHash.numBytes.Get()
}

// CountTx gets the number of transactions in the cache
func (cache *TxCache) CountTx() int64 {
	return cache.txByHash.counter.Get()
}

// Len is an alias for CountTx
func (cache *TxCache) Len() int {
	return int(cache.CountTx())
}

// CountSenders gets the number of senders in the cache
func (cache *TxCache) CountSenders() int64 {
	return cache.txListBySender.counter.Get()
}

// ForEachTransaction iterates over the transactions in the cache
func (cache *TxCache) ForEachTransaction(function ForEachTransaction) {
	cache.txByHash.forEach(function)
}

// Clear clears the cache
func (cache *TxCache) Clear() {
	cache.txListBySender.clear()
	cache.txByHash.clear()
}

// Put is not implemented
func (cache *TxCache) Put(key []byte, value interface{}) (evicted bool) {
	log.Error("TxCache.Put is not implemented")
	return false
}

// Get gets a transaction by hash
func (cache *TxCache) Get(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// Has is not implemented
func (cache *TxCache) Has(key []byte) bool {
	log.Error("TxCache.Has is not implemented")
	return false
}

// Peek gets a transaction by hash
func (cache *TxCache) Peek(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// HasOrAdd is not implemented
func (cache *TxCache) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	log.Error("TxCache.HasOrAdd is not implemented")
	return false, false
}

// Remove removes tx by hash
func (cache *TxCache) Remove(key []byte) {
	_ = cache.RemoveTxByHash(key)
}

// RemoveOldest is not implemented
func (cache *TxCache) RemoveOldest() {
	log.Error("TxCache.RemoveOldest is not implemented")
}

// Keys returns the tx hashes in the cache
func (cache *TxCache) Keys() txHashes {
	return cache.txByHash.keys()
}

// MaxSize is not implemented
func (cache *TxCache) MaxSize() int {
	//TODO: Should be analyzed if the returned value represents the max size of one cache in sharded cache configuration
	return int(cache.config.CountThreshold)
}

// RegisterHandler is not implemented
func (cache *TxCache) RegisterHandler(func(key []byte, value interface{})) {
	log.Error("TxCache.RegisterHandler is not implemented")
}

// NotifyAccountNonce should be called by external components (such as interceptors and transactions processor)
// in order to inform the cache about initial nonce gap phenomena
func (cache *TxCache) NotifyAccountNonce(accountKey []byte, nonce uint64) {
	cache.txListBySender.notifyAccountNonce(accountKey, nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *TxCache) IsInterfaceNil() bool {
	return cache == nil
}
