package testscommon

import "github.com/multiversx/mx-chain-go/config"

// AntifloodConfigsHandlerStub -
type AntifloodConfigsHandlerStub struct {
	GetCurrentConfigCalled func() config.AntifloodConfigByRound
	IsEnabledCalled        func() bool
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

// IsInterfaceNil -
func (stub *AntifloodConfigsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
