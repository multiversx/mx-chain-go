package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// InterceptorsContainerFactoryMock -
type InterceptorsContainerFactoryMock struct {
	CreateInterceptorsContainerFactoryCalled func(args interceptorscontainer.CommonInterceptorsContainerFactoryArgs) (process.InterceptorsContainerFactory, error)
}

// CreateInterceptorsContainerFactory -
func (i *InterceptorsContainerFactoryMock) CreateInterceptorsContainerFactory(args interceptorscontainer.CommonInterceptorsContainerFactoryArgs) (process.InterceptorsContainerFactory, error) {
	if i.CreateInterceptorsContainerFactoryCalled != nil {
		return i.CreateInterceptorsContainerFactory(args)
	}
	return &testscommon.ShardInterceptorsContainerFactoryMock{}, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (i *InterceptorsContainerFactoryMock) IsInterfaceNil() bool {
	return i == nil
}
