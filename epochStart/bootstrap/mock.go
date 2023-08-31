package bootstrap

import (
	"github.com/multiversx/mx-chain-go/common"
)

// EpochStartBootstrapperFactoryStub -
type EpochStartBootstrapperFactoryStub struct {
	CreateEpochStartBootstrapperCalled func(args ArgsEpochStartBootstrap) (EpochStartBootstrapper, error)
}

// CreateEpochStartBootstrapper -
func (e *EpochStartBootstrapperFactoryStub) CreateEpochStartBootstrapper(args ArgsEpochStartBootstrap) (EpochStartBootstrapper, error) {
	if e.CreateEpochStartBootstrapperCalled != nil {
		return e.CreateEpochStartBootstrapperCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (e *EpochStartBootstrapperFactoryStub) IsInterfaceNil() bool {
	return e == nil
}

// EpochStartBootstrapperStub -
type EpochStartBootstrapperStub struct {
	TrieHolder      common.TriesHolder
	StorageManagers map[string]common.StorageManager
	BootstrapCalled func() (Parameters, error)
}

// Bootstrap -
func (esbs *EpochStartBootstrapperStub) Bootstrap() (Parameters, error) {
	if esbs.BootstrapCalled != nil {
		return esbs.BootstrapCalled()
	}

	return Parameters{}, nil
}

// IsInterfaceNil -
func (esbs *EpochStartBootstrapperStub) IsInterfaceNil() bool {
	return esbs == nil
}

// Close -
func (esbs *EpochStartBootstrapperStub) Close() error {
	return nil
}
