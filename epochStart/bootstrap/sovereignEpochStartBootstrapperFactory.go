package bootstrap

// sovereignEpochStartBootstrapperFactory defines the epoch start bootstrapper factory for chain run sovereign
type sovereignEpochStartBootstrapperFactory struct {
}

// NewSovereignEpochStartBootstrapperFactory creates a new epoch start bootstrapper factory for chain run sovereign
func NewSovereignEpochStartBootstrapperFactory() EpochStartBootstrapperCreator {
	return &sovereignEpochStartBootstrapperFactory{}
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
