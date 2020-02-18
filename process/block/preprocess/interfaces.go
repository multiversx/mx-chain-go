package preprocess

import "github.com/ElrondNetwork/elrond-go/data"

// SortedTransactionsProvider defines the public API of the transactions cache
type SortedTransactionsProvider interface {
	GetSortedTransactions() ([]data.TransactionHandler, [][]byte)
	NotifyAccountNonce(accountKey []byte, nonce uint64)
	IsInterfaceNil() bool
}

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	IsShardStuck(shardID uint32) bool
	IsInterfaceNil() bool
}
