package testscommon

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

// AntifloodConfigsHandlerStub -
type AntifloodConfigsHandlerStub struct {
	GetCurrentConfigCalled              func() config.AntifloodConfigByRound
	IsEnabledCalled                     func() bool
	GetFloodPreventerConfigByTypeCalled func(configType common.FloodPreventerType) config.FloodPreventerConfig
}

// GetCurrentConfig -
func (stub *AntifloodConfigsHandlerStub) GetCurrentConfig() config.AntifloodConfigByRound {
	if stub.GetCurrentConfigCalled != nil {
		return stub.GetCurrentConfigCalled()
	}

	return GetDefaultAntifloodConfig().ConfigsByRound[0]
}

// IsEnabled -
func (stub *AntifloodConfigsHandlerStub) IsEnabled() bool {
	if stub.IsEnabledCalled != nil {
		return stub.IsEnabledCalled()
	}

	return false
}

// GetFloodPreventerConfigByType -
func (stub *AntifloodConfigsHandlerStub) GetFloodPreventerConfigByType(configType common.FloodPreventerType) config.FloodPreventerConfig {
	if stub.GetFloodPreventerConfigByTypeCalled != nil {
		return stub.GetFloodPreventerConfigByTypeCalled(configType)
	}

	return config.FloodPreventerConfig{}
}

// IsInterfaceNil -
func (stub *AntifloodConfigsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
