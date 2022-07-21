package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

// PeerBlacklistHandlerStub -
type PeerBlacklistHandlerStub struct {
	BlacklistPeerCalled          func(peer core.PeerID, duration time.Duration)
	StartSweepingTimeCacheCalled func()
}

// BlacklistPeer -
func (stub *PeerBlacklistHandlerStub) BlacklistPeer(peer core.PeerID, duration time.Duration) {
	if stub.BlacklistPeerCalled != nil {
		stub.BlacklistPeerCalled(peer, duration)
	}

	return
}

// StartSweepingTimeCache -
func (stub *PeerBlacklistHandlerStub) StartSweepingTimeCache() {
	if stub.StartSweepingTimeCacheCalled != nil {
		stub.StartSweepingTimeCacheCalled()
	}
}

// IsInterfaceNil -
func (stub *PeerBlacklistHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
