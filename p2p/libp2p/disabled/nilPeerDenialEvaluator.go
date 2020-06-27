package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

// NilPeerDenialEvaluator is a mock implementation of PeerDenialEvaluator that does not manage black listed keys
// (all keys [peers] are whitelisted)
type NilPeerDenialEvaluator struct {
}

// IsDenied outputs false (all peers are white listed)
func (npde *NilPeerDenialEvaluator) IsDenied(_ core.PeerID) bool {
	return false
}

// UpsertPeerID returns nil and does nothing
func (npde *NilPeerDenialEvaluator) UpsertPeerID(_ core.PeerID, _ time.Duration) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (npde *NilPeerDenialEvaluator) IsInterfaceNil() bool {
	return npde == nil
}
