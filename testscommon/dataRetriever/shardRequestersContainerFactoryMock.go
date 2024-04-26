package dataRetriever

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// ShardRequestersContainerFactoryMock -
type ShardRequestersContainerFactoryMock struct {
	CreateCalled func() (dataRetriever.RequestersContainer, error)
}

// Create -
func (s *ShardRequestersContainerFactoryMock) Create() (dataRetriever.RequestersContainer, error) {
	if s.CreateCalled != nil {
		return s.Create()
	}
	return &RequestersContainerStub{}, nil
}

// IsInterfaceNil -
func (s *ShardRequestersContainerFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
