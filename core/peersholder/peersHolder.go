package peersholder

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

type peerInfo struct {
	pid     core.PeerID
	shardID uint32
}

type peersHolder struct {
	pubKeysToPeerIDs map[string]*peerInfo
	peersIDsPerShard map[uint32][]core.PeerID
	peerIDs          map[core.PeerID]struct{}
	sync.RWMutex
}

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder(preferredPublicKeys [][]byte) *peersHolder {
	pubKeysToPeerIDs := make(map[string]*peerInfo)

	for _, pubKey := range preferredPublicKeys {
		pubKeysToPeerIDs[string(pubKey)] = nil
	}

	return &peersHolder{
		pubKeysToPeerIDs: pubKeysToPeerIDs,
		peersIDsPerShard: make(map[uint32][]core.PeerID),
		peerIDs:          make(map[core.PeerID]struct{}),
	}
}

// Put will perform the insert or the upgrade operation if the provided public key is inside the preferred peers list
func (ph *peersHolder) Put(pubKey []byte, peerID core.PeerID, shardID uint32) {
	ph.Lock()
	defer ph.Unlock()

	pubKeyStr := string(pubKey)

	pInfo, pubKeyExists := ph.pubKeysToPeerIDs[pubKeyStr]
	if !pubKeyExists {
		// early exit - the public key is not a preferred peer
		return
	}

	if pInfo == nil {
		ph.addPeerInMaps(pubKeyStr, peerID, shardID)
		return
	}

	ph.updatePeerInMaps(pubKeyStr, peerID, shardID)
}

// Get will return a map containing the preferred peer IDs, split by shard ID
func (ph *peersHolder) Get() (map[uint32][]core.PeerID, error) {
	ph.RLock()
	defer ph.RUnlock()

	if len(ph.peersIDsPerShard) == 0 {
		return nil, core.ErrEmptyPreferredPeersList
	}

	return ph.peersIDsPerShard, nil
}

// Contains returns true if the provided peer id is a preferred connection
func (ph *peersHolder) Contains(peerID core.PeerID) bool {
	ph.RLock()
	defer ph.RUnlock()

	_, found := ph.peerIDs[peerID]
	return found
}

// Remove will remove the provided peer ID from the inner members
func (ph *peersHolder) Remove(peerID core.PeerID) {
	ph.Lock()
	defer ph.Unlock()

	delete(ph.peerIDs, peerID)

	shard, index, found := ph.getShardAndIndexForPeer(peerID)
	if found {
		ph.removePeerFromMapAtIndex(shard, index)
	}

	pubKeyForPeerID := ""
	for pubKey, peerInfo := range ph.pubKeysToPeerIDs {
		if peerInfo == nil {
			continue
		}

		if peerInfo.pid == peerID {
			pubKeyForPeerID = pubKey
			break
		}
	}

	if len(pubKeyForPeerID) > 0 {
		// don't remove the entry because all the keys in this map refer to preferred connections and a reconnection might
		// be done later
		ph.pubKeysToPeerIDs[pubKeyForPeerID] = nil
	}
}

// this function must be called under mutex protection
func (ph *peersHolder) addPeerInMaps(pubKey string, peerID core.PeerID, shardID uint32) {
	ph.pubKeysToPeerIDs[pubKey] = &peerInfo{
		pid:     peerID,
		shardID: shardID,
	}

	ph.peersIDsPerShard[shardID] = append(ph.peersIDsPerShard[shardID], peerID)

	ph.peerIDs[peerID] = struct{}{}
}

// this function must be called under mutex protection
func (ph *peersHolder) updatePeerInMaps(pubKey string, peerID core.PeerID, shardID uint32) {
	ph.pubKeysToPeerIDs[pubKey].pid = peerID
	ph.pubKeysToPeerIDs[pubKey].shardID = shardID

	ph.updatePeerInShardedMap(peerID, shardID)

	ph.peerIDs[peerID] = struct{}{}
}

// this function must be called under mutex protection
func (ph *peersHolder) updatePeerInShardedMap(peerID core.PeerID, newShardID uint32) {
	shardID, index, found := ph.getShardAndIndexForPeer(peerID)

	isDifferentShardID := shardID != newShardID
	shouldRemoveOldEntry := found && isDifferentShardID
	shouldAddNewEntry := !found

	if shouldRemoveOldEntry {
		ph.removePeerFromMapAtIndex(shardID, index)
		shouldAddNewEntry = true
	}

	if shouldAddNewEntry {
		ph.peersIDsPerShard[newShardID] = append(ph.peersIDsPerShard[newShardID], peerID)
	}
}

func (ph *peersHolder) removePeerFromMapAtIndex(shardID uint32, index int) {
	ph.peersIDsPerShard[shardID] = append(ph.peersIDsPerShard[shardID][:index], ph.peersIDsPerShard[shardID][index+1:]...)
	if len(ph.peersIDsPerShard[shardID]) == 0 {
		delete(ph.peersIDsPerShard, shardID)
	}
}

// this function must be called under mutex protection
func (ph *peersHolder) getShardAndIndexForPeer(peerID core.PeerID) (uint32, int, bool) {
	for shardID, peerIDsInShard := range ph.peersIDsPerShard {
		for idx, peer := range peerIDsInShard {
			if peer == peerID {
				return shardID, idx, true
			}
		}
	}

	return 0, 0, false
}

// Clear will delete all the entries from the inner map
func (ph *peersHolder) Clear() {
	ph.Lock()
	ph.pubKeysToPeerIDs = make(map[string]*peerInfo)
	ph.peersIDsPerShard = make(map[uint32][]core.PeerID)
	ph.peerIDs = make(map[core.PeerID]struct{})
	ph.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ph *peersHolder) IsInterfaceNil() bool {
	return ph == nil
}
