package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	storagerequesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
)

type sovereignShardRequestersContainerCreator struct {
}

// NewSovereignShardRequestersContainerCreator creates a storage sovereign shard requesters container factory creator
func NewSovereignShardRequestersContainerCreator() *sovereignShardRequestersContainerCreator {
	return &sovereignShardRequestersContainerCreator{}
}

// CreateShardRequestersContainerFactory creates a storage sovereign shard requesters container factory
func (f *sovereignShardRequestersContainerCreator) CreateShardRequestersContainerFactory(args storagerequesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	shardFactory, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)
	if err != nil {
		return nil, err
	}

	return storagerequesterscontainer.NewSovereignShardRequestersContainerFactory(shardFactory)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *sovereignShardRequestersContainerCreator) IsInterfaceNil() bool {
	return f == nil
}
