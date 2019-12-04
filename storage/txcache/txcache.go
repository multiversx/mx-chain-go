package txcache

import (
	cmap "github.com/ElrondNetwork/concurrent-map"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TxCache is
type TxCache struct {
	TxListBySender *cmap.ConcurrentMap
	TxByHash       *cmap.ConcurrentMap
}

// NewTxCache creates
func NewTxCache(size int, shards int) *TxCache {
	// todo: size and shards given differently.
	txBySender := cmap.New(size, shards)
	txByHash := cmap.New(size, shards)

	txCache := &TxCache{
		TxListBySender: txBySender,
		TxByHash:       txByHash,
	}

	return txCache
}

// AddTx adds
func (cache *TxCache) AddTx(txHash []byte, tx data.TransactionHandler) {
	sender := string(tx.GetSndAddress())

	cache.TxByHash.Set(string(txHash), tx)

	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		listForSender = &TxListForSender{}
		cache.TxListBySender.Set(sender, listForSender)
	}

	listForSender.addTransaction(tx)
}

func (cache *TxCache) getListForSender(sender string) (*TxListForSender, bool) {
	listForSenderUntyped, ok := cache.TxListBySender.Get(sender)
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
	txUntyped, ok := cache.TxByHash.Get(string(txHash))
	if !ok {
		return nil, false
	}

	tx := txUntyped.(data.TransactionHandler)
	return tx, true
}

// GetSorted gets
func (cache *TxCache) GetSorted(count int) {
}

// RemoveByTxHash removes
func (cache *TxCache) RemoveByTxHash(txHash []byte) {
	tx, ok := cache.getTx(string(txHash))
	if !ok {
		return
	}

	cache.TxByHash.Remove(string(txHash))
	sender := string(tx.GetSndAddress())
	listForSender, ok := cache.getListForSender(sender)
	if !ok {
		return
	}

	listForSender.removeTransaction(tx)
}

// TxListForSender is
type TxListForSender struct {
}

func (list *TxListForSender) addTransaction(tx data.TransactionHandler) {
}

func (list *TxListForSender) removeTransaction(tx data.TransactionHandler) {
}
