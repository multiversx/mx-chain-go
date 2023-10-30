package factory

import (
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
)

// EpochStartBootstrapperFactoryMock -
type EpochStartBootstrapperFactoryMock struct {
	CreateEpochStartBootstrapperCalled func(args bootstrap.ArgsEpochStartBootstrap) (bootstrap.EpochStartBootstrapper, error)
}

// CreateEpochStartBootstrapper -
func (e *EpochStartBootstrapperFactoryMock) CreateEpochStartBootstrapper(args bootstrap.ArgsEpochStartBootstrap) (bootstrap.EpochStartBootstrapper, error) {
	if e.CreateEpochStartBootstrapperCalled != nil {
		return e.CreateEpochStartBootstrapperCalled(args)
	}
	return &bootstrapMocks.EpochStartBootstrapperStub{}, nil
}

// IsInterfaceNil -
func (e *EpochStartBootstrapperFactoryMock) IsInterfaceNil() bool {
	return e == nil
}
