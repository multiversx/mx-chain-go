package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

// RequestersContainerFactoryMock -
type RequestersContainerFactoryMock struct {
	CreateRequesterContainerFactoryCalled func(args requesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error)
}

// CreateRequesterContainerFactory -
func (r *RequestersContainerFactoryMock) CreateRequesterContainerFactory(args requesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	if r.CreateRequesterContainerFactoryCalled != nil {
		return r.CreateRequesterContainerFactory(args)
	}
	return &dataRetrieverMocks.ShardRequestersContainerFactoryMock{}, nil
}

// IsInterfaceNil checks if underlying pointer is nil
func (r *RequestersContainerFactoryMock) IsInterfaceNil() bool {
	return r == nil
}
