package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	storagerequesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
)

type shardRequestersContainerCreator struct {
}

// NewShardRequestersContainerCreator creates a storage shard requesters container factory creator
func NewShardRequestersContainerCreator() *shardRequestersContainerCreator {
	return &shardRequestersContainerCreator{}
}

// CreateShardRequestersContainerFactory creates a storage shard requesters container factory
func (f *shardRequestersContainerCreator) CreateShardRequestersContainerFactory(args storagerequesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	return storagerequesterscontainer.NewShardRequestersContainerFactory(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *shardRequestersContainerCreator) IsInterfaceNil() bool {
	return f == nil
}
