package factory

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgBaseProcessor holds all dependencies required by the process data factory in order to create
// new instances
type ArgBaseProcessor struct {
	CoreComponents       CoreComponentsHolder
	DataComponents       DataComponentsHolder
	BootstrapComponents  BootstrapComponentsHolder
	StatusComponents     StatusComponentsHolder
	StatusCoreComponents StatusCoreComponentsHolder

	Config                         config.Config
	PrefsConfig                    config.Preferences
	AccountsDB                     map[state.AccountsDbIdentifier]state.AccountsAdapter
	ForkDetector                   process.ForkDetector
	NodesCoordinator               nodesCoordinator.NodesCoordinator
	FeeHandler                     process.TransactionFeeHandler
	RequestHandler                 process.RequestHandler
	BlockChainHook                 process.BlockChainHookHandler
	TxCoordinator                  process.TransactionCoordinator
	EpochStartTrigger              process.EpochStartTriggerHandler
	HeaderValidator                process.HeaderConstructionValidator
	BootStorer                     process.BootStorer
	BlockTracker                   process.BlockTracker
	BlockSizeThrottler             process.BlockSizeThrottler
	Version                        string
	HistoryRepository              dblookupext.HistoryRepository
	EnableRoundsHandler            process.EnableRoundsHandler
	VMContainersFactory            process.VirtualMachinesContainerFactory
	VmContainer                    process.VirtualMachinesContainer
	GasHandler                     process.GasHandler
	OutportDataProvider            outport.DataProviderOutport
	ScheduledTxsExecutionHandler   process.ScheduledTxsExecutionHandler
	ScheduledMiniBlocksEnableEpoch uint32
	ProcessedMiniBlocksTracker     process.ProcessedMiniBlocksTracker
	ReceiptsRepository             ReceiptsRepository
	BlockProcessingCutoffHandler   cutoff.BlockProcessingCutoffHandler
	ValidatorStatisticsProcessor   process.ValidatorStatisticsProcessor
}

// ResolverRequestArgs holds all dependencies required by the process data factory to create components
type ResolverRequestArgs struct {
	RequestersFinder      dataRetriever.RequestersFinder
	RequestedItemsHandler dataRetriever.RequestedItemsHandler
	WhiteListHandler      process.WhiteListHandler
	MaxTxsToRequest       int
	ShardID               uint32
	RequestInterval       time.Duration
}

// ScheduledTxsExecutionFactoryArgs holds all dependencies required by the process data factory to create components
type ScheduledTxsExecutionFactoryArgs struct {
	Storer           storage.Storer
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}

// ArgsHeaderValidator are the arguments needed to create a new header validator
type ArgsHeaderValidator struct {
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
}

// ShardForkDetectorFactoryArgs are the arguments needed to create a new fork detector
type ShardForkDetectorFactoryArgs struct {
	RoundHandler    consensus.RoundHandler
	HeaderBlackList process.TimeCacher
	BlockTracker    process.BlockTracker
	GenesisTime     int64
}
