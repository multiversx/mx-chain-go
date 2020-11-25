package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

// PeerBlackListHandlerStub -
type PeerBlackListHandlerStub struct {
	UpsertCalled func(pid core.PeerID, span time.Duration) error
	HasCalled    func(pid core.PeerID) bool
	SweepCalled  func()
}

// Upsert -
func (pblhs *PeerBlackListHandlerStub) Upsert(pid core.PeerID, span time.Duration) error {
	if pblhs.UpsertCalled == nil {
		return nil
	}

	return pblhs.UpsertCalled(pid, span)
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
