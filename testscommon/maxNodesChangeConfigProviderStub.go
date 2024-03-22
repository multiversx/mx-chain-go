package testscommon

import "github.com/multiversx/mx-chain-go/config"

// MaxNodesChangeConfigProviderStub -
type MaxNodesChangeConfigProviderStub struct {
	GetAllNodesConfigCalled     func() []config.MaxNodesChangeConfig
	GetCurrentNodesConfigCalled func() config.MaxNodesChangeConfig
	EpochConfirmedCalled        func(epoch uint32, round uint64)
}

// GetAllNodesConfig -
func (stub *MaxNodesChangeConfigProviderStub) GetAllNodesConfig() []config.MaxNodesChangeConfig {
	if stub.GetAllNodesConfigCalled != nil {
		return stub.GetAllNodesConfigCalled()
	}

	return nil
}

// GetCurrentNodesConfig -
func (stub *MaxNodesChangeConfigProviderStub) GetCurrentNodesConfig() config.MaxNodesChangeConfig {
	if stub.GetCurrentNodesConfigCalled != nil {
		return stub.GetCurrentNodesConfigCalled()
	}

	return config.MaxNodesChangeConfig{}
}

// EpochConfirmed -
func (stub *MaxNodesChangeConfigProviderStub) EpochConfirmed(epoch uint32, round uint64) {
	if stub.EpochConfirmedCalled != nil {
		stub.EpochConfirmedCalled(epoch, round)
	}
}

// IsInterfaceNil -
func (stub *MaxNodesChangeConfigProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
