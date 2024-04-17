package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	dataRetriever2 "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

type ShardRequestersContainerFactoryMock struct {
	CreateCalled func() (dataRetriever.RequestersContainer, error)
}

func (s *ShardRequestersContainerFactoryMock) Create() (dataRetriever.RequestersContainer, error) {
	if s.CreateCalled != nil {
		return s.Create()
	}
	return &dataRetriever2.RequestersContainerStub{}, nil
}

func (s *ShardRequestersContainerFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
