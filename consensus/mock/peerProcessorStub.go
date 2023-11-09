package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ValidatorStatisticsProcessorStub -
type ValidatorStatisticsProcessorStub struct {
	LoadInitialStateCalled func(in []*sharding.InitialNode) error
	UpdatePeerStateCalled  func(header, previousHeader data.HeaderHandler) error
	IsInterfaceNilCalled   func() bool
}

// LoadInitialState -
func (pm *ValidatorStatisticsProcessorStub) LoadInitialState(in []*sharding.InitialNode) error {
	if pm.LoadInitialStateCalled != nil {
		return pm.LoadInitialStateCalled(in)
	}
	return nil
}

// UpdatePeerState -
func (pm *ValidatorStatisticsProcessorStub) UpdatePeerState(header, previousHeader data.HeaderHandler) error {
	if pm.UpdatePeerStateCalled != nil {
		return pm.UpdatePeerStateCalled(header, previousHeader)
	}
	return nil
}

// IsInterfaceNil -
func (pm *ValidatorStatisticsProcessorStub) IsInterfaceNil() bool {
	if pm.IsInterfaceNilCalled != nil {
		return pm.IsInterfaceNilCalled()
	}
	return false
}
