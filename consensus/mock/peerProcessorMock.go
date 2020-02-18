package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ValidatorStatisticsProcessorMock -
type ValidatorStatisticsProcessorMock struct {
	LoadInitialStateCalled func(in []*sharding.InitialNode) error
	UpdatePeerStateCalled  func(header, previousHeader data.HeaderHandler) error
	IsInterfaceNilCalled   func() bool
}

// LoadInitialState -
func (pm *ValidatorStatisticsProcessorMock) LoadInitialState(in []*sharding.InitialNode) error {
	if pm.LoadInitialStateCalled != nil {
		return pm.LoadInitialStateCalled(in)
	}
	return nil
}

// UpdatePeerState -
func (pm *ValidatorStatisticsProcessorMock) UpdatePeerState(header, previousHeader data.HeaderHandler) error {
	if pm.UpdatePeerStateCalled != nil {
		return pm.UpdatePeerStateCalled(header, previousHeader)
	}
	return nil
}

// IsInterfaceNil -
func (pm *ValidatorStatisticsProcessorMock) IsInterfaceNil() bool {
	if pm.IsInterfaceNilCalled != nil {
		return pm.IsInterfaceNilCalled()
	}
	return false
}
