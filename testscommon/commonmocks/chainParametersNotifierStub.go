package commonmocks

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

// ChainParametersNotifierStub -
type ChainParametersNotifierStub struct {
	ChainParametersChangedCalled       func(chainParameters config.ChainParametersByEpochConfig)
	UpdateCurrentChainParametersCalled func(params config.ChainParametersByEpochConfig)
	RegisterNotifyHandlerCalled        func(handler common.ChainParametersSubscriptionHandler)
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

// RegisterNotifyHandler -
func (c *ChainParametersNotifierStub) RegisterNotifyHandler(handler common.ChainParametersSubscriptionHandler) {
	if c.RegisterNotifyHandlerCalled != nil {
		c.RegisterNotifyHandlerCalled(handler)
	}
}

func (c *ChainParametersNotifierStub) IsInterfaceNil() bool {
	return c == nil
}
