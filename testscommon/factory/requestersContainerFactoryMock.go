package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	requesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/requestersContainer"
	dataRetriever2 "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

// RequestersContainerFactoryMock -
type RequestersContainerFactoryMock struct {
	CreateRequesterContainerFactoryCalled func(args requesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error)
}

// CreateRequesterContainerFactory creates a requester container factory for regular shards
func (r *RequestersContainerFactoryMock) CreateRequesterContainerFactory(args requesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error) {
	if r.CreateRequesterContainerFactoryCalled != nil {
		return r.CreateRequesterContainerFactory(args)
	}
	return &dataRetriever2.ShardRequestersContainerFactoryMock{}, nil
}

// IsInterfaceNil checks if underlying pointer is nil
func (r *RequestersContainerFactoryMock) IsInterfaceNil() bool {
	return r == nil
}
