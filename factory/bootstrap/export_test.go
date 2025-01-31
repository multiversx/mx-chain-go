package bootstrap

import "github.com/multiversx/mx-chain-go/factory"

func (bc *bootstrapComponents) EpochStartBootstrapper() factory.EpochStartBootstrapper {
	return bc.epochStartBootstrapper
}
