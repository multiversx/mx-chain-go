package txcache

import (
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
	txUntyped, ok := cache.txByHash.Get(string(txHash))
	if !ok {
		return nil, false
	}

	tx := txUntyped.(data.TransactionHandler)
	return tx, true
}

// GetSorted gets
func (cache *TxCache) GetSorted(count int) []data.TransactionHandler {
	result := make([]data.TransactionHandler, 0)

	cache.txListBySender.IterCb(func(key string, txListUntyped interface{}) {
		txList := txListUntyped.(*TxListForSender)
		txList.sortTransactions()
		result = append(result, txList.Items...)
	})

	return result
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
	Items []data.TransactionHandler
}

func (list *TxListForSender) addTransaction(tx data.TransactionHandler) {
	list.Items = append(list.Items, tx)
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
	items := list.Items

	sort.Slice(items, func(i, j int) bool {
		return items[i].GetNonce() < items[j].GetNonce()
	})
}
