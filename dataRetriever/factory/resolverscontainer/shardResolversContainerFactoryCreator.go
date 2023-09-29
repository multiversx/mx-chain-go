package resolverscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type shardResolversContainerFactoryCreator struct {
}

// NewShardResolversContainerFactoryCreator creates a new shard resolvers container factory creator
func NewShardResolversContainerFactoryCreator() *shardResolversContainerFactoryCreator {
	return &shardResolversContainerFactoryCreator{}
}

// CreateShardResolversContainerFactory creates a shard resolver container factory for regular shards
func (f *shardResolversContainerFactoryCreator) CreateShardResolversContainerFactory(args FactoryArgs,
) (dataRetriever.ResolversContainerFactory, error) {
	return NewShardResolversContainerFactory(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *shardResolversContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
