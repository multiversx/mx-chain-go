package resolverscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type shardResolversContainerFactoryCreator struct {
}

func NewShardResolversContainerFactoryCreator() *shardResolversContainerFactoryCreator {
	return &shardResolversContainerFactoryCreator{}
}

func (f *shardResolversContainerFactoryCreator) CreateShardResolversContainerFactory(args FactoryArgs,
) (dataRetriever.ResolversContainerFactory, error) {
	return NewShardResolversContainerFactory(args)
}

func (f *shardResolversContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
