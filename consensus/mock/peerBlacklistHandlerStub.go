package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

// PeerBlacklistHandlerStub -
type PeerBlacklistHandlerStub struct {
	BlacklistPeerCalled func(peer core.PeerID, duration time.Duration)
}

// BlacklistPeer -
func (stub *PeerBlacklistHandlerStub) BlacklistPeer(peer core.PeerID, duration time.Duration) {
	if stub.BlacklistPeerCalled != nil {
		stub.BlacklistPeerCalled(peer, duration)
	}

	return
}

// IsInterfaceNil -
func (stub *PeerBlacklistHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
