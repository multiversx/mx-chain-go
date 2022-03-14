package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

// PeerDenialEvaluator is a disabled implementation of PeerDenialEvaluator that does not manage black listed keys
// (all keys [peers] are whitelisted)
type PeerDenialEvaluator struct {
}

// IsDenied outputs false (all peers are white listed)
func (pde *PeerDenialEvaluator) IsDenied(_ core.PeerID) bool {
	return false
}

// UpsertPeerID returns nil and does nothing
func (pde *PeerDenialEvaluator) UpsertPeerID(_ core.PeerID, _ time.Duration) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pde *PeerDenialEvaluator) IsInterfaceNil() bool {
	return pde == nil
}
