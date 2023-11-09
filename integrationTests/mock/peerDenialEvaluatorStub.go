package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

// PeerDenialEvaluatorStub -
type PeerDenialEvaluatorStub struct {
	IsDeniedCalled    func(pid core.PeerID) bool
	UpsertPeerIDClled func(pid core.PeerID, duration time.Duration) error
}

// UpsertPeerID -
func (pdes *PeerDenialEvaluatorStub) UpsertPeerID(pid core.PeerID, duration time.Duration) error {
	if pdes.UpsertPeerIDClled != nil {
		return pdes.UpsertPeerIDClled(pid, duration)
	}

	return nil
}

// IsDenied -
func (pdes *PeerDenialEvaluatorStub) IsDenied(pid core.PeerID) bool {
	if pdes.IsDeniedCalled != nil {
		return pdes.IsDeniedCalled(pid)
	}

	return false
}

// IsInterfaceNil -
func (pdes *PeerDenialEvaluatorStub) IsInterfaceNil() bool {
	return pdes == nil
}
