package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// MetaBlockInterceptorHandler defines what a component which will handle receiving the meta blocks should do
type MetaBlockInterceptorHandler interface {
	process.Interceptor
	GetMetaBlock(target int, epoch uint32) (*block.MetaBlock, error)
}

// ShardHeaderInterceptorHandler defines what a component which will handle receiving the the shard headers should do
type ShardHeaderInterceptorHandler interface {
	process.Interceptor
	GetAllReceivedShardHeaders() []block.ShardData
}

// MetaBlockResolverHandler defines what a component which will handle requesting a metablock should do
type MetaBlockResolverHandler interface {
	RequestEpochStartMetaBlock(epoch uint32) error
	IsInterfaceNil() bool
}

// NodesConfigProviderHandler defines what a component which will handle the nodes config should be able to do
type NodesConfigProviderHandler interface {
	GetNodesConfigForMetaBlock(metaBlock *block.MetaBlock) (*sharding.NodesSetup, error)
	IsInterfaceNil() bool
}

// EpochStartDataProviderHandler defines what a component which fetches the data needed for starting in an epoch should do
type EpochStartDataProviderHandler interface {
	Bootstrap() (*ComponentsNeededForBootstrap, error)
}
