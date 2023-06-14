package track

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgBaseTracker holds all dependencies required by the process data factory in order to create
// new instances of shard/meta block tracker
type ArgBaseTracker struct {
	Hasher                 hashing.Hasher
	HeaderValidator        process.HeaderConstructionValidator
	Marshalizer            marshal.Marshalizer
	RequestHandler         process.RequestHandler
	RoundHandler           process.RoundHandler
	ShardCoordinator       sharding.Coordinator
	Store                  dataRetriever.StorageService
	StartHeaders           map[uint32]data.HeaderHandler
	PoolsHolder            dataRetriever.PoolsHolder
	WhitelistHandler       process.WhiteListHandler
	FeeHandler             process.FeeHandler
	ChainParametersHandler process.ChainParametersHandler
}

// ArgShardTracker holds all dependencies required by the process data factory in order to create
// new instances of shard block tracker
type ArgShardTracker struct {
	ArgBaseTracker
}

// ArgMetaTracker holds all dependencies required by the process data factory in order to create
// new instances of meta block tracker
type ArgMetaTracker struct {
	ArgBaseTracker
}
