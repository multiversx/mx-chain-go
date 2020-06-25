package disabled

import (
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

// IsInterfaceNil returns true if there is no value under the interface
func (npde *NilPeerDenialEvaluator) IsInterfaceNil() bool {
	return npde == nil
}
