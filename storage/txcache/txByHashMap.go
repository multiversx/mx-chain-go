package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TxByHashMap is a new map-like structure for holding and accessing transactions by txHash
type TxByHashMap struct {
	backingMap *ConcurrentMap
	counter    core.AtomicCounter
}

// NewTxByHashMap creates a new TxByHashMap instance
func NewTxByHashMap(size uint32, noChunksHint uint32) TxByHashMap {
	// We'll hold at most "size" transactions
	backingMap := NewConcurrentMap(size, noChunksHint)

	return TxByHashMap{
		backingMap: backingMap,
		counter:    0,
	}
}

// AddTx adds a transaction to the map
func (txMap *TxByHashMap) AddTx(txHash []byte, tx data.TransactionHandler) {
	txMap.backingMap.Set(string(txHash), tx)
	txMap.counter.Increment()
}

// RemoveTx removes a transaction from the map
func (txMap *TxByHashMap) RemoveTx(txHash string) (data.TransactionHandler, bool) {
	tx, ok := txMap.GetTx(txHash)
	if !ok {
		return nil, false
	}

	txMap.backingMap.Remove(txHash)
	txMap.counter.Decrement()
	return tx, true
}

// GetTx gets a transaction from the map
func (txMap *TxByHashMap) GetTx(txHash string) (data.TransactionHandler, bool) {
	txUntyped, ok := txMap.backingMap.Get(txHash)
	if !ok {
		return nil, false
	}

	tx := txUntyped.(data.TransactionHandler)
	return tx, true
}

// RemoveTxsBulk removes transactions, in bulk
func (txMap *TxByHashMap) RemoveTxsBulk(txHashes [][]byte) uint32 {
	for _, txHash := range txHashes {
		txMap.backingMap.Remove(string(txHash))
	}

	oldCount := uint32(txMap.counter.Get())
	newCount := uint32(txMap.backingMap.Count())
	noRemoved := oldCount - newCount

	txMap.counter.Set(int64(newCount))
	return noRemoved
}
