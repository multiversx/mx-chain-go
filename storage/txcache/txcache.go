package txcache

import (
	"fmt"
	"sort"

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

// TxListForSender is
type TxListForSender struct {
	IsSorted       bool
	CopyBatchIndex int
	CopyBatchSize  int
	Items          []data.TransactionHandler
}

func (list *TxListForSender) addTransaction(tx data.TransactionHandler) {
	list.Items = append(list.Items, tx)
	list.IsSorted = false
}

func (list *TxListForSender) removeTransaction(tx data.TransactionHandler) {
	index := list.findTx(tx)
	list.Items = append(list.Items[:index], list.Items[index+1:]...)
}

func (list *TxListForSender) findTx(tx data.TransactionHandler) int {
	for i, item := range list.Items {
		if item == tx {
			return i
		}
	}

	return -1
}

func (list *TxListForSender) sortTransactions() {
	if list.IsSorted {
		return
	}

	items := list.Items

	sort.Slice(items, func(i, j int) bool {
		return items[i].GetNonce() < items[j].GetNonce()
	})

	list.IsSorted = true
}

func (list *TxListForSender) restartBatchCopying(batchSize int) {
	list.CopyBatchIndex = 0
	list.CopyBatchSize = batchSize
}

func (list *TxListForSender) copyBatchTo(destination []data.TransactionHandler) int {
	length := len(list.Items)
	index := list.CopyBatchIndex

	if index == length {
		return 0
	}

	batchEnd := index + list.CopyBatchSize

	if batchEnd > length {
		batchEnd = length
	}

	copied := copy(destination, list.Items[index:batchEnd])
	list.CopyBatchIndex += copied
	return copied
}
