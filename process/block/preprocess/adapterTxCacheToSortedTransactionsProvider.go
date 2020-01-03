package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

type txCacheToSortedTransactionsProviderAdapter struct {
	txCache *txcache.TxCache
}

func newTxCacheToSortedTransactionsProviderAdapter(txCache *txcache.TxCache) *txCacheToSortedTransactionsProviderAdapter {
	adapter := &txCacheToSortedTransactionsProviderAdapter{
		txCache: txCache,
	}

	return adapter
}

// GetSortedTransactions gets the transactions from the cache
func (adapter *txCacheToSortedTransactionsProviderAdapter) GetSortedTransactions() ([]data.TransactionHandler, [][]byte) {
	txs, txHashes := adapter.txCache.GetTransactions(process.MaxItemsInBlock, process.NumTxPerSenderBatchForFillingMiniblock)
	return txs, txHashes
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *txCacheToSortedTransactionsProviderAdapter) IsInterfaceNil() bool {
	return adapter == nil
}
