package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type PeerProcessorMock struct {
	LoadInitialStateCalled func(in []*sharding.InitialNode) error
	UpdatePeerStateCalled func(header, perviousHeader data.HeaderHandler) error
	IsInterfaceNilCalled func() bool
}

func (pm *PeerProcessorMock) LoadInitialState(in []*sharding.InitialNode) error {
	if pm.LoadInitialStateCalled != nil {
		return pm.LoadInitialStateCalled(in)
	}
	return nil
}

func (pm *PeerProcessorMock) UpdatePeerState(header, previousHeader data.HeaderHandler) error {
	if pm.UpdatePeerStateCalled != nil {
		return pm.UpdatePeerStateCalled(header, previousHeader)
	}
	return nil
}

func (pm *PeerProcessorMock) IsInterfaceNil() bool {
	if pm.IsInterfaceNilCalled != nil {
		return pm.IsInterfaceNilCalled()
	}
	return false
}
