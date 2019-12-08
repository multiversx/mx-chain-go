package txcache

import "github.com/ElrondNetwork/elrond-go/data/transaction"

// TxByHashMap is
type TxByHashMap struct {
	Map     *ConcurrentMap
	Counter AtomicCounter
}

// NewTxByHashMap creates a new map-like structure for holding and accessing transactions by txHash
func NewTxByHashMap(size int, shardsHint int) TxByHashMap {
	// We'll hold at most "size" transactions
	backingMap := NewConcurrentMap(size, shardsHint)

	return TxByHashMap{
		Map:     backingMap,
		Counter: 0,
	}
}

func (txMap *TxByHashMap) addTransaction(txHash []byte, tx *transaction.Transaction) {
	txMap.Map.Set(string(txHash), tx)
	txMap.Counter.Increment()
}

func (txMap *TxByHashMap) getTransaction(txHash string) (*transaction.Transaction, bool) {
	txUntyped, ok := txMap.Map.Get(txHash)
	if !ok {
		return nil, false
	}

	tx := txUntyped.(*transaction.Transaction)
	return tx, true
}

func (txMap *TxByHashMap) removeTransaction(txHash string) (*transaction.Transaction, bool) {
	tx, ok := txMap.getTransaction(txHash)
	if !ok {
		return nil, false
	}

	txMap.Map.Remove(txHash)
	txMap.Counter.Decrement()
	return tx, true
}
