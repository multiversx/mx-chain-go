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

// IsInterfaceNil returns true if there is no value under the interface
func (bcf *epochStartBootstrapperFactory) IsInterfaceNil() bool {
	return bcf == nil
}
