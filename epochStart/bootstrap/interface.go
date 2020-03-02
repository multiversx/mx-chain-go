package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type metaBlockInterceptorHandler interface {
	process.Interceptor
	GetMetaBlock(target int, epoch uint32) (*block.MetaBlock, error)
}

type shardHeaderInterceptorHandler interface {
	process.Interceptor
	GetAllReceivedShardHeaders() []block.ShardData
}

type metaBlockResolverHandler interface {
	RequestEpochStartMetaBlock(epoch uint32) error
}

// NodesConfigProviderHandler defines what a component which will handle the nodes config should be able to do
type NodesConfigProviderHandler interface {
	GetNodesConfigForMetaBlock(metaBlock *block.MetaBlock) (*sharding.NodesSetup, error)
}
