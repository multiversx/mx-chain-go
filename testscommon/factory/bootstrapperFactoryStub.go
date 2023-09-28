package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// BootstrapperFactoryStub -
type BootstrapperFactoryStub struct {
}

// CreateBootstrapper -
func (b *BootstrapperFactoryStub) CreateBootstrapper(_ sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	return &testscommon.BootstrapperMock{}, nil
}

// IsInterfaceNil -
func (b *BootstrapperFactoryStub) IsInterfaceNil() bool {
	return b == nil
}
