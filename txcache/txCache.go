package txcache

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-storage"
)

var _ elrond_go_storage.Cacher = (*TxCache)(nil)

// TxCache represents a cache-like structure (it has a fixed capacity and implements an eviction mechanism) for holding transactions
type TxCache struct {
	name                      string
	txListBySender            *txListBySenderMap
	txByHash                  *txByHashMap
	config                    ConfigSourceMe
	evictionMutex             sync.Mutex
	evictionJournal           evictionJournal
	evictionSnapshotOfSenders []*txListForSender
	isEvictionInProgress      atomic.Flag
	numSendersSelected        atomic.Counter
	numSendersWithInitialGap  atomic.Counter
	numSendersWithMiddleGap   atomic.Counter
	numSendersInGracePeriod   atomic.Counter
	sweepingMutex             sync.Mutex
	sweepingListOfSenders     []*txListForSender
	mutTxOperation            sync.Mutex
}

// NewTxCache creates a new transaction cache
func NewTxCache(config ConfigSourceMe, txGasHandler TxGasHandler) (*TxCache, error) {
	log.Debug("NewTxCache", "config", config.String())
	elrond_go_storage.MonitorNewCache(config.Name, uint64(config.NumBytesThreshold))

	err := config.verify()
	if err != nil {
		return nil, err
	}
	if check.IfNil(txGasHandler) {
		return nil, elrond_go_storage.ErrNilTxGasHandler
	}

	// Note: for simplicity, we use the same "numChunks" for both internal concurrent maps
	numChunks := config.NumChunks
	senderConstraintsObj := config.getSenderConstraints()
	txFeeHelper := newFeeComputationHelper(txGasHandler.MinGasPrice(), txGasHandler.MinGasLimit(), txGasHandler.MinGasPriceForProcessing())
	scoreComputerObj := newDefaultScoreComputer(txFeeHelper)

	txCache := &TxCache{
		name:            config.Name,
		txListBySender:  newTxListBySenderMap(numChunks, senderConstraintsObj, scoreComputerObj, txGasHandler, txFeeHelper),
		txByHash:        newTxByHashMap(numChunks),
		config:          config,
		evictionJournal: evictionJournal{},
	}

	txCache.initSweepable()
	return txCache, nil
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	if tx == nil || check.IfNil(tx.Tx) {
		return false, false
	}

	if cache.config.EvictionEnabled {
		cache.doEviction()
	}

	cache.mutTxOperation.Lock()
	addedInByHash := cache.txByHash.addTx(tx)
	addedInBySender, evicted := cache.txListBySender.addTx(tx)
	cache.mutTxOperation.Unlock()
	if addedInByHash != addedInBySender {
		// This can happen  when two go-routines concur to add the same transaction:
		// - A adds to "txByHash"
		// - B won't add to "txByHash" (duplicate)
		// - B adds to "txListBySender"
		// - A won't add to "txListBySender" (duplicate)
		log.Trace("TxCache.AddTx(): slight inconsistency detected:", "name", cache.name, "tx", tx.TxHash, "sender", tx.Tx.GetSndAddr(), "addedInByHash", addedInByHash, "addedInBySender", addedInBySender)
	}

	if len(evicted) > 0 {
		cache.monitorEvictionWrtSenderLimit(tx.Tx.GetSndAddr(), evicted)
		cache.txByHash.RemoveTxsBulk(evicted)
	}

	// The return value "added" is true even if transaction added, but then removed due to limits be sender.
	// This it to ensure that onAdded() notification is triggered.
	return true, addedInByHash || addedInBySender
}

// GetByTxHash gets the transaction by hash
func (cache *TxCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	tx, ok := cache.txByHash.getTx(string(txHash))
	return tx, ok
}

// SelectTransactionsWithBandwidth selects a reasonably fair list of transactions to be included in the next miniblock
// It returns at most "numRequested" transactions
// Each sender gets the chance to give at least bandwidthPerSender gas worth of transactions, unless "numRequested" limit is reached before iterating over all senders
func (cache *TxCache) SelectTransactionsWithBandwidth(numRequested int, batchSizePerSender int, bandwidthPerSender uint64) []*WrappedTransaction {
	result := cache.doSelectTransactions(numRequested, batchSizePerSender, bandwidthPerSender)
	go cache.doAfterSelection()
	return result
}

func (cache *TxCache) doSelectTransactions(numRequested int, batchSizePerSender int, bandwidthPerSender uint64) []*WrappedTransaction {
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
			journal := txList.selectBatchTo(isFirstBatch, result[resultFillIndex:], batchSizeWithScoreCoefficient, bandwidthPerSender)
			cache.monitorBatchSelectionEnd(journal)

			if isFirstBatch {
				cache.collectSweepable(txList)
			}

			resultFillIndex += journal.copied
			copiedInThisPass += journal.copied
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
	cache.Diagnose(false)
}

