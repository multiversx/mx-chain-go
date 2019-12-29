package block

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
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
	BlockChainHook               process.BlockChainHookHandler
	TxCoordinator                process.TransactionCoordinator
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
	EpochStartTrigger            process.EpochStartTriggerHandler
	HeaderValidator              process.HeaderConstructionValidator
	Rounder                      consensus.Rounder
	BootStorer                   process.BootStorer
}

// ArgShardProcessor holds all dependencies required by the process data factory in order to create
// new instances of shard processor
type ArgShardProcessor struct {
	ArgBaseProcessor
	DataPool        dataRetriever.PoolsHolder
	TxsPoolsCleaner process.PoolsCleaner
}

// ArgMetaProcessor holds all dependencies required by the process data factory in order to create
// new instances of meta processor
type ArgMetaProcessor struct {
	ArgBaseProcessor
	DataPool           dataRetriever.MetaPoolsHolder
	PendingMiniBlocks  process.PendingMiniBlocksHandler
	SCDataGetter       external.SCQueryService
	PeerChangesHandler process.PeerChangesHandler
	SCToProtocol       process.SmartContractToProtocolHandler
}
