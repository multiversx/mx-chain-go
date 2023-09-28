package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BootstrapperFromStorageFactoryStub -
type BootstrapperFromStorageFactoryStub struct {
}

// CreateBootstrapperFromStorage -
func (b *BootstrapperFromStorageFactoryStub) CreateBootstrapperFromStorage(_ storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	return &testscommon.StorageBootstrapperMock{}, nil
}

// IsInterfaceNil -
func (b *BootstrapperFromStorageFactoryStub) IsInterfaceNil() bool {
	return b == nil
}
