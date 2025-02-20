package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	storagerequesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

// ShardRequestersContainerCreatorMock -
type ShardRequestersContainerCreatorMock struct {
	CreateShardRequestersContainerFactoryCalled func(args storagerequesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error)
}

// CreateShardRequestersContainerFactory -
func (mock *ShardRequestersContainerCreatorMock) CreateShardRequestersContainerFactory(args storagerequesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	if mock.CreateShardRequestersContainerFactoryCalled != nil {
		return mock.CreateShardRequestersContainerFactoryCalled(args)
	}

	return &dataRetrieverMock.ShardRequestersContainerFactoryMock{}, nil
}

// IsInterfaceNil -
func (mock *ShardRequestersContainerCreatorMock) IsInterfaceNil() bool {
	return mock == nil
}
