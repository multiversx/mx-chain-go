package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type ValidatorStatisticsProcessorMock struct {
	SaveInitialStateCalled func(in []*sharding.InitialNode) error
	UpdatePeerStateCalled  func(header data.HeaderHandler) ([]byte, error)
	RevertPeerStateCalled  func(header data.HeaderHandler) error
	IsInterfaceNilCalled   func() bool
}

func (pm *ValidatorStatisticsProcessorMock) SaveInitialState(in []*sharding.InitialNode) error {
	if pm.SaveInitialStateCalled != nil {
		return pm.SaveInitialStateCalled(in)
	}
	return nil
}

func (pm *ValidatorStatisticsProcessorMock) UpdatePeerState(header data.HeaderHandler) ([]byte, error) {
	if pm.UpdatePeerStateCalled != nil {
		return pm.UpdatePeerStateCalled(header)
	}
	return nil, nil
}

func (pm *ValidatorStatisticsProcessorMock) RevertPeerState(header data.HeaderHandler) error {
	if pm.RevertPeerStateCalled != nil {
		return pm.RevertPeerStateCalled(header)
	}
	return nil
}

func (pm *ValidatorStatisticsProcessorMock) IsInterfaceNil() bool {
	if pm.IsInterfaceNilCalled != nil {
		return pm.IsInterfaceNilCalled()
	}
	return false
}
