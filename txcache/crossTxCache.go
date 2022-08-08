package txcache

import (
	"github.com/ElrondNetwork/elrond-go-storage"
	"github.com/ElrondNetwork/elrond-go-storage/immunitycache"
)

var _ elrond_go_storage.Cacher = (*CrossTxCache)(nil)

// CrossTxCache holds cross-shard transactions (where destination == me)
type CrossTxCache struct {
	*immunitycache.ImmunityCache
	config ConfigDestinationMe
}

// NewCrossTxCache creates a new transactions cache
func NewCrossTxCache(config ConfigDestinationMe) (*CrossTxCache, error) {
	log.Debug("NewCrossTxCache", "config", config.String())

	err := config.verify()
	if err != nil {
		return nil, err
	}

	immunityCacheConfig := immunitycache.CacheConfig{
		Name:                        config.Name,
		NumChunks:                   config.NumChunks,
		MaxNumBytes:                 config.MaxNumBytes,
		MaxNumItems:                 config.MaxNumItems,
		NumItemsToPreemptivelyEvict: config.NumItemsToPreemptivelyEvict,
	}

	immunityCache, err := immunitycache.NewImmunityCache(immunityCacheConfig)
	if err != nil {
		return nil, err
	}

	cache := CrossTxCache{
		ImmunityCache: immunityCache,
		config:        config,
	}

	return &cache, nil
}

// ImmunizeTxsAgainstEviction marks items as non-evictable
func (cache *CrossTxCache) ImmunizeTxsAgainstEviction(keys [][]byte) {
	numNow, numFuture := cache.ImmunityCache.ImmunizeKeys(keys)
	log.Trace("CrossTxCache.ImmunizeTxsAgainstEviction()",
		"name", cache.config.Name,
		"len(keys)", len(keys),
		"numNow", numNow,
		"numFuture", numFuture,
	)
	cache.Diagnose(false)
}

// AddTx adds a transaction in the cache
func (cache *CrossTxCache) AddTx(tx *WrappedTransaction) (has, added bool) {
	return cache.HasOrAdd(tx.TxHash, tx, int(tx.Size))
}

// GetByTxHash gets the transaction by hash
func (cache *CrossTxCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	item, ok := cache.ImmunityCache.Get(txHash)
	if !ok {
		return nil, false
	}
	tx, ok := item.(*WrappedTransaction)
	if !ok {
		return nil, false
	}

	return tx, true
}

// Get returns the unwrapped payload of a TransactionWrapper
// Implemented for compatibility reasons (see txPoolsCleaner.go).
func (cache *CrossTxCache) Get(key []byte) (value interface{}, ok bool) {
	wrapped, ok := cache.GetByTxHash(key)
	if !ok {
		return nil, false
	}

	return wrapped.Tx, true
}

// Peek returns the unwrapped payload of a TransactionWrapper
// Implemented for compatibility reasons (see transactions.go, common.go).
func (cache *CrossTxCache) Peek(key []byte) (value interface{}, ok bool) {
	return cache.Get(key)
}

// RemoveTxByHash removes tx by hash
func (cache *CrossTxCache) RemoveTxByHash(txHash []byte) bool {
	return cache.RemoveWithResult(txHash)
}

// ForEachTransaction iterates over the transactions in the cache
func (cache *CrossTxCache) ForEachTransaction(function ForEachTransaction) {
	cache.ForEachItem(func(key []byte, item interface{}) {
		tx, ok := item.(*WrappedTransaction)
		if !ok {
			return
		}

		function(key, tx)
	})
}

// GetTransactionsPoolForSender returns an empty slice, only to respect the interface
// CrossTxCache does not support transaction selection (not applicable, since transactions are already half-executed),
// thus does not handle nonces, nonce gaps etc.
func (cache *CrossTxCache) GetTransactionsPoolForSender(_ string) []*WrappedTransaction {
	return make([]*WrappedTransaction, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *CrossTxCache) IsInterfaceNil() bool {
	return cache == nil
}
