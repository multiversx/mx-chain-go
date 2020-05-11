package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/storage/txcache/maps"
)

// txByHashMap is a new map-like structure for holding and accessing transactions by txHash
type txByHashMap struct {
	backingMap *maps.ConcurrentMap
	counter    atomic.Counter
	numBytes   atomic.Counter
}

// newTxByHashMap creates a new TxByHashMap instance
func newTxByHashMap(nChunksHint uint32) txByHashMap {
	backingMap := maps.NewConcurrentMap(nChunksHint)

	return txByHashMap{
		backingMap: backingMap,
	}
}

// addTx adds a transaction to the map
func (txMap *txByHashMap) addTx(tx *WrappedTransaction) bool {
	added := txMap.backingMap.SetIfAbsent(string(tx.TxHash), tx)
	if added {
		txMap.counter.Increment()
		txMap.numBytes.Add(int64(estimateTxSize(tx)))
	}

	return added
}

// removeTx removes a transaction from the map
func (txMap *txByHashMap) removeTx(txHash string) (*WrappedTransaction, bool) {
	tx, ok := txMap.getTx(txHash)
	if !ok {
		return nil, false
	}

	txMap.backingMap.Remove(txHash)
	txMap.counter.Decrement()
	txMap.numBytes.Subtract(int64(estimateTxSize(tx)))
	return tx, true
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

// RemoveTxsBulk removes transactions, in bulk
func (txMap *txByHashMap) RemoveTxsBulk(txHashes txHashes) uint32 {
	oldCount := uint32(txMap.counter.Get())

	for _, txHash := range txHashes {
		txMap.removeTx(string(txHash))
	}

	newCount := uint32(txMap.counter.Get())
	numRemoved := oldCount - newCount
	return numRemoved
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)

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

func (txMap *txByHashMap) keys() txHashes {
	keys := txMap.backingMap.Keys()
	keysAsBytes := make(txHashes, len(keys))
	for i := 0; i < len(keys); i++ {
		keysAsBytes[i] = []byte(keys[i])
	}

	return keysAsBytes
}
