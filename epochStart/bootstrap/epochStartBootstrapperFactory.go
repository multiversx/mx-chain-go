package bootstrap

// epochStartBootstrapperFactory defines the epoch start bootstrapper factory for chain run normal
type epochStartBootstrapperFactory struct {
}

// NewEpochStartBootstrapperFactory creates a new epoch start bootstrapper factory for chain run normal
func NewEpochStartBootstrapperFactory() EpochStartBootstrapperCreator {
	return &epochStartBootstrapperFactory{}
}

// CreateEpochStartBootstrapper creates a new epoch start bootstrapper for chain run normal
func (bcf *epochStartBootstrapperFactory) CreateEpochStartBootstrapper(epochStartBootstrapArgs ArgsEpochStartBootstrap) (EpochStartBootstrapper, error) {
	return NewEpochStartBootstrap(epochStartBootstrapArgs)
}

// CreateStorageEpochStartBootstrapper creates a new storage epoch start bootstrapper for normal chain operations.
func (bcf *epochStartBootstrapperFactory) CreateStorageEpochStartBootstrapper(epochStartBootstrapArgs ArgsStorageEpochStartBootstrap) (EpochStartBootstrapper, error) {
	esb, err := NewEpochStartBootstrap(epochStartBootstrapArgs.ArgsEpochStartBootstrap)
	if err != nil {
		return nil, err
	}
	epochStartBootstrapArgs.EpochStartBootStrap = esb
	return NewStorageEpochStartBootstrap(epochStartBootstrapArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bcf *epochStartBootstrapperFactory) IsInterfaceNil() bool {
	return bcf == nil
}
