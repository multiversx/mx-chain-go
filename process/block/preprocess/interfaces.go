package preprocess

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

// SortedTransactionsProvider defines the public API of the transactions cache
type SortedTransactionsProvider interface {
	GetSortedTransactions() []*txcache.WrappedTransaction
	NotifyAccountNonce(accountKey []byte, nonce uint64)
	IsInterfaceNil() bool
}

// TxCache defines the functionality for the transactions cache
type TxCache interface {
	SelectTransactions(numRequested int, batchSizePerSender int) []*txcache.WrappedTransaction
	NotifyAccountNonce(accountKey []byte, nonce uint64)
	IsInterfaceNil() bool
}

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	IsShardStuck(shardID uint32) bool
	IsInterfaceNil() bool
}

// BlockSizeComputationHandler defines the functionality for block size computation
type BlockSizeComputationHandler interface {
	Init()
	AddNumMiniBlocks(numMiniBlocks int)
	AddNumTxs(numTxs int)
	IsMaxBlockSizeReached(numNewMiniBlocks int, numNewTxs int) bool
	IsMaxBlockSizeWithoutThrottleReached(numNewMiniBlocks int, numNewTxs int) bool
	IsInterfaceNil() bool
}

// BlockSizeThrottler defines the functionality of adapting the node to the network speed/latency when it should send a
// block to its peers which should be received in a limited time frame
type BlockSizeThrottler interface {
	GetCurrentMaxSize() uint32
	IsInterfaceNil() bool
}

// BalanceComputationHandler defines the functionality for addresses balances computation, used in preventing executing
// too many debit transactions, after the proposer executed a credit transaction on the same account in the same block
type BalanceComputationHandler interface {
	Init()
	SetBalanceToAddress(address []byte, value *big.Int)
	AddBalanceToAddress(address []byte, value *big.Int) bool
	SubBalanceFromAddress(address []byte, value *big.Int) bool
	IsAddressSet(address []byte) bool
	AddressHasEnoughBalance(address []byte, value *big.Int) bool
	IsInterfaceNil() bool
}

// ScheduledTxsExecutionHandler defines the functionality for execution of scheduled transactions
type ScheduledTxsExecutionHandler interface {
	Init()
	Add(txHash []byte, tx data.TransactionHandler) bool
	Execute(txHash []byte) error
	ExecuteAll() error
	IsInterfaceNil() bool
}
