package preprocess

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

// SmartContractResultPreProcessorCreatorArgs is the struct containing the data needed to create a SmartContractResultPreProcessor
type SmartContractResultPreProcessorCreatorArgs struct {
	ScrDataPool                  dataRetriever.ShardedDataCacherNotifier
	Store                        dataRetriever.StorageService
	Hasher                       hashing.Hasher
	Marshalizer                  marshal.Marshalizer
	ScrProcessor                 process.SmartContractResultProcessor
	ShardCoordinator             sharding.Coordinator
	Accounts                     state.AccountsAdapter
	OnRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	GasHandler                   process.GasHandler
	EconomicsFee                 process.FeeHandler
	PubkeyConverter              core.PubkeyConverter
	BlockSizeComputation         BlockSizeComputationHandler
	BalanceComputation           BalanceComputationHandler
	EnableEpochsHandler          common.EnableEpochsHandler
	ProcessedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
}
