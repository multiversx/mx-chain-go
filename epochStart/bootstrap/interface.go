package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// EpochStartMetaBlockInterceptorHandler defines what a component which will handle receiving the epoch start meta blocks should do
type EpochStartMetaBlockInterceptorHandler interface {
	process.Interceptor
	GetEpochStartMetaBlock(target int, epoch uint32) (*block.MetaBlock, error)
}

// MetaBlockInterceptorHandler defines what a component which will handle receiving the meta blocks should do
type MetaBlockInterceptorHandler interface {
	process.Interceptor
	GetMetaBlock(hash []byte, target int) (*block.MetaBlock, error)
}

// ShardHeaderInterceptorHandler defines what a component which will handle receiving the the shard headers should do
type ShardHeaderInterceptorHandler interface {
	process.Interceptor
	GetShardHeader(hash []byte, target int) (*block.Header, error)
}

// MiniBlockInterceptorHandler defines what a component which will handle receiving the mini blocks should do
type MiniBlockInterceptorHandler interface {
	process.Interceptor
	GetMiniBlock(hash []byte, target int) (*block.MiniBlock, error)
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
