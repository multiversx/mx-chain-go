package requesterscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type sovereignShardRequestersContainerFactoryCreator struct {
}

// NewSovereignShardRequestersContainerFactoryCreator creates a new sovereign shard requester container factory creator
func NewSovereignShardRequestersContainerFactoryCreator() *sovereignShardRequestersContainerFactoryCreator {
	return &sovereignShardRequestersContainerFactoryCreator{}
}

// CreateRequesterContainerFactory creates a requester container factory for sovereign shards
func (f *sovereignShardRequestersContainerFactoryCreator) CreateRequesterContainerFactory(args FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	shardFactory, err := NewShardRequestersContainerFactory(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignShardRequestersContainerFactory(shardFactory)
}

// IsInterfaceNil checks if underlying pointer is nil
func (f *sovereignShardRequestersContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
