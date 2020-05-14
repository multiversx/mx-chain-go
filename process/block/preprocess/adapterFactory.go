package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

// createSortedTransactionsProvider creates a sorted transactions provider for a given cache
func createSortedTransactionsProvider(transactionsPreprocessor *transactions, cache storage.Cacher, cacheKey string) SortedTransactionsProvider {
	txCache, isTxCache := cache.(TxCache)
	if isTxCache {
		return newAdapterTxCacheToSortedTransactionsProvider(txCache)
	}

	return newAdapterGenericCacheToSortedTransactionsProvider(transactionsPreprocessor, cache, cacheKey)
}
