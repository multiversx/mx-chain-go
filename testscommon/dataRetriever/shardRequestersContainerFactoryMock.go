package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

type ShardRequestersContainerFactoryMock struct {
	CreateCalled func() (dataRetriever.RequestersContainer, error)
}

func (s *ShardRequestersContainerFactoryMock) Create() (dataRetriever.RequestersContainer, error) {
	if s.CreateCalled != nil {
		return s.Create()
	}
	return &RequestersContainerStub{}, nil
}

func (s *ShardRequestersContainerFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
