package requesterscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type shardRequestersContainerFactoryCreator struct {
}

// NewShardRequestersContainerFactoryCreator creates a new shard requester container factory creator
func NewShardRequestersContainerFactoryCreator() *shardRequestersContainerFactoryCreator {
	return &shardRequestersContainerFactoryCreator{}
}

// CreateRequesterContainerFactory creates a requester container factory for regular shards
func (f *shardRequestersContainerFactoryCreator) CreateRequesterContainerFactory(args FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	return NewShardRequestersContainerFactory(args)
}

// IsInterfaceNil checks if underlying pointer is nil
func (f *shardRequestersContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
