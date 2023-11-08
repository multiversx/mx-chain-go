package sync

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

// ArgBaseBootstrapper holds all dependencies required by the bootstrap data factory in order to create
// new instances
type ArgBaseBootstrapper struct {
	HistoryRepo                  dblookupext.HistoryRepository
	PoolsHolder                  dataRetriever.PoolsHolder
	Store                        dataRetriever.StorageService
	ChainHandler                 data.ChainHandler
	RoundHandler                 consensus.RoundHandler
	BlockProcessor               process.BlockProcessor
	BlockProcessorWithRevert     process.BlockProcessor
	WaitTime                     time.Duration
	Hasher                       hashing.Hasher
	Marshalizer                  marshal.Marshalizer
	ForkDetector                 process.ForkDetector
	RequestHandler               process.RequestHandler
	ShardCoordinator             sharding.Coordinator
	Accounts                     state.AccountsAdapter
	BlackListHandler             process.TimeCacher
	NetworkWatcher               process.NetworkConnectionWatcher
	BootStorer                   process.BootStorer
	StorageBootstrapper          process.BootstrapperFromStorage
	EpochHandler                 dataRetriever.EpochHandler
	MiniblocksProvider           process.MiniBlockProvider
	Uint64Converter              typeConverters.Uint64ByteSliceConverter
	AppStatusHandler             core.AppStatusHandler
	OutportHandler               outport.OutportHandler
	AccountsDBSyncer             process.AccountsDBSyncer
	CurrentEpochProvider         process.CurrentNetworkEpochProviderHandler
	IsInImportMode               bool
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	ProcessWaitTime              time.Duration
	RepopulateTokensSupplies     bool
}

// ArgShardBootstrapper holds all dependencies required by the bootstrap data factory in order to create
// new instances of shard bootstrapper
type ArgShardBootstrapper struct {
	ArgBaseBootstrapper
}

// ArgMetaBootstrapper holds all dependencies required by the bootstrap data factory in order to create
// new instances of meta bootstrapper
type ArgMetaBootstrapper struct {
	ArgBaseBootstrapper
	EpochBootstrapper           process.EpochBootstrapper
	ValidatorStatisticsDBSyncer process.AccountsDBSyncer
	ValidatorAccountsDB         state.AccountsAdapter
}
