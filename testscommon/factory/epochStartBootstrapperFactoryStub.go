package factory

import "github.com/multiversx/mx-chain-go/epochStart/bootstrap"

// EpochStartBootstrapperFactoryStub -
type EpochStartBootstrapperFactoryStub struct {
	CreateEpochStartBootstrapperCalled func(args bootstrap.ArgsEpochStartBootstrap) (bootstrap.EpochStartBootstrapper, error)
}

// NewEpochStartBootstrapperFactoryStub -
func NewEpochStartBootstrapperFactoryStub() *EpochStartBootstrapperFactoryStub {
	return &EpochStartBootstrapperFactoryStub{}
}

// CreateEpochStartBootstrapper -
func (e *EpochStartBootstrapperFactoryStub) CreateEpochStartBootstrapper(args bootstrap.ArgsEpochStartBootstrap) (bootstrap.EpochStartBootstrapper, error) {
	if e.CreateEpochStartBootstrapperCalled != nil {
		return e.CreateEpochStartBootstrapperCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (e *EpochStartBootstrapperFactoryStub) IsInterfaceNil() bool {
	return false
}
