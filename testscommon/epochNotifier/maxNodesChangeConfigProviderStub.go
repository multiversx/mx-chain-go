package epochNotifier

import "github.com/ElrondNetwork/elrond-go/config"

// NodesConfigProviderMock -
type NodesConfigProviderMock struct {
	AllConfigs    []config.MaxNodesChangeConfig
	CurrentConfig config.MaxNodesChangeConfig
}

// GetAllNodesConfig -
func (ncm *NodesConfigProviderMock) GetAllNodesConfig() []config.MaxNodesChangeConfig {
	return ncm.AllConfigs
}

// GetCurrentNodesConfig -
func (ncm *NodesConfigProviderMock) GetCurrentNodesConfig() config.MaxNodesChangeConfig {
	return ncm.CurrentConfig
}

// EpochConfirmed -
func (ncm *NodesConfigProviderMock) EpochConfirmed(uint32, uint64) {

}

// IsInterfaceNil -
func (ncm *NodesConfigProviderMock) IsInterfaceNil() bool {
	return ncm == nil
}
