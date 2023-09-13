package resolverscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type ShardResolversContainerFactoryCreator interface {
	CreateShardResolversContainerFactory(args FactoryArgs) (dataRetriever.ResolversContainerFactory, error)
	IsInterfaceNil() bool
}
