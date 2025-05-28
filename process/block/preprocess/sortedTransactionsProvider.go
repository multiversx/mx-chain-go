package preprocess

import (
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

// TODO: Refactor "transactions.go" to not require the components in this file anymore
// createSortedTransactionsProvider is a "simple factory" for "SortedTransactionsProvider" objects
func createSortedTransactionsProvider(cache storage.Cacher, sortedTransactionsConfig config.SortedTransactionsConfig) SortedTransactionsProvider {
	txCache, isTxCache := cache.(TxCache)
	if isTxCache {
		return newAdapterTxCacheToSortedTransactionsProvider(txCache, sortedTransactionsConfig)
	}

	log.Error("Could not create a real [SortedTransactionsProvider], will create a disabled one")
	return &disabledSortedTransactionsProvider{}
}

// adapterTxCacheToSortedTransactionsProvider adapts a "TxCache" to the "SortedTransactionsProvider" interface
type adapterTxCacheToSortedTransactionsProvider struct {
	txCache                  TxCache
	sortedTransactionsConfig config.SortedTransactionsConfig
}

func newAdapterTxCacheToSortedTransactionsProvider(txCache TxCache, sortedTransactionsConfig config.SortedTransactionsConfig) *adapterTxCacheToSortedTransactionsProvider {
	adapter := &adapterTxCacheToSortedTransactionsProvider{
		txCache:                  txCache,
		sortedTransactionsConfig: sortedTransactionsConfig,
	}

	return adapter
}

// GetSortedTransactions gets the transactions from the cache
func (adapter *adapterTxCacheToSortedTransactionsProvider) GetSortedTransactions(session txcache.SelectionSession) []*txcache.WrappedTransaction {
	txs, _ := adapter.txCache.SelectTransactions(session, process.TxCacheSelectionGasRequested, adapter.sortedTransactionsConfig.TxCacheSelectionMaxNumTxs, time.Duration(adapter.sortedTransactionsConfig.TxCacheSelectionLoopMaximumDuration)*time.Millisecond)
	return txs
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *adapterTxCacheToSortedTransactionsProvider) IsInterfaceNil() bool {
	return adapter == nil
}

// disabledSortedTransactionsProvider is a disabled "SortedTransactionsProvider" (should never be used in practice)
type disabledSortedTransactionsProvider struct {
}

// GetSortedTransactions returns an empty slice
func (adapter *disabledSortedTransactionsProvider) GetSortedTransactions(_ txcache.SelectionSession) []*txcache.WrappedTransaction {
	return make([]*txcache.WrappedTransaction, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *disabledSortedTransactionsProvider) IsInterfaceNil() bool {
	return adapter == nil
}
