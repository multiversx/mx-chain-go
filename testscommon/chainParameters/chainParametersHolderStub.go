package chainParameters

import (
	"errors"

	"github.com/multiversx/mx-chain-go/config"
)

// ChainParametersHandlerStub -
type ChainParametersHandlerStub struct {
	CurrentChainParametersCalled  func() config.ChainParametersByEpochConfig
	AllChainParametersCalled      func() []config.ChainParametersByEpochConfig
	ChainParametersForEpochCalled func(epoch uint32) (config.ChainParametersByEpochConfig, error)
}

// NewChainParametersHandlerStubWithRealConfig -
func NewChainParametersHandlerStubWithRealConfig(cfg []config.ChainParametersByEpochConfig) *ChainParametersHandlerStub {
	return &ChainParametersHandlerStub{
		ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
			for _, chainParams := range cfg {
				if chainParams.EnableEpoch <= epoch {
					return chainParams, nil
				}
			}

			return config.ChainParametersByEpochConfig{}, errors.New("epoch not found")
		},
	}
}

// CurrentChainParameters -
func (stub *ChainParametersHandlerStub) CurrentChainParameters() config.ChainParametersByEpochConfig {
	if stub.CurrentChainParametersCalled != nil {
		return stub.CurrentChainParametersCalled()
	}

	return config.ChainParametersByEpochConfig{}
}

// AllChainParameters -
func (stub *ChainParametersHandlerStub) AllChainParameters() []config.ChainParametersByEpochConfig {
	if stub.AllChainParametersCalled != nil {
		return stub.AllChainParametersCalled()
	}

	return nil
}

// ChainParametersForEpoch -
func (stub *ChainParametersHandlerStub) ChainParametersForEpoch(epoch uint32) (config.ChainParametersByEpochConfig, error) {
	if stub.ChainParametersForEpochCalled != nil {
		return stub.ChainParametersForEpochCalled(epoch)
	}

	return config.ChainParametersByEpochConfig{}, nil
}

// IsInterfaceNil -
func (stub *ChainParametersHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
