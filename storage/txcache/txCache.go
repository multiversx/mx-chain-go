package txcache

import (
	cmap "github.com/ElrondNetwork/concurrent-map"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxCache is
type TxCache struct {
	txListBySender *cmap.ConcurrentMap
	txByHash       *cmap.ConcurrentMap
}

// NewTxCache creates
func NewTxCache(size int, shards int) *TxCache {
	// todo - comment, implications of size value, shards value

	// todo-fix: size and shards do not have to be the same (that's an arbitrary constraint)
	txBySender := cmap.New(size, shards)
	txByHash := cmap.New(size, shards)

	txCache := &TxCache{
		txListBySender: txBySender,
		txByHash:       txByHash,
	}

	return txCache
}

// AddTx adds
func (cache *TxCache) AddTx(txHash []byte, tx *transaction.Transaction) {
	sender := string(tx.SndAddr)

	cache.txByHash.Set(string(txHash), tx)

	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		listForSender = &TxListForSender{}
		cache.txListBySender.Set(sender, listForSender)
	}

	// todo protect / mutex
	listForSender.addTransaction(tx)

	// todo: implement eviction
	// option 1: catch eviction event of cache.txByHash and remove the tx from the other collection as well
	// option 2:
}

func (cache *TxCache) getListForSender(sender string) (*TxListForSender, bool) {
	listForSenderUntyped, ok := cache.txListBySender.Get(sender)
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
	txUntyped, ok := cache.txByHash.Get(txHash)
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

		cache.txListBySender.IterCb(func(key string, txListUntyped interface{}) {
			if resultIsFull {
				return
			}

			txList := txListUntyped.(*TxListForSender)

			// Do this on first pass only
			if pass == 0 {
				txList.restartBatchCopying(batchSizePerSender)
			}

			copied := txList.copyBatchTo(result[resultFillIndex:])

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

	cache.txByHash.Remove(string(txHash))
	sender := string(tx.SndAddr)
	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		return
	}

	// todo protect / mutex
	listForSender.removeTransaction(tx)
	if listForSender.isEmpty() {
		cache.txListBySender.Remove(sender)
	}
}

// CountTx counts
func (cache *TxCache) CountTx() int {
	count := 0

	cache.txListBySender.IterCb(func(key string, txListUntyped interface{}) {
		txList := txListUntyped.(*TxListForSender)
		count += txList.Items.Len()
	})

	return count
}
