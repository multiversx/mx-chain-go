package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

type preferredPeersHolder struct {
}

// NewPreferredPeersHolder returns a new instance of preferredPeersHolder
func NewPreferredPeersHolder() *preferredPeersHolder {
	return &preferredPeersHolder{}
}

// PutConnectionAddress does nothing as it is disabled
func (holder *preferredPeersHolder) PutConnectionAddress(_ core.PeerID, _ string) {
}

// PutShardID does nothing as it is disabled
func (holder *preferredPeersHolder) PutShardID(_ core.PeerID, _ uint32) {
}

// Get returns an empty map as it is disabled
func (holder *preferredPeersHolder) Get() map[uint32][]core.PeerID {
	return make(map[uint32][]core.PeerID)
}

// Contains returns false
func (holder *preferredPeersHolder) Contains(_ core.PeerID) bool {
	return false
}

// Remove does nothing as it is disabled
func (holder *preferredPeersHolder) Remove(_ core.PeerID) {
}

// Clear does nothing as it is disabled
func (holder *preferredPeersHolder) Clear() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *preferredPeersHolder) IsInterfaceNil() bool {
	return holder == nil
}
