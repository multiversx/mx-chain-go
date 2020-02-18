package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

type adapterTxCacheToSortedTransactionsProvider struct {
	txCache *txcache.TxCache
}

func newAdapterTxCacheToSortedTransactionsProvider(txCache *txcache.TxCache) *adapterTxCacheToSortedTransactionsProvider {
	adapter := &adapterTxCacheToSortedTransactionsProvider{
		txCache: txCache,
	}

	return adapter
}

// GetSortedTransactions gets the transactions from the cache
func (adapter *adapterTxCacheToSortedTransactionsProvider) GetSortedTransactions() ([]data.TransactionHandler, [][]byte) {
	txs, txHashes := adapter.txCache.SelectTransactions(process.MaxItemsInBlock, process.NumTxPerSenderBatchForFillingMiniblock)
	return txs, txHashes
}

// NotifyAccountNonce notifies the cache about the current nonce of an account
func (adapter *adapterTxCacheToSortedTransactionsProvider) NotifyAccountNonce(accountKey []byte, nonce uint64) {
	adapter.txCache.NotifyAccountNonce(accountKey, nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *adapterTxCacheToSortedTransactionsProvider) IsInterfaceNil() bool {
	return adapter == nil
}
