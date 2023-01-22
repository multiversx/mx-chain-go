package txcache

import (
	"github.com/multiversx/mx-chain-storage-go/txcache"
)

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction = txcache.WrappedTransaction

// TxGasHandler handles a transaction gas and gas cost
type TxGasHandler = txcache.TxGasHandler

// ForEachTransaction is an iterator callback
type ForEachTransaction = txcache.ForEachTransaction

// ConfigDestinationMe holds cache configuration
type ConfigDestinationMe = txcache.ConfigDestinationMe

// ConfigSourceMe holds cache configuration
type ConfigSourceMe = txcache.ConfigSourceMe

// TxCache represents a cache-like structure (it has a fixed capacity and implements an eviction mechanism) for holding transactions
type TxCache = txcache.TxCache

// DisabledCache represents a disabled cache
type DisabledCache = txcache.DisabledCache

// CrossTxCache holds cross-shard transactions (where destination == me)
type CrossTxCache = txcache.CrossTxCache

// NewTxCache creates a new transaction cache
func NewTxCache(config ConfigSourceMe, txGasHandler TxGasHandler) (*TxCache, error) {
	return txcache.NewTxCache(config, txGasHandler)
}

// NewDisabledCache creates a new disabled cache
func NewDisabledCache() *DisabledCache {
	return txcache.NewDisabledCache()
}

// NewCrossTxCache creates a new transactions cache
func NewCrossTxCache(config ConfigDestinationMe) (*CrossTxCache, error) {
	return txcache.NewCrossTxCache(config)
}
