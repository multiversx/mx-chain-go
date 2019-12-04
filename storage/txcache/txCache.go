package txcache

import (
	"fmt"

	cmap "github.com/ElrondNetwork/concurrent-map"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TxCache is
type TxCache struct {
	txListBySender *cmap.ConcurrentMap
	txByHash       *cmap.ConcurrentMap
}

// NewTxCache creates
func NewTxCache(size int, shards int) *TxCache {
	// todo: size and shards given differently.
	txBySender := cmap.New(size, shards)
	txByHash := cmap.New(size, shards)

	txCache := &TxCache{
		txListBySender: txBySender,
		txByHash:       txByHash,
	}

	return txCache
}

// AddTx adds
func (cache *TxCache) AddTx(txHash []byte, tx data.TransactionHandler) {
	sender := string(tx.GetSndAddress())

	cache.txByHash.Set(string(txHash), tx)

	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		listForSender = &TxListForSender{}
		cache.txListBySender.Set(sender, listForSender)
	}

	listForSender.addTransaction(tx)
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
func (cache *TxCache) GetByTxHash(txHash []byte) (data.TransactionHandler, bool) {
	return cache.getTx(string(txHash))
}

func (cache *TxCache) getTx(txHash string) (data.TransactionHandler, bool) {
	txUntyped, ok := cache.txByHash.Get(txHash)
	if !ok {
		return nil, false
	}

	tx := txUntyped.(data.TransactionHandler)
	return tx, true
}

// GetSorted gets
// We do multiple passes in order to fill the result with sorted transactions
// batchSizePerSender specifies how many transactions to copy for a sender in a pass
func (cache *TxCache) GetSorted(count int, batchSizePerSender int) []data.TransactionHandler {
	result := make([]data.TransactionHandler, count)
	resultFillIndex := 0
	resultIsFull := false

	for pass := 0; ; pass++ {
		fmt.Println("Pass", pass, "fill index", resultFillIndex)

		copiedInThisPass := 0

		cache.txListBySender.IterCb(func(key string, txListUntyped interface{}) {
			if resultIsFull {
				return
			}

			txList := txListUntyped.(*TxListForSender)

			// Do this on first pass only
			if pass == 0 {
				txList.sortTransactions()
				txList.restartBatchCopying(batchSizePerSender)
			}

			copied := txList.copyBatchTo(result[resultFillIndex:])

			resultFillIndex += copied
			copiedInThisPass += copied
			resultIsFull = resultFillIndex == count
		})

		nothingCopiedThisPass := copiedInThisPass == 0

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
	sender := string(tx.GetSndAddress())
	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		return
	}

	listForSender.removeTransaction(tx)
}
