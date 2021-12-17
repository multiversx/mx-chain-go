package track

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgBaseTracker holds all dependencies required by the process data factory in order to create
// new instances of shard/meta block tracker
type ArgBaseTracker struct {
	Hasher           hashing.Hasher
	HeaderValidator  process.HeaderConstructionValidator
	Marshalizer      marshal.Marshalizer
	RequestHandler   process.RequestHandler
	RoundHandler     process.RoundHandler
	ShardCoordinator sharding.Coordinator
	Store            dataRetriever.StorageService
	StartHeaders     map[uint32]data.HeaderHandler
	PoolsHolder      dataRetriever.PoolsHolder
	WhitelistHandler process.WhiteListHandler
	FeeHandler       process.FeeHandler
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
