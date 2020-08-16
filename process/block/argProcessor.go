package block

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgBaseProcessor holds all dependencies required by the process data factory in order to create
// new instances
type ArgBaseProcessor struct {
	AccountsDB             map[state.AccountsDbIdentifier]state.AccountsAdapter
	ForkDetector           process.ForkDetector
	Hasher                 hashing.Hasher
	Marshalizer            marshal.Marshalizer
	Store                  dataRetriever.StorageService
	ShardCoordinator       sharding.Coordinator
	NodesCoordinator       sharding.NodesCoordinator
	FeeHandler             process.TransactionFeeHandler
	Uint64Converter        typeConverters.Uint64ByteSliceConverter
	RequestHandler         process.RequestHandler
	BlockChainHook         process.BlockChainHookHandler
	TxCoordinator          process.TransactionCoordinator
	EpochStartTrigger      process.EpochStartTriggerHandler
	HeaderValidator        process.HeaderConstructionValidator
	Rounder                consensus.Rounder
	BootStorer             process.BootStorer
	BlockTracker           process.BlockTracker
	DataPool               dataRetriever.PoolsHolder
	BlockChain             data.ChainHandler
	StateCheckpointModulus uint
	BlockSizeThrottler     process.BlockSizeThrottler
	Indexer                indexer.Indexer
	TpsBenchmark           statistics.TPSBenchmark
	Version                string
	HistoryRepository      fullHistory.HistoryRepository
	EpochNotifier          process.EpochNotifier
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
	SCDataGetter                 external.SCQueryService
	SCToProtocol                 process.SmartContractToProtocolHandler
	EpochStartDataCreator        process.EpochStartDataCreator
	EpochEconomics               process.EndOfEpochEconomics
	EpochRewardsCreator          process.EpochStartRewardsCreator
	EpochValidatorInfoCreator    process.EpochStartValidatorInfoCreator
	EpochSystemSCProcessor       process.EpochStartSystemSCProcessor
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
}
