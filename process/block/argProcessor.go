package block

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

type coreComponentsHolder interface {
	Hasher() hashing.Hasher
	InternalMarshalizer() marshal.Marshalizer
	Uint64ByteSliceConverter() typeConverters.Uint64ByteSliceConverter
	EpochNotifier() process.EpochNotifier
	EnableEpochsHandler() common.EnableEpochsHandler
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

// ArgShardProcessor holds all dependencies required by the process data factory in order to create
// new instances of shard processor
type ArgShardProcessor struct {
	factory.ArgBaseProcessor
}

// ArgMetaProcessor holds all dependencies required by the process data factory in order to create
// new instances of meta processor
type ArgMetaProcessor struct {
	factory.ArgBaseProcessor
	PendingMiniBlocksHandler     process.PendingMiniBlocksHandler
	SCToProtocol                 process.SmartContractToProtocolHandler
	EpochStartDataCreator        process.EpochStartDataCreator
	EpochEconomics               process.EndOfEpochEconomics
	EpochRewardsCreator          process.RewardsCreator
	EpochValidatorInfoCreator    process.EpochStartValidatorInfoCreator
	EpochSystemSCProcessor       process.EpochStartSystemSCProcessor
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
}
