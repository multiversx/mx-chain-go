package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
)

// BootstrapperFromStorageFactoryStub -
type BootstrapperFromStorageFactoryStub struct {
	CreateBootstrapperFromStorageCalled func(args storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error)
}

// CreateBootstrapperFromStorage -
func (b *BootstrapperFromStorageFactoryStub) CreateBootstrapperFromStorage(args storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	if b.CreateBootstrapperFromStorageCalled != nil {
		return b.CreateBootstrapperFromStorageCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (b *BootstrapperFromStorageFactoryStub) IsInterfaceNil() bool {
	return b == nil
}
