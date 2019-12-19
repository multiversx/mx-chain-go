package memp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// Network provides in-memory connectivity for the Messenger
// struct. It simulates a network where each peer is connected to all the other
// peers. The peers are connected to the network if they are in the internal
// `peers` map; otherwise, they are disconnected.
type Network struct {
	mutex sync.RWMutex
	peers map[p2p.PeerID]*Messenger
}

// NewNetwork constructs a new Network instance with an empty
// internal map of peers.
func NewNetwork() *Network {
	network := Network{
		mutex: sync.RWMutex{},
		peers: make(map[p2p.PeerID]*Messenger),
	}

	return &network
}

// ListAddressesExceptOne provides the addresses of the known peers, except a specified one.
func (network *Network) ListAddressesExceptOne(peerIDToExclude p2p.PeerID) []string {
	network.mutex.RLock()
	resultingLength := len(network.peers) - 1
	addresses := make([]string, resultingLength)
	idx := 0
	for _, peer := range network.peers {
		if peer.ID() == peerIDToExclude {
			continue
		}
		addresses[idx] = fmt.Sprintf("/memp2p/%s", peer.ID())
		idx++
	}
	network.mutex.RUnlock()

	return addresses
}

// Peers provides a copy of its internal map of peers
func (network *Network) Peers() map[p2p.PeerID]*Messenger {
	peersCopy := make(map[p2p.PeerID]*Messenger)

	network.mutex.RLock()
	for peerID, peer := range network.peers {
		peersCopy[peerID] = peer
	}
	network.mutex.RUnlock()

	return peersCopy
}

// PeersExceptOne provides a copy of its internal map of peers, excluding a specific peer.
func (network *Network) PeersExceptOne(peerIDToExclude p2p.PeerID) map[p2p.PeerID]*Messenger {
	peersCopy := make(map[p2p.PeerID]*Messenger)

	network.mutex.RLock()
	for peerID, peer := range network.peers {
		if peerID == peerIDToExclude {
			continue
		}
		peersCopy[peerID] = peer
	}
	network.mutex.RUnlock()

	return peersCopy
}

// PeerIDs provides a copy of its internal slice of peerIDs
func (network *Network) PeerIDs() []p2p.PeerID {
	network.mutex.RLock()
	peerIDsCopy := make([]p2p.PeerID, len(network.peers))
	idx := 0
	for peerID := range network.peers {
		peerIDsCopy[idx] = peerID
		idx++
	}
	network.mutex.RUnlock()

	return peerIDsCopy
}

//PeerIDsExceptOne provides a copy of its internal slice of peerIDs, excluding a specific peer.
func (network *Network) PeerIDsExceptOne(peerIDToExclude p2p.PeerID) []p2p.PeerID {
	network.mutex.RLock()
	peerIDsCopy := make([]p2p.PeerID, len(network.peers)-1)
	idx := 0
	for peerID := range network.peers {
		if peerID == peerIDToExclude {
			continue
		}
		peerIDsCopy[idx] = peerID
		idx++
	}
	network.mutex.RUnlock()
	return peerIDsCopy
}

// RegisterPeer adds a messenger to the Peers map and its PeerID to the peerIDs
// slice.
func (network *Network) RegisterPeer(messenger *Messenger) {
	network.mutex.Lock()
	network.peers[messenger.ID()] = messenger
	network.mutex.Unlock()
}

// UnregisterPeer removes a messenger from the Peers map and its PeerID from
// the peerIDs slice.
func (network *Network) UnregisterPeer(peerID p2p.PeerID) {
	network.mutex.Lock()
	delete(network.peers, peerID)
	network.mutex.Unlock()
}

// IsPeerConnected returns true if the peer represented by the provided ID is
// found in the inner `peers` map of the Network instance, which
// determines whether it is connected to the network or not.
func (network *Network) IsPeerConnected(peerID p2p.PeerID) bool {
	network.mutex.RLock()
	_, found := network.peers[peerID]
	network.mutex.RUnlock()
	return found
}
