package testscommon

import (
	"github.com/multiversx/mx-chain-go/process"
)

// ShardInterceptorsContainerFactoryMock -
type ShardInterceptorsContainerFactoryMock struct {
	CreateCalled func() (process.InterceptorsContainer, process.InterceptorsContainer, error)
}

// Create -
func (s *ShardInterceptorsContainerFactoryMock) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	if s.CreateCalled != nil {
		return s.Create()
	}
	return &InterceptorsContainerStub{}, &InterceptorsContainerStub{}, nil
}

// IsInterfaceNil - checks if underlying pointer is nil
func (s *ShardInterceptorsContainerFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
