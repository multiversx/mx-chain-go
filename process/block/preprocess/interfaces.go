package preprocess

import (
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

// SortedTransactionsProvider defines the public API of the transactions cache
type SortedTransactionsProvider interface {
	GetSortedTransactions(session txcache.SelectionSession) []*txcache.WrappedTransaction
	IsInterfaceNil() bool
}

// TxCache defines the functionality for the transactions cache
type TxCache interface {
	SelectTransactions(session txcache.SelectionSession, gasRequested uint64, maxNum int, selectionLoopMaximumDuration time.Duration) ([]*txcache.WrappedTransaction, uint64)
	IsInterfaceNil() bool
}

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	IsShardStuck(shardID uint32) bool
	ShouldSkipMiniBlocksCreationFromSelf() bool
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

// TxsForBlockHandler defines the functionality for handling transactions for a block
type TxsForBlockHandler interface {
	Reset()
	AddTransaction(
		txHash []byte,
		tx data.TransactionHandler,
		senderShardID uint32,
		receiverShardID uint32,
	)
	GetTxInfoByHash(hash []byte) (*txInfo, bool)
	GetAllCurrentUsedTxs() map[string]data.TransactionHandler
	GetMissingTxsCount() int
	ReceivedTransaction(txHash []byte, tx data.TransactionHandler)
	HasMissingTransactions() bool
	ComputeExistingAndRequestMissing(
		body *block.Body,
		isMiniBlockCorrect func(block.Type) bool,
		txPool dataRetriever.ShardedDataCacherNotifier,
		onRequestTxs func(shardID uint32, txHashes [][]byte),
	) int
	IsInterfaceNil() bool
}
