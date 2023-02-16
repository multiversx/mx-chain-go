package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

// PeerBlackListCacherStub -
type PeerBlackListCacherStub struct {
	UpsertCalled func(pid core.PeerID, span time.Duration) error
	HasCalled    func(pid core.PeerID) bool
	SweepCalled  func()
}

// Upsert -
func (stub *PeerBlackListCacherStub) Upsert(pid core.PeerID, span time.Duration) error {
	if stub.UpsertCalled == nil {
		return nil
	}

	return stub.UpsertCalled(pid, span)
}

// Has -
func (stub *PeerBlackListCacherStub) Has(pid core.PeerID) bool {
	if stub.HasCalled == nil {
		return false
	}

	return stub.HasCalled(pid)
}

// Sweep -
func (stub *PeerBlackListCacherStub) Sweep() {
	if stub.SweepCalled == nil {
		return
	}

	stub.SweepCalled()
}

// IsInterfaceNil -
func (stub *PeerBlackListCacherStub) IsInterfaceNil() bool {
	return stub == nil
}
