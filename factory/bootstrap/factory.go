package bootstrap

import (
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/factory"
)

type EpochStartBootstrapperFactory struct {
}

func (bcf *EpochStartBootstrapperFactory) CreateEpochStartBootstrapper(epochStartBootstrapArgs bootstrap.ArgsEpochStartBootstrap) (factory.EpochStartBootstrapper, error) {
	return bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
}
