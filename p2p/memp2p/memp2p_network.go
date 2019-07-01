package memp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MemP2PNetwork provides in-memory connectivity for the MemP2PMessenger
// struct. It simulates a network where each peer is connected to all the other
// peers. The peers are connected to the network if they are in the internal
// `peers` map; otherwise, they are disconnected.
type MemP2PNetwork struct {
	mutex           sync.RWMutex
	messageLogMutex sync.RWMutex
	peerIDs         []p2p.PeerID
	peers           map[p2p.PeerID]*MemP2PMessenger
	LogMessages     bool
	Messages        []p2p.MessageP2P
}

// NewMemP2PNetwork constructs a new MemP2PNetwork instance with an empty
// internal map of peers.
func NewMemP2PNetwork() (*MemP2PNetwork, error) {
	var peerIDs []p2p.PeerID
	var messages []p2p.MessageP2P

	network := MemP2PNetwork{
		mutex:           sync.RWMutex{},
		messageLogMutex: sync.RWMutex{},
		peerIDs:         peerIDs,
		peers:           make(map[p2p.PeerID]*MemP2PMessenger),
		LogMessages:     false,
		Messages:        messages,
	}

	return &network, nil
}

// ListAddresses provides the addresses of the known peers.
func (network *MemP2PNetwork) ListAddresses() []string {
	network.mutex.RLock()
	addresses := make([]string, len(network.peerIDs))
	i := 0
	for _, peerID := range network.peerIDs {
		addresses[i] = fmt.Sprintf("/memp2p/%s", peerID)
		i++
	}
	network.mutex.RUnlock()
	return addresses
}

// ListAddressesExceptOne provides the addresses of the known peers, except a specified one.
func (network *MemP2PNetwork) ListAddressesExceptOne(peerIDToExclude p2p.PeerID) []string {
	network.mutex.RLock()
	resultingLength := len(network.peerIDs) - 1
	if resultingLength <= 0 {
		network.mutex.RUnlock()
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
	network.mutex.RUnlock()
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

// PeersExceptOne provides a copy of its internal map of peers, excluding a specific peer.
func (network *MemP2PNetwork) PeersExceptOne(peerIDToExclude p2p.PeerID) map[p2p.PeerID]*MemP2PMessenger {
	network.mutex.RLock()
	peersCopy := make(map[p2p.PeerID]*MemP2PMessenger)
	for _, peerID := range network.peerIDs {
		if peerID == peerIDToExclude {
			continue
		}
		peersCopy[peerID] = network.peers[peerID]
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

// LogMessage adds a message to its internal log of messages.
func (network *MemP2PNetwork) LogMessage(message p2p.MessageP2P) {
	network.messsageLogMutex.RLock()
	network.Messages = append(network.Messages, message)
	network.messsageLogMutex.RUnlock()
}

// IsPeerConnected returns true if the peer represented by the provided ID is
// found in the inner `peers` map of the MemP2PNetwork instance, which
// determines whether it is connected to the network or not.
func (network *MemP2PNetwork) IsPeerConnected(peerID p2p.PeerID) bool {
	network.mutex.RLock()
	_, found := network.peers[peerID]
	network.mutex.RUnlock()
	return found
}
