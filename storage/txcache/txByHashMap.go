package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxByHashMap is a new map-like structure for holding and accessing transactions by txHash
type TxByHashMap struct {
	Map     *ConcurrentMap
	Counter core.AtomicCounter
}

// NewTxByHashMap creates a new TxByHashMap instance
func NewTxByHashMap(size uint32, shardsHint uint32) TxByHashMap {
	// We'll hold at most "size" transactions
	backingMap := NewConcurrentMap(size, shardsHint)

	return TxByHashMap{
		Map:     backingMap,
		Counter: 0,
	}
}

// AddTx adds a transaction to the map
func (txMap *TxByHashMap) AddTx(txHash []byte, tx *transaction.Transaction) {
	txMap.Map.Set(string(txHash), tx)
	txMap.Counter.Increment()
}

// RemoveTx removes a transaction from the map
func (txMap *TxByHashMap) RemoveTx(txHash string) (*transaction.Transaction, bool) {
	tx, ok := txMap.GetTx(txHash)
	if !ok {
		return nil, false
	}

	txMap.Map.Remove(txHash)
	txMap.Counter.Decrement()
	return tx, true
}

// GetTx gets a transaction from the map
func (txMap *TxByHashMap) GetTx(txHash string) (*transaction.Transaction, bool) {
	txUntyped, ok := txMap.Map.Get(txHash)
	if !ok {
		return nil, false
	}

	tx := txUntyped.(*transaction.Transaction)
	return tx, true
}

// RemoveTxsBulk removes transactions, in bulk
func (txMap *TxByHashMap) RemoveTxsBulk(txHashes [][]byte) uint32 {
	for _, txHash := range txHashes {
		txMap.Map.Remove(string(txHash))
	}

	oldCount := uint32(txMap.Counter.Get())
	newCount := uint32(txMap.Map.Count())
	noRemoved := oldCount - newCount

	txMap.Counter.Set(int64(newCount))
	return noRemoved
}
