package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

type ShardResolversContainerFactoryMock struct {
	CreateCalled func() (dataRetriever.ResolversContainer, error)
}

func (s *ShardResolversContainerFactoryMock) Create() (dataRetriever.ResolversContainer, error) {
	if s.CreateCalled != nil {
		return s.Create()
	}
	return &ResolversContainerStub{}, nil
}

func (s *ShardResolversContainerFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
