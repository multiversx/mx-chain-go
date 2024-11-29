package txcache

import (
	"github.com/multiversx/mx-chain-storage-go/txcache"
	"github.com/multiversx/mx-chain-storage-go/types"
)

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction = txcache.WrappedTransaction

// AccountState represents the state of an account (as seen by the mempool)
type AccountState = types.AccountState

// MempoolHost provides blockchain information for mempool operations
type MempoolHost = txcache.MempoolHost

// SelectionSession provides blockchain information for transaction selection
type SelectionSession = txcache.SelectionSession

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
func NewTxCache(config ConfigSourceMe, host MempoolHost) (*TxCache, error) {
	return txcache.NewTxCache(config, host)
}

// NewDisabledCache creates a new disabled cache
func NewDisabledCache() *DisabledCache {
	return txcache.NewDisabledCache()
}

// NewCrossTxCache creates a new transactions cache
func NewCrossTxCache(config ConfigDestinationMe) (*CrossTxCache, error) {
	return txcache.NewCrossTxCache(config)
}
