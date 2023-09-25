package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

// BootstrapperFactoryStub -
type BootstrapperFactoryStub struct {
	CreateBootstrapperCalled func(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error)
}

// CreateBootstrapper -
func (b *BootstrapperFactoryStub) CreateBootstrapper(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	if b.CreateBootstrapperCalled != nil {
		return b.CreateBootstrapperCalled(argsBaseBootstrapper)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BootstrapperFactoryStub) IsInterfaceNil() bool {
	return b == nil
}
