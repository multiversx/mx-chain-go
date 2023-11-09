package disabled

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.PeerBlackListCacher = (*PeerBlacklistCacher)(nil)

// PeerBlacklistCacher is a mock implementation of PeerBlacklistHandler that does not manage black listed keys
// (all keys [peers] are whitelisted)
type PeerBlacklistCacher struct {
}

// Upsert does nothing
func (pbc *PeerBlacklistCacher) Upsert(_ core.PeerID, _ time.Duration) error {
	return nil
}

// Sweep does nothing
func (pbc *PeerBlacklistCacher) Sweep() {
}

// Has outputs false (all peers are white listed)
func (pbc *PeerBlacklistCacher) Has(_ core.PeerID) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (pbc *PeerBlacklistCacher) IsInterfaceNil() bool {
	return pbc == nil
}
