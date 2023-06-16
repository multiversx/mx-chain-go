package preprocess

import (
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

// SortedTransactionsProvider defines the public API of the transactions cache
type SortedTransactionsProvider interface {
	GetSortedTransactions() []*txcache.WrappedTransaction
	NotifyAccountNonce(accountKey []byte, nonce uint64)
	IsInterfaceNil() bool
}

// TxCache defines the functionality for the transactions cache
type TxCache interface {
	SelectTransactionsWithBandwidth(numRequested int, batchSizePerSender int, bandwidthPerSender uint64) []*txcache.WrappedTransaction
	NotifyAccountNonce(accountKey []byte, nonce uint64)
	IsInterfaceNil() bool
}

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	IsShardStuck(shardID uint32) bool
	ShouldSkipMiniBlocksCreationFromSelf() bool
	IsInterfaceNil() bool
}

// BlockSizeThrottler defines the functionality of adapting the node to the network speed/latency when it should send a
// block to its peers which should be received in a limited time frame
type BlockSizeThrottler interface {
	GetCurrentMaxSize() uint32
	IsInterfaceNil() bool
}
