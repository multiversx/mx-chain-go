package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
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
