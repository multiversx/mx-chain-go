package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

// PeerBlackListCacherStub -
type PeerBlackListCacherStub struct {
	AddCalled    func(pid core.PeerID) error
	UpsertCalled func(pid core.PeerID, span time.Duration) error
	HasCalled    func(pid core.PeerID) bool
	SweepCalled  func()
}

// Add -
func (pblhs *PeerBlackListCacherStub) Add(pid core.PeerID) error {
	if pblhs.AddCalled == nil {
		return nil
	}

	return pblhs.AddCalled(pid)
}

// Upsert -
func (pblhs *PeerBlackListCacherStub) Upsert(pid core.PeerID, span time.Duration) error {
	if pblhs.UpsertCalled == nil {
		return nil
	}

	return pblhs.UpsertCalled(pid, span)
}

// Has -
func (pblhs *PeerBlackListCacherStub) Has(pid core.PeerID) bool {
	if pblhs.HasCalled == nil {
		return false
	}

	return pblhs.HasCalled(pid)
}

// Sweep -
func (pblhs *PeerBlackListCacherStub) Sweep() {
	if pblhs.SweepCalled == nil {
		return
	}

	pblhs.SweepCalled()
}

// IsInterfaceNil -
func (pblhs *PeerBlackListCacherStub) IsInterfaceNil() bool {
	return pblhs == nil
}
