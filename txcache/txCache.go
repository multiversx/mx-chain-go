package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-chain-storage-go/monitoring"
	"github.com/multiversx/mx-chain-storage-go/types"
)

var _ types.Cacher = (*TxCache)(nil)

// TxCache represents a cache-like structure (it has a fixed capacity and implements an eviction mechanism) for holding transactions
type TxCache struct {
	name                 string
	txListBySender       *txListBySenderMap
	txByHash             *txByHashMap
	config               ConfigSourceMe
	evictionMutex        sync.Mutex
	isEvictionInProgress atomic.Flag
	mutTxOperation       sync.Mutex
}

// NewTxCache creates a new transaction cache
func NewTxCache(config ConfigSourceMe, txGasHandler TxGasHandler) (*TxCache, error) {
	log.Debug("NewTxCache", "config", config.String())
	monitoring.MonitorNewCache(config.Name, uint64(config.NumBytesThreshold))

	err := config.verify()
	if err != nil {
		return nil, err
	}
	if check.IfNil(txGasHandler) {
		return nil, common.ErrNilTxGasHandler
	}

	// Note: for simplicity, we use the same "numChunks" for both internal concurrent maps
	numChunks := config.NumChunks
	senderConstraintsObj := config.getSenderConstraints()
	scoreComputerObj := newDefaultScoreComputer(txGasHandler)

	txCache := &TxCache{
		name:           config.Name,
		txListBySender: newTxListBySenderMap(numChunks, senderConstraintsObj, scoreComputerObj, txGasHandler),
		txByHash:       newTxByHashMap(numChunks),
		config:         config,
	}

	return txCache, nil
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	if tx == nil || check.IfNil(tx.Tx) {
		return false, false
	}

	logAdd.Trace("AddTx()", "tx", tx.TxHash)

	if cache.config.EvictionEnabled {
		_ = cache.doEviction()
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
		logAdd.Debug("AddTx(): slight inconsistency detected:", "tx", tx.TxHash, "sender", tx.Tx.GetSndAddr(), "addedInByHash", addedInByHash, "addedInBySender", addedInBySender)
	}

	if len(evicted) > 0 {
		logRemove.Debug("AddTx() with eviction", "sender", tx.Tx.GetSndAddr(), "num evicted txs", len(evicted))
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

// SelectTransactions selects a reasonably fair list of transactions to be included in the next miniblock
// It returns at most "numRequested" transactions, with total gas ~ "gasRequested".
//
// Selection is performed in more passes.
// In each pass, each sender is allowed to contribute a batch of transactions,
// with a number of transactions and total gas proportional to the sender's score.
func (cache *TxCache) SelectTransactions(numRequested int, gasRequested uint64, baseNumPerSenderBatch int, baseGasPerSenderBatch uint64) []*WrappedTransaction {
	senders, transactions := cache.doSelectTransactions(
		logSelect,
		numRequested,
		gasRequested,
		baseNumPerSenderBatch,
		baseGasPerSenderBatch,
	)

	go cache.diagnoseCounters()
	go displaySelectionOutcome(logSelect, senders, transactions)

	return transactions
}

func (cache *TxCache) doSelectTransactions(contextualLogger logger.Logger, numRequested int, gasRequested uint64, baseNumPerSenderBatch int, baseGasPerSenderBatch uint64) ([]*txListForSender, []*WrappedTransaction) {
	stopWatch := core.NewStopWatch()
	stopWatch.Start("selection")

	contextualLogger.Debug(
		"doSelectTransactions(): begin",
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)

	senders := cache.getSendersEligibleForSelection()
	transactions := make([]*WrappedTransaction, numRequested)

	shouldContinueSelection := true
	selectedGas := uint64(0)
	selectedNum := 0

	for pass := 0; shouldContinueSelection; pass++ {
		selectedNumInThisPass := 0

		for _, txList := range senders {
			score := txList.getScore()

			// Slighly suboptimal: we recompute the constraints for each pass,
			// even though they are constant with respect to a sender, in the scope of a selection.
			// However, this is not a performance bottleneck.
			numPerBatch, gasPerBatch := cache.computeSelectionSenderConstraints(score, baseNumPerSenderBatch, baseGasPerSenderBatch)

			isFirstBatch := pass == 0
			batchSelectionJournal := txList.selectBatchTo(isFirstBatch, transactions[selectedNum:], numPerBatch, gasPerBatch)
			selectedGas += batchSelectionJournal.selectedGas
			selectedNum += batchSelectionJournal.selectedNum
			selectedNumInThisPass += batchSelectionJournal.selectedNum

			shouldContinueSelection := selectedNum < numRequested && selectedGas < gasRequested
			if !shouldContinueSelection {
				break
			}
		}

		nothingSelectedInThisPass := selectedNumInThisPass == 0
		if nothingSelectedInThisPass {
			// No more passes needed
			break
		}
	}

	transactions = transactions[:selectedNum]

	contextualLogger.Debug(
		"doSelectTransactions(): end",
		"duration", stopWatch.GetMeasurement("selection"),
		"num txs selected", selectedNum,
	)

	return senders, transactions
}

func (cache *TxCache) getSendersEligibleForSelection() []*txListForSender {
	return cache.txListBySender.getSnapshotDescending()
}

func (cache *TxCache) computeSelectionSenderConstraints(score int, baseNumPerBatch int, baseGasPerBatch uint64) (int, uint64) {
	if score == 0 {
		return 1, 1
	}

	scoreDivision := float64(score) / float64(maxSenderScore)
	numPerBatch := int(float64(baseNumPerBatch) * scoreDivision)
	gasPerBatch := uint64(float64(baseGasPerBatch) * scoreDivision)

	return numPerBatch, gasPerBatch
}

// RemoveTxByHash removes tx by hash
func (cache *TxCache) RemoveTxByHash(txHash []byte) bool {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()

	logRemove.Trace("RemoveTxByHash()", "tx", txHash)

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
		logRemove.Debug("RemoveTxByHash(): slight inconsistency detected: !foundInBySender", "tx", txHash)
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

func (cache *TxCache) getAllTransactions() []*WrappedTransaction {
	transactions := make([]*WrappedTransaction, 0, cache.Len())

	cache.ForEachTransaction(func(_ []byte, tx *WrappedTransaction) {
		transactions = append(transactions, tx)
	})

	return transactions
}

// GetTransactionsPoolForSender returns the list of transaction hashes for the sender
func (cache *TxCache) GetTransactionsPoolForSender(sender string) []*WrappedTransaction {
	listForSender, ok := cache.txListBySender.getListForSender(sender)
	if !ok {
		return nil
	}

	return listForSender.getTxs()
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
	evicted := cache.txListBySender.notifyAccountNonce(accountKey, nonce)

	if len(evicted) > 0 {
		logRemove.Trace("NotifyAccountNonce() with eviction", "sender", accountKey, "nonce", nonce, "num evicted txs", len(evicted))
		cache.txByHash.RemoveTxsBulk(evicted)
	}
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
