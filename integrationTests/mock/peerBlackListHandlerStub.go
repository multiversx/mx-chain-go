package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

// PeerBlackListHandlerStub -
type PeerBlackListHandlerStub struct {
	AddCalled         func(pid core.PeerID) error
	AddWithSpanCalled func(pid core.PeerID, span time.Duration) error
	UpdateCalled      func(pid core.PeerID, span time.Duration) error
	HasCalled         func(pid core.PeerID) bool
	SweepCalled       func()
}

// Add -
func (pblhs *PeerBlackListHandlerStub) Add(pid core.PeerID) error {
	if pblhs.AddCalled == nil {
		return nil
	}

	return pblhs.AddCalled(pid)
}

// AddWithSpan -
func (pblhs *PeerBlackListHandlerStub) AddWithSpan(pid core.PeerID, span time.Duration) error {
	if pblhs.AddWithSpanCalled == nil {
		return nil
	}

	return pblhs.AddWithSpanCalled(pid, span)
}

// Update -
func (pblhs *PeerBlackListHandlerStub) Update(pid core.PeerID, span time.Duration) error {
	if pblhs.UpdateCalled == nil {
		return nil
	}

	return pblhs.UpdateCalled(pid, span)
}

// Has -
func (pblhs *PeerBlackListHandlerStub) Has(pid core.PeerID) bool {
	if pblhs.HasCalled == nil {
		return false
	}

	return pblhs.HasCalled(pid)
}

// Sweep -
func (pblhs *PeerBlackListHandlerStub) Sweep() {
	if pblhs.SweepCalled == nil {
		return
	}

	pblhs.SweepCalled()
}

// IsInterfaceNil -
func (pblhs *PeerBlackListHandlerStub) IsInterfaceNil() bool {
	return pblhs == nil
}
