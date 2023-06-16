package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"math/big"
)

type PreProcessorsFactoryArgs struct {
	UnsignedTransactions       dataRetriever.ShardedDataCacherNotifier
	Store                      dataRetriever.StorageService
	Marshaller                 marshal.Marshalizer
	Hasher                     hashing.Hasher
	ScResultProcessor          SmartContractResultProcessor
	ShardCoordinator           sharding.Coordinator
	Accounts                   state.AccountsAdapter
	RequestHandler             RequestHandler
	GasHandler                 GasHandler
	EconomicsFee               FeeHandler
	PubkeyConverter            core.PubkeyConverter
	BlockSizeComputation       BlockSizeComputationHandler
	BalanceComputation         BalanceComputationHandler
	EnableEpochsHandler        common.EnableEpochsHandler
	ProcessedMiniBlocksTracker ProcessedMiniBlocksTracker
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
