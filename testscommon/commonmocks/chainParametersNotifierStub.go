package commonmocks

import (
	"github.com/multiversx/mx-chain-go/config"
)

// ChainParametersNotifierStub -
type ChainParametersNotifierStub struct {
	ChainParametersChangedCalled       func(chainParameters config.ChainParametersByEpochConfig)
	UpdateCurrentChainParametersCalled func(params config.ChainParametersByEpochConfig)
}

// ChainParametersChanged -
func (c *ChainParametersNotifierStub) ChainParametersChanged(chainParameters config.ChainParametersByEpochConfig) {
	if c.ChainParametersChangedCalled != nil {
		c.ChainParametersChangedCalled(chainParameters)
	}
}

// UpdateCurrentChainParameters -
func (c *ChainParametersNotifierStub) UpdateCurrentChainParameters(params config.ChainParametersByEpochConfig) {
	if c.UpdateCurrentChainParametersCalled != nil {
		c.UpdateCurrentChainParametersCalled(params)
	}
}

func (c *ChainParametersNotifierStub) IsInterfaceNil() bool {
	return c == nil
}
