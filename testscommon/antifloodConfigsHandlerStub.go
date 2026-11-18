package testscommon

import (
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

// AntifloodConfigsHandlerStub -
type AntifloodConfigsHandlerStub struct {
	GetCurrentConfigCalled              func() config.AntifloodConfigByRound
	IsEnabledCalled                     func() bool
	GetFloodPreventerConfigByTypeCalled func(configType common.FloodPreventerType) config.FloodPreventerConfig
	SetActivationRoundCalled            func(round uint64, log logger.Logger)
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

// SetActivationRound -
func (stub *AntifloodConfigsHandlerStub) SetActivationRound(round uint64, log logger.Logger) {
	if stub.SetActivationRoundCalled != nil {
		stub.SetActivationRoundCalled(round, log)
	}
}

// IsInterfaceNil -
func (stub *AntifloodConfigsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
