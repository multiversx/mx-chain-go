package txcache

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/data"
)

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

// todo: optimize this, don't use append perhaps()
// todo: implement delete without memory leak: https://github.com/golang/go/wiki/SliceTricks
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
