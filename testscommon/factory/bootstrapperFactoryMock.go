package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BootstrapperFactoryMock -
type BootstrapperFactoryMock struct {
	CreateBootstrapperCalled func(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error)
}

// CreateBootstrapper -
func (b *BootstrapperFactoryMock) CreateBootstrapper(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	if b.CreateBootstrapperCalled != nil {
		return b.CreateBootstrapperCalled(argsBaseBootstrapper)
	}
	return &testscommon.BootstrapperMock{}, nil
}

// IsInterfaceNil -
func (b *BootstrapperFactoryMock) IsInterfaceNil() bool {
	return b == nil
}
