package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BootstrapperFromStorageFactoryMock -
type BootstrapperFromStorageFactoryMock struct {
	CreateBootstrapperFromStorageCalled func(args storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error)
}

// CreateBootstrapperFromStorage -
func (b *BootstrapperFromStorageFactoryMock) CreateBootstrapperFromStorage(args storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	if b.CreateBootstrapperFromStorageCalled != nil {
		return b.CreateBootstrapperFromStorageCalled(args)
	}
	return &testscommon.StorageBootstrapperMock{}, nil
}

// IsInterfaceNil -
func (b *BootstrapperFromStorageFactoryMock) IsInterfaceNil() bool {
	return b == nil
}
