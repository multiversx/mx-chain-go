package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

type disabledPreferredPeersHolder struct {
}

// NewPreferredPeersHolder returns a new instance of disabledPreferredPeersHolder
func NewPreferredPeersHolder() *disabledPreferredPeersHolder {
	return &disabledPreferredPeersHolder{}
}

// PutConnectionAddress does nothing as it is disabled
func (d *disabledPreferredPeersHolder) PutConnectionAddress(_ core.PeerID, _ string) {
}

// PutShardID does nothing as it is disabled
func (d *disabledPreferredPeersHolder) PutShardID(_ core.PeerID, _ uint32) {
}

// Get returns an empty map
func (d *disabledPreferredPeersHolder) Get() map[uint32][]core.PeerID {
	return make(map[uint32][]core.PeerID)
}

// Contains returns false
func (d *disabledPreferredPeersHolder) Contains(_ core.PeerID) bool {
	return false
}

// Remove does nothing as it is disabled
func (d *disabledPreferredPeersHolder) Remove(_ core.PeerID) {
}

// Clear does nothing as it is disabled
func (d *disabledPreferredPeersHolder) Clear() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledPreferredPeersHolder) IsInterfaceNil() bool {
	return d == nil
}
