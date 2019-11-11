package block

import (
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgBaseProcessor holds all dependencies required by the process data factory in order to create
// new instances
type ArgBaseProcessor struct {
	Accounts                     state.AccountsAdapter
	ForkDetector                 process.ForkDetector
	Hasher                       hashing.Hasher
	Marshalizer                  marshal.Marshalizer
	Store                        dataRetriever.StorageService
	ShardCoordinator             sharding.Coordinator
	NodesCoordinator             sharding.NodesCoordinator
	SpecialAddressHandler        process.SpecialAddressHandler
	Uint64Converter              typeConverters.Uint64ByteSliceConverter
	StartHeaders                 map[uint32]data.HeaderHandler
	RequestHandler               process.RequestHandler
	Core                         serviceContainer.Core
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
	EndOfEpochTrigger            process.EndOfEpochTriggerHandler
	HeaderValidator              process.HeaderConstructionValidator
}

// ArgShardProcessor holds all dependencies required by the process data factory in order to create
// new instances of shard processor
type ArgShardProcessor struct {
	ArgBaseProcessor
	DataPool        dataRetriever.PoolsHolder
	TxCoordinator   process.TransactionCoordinator
	TxsPoolsCleaner process.PoolsCleaner
}

// ArgMetaProcessor holds all dependencies required by the process data factory in order to create
// new instances of meta processor
type ArgMetaProcessor struct {
	ArgBaseProcessor
	DataPool          dataRetriever.MetaPoolsHolder
	PendingMiniBlocks process.PendingMiniBlocksHandler
}
