package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

type disabledPreferredPeersHolder struct {
}

// NewPreferredPeersHolder returns a new instance of disabledPreferredPeersHolder
func NewPreferredPeersHolder() *disabledPreferredPeersHolder {
	return &disabledPreferredPeersHolder{}
}

// Put won't do anything
func (d *disabledPreferredPeersHolder) Put(_ []byte, _ core.PeerID, _ uint32) {
}

// Get will return an empty map
func (d *disabledPreferredPeersHolder) Get() map[uint32][]core.PeerID {
	return make(map[uint32][]core.PeerID)
}

// Contains returns false
func (d *disabledPreferredPeersHolder) Contains(_ core.PeerID) bool {
	return false
}

// Remove won't do anything
func (d *disabledPreferredPeersHolder) Remove(_ core.PeerID) {
}

// Clear won't do anything
func (d *disabledPreferredPeersHolder) Clear() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledPreferredPeersHolder) IsInterfaceNil() bool {
	return d == nil
}
