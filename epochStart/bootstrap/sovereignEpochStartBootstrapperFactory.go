package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
)

// sovereignEpochStartBootstrapperFactory defines the epoch start bootstrapper factory for chain run sovereign
type sovereignEpochStartBootstrapperFactory struct {
	epochStartBootstrapperFactory EpochStartBootstrapperCreator
}

// NewSovereignEpochStartBootstrapperFactory creates a new epoch start bootstrapper factory for chain run sovereign
func NewSovereignEpochStartBootstrapperFactory(esbf EpochStartBootstrapperCreator) (EpochStartBootstrapperCreator, error) {
	if check.IfNil(esbf) {
		return nil, errors.ErrNilEpochStartBootstrapperFactory
	}
	return &sovereignEpochStartBootstrapperFactory{
		epochStartBootstrapperFactory: esbf,
	}, nil
}

// CreateEpochStartBootstrapper creates a new epoch start bootstrapper for sovereign chain operations
func (bcf *sovereignEpochStartBootstrapperFactory) CreateEpochStartBootstrapper(epochStartBootstrapArgs ArgsEpochStartBootstrap) (EpochStartBootstrapper, error) {
	return bcf.createSovereignEpochStartBootStrapper(epochStartBootstrapArgs)
}

// CreateStorageEpochStartBootstrapper creates a new storage epoch start bootstrapper for sovereign chain operations
func (bcf *sovereignEpochStartBootstrapperFactory) CreateStorageEpochStartBootstrapper(epochStartBootstrapArgs ArgsStorageEpochStartBootstrap) (EpochStartBootstrapper, error) {
	esb, err := bcf.createSovereignEpochStartBootStrapper(epochStartBootstrapArgs.ArgsEpochStartBootstrap)
	if err != nil {
		return nil, err
	}

	epochStartBootstrapArgs.EpochStartBootStrap = esb.epochStartBootstrap
	return NewStorageEpochStartBootstrap(epochStartBootstrapArgs)
}

func (bcf *sovereignEpochStartBootstrapperFactory) createSovereignEpochStartBootStrapper(epochStartBootstrapArgs ArgsEpochStartBootstrap) (*sovereignChainEpochStartBootstrap, error) {
	epochStartBootstrapper, err := NewEpochStartBootstrap(epochStartBootstrapArgs)
	if err != nil {
		return nil, err
	}

	return NewSovereignChainEpochStartBootstrap(epochStartBootstrapper)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bcf *sovereignEpochStartBootstrapperFactory) IsInterfaceNil() bool {
	return bcf == nil
}
