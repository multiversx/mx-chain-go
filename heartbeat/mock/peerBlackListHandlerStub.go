package mock

import "github.com/ElrondNetwork/elrond-go/core"

// PeerBlackListHandlerStub -
type PeerBlackListHandlerStub struct {
	AddCalled   func(pid core.PeerID) error
	HasCalled   func(pid core.PeerID) bool
	SweepCalled func()
}

// Sweep -
func (pblhs *PeerBlackListHandlerStub) Sweep() {
	if pblhs.SweepCalled != nil {
		pblhs.SweepCalled()
	}
}

// Add -
func (pblhs *PeerBlackListHandlerStub) Add(pid core.PeerID) error {
	if pblhs.AddCalled != nil {
		return pblhs.AddCalled(pid)
	}

	return nil
}

// Has -
func (pblhs *PeerBlackListHandlerStub) Has(pid core.PeerID) bool {
	if pblhs.HasCalled != nil {
		return pblhs.HasCalled(pid)
	}

	return false
}

// IsInterfaceNil -
func (pblhs *PeerBlackListHandlerStub) IsInterfaceNil() bool {
	return pblhs == nil
}
