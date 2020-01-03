package preprocess

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type cacheToSortedTransactionsProviderAdapter struct {
	transactionsPreprocessor *transactions
	cache                    storage.Cacher
	cacheKey                 string
}

func newCacheToSortedTransactionsProviderAdapter(transactionsPreprocessor *transactions, cache storage.Cacher, cacheKey string) *cacheToSortedTransactionsProviderAdapter {
	adapter := &cacheToSortedTransactionsProviderAdapter{
		transactionsPreprocessor: transactionsPreprocessor,
		cache:                    cache,
		cacheKey:                 cacheKey,
	}

	return adapter
}

// GetTransactions gets the transactions from the cache, sorted by nonce
func (adapter *cacheToSortedTransactionsProviderAdapter) GetTransactions(numRequested int, batchSizePerSender int) ([]data.TransactionHandler, [][]byte) {
	txsAsStruct, txHashes := adapter.getOrderedTx()
	txsAsInterface := castToTxsAsInterface(txsAsStruct)
	return txsAsInterface, txHashes
}

// getOrderedTx was moved here from the previous implementation
func (adapter *cacheToSortedTransactionsProviderAdapter) getOrderedTx() ([]*transaction.Transaction, [][]byte) {
	txs := adapter.transactionsPreprocessor
	strCache := adapter.cacheKey

	txs.mutOrderedTxs.RLock()
	orderedTxs := txs.orderedTxs[strCache]
	orderedTxHashes := txs.orderedTxHashes[strCache]
	txs.mutOrderedTxs.RUnlock()

	alreadyOrdered := len(orderedTxs) > 0
	if !alreadyOrdered {
		orderedTxs, orderedTxHashes = sortTxByNonce(adapter.cache)

		log.Debug("creating mini blocks has been started",
			"have num txs", len(orderedTxs),
			"strCache", strCache,
		)

		txs.mutOrderedTxs.Lock()
		txs.orderedTxs[strCache] = orderedTxs
		txs.orderedTxHashes[strCache] = orderedTxHashes
		txs.mutOrderedTxs.Unlock()
	}

	return orderedTxs, orderedTxHashes
}

// sortTxByNonce was moved here from the previous implementation
func sortTxByNonce(cache storage.Cacher) ([]*transaction.Transaction, [][]byte) {
	txShardPool := cache

	keys := txShardPool.Keys()
	transactions := make([]*transaction.Transaction, 0, len(keys))
	txHashes := make([][]byte, 0, len(keys))

	mTxHashes := make(map[uint64][][]byte, len(keys))
	mTransactions := make(map[uint64][]*transaction.Transaction, len(keys))

	nonces := make([]uint64, 0, len(keys))

	for _, key := range keys {
		val, _ := txShardPool.Peek(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*transaction.Transaction)
		if !ok {
			continue
		}

		if mTxHashes[tx.Nonce] == nil {
			nonces = append(nonces, tx.Nonce)
			mTxHashes[tx.Nonce] = make([][]byte, 0)
			mTransactions[tx.Nonce] = make([]*transaction.Transaction, 0)
		}

		mTxHashes[tx.Nonce] = append(mTxHashes[tx.Nonce], key)
		mTransactions[tx.Nonce] = append(mTransactions[tx.Nonce], tx)
	}

	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i] < nonces[j]
	})

	for _, nonce := range nonces {
		keys := mTxHashes[nonce]

		for idx, key := range keys {
			txHashes = append(txHashes, key)
			transactions = append(transactions, mTransactions[nonce][idx])
		}
	}

	return transaction.TrimSlicePtr(transactions), sliceUtil.TrimSliceSliceByte(txHashes)
}

func castToTxsAsInterface(txsAsStruct []*transaction.Transaction) []data.TransactionHandler {
	txsAsInterface := make([]data.TransactionHandler, 0)
	for _, tx := range txsAsStruct {
		txCasted := data.TransactionHandler(tx)
		txsAsInterface = append(txsAsInterface, txCasted)
	}

	return txsAsInterface
}

// IsInterfaceNil returns true if there is no value under the interface
func (adapter *cacheToSortedTransactionsProviderAdapter) IsInterfaceNil() bool {
	return adapter == nil
}

// getSortedTransactionsProvider gets a sorted transactions provider given a generic cache
func getSortedTransactionsProvider(transactionsPreprocessor *transactions, cache storage.Cacher, cacheKey string) SortedTransactionsProvider {
	sortedTransactionsProvider, isSortedTransactionsProvider := cache.(SortedTransactionsProvider)
	if isSortedTransactionsProvider {
		return sortedTransactionsProvider
	}

	adapter := newCacheToSortedTransactionsProviderAdapter(transactionsPreprocessor, cache, cacheKey)
	return adapter
}
