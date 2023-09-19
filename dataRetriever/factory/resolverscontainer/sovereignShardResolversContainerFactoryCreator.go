package resolverscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type sovereignShardResolversContainerFactoryCreator struct {
}

// NewSovereignShardResolversContainerFactoryCreator creates a new sovereign shard resolvers container factory creator
func NewSovereignShardResolversContainerFactoryCreator() *sovereignShardResolversContainerFactoryCreator {
	return &sovereignShardResolversContainerFactoryCreator{}
}

// CreateShardResolversContainerFactory creates a shard resolver container factory for sovereign shards
func (f *sovereignShardResolversContainerFactoryCreator) CreateShardResolversContainerFactory(args FactoryArgs,
) (dataRetriever.ResolversContainerFactory, error) {
	shardContainer, err := NewShardResolversContainerFactory(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignShardResolversContainerFactory(shardContainer)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignShardResolversContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
