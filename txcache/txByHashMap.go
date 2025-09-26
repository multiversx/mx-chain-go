package txcache

import (
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/txcache/maps"
)

// txByHashMap is a new map-like structure for holding and accessing transactions by txHash
type txByHashMap struct {
	backingMap *maps.ConcurrentMap
	counter    atomic.Counter
	numBytes   atomic.Counter
}

// newTxByHashMap creates a new TxByHashMap instance
func newTxByHashMap(nChunksHint uint32) *txByHashMap {
	backingMap := maps.NewConcurrentMap(nChunksHint)

	return &txByHashMap{
		backingMap: backingMap,
	}
}

// addTx adds a transaction to the map
func (txMap *txByHashMap) addTx(tx *WrappedTransaction) bool {
	added := txMap.backingMap.SetIfAbsent(string(tx.TxHash), tx)
	if added {
		txMap.counter.Increment()
		txMap.numBytes.Add(tx.Size)
	}

	return added
}

// removeTx removes a transaction from the map
func (txMap *txByHashMap) removeTx(txHash string) (*WrappedTransaction, bool) {
	item, removed := txMap.backingMap.Remove(txHash)
	if !removed {
		return nil, false
	}

	tx, ok := item.(*WrappedTransaction)
	if !ok {
		return nil, false
	}

	if removed {
		txMap.counter.Decrement()
		txMap.numBytes.Subtract(tx.Size)
	}

	return tx, true
}

// removeTxWithTrackingCheck removes a transaction from the map but checks if it is still tracked
// TODO should do the check at a higher level
func (txMap *txByHashMap) removeTxWithTrackingCheck(txHash string, txTracker txsTracker) bool {
	tx, ok := txMap.getTx(txHash)
	if ok && txTracker.IsTransactionTracked(tx) {
		return false
	}

	_, ok = txMap.removeTx(txHash)
	return ok
}

// getTx gets a transaction from the map
func (txMap *txByHashMap) getTx(txHash string) (*WrappedTransaction, bool) {
	txUntyped, ok := txMap.backingMap.Get(txHash)
	if !ok {
		return nil, false
	}

	tx := txUntyped.(*WrappedTransaction)
	return tx, true
}

// GetTxsBulk gets a bulk of transactions from map
func (txMap *txByHashMap) GetTxsBulk(txHashes [][]byte) []*WrappedTransaction {
	txs := make([]*WrappedTransaction, 0, len(txHashes))
	for _, txHash := range txHashes {
		txUntyped, ok := txMap.backingMap.Get(string(txHash))
		if !ok {
			continue
		}

		tx := txUntyped.(*WrappedTransaction)

		txs = append(txs, tx)
	}

	return txs
}

// RemoveTxsBulk removes transactions, in bulk
func (txMap *txByHashMap) RemoveTxsBulk(txHashes [][]byte) uint32 {
	numRemoved := uint32(0)

	for _, txHash := range txHashes {
		_, removed := txMap.removeTx(string(txHash))
		if removed {
			numRemoved++
		}
	}

	return numRemoved
}

// RemoveTxsBulkWithTrackingCheck removes transactions, in bulk, but checks if each tx can be removed
// TODO should do the check at a higher level
func (txMap *txByHashMap) RemoveTxsBulkWithTrackingCheck(txHashes [][]byte, txTracker txsTracker) uint32 {
	if check.IfNil(txTracker) {
		return 0
	}

	numRemoved := uint32(0)
	for _, txHash := range txHashes {
		removed := txMap.removeTxWithTrackingCheck(string(txHash), txTracker)
		if removed {
			numRemoved++
		}
	}

	return numRemoved
}

// forEach iterates over the senders
func (txMap *txByHashMap) forEach(function ForEachTransaction) {
	txMap.backingMap.IterCb(func(key string, item interface{}) {
		tx := item.(*WrappedTransaction)
		function([]byte(key), tx)
	})
}

func (txMap *txByHashMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}

func (txMap *txByHashMap) keys() [][]byte {
	keys := txMap.backingMap.Keys()
	keysAsBytes := make([][]byte, len(keys))
	for i := 0; i < len(keys); i++ {
		keysAsBytes[i] = []byte(keys[i])
	}

	return keysAsBytes
}
