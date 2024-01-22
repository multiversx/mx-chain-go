package resolverscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

// ShardResolversContainerFactoryCreator defines a shard resolvers container factory creator
type ShardResolversContainerFactoryCreator interface {
	CreateShardResolversContainerFactory(args FactoryArgs) (dataRetriever.ResolversContainerFactory, error)
	IsInterfaceNil() bool
}
