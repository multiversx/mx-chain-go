package block

import (
	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type coreComponentsHolder interface {
	Hasher() hashing.Hasher
	AddressPubKeyConverter() core.PubkeyConverter
	InternalMarshalizer() marshal.Marshalizer
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	EpochNotifier() process.EpochNotifier
	EnableEpochsHandler() common.EnableEpochsHandler
	RoundNotifier() process.RoundNotifier
	EnableRoundsHandler() process.EnableRoundsHandler
	RoundHandler() consensus.RoundHandler
	EconomicsData() process.EconomicsDataHandler
	ProcessStatusHandler() common.ProcessStatusHandler
	IsInterfaceNil() bool
}

type dataComponentsHolder interface {
	StorageService() dataRetriever.StorageService
	Datapool() dataRetriever.PoolsHolder
	Blockchain() data.ChainHandler
	IsInterfaceNil() bool
}

type bootstrapComponentsHolder interface {
	ShardCoordinator() sharding.Coordinator
	VersionedHeaderFactory() nodeFactory.VersionedHeaderFactory
	HeaderIntegrityVerifier() nodeFactory.HeaderIntegrityVerifierHandler
	IsInterfaceNil() bool
}

type statusComponentsHolder interface {
	OutportHandler() outport.OutportHandler
	IsInterfaceNil() bool
}

type statusCoreComponentsHolder interface {
	AppStatusHandler() core.AppStatusHandler
	IsInterfaceNil() bool
}

// ArgBaseProcessor holds all dependencies required by the process data factory in order to create
// new instances
type ArgBaseProcessor struct {
	CoreComponents       coreComponentsHolder
	DataComponents       dataComponentsHolder
	BootstrapComponents  bootstrapComponentsHolder
	StatusComponents     statusComponentsHolder
	StatusCoreComponents statusCoreComponentsHolder
	RunTypeComponents    runTypeComponentsHolder

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
	VMContainersFactory            process.VirtualMachinesContainerFactory
	VmContainer                    process.VirtualMachinesContainer
	GasHandler                     gasConsumedProvider
	OutportDataProvider            outport.DataProviderOutport
	ScheduledTxsExecutionHandler   process.ScheduledTxsExecutionHandler
	ScheduledMiniBlocksEnableEpoch uint32
	ProcessedMiniBlocksTracker     process.ProcessedMiniBlocksTracker
	ReceiptsRepository             receiptsRepository
	BlockProcessingCutoffHandler   cutoff.BlockProcessingCutoffHandler
	ManagedPeersHolder             common.ManagedPeersHolder
	ValidatorStatisticsProcessor   process.ValidatorStatisticsProcessor
	OutGoingOperationsPool         OutGoingOperationsPool
	DataCodec                      sovereign.DataDecoderHandler
	TopicsChecker                  sovereign.TopicsCheckerHandler
}

// ArgShardProcessor holds all dependencies required by the process data factory in order to create
// new instances of shard processor
type ArgShardProcessor struct {
	ArgBaseProcessor
}

// ArgMetaProcessor holds all dependencies required by the process data factory in order to create
// new instances of meta processor
type ArgMetaProcessor struct {
	ArgBaseProcessor
	PendingMiniBlocksHandler     process.PendingMiniBlocksHandler
	SCToProtocol                 process.SmartContractToProtocolHandler
	EpochStartDataCreator        process.EpochStartDataCreator
	EpochEconomics               process.EndOfEpochEconomics
	EpochRewardsCreator          process.RewardsCreator
	EpochValidatorInfoCreator    process.EpochStartValidatorInfoCreator
	EpochSystemSCProcessor       process.EpochStartSystemSCProcessor
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
}
