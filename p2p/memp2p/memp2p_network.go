package memp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MemP2PNetwork struct {
	mutex       sync.RWMutex
	peerIDs     []p2p.PeerID
	peers       map[p2p.PeerID]*MemP2PMessenger
	LogMessages bool
	Messages    []p2p.MessageP2P
}

func NewMemP2PNetwork() *MemP2PNetwork {
	var peerIDs []p2p.PeerID
	var messages []p2p.MessageP2P
	return &MemP2PNetwork{
		mutex:       sync.RWMutex{},
		peerIDs:     peerIDs,
		peers:       make(map[p2p.PeerID]*MemP2PMessenger),
		LogMessages: false,
		Messages:    messages,
	}
}

// ListAddresses provides the addresses of the known peers.
func (network *MemP2PNetwork) ListAddresses() []string {
	network.mutex.Lock()
	addresses := make([]string, len(network.peerIDs))
	i := 0
	for _, peerID := range network.peerIDs {
		addresses[i] = fmt.Sprintf("/memp2p/%s", peerID)
		i++
	}
	network.mutex.Unlock()
	return addresses
}

// ListAddresses provides the addresses of the known peers, except a specified one.
func (network *MemP2PNetwork) ListAddressesExceptOne(peerIDToExclude p2p.PeerID) []string {
	network.mutex.Lock()
	resultingLength := len(network.peerIDs) - 1
	if resultingLength <= 0 {
		network.mutex.Unlock()
		return []string{}
	}
	addresses := make([]string, resultingLength)
	k := 0
	for _, peerID := range network.peerIDs {
		if peerID == peerIDToExclude {
			continue
		}
		addresses[k] = fmt.Sprintf("/memp2p/%s", peerID)
		k++
	}
	network.mutex.Unlock()
	return addresses
}

// Peers provides a copy of its internal map of peers
func (network *MemP2PNetwork) Peers() map[p2p.PeerID]*MemP2PMessenger {
	network.mutex.RLock()
	peersCopy := make(map[p2p.PeerID]*MemP2PMessenger)
	for peerID, peer := range network.peers {
		peersCopy[peerID] = peer
	}
	network.mutex.RUnlock()
	return peersCopy
}

// PeersExcludingMessenger provides a copy of its internal map of peers, excluding a specific peer.
func (network *MemP2PNetwork) PeersExceptOne(peerIDToExclude p2p.PeerID) map[p2p.PeerID]*MemP2PMessenger {
	network.mutex.RLock()
	peersCopy := make(map[p2p.PeerID]*MemP2PMessenger)
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
func (network *MemP2PNetwork) PeerIDs() []p2p.PeerID {
	network.mutex.RLock()
	peerIDsCopy := make([]p2p.PeerID, len(network.peerIDs))
	for i, peer := range network.peerIDs {
		peerIDsCopy[i] = peer
	}
	network.mutex.RUnlock()
	return peerIDsCopy
}

//PeerIDsExceptOne provides a copy of its internal slice of peerIDs, excluding a specific peer.
func (network *MemP2PNetwork) PeerIDsExceptOne(peerIDToExclude p2p.PeerID) []p2p.PeerID {
	network.mutex.RLock()
	resultingLength := len(network.peerIDs) - 1
	if resultingLength <= 0 {
		network.mutex.RUnlock()
		return []p2p.PeerID{}
	}

	peerIDsCopy := make([]p2p.PeerID, resultingLength)
	k := 0
	for _, peerID := range network.peerIDs {
		if peerID == peerIDToExclude {
			continue
		}
		peerIDsCopy[k] = peerID
		k++
	}
	network.mutex.RUnlock()
	return peerIDsCopy
}

// RegisterPeer adds a messenger to the Peers map and its PeerID to the peerIDs
// slice.
func (network *MemP2PNetwork) RegisterPeer(messenger *MemP2PMessenger) {
	network.mutex.RLock()
	network.peerIDs = append(network.peerIDs, messenger.ID())
	network.peers[messenger.ID()] = messenger
	network.mutex.RUnlock()
}

// UnregisterPeer removes a messenger from the Peers map and its PeerID from
// the peerIDs slice.
func (network *MemP2PNetwork) UnregisterPeer(peerID p2p.PeerID) {
	network.mutex.RLock()
	// Delete from the Peers map.
	delete(network.peers, peerID)
	// Remove from the peerIDs slice, maintaining the order of the slice.
	index := -1
	for i, id := range network.peerIDs {
		if id == peerID {
			index = i
		}
	}
	network.peerIDs = append(network.peerIDs[0:index], network.peerIDs[index+1:]...)
	network.mutex.RUnlock()
}

func (network *MemP2PNetwork) LogMessage(message p2p.MessageP2P) {
	network.mutex.RLock()
	network.Messages = append(network.Messages, message)
	network.mutex.RUnlock()
}

func (network *MemP2PNetwork) IsPeerConnected(peerID p2p.PeerID) bool {
	network.mutex.Lock()
	_, found := network.peers[peerID]
	network.mutex.Unlock()
	return found
}
