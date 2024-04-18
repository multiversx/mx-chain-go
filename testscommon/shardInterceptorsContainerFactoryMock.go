package testscommon

import (
	"github.com/multiversx/mx-chain-go/process"
)

type ShardInterceptorsContainerFactoryMock struct {
	CreateCalled func() (process.InterceptorsContainer, process.InterceptorsContainer, error)
}

func (s *ShardInterceptorsContainerFactoryMock) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	if s.CreateCalled != nil {
		return s.Create()
	}
	return &InterceptorsContainerStub{}, &InterceptorsContainerStub{}, nil
}

func (s *ShardInterceptorsContainerFactoryMock) IsInterfaceNil() bool {
	return s == nil
}
