package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

// createSortedTransactionsProvider creates a sorted transactions provider for a given cache
func createSortedTransactionsProvider(transactionsPreprocessor *transactions, cache storage.Cacher, cacheKey string) SortedTransactionsProvider {
	txCache, isTxCache := cache.(TxCache)
	if isTxCache {
		return newAdapterTxCacheToSortedTransactionsProvider(txCache)
	}

	log.Error("Could not create a real [SortedTransactionsProvider], will create a disabled one")
	return &disabledSortedTransactionsProvider{}
}

type disabledSortedTransactionsProvider struct {
}

// GetSortedTransactions returns an empty slice
func (adapter *disabledSortedTransactionsProvider) GetSortedTransactions() []*txcache.WrappedTransaction {
	return make([]*txcache.WrappedTransaction, 0)
}

// NotifyAccountNonce does nothing
func (adapter *disabledSortedTransactionsProvider) NotifyAccountNonce(_ []byte, _ uint64) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *disabledSortedTransactionsProvider) IsInterfaceNil() bool {
	return adapter == nil
}
