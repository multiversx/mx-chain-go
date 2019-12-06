package txcache

import (
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxCache is
type TxCache struct {
	txListBySender   TxListBySenderMap
	txByHash         TxByHashMap
	evictionStrategy *EvictionStrategy
}

// TxListBySenderMap is
type TxListBySenderMap struct {
	Map     *ConcurrentMap
	Counter AtomicCounter
}

// TxByHashMap is
type TxByHashMap struct {
	Map     *ConcurrentMap
	Counter AtomicCounter
}

// NewTxCache creates a new transaction cache
// "size" dictates the maximum number of transactions to hold in this cache at a given time
// "shardsHint" is used to configure the internal concurrent maps on which the implementation relies
func NewTxCache(size int, shardsHint int) *TxCache {
	// Note: for simplicity, we use the same "shardsHint" for both internal concurrent maps

	// We'll hold at most "size" lists of 1 transaction
	txBySender := NewConcurrentMap(size, shardsHint)
	// We'll hold at most "size" transactions
	txByHash := NewConcurrentMap(size, shardsHint)

	txCache := &TxCache{
		txListBySender: TxListBySenderMap{
			Map:     txBySender,
			Counter: 0,
		},
		txByHash: TxByHashMap{
			Map:     txByHash,
			Counter: 0,
		},
	}

	return txCache
}

// AddTx adds a transaction in the cache
// Eviction happens if maximum capacity is reached
func (cache *TxCache) AddTx(txHash []byte, tx *transaction.Transaction) {
	sender := string(tx.SndAddr)

	cache.txByHash.Map.Set(string(txHash), tx)
	cache.txByHash.Counter.Increment()

	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		listForSender = cache.addSender(sender)
	}

	if cache.evictionStrategy != nil {
		cache.evictionStrategy.DoEvictionIfNecessary(tx)
	}

	listForSender.AddTransaction(txHash, tx)

}

func (cache *TxCache) getListForSender(sender string) (*TxListForSender, bool) {
	listForSenderUntyped, ok := cache.txListBySender.Map.Get(sender)
	if !ok {
		return nil, false
	}

	listForSender := listForSenderUntyped.(*TxListForSender)
	return listForSender, true
}

// GetByTxHash gets
func (cache *TxCache) GetByTxHash(txHash []byte) (*transaction.Transaction, bool) {
	return cache.getTx(string(txHash))
}

func (cache *TxCache) getTx(txHash string) (*transaction.Transaction, bool) {
	txUntyped, ok := cache.txByHash.Map.Get(txHash)
	if !ok {
		return nil, false
	}

	tx := txUntyped.(*transaction.Transaction)
	return tx, true
}

// GetSorted gets
func (cache *TxCache) GetSorted(noRequested int, batchSizePerSender int) []*transaction.Transaction {
	result := make([]*transaction.Transaction, noRequested)
	resultFillIndex := 0
	resultIsFull := false

	for pass := 0; ; pass++ {
		copiedInThisPass := 0

		cache.txListBySender.Map.IterCb(func(key string, txListUntyped interface{}) {
			if resultIsFull {
				return
			}

			txList := txListUntyped.(*TxListForSender)

			// Do this on first pass only
			if pass == 0 {
				txList.RestartBatchCopying(batchSizePerSender)
			}

			copied := txList.CopyBatchTo(result[resultFillIndex:])

			resultFillIndex += copied
			copiedInThisPass += copied
			resultIsFull = resultFillIndex == noRequested
		})

		nothingCopiedThisPass := copiedInThisPass == 0

		// No more passes needed
		if nothingCopiedThisPass || resultIsFull {
			break
		}
	}

	return result[:resultFillIndex]
}

// RemoveByTxHash removes
func (cache *TxCache) RemoveByTxHash(txHash []byte) {
	tx, ok := cache.getTx(string(txHash))
	if !ok {
		return
	}

	cache.txByHash.Map.Remove(string(txHash))
	cache.txByHash.Counter.Decrement()

	sender := string(tx.SndAddr)
	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
		return
	}

	listForSender.RemoveTransaction(tx)
	if listForSender.IsEmpty() {
		cache.removeSender(sender)
	}
}

// CountTx counts
func (cache *TxCache) CountTx() int {
	count := 0

	cache.txListBySender.Map.IterCb(func(key string, txListUntyped interface{}) {
		txList := txListUntyped.(*TxListForSender)
		count += txList.Items.Len()
	})

	return count
}

func (cache *TxCache) addSender(sender string) *TxListForSender {
	listForSender := NewTxListForSender()
	cache.txListBySender.Map.Set(sender, listForSender)
	cache.txListBySender.Counter.Increment()
	return listForSender
}

func (cache *TxCache) removeSender(sender string) {
	cache.txListBySender.Map.Remove(sender)
	cache.txListBySender.Counter.Decrement()
}

func (cache *TxCache) removeSenders(senders []string) {
	for _, senderKey := range senders {
		cache.removeSender(senderKey)
	}
}

func (cache *TxCache) selfCheck() {
	// todo check sync between the two maps
}
