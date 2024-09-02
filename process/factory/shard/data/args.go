package data

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

// RunTypeComponents defines the needed run type components for pre-processor factory
type RunTypeComponents interface {
	TxPreProcessorCreator() preprocess.TxPreProcessorCreator
	SCResultsPreProcessorCreator() preprocess.SmartContractResultPreProcessorCreator
	IsInterfaceNil() bool
}

// ArgPreProcessorsContainerFactory defines the arguments needed by the pre-processor container factory
type ArgPreProcessorsContainerFactory struct {
	ShardCoordinator             sharding.Coordinator
	Store                        dataRetriever.StorageService
	Marshaller                   marshal.Marshalizer
	Hasher                       hashing.Hasher
	DataPool                     dataRetriever.PoolsHolder
	PubkeyConverter              core.PubkeyConverter
	Accounts                     state.AccountsAdapter
	RequestHandler               process.RequestHandler
	TxProcessor                  process.TransactionProcessor
	ScProcessor                  process.SmartContractProcessor
	ScResultProcessor            process.SmartContractResultProcessor
	RewardsTxProcessor           process.RewardTransactionProcessor
	EconomicsFee                 process.FeeHandler
	GasHandler                   process.GasHandler
	BlockTracker                 preprocess.BlockTracker
	BlockSizeComputation         preprocess.BlockSizeComputationHandler
	BalanceComputation           preprocess.BalanceComputationHandler
	EnableEpochsHandler          common.EnableEpochsHandler
	TxTypeHandler                process.TxTypeHandler
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	ProcessedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler      common.TxExecutionOrderHandler
	RunTypeComponents            RunTypeComponents
}

// PreProcessorsContainerFactoryCreator defines a pre-processors container factory creator
type PreProcessorsContainerFactoryCreator interface {
	CreatePreProcessorContainerFactory(
		args ArgPreProcessorsContainerFactory,
	) (process.PreProcessorsContainerFactory, error)
	IsInterfaceNil() bool
}
