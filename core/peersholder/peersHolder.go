package peersholder

import (
	"sync"
)

type peersHolder struct {
	pubKeysToPeerIDs map[string]string
	sync.RWMutex
}

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder() *peersHolder {
	return &peersHolder{
		pubKeysToPeerIDs: make(map[string]string),
	}
}

// Add will add a pair of public key and peer ID
func (ph *peersHolder) Add(publicKey string, peerID string) {
	ph.Lock()
	ph.pubKeysToPeerIDs[publicKey] = peerID
	ph.Unlock()
}

// GetPeerIDForPublicKey returns the peer ID if it's found for the given public key
func (ph *peersHolder) GetPeerIDForPublicKey(publicKey string) (string, bool) {
	ph.RLock()
	peerID, found := ph.pubKeysToPeerIDs[publicKey]
	ph.RUnlock()

	return peerID, found
}

// GetPublicKeyForPeerID returns the public key if it's found for the given peer id
func (ph *peersHolder) GetPublicKeyForPeerID(peerID string) (string, bool) {
	// TODO: analyze if implementing a O(1) access to get the key based on the value is needed
	ph.RLock()
	for pubKey, pID := range ph.pubKeysToPeerIDs {
		if peerID == pID {
			ph.RUnlock()
			return pubKey, true
		}
	}
	ph.RUnlock()

	return "", false
}

// DeletePublicKey will remove the entry for the given public key, if found
func (ph *peersHolder) DeletePublicKey(pubKey string) bool {
	ph.Lock()
	defer ph.Unlock()
	_, found := ph.pubKeysToPeerIDs[pubKey]
	if !found {
		return false
	}

	delete(ph.pubKeysToPeerIDs, pubKey)

	return true
}

// DeletePeerID will remove the entry for the given peer id, if found
func (ph *peersHolder) DeletePeerID(pID string) bool {
	key, found := ph.GetPublicKeyForPeerID(pID)
	if !found {
		return false
	}

	ph.Lock()
	delete(ph.pubKeysToPeerIDs, key)
	ph.Unlock()

	return true
}

// Len returns the length of the inner map
func (ph *peersHolder) Len() int {
	ph.RLock()
	mapLen := len(ph.pubKeysToPeerIDs)
	ph.RUnlock()

	return mapLen
}

// Clear will delete all the entries from the inner map
func (ph *peersHolder) Clear() {
	ph.Lock()
	ph.pubKeysToPeerIDs = make(map[string]string)
	ph.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ph *peersHolder) IsInterfaceNil() bool {
	return ph == nil
}