// RemoveTxByHash removes tx by hash
func (cache *TxCache) RemoveTxByHash(txHash []byte) bool {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()

	tx, foundInByHash := cache.txByHash.removeTx(string(txHash))
	if !foundInByHash {
		return false
	}

	foundInBySender := cache.txListBySender.removeTx(tx)
	if !foundInBySender {
		// This condition can arise often at high load & eviction, when two go-routines concur to remove the same transaction:
		// - A = remove transactions upon commit / final
		// - B = remove transactions due to high load (eviction)
		//
		// - A reaches "RemoveTxByHash()", then "cache.txByHash.removeTx()".
		// - B reaches "cache.txByHash.RemoveTxsBulk()"
		// - B reaches "cache.txListBySender.RemoveSendersBulk()"
		// - A reaches "cache.txListBySender.removeTx()", but sender does not exist anymore
		log.Trace("TxCache.RemoveTxByHash(): slight inconsistency detected: !foundInBySender", "name", cache.name, "tx", txHash)
	}

	return true
}

// NumBytes gets the approximate number of bytes stored in the cache
func (cache *TxCache) NumBytes() int {
	return int(cache.txByHash.numBytes.GetUint64())
}

// CountTx gets the number of transactions in the cache
func (cache *TxCache) CountTx() uint64 {
	return cache.txByHash.counter.GetUint64()
}

// Len is an alias for CountTx
func (cache *TxCache) Len() int {
	return int(cache.CountTx())
}

// SizeInBytesContained returns 0
func (cache *TxCache) SizeInBytesContained() uint64 {
	return 0
}

// CountSenders gets the number of senders in the cache
func (cache *TxCache) CountSenders() uint64 {
	return cache.txListBySender.counter.GetUint64()
}

// ForEachTransaction iterates over the transactions in the cache
func (cache *TxCache) ForEachTransaction(function ForEachTransaction) {
	cache.txByHash.forEach(function)
}

// GetTransactionsPoolForSender returns the list of transaction hashes for the sender
func (cache *TxCache) GetTransactionsPoolForSender(sender string) []*WrappedTransaction {
	listForSender, ok := cache.txListBySender.getListForSender(sender)
	if !ok {
		return nil
	}

	wrappedTxs := make([]*WrappedTransaction, listForSender.items.Len())
	for element, i := listForSender.items.Front(), 0; element != nil; element, i = element.Next(), i+1 {
		tx := element.Value.(*WrappedTransaction)
		wrappedTxs[i] = tx
	}

	return wrappedTxs
}

// Clear clears the cache
func (cache *TxCache) Clear() {
	cache.mutTxOperation.Lock()
	cache.txListBySender.clear()
	cache.txByHash.clear()
	cache.mutTxOperation.Unlock()
}

// Put is not implemented
func (cache *TxCache) Put(_ []byte, _ interface{}, _ int) (evicted bool) {
	log.Error("TxCache.Put is not implemented")
	return false
}

// Get gets a transaction (unwrapped) by hash
// Implemented for compatibility reasons (see txPoolsCleaner.go).
func (cache *TxCache) Get(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// Has checks if a transaction exists
func (cache *TxCache) Has(key []byte) bool {
	_, ok := cache.GetByTxHash(key)
	return ok
}

// Peek gets a transaction (unwrapped) by hash
// Implemented for compatibility reasons (see transactions.go, common.go).
func (cache *TxCache) Peek(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// HasOrAdd is not implemented
func (cache *TxCache) HasOrAdd(_ []byte, _ interface{}, _ int) (has, added bool) {
	log.Error("TxCache.HasOrAdd is not implemented")
	return false, false
}

// Remove removes tx by hash
func (cache *TxCache) Remove(key []byte) {
	_ = cache.RemoveTxByHash(key)
}

// Keys returns the tx hashes in the cache
func (cache *TxCache) Keys() [][]byte {
	return cache.txByHash.keys()
}

// MaxSize is not implemented
func (cache *TxCache) MaxSize() int {
	// TODO: Should be analyzed if the returned value represents the max size of one cache in sharded cache configuration
	return int(cache.config.CountThreshold)
}

// RegisterHandler is not implemented
func (cache *TxCache) RegisterHandler(func(key []byte, value interface{}), string) {
	log.Error("TxCache.RegisterHandler is not implemented")
}

// UnRegisterHandler is not implemented
func (cache *TxCache) UnRegisterHandler(string) {
	log.Error("TxCache.UnRegisterHandler is not implemented")
}

// NotifyAccountNonce should be called by external components (such as interceptors and transactions processor)
// in order to inform the cache about initial nonce gap phenomena
func (cache *TxCache) NotifyAccountNonce(accountKey []byte, nonce uint64) {
	cache.txListBySender.notifyAccountNonce(accountKey, nonce)
}

// ImmunizeTxsAgainstEviction does nothing for this type of cache
func (cache *TxCache) ImmunizeTxsAgainstEviction(_ [][]byte) {
}

// Close does nothing for this cacher implementation
func (cache *TxCache) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *TxCache) IsInterfaceNil() bool {
	return cache == nil
}
