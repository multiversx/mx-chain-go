package peersholder

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

type peerInfo struct {
	pid     core.PeerID
	shardID uint32
}

type peerIDPosition struct {
	shardID uint32
	index   int
}

type peersHolder struct {
	pubKeysToPeerIDs map[string]*peerInfo
	peersIDsPerShard map[uint32][]core.PeerID
	peerIDs          map[core.PeerID]*peerIDPosition
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
		peerIDs:          make(map[core.PeerID]*peerIDPosition),
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

	if pInfo.shardID == shardID {
		return
	}

	ph.updatePeerInMaps(pubKeyStr, peerID, shardID)
}

// Get will return a map containing the preferred peer IDs, split by shard ID
func (ph *peersHolder) Get() map[uint32][]core.PeerID {
	ph.RLock()
	defer ph.RUnlock()

	return ph.peersIDsPerShard
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

	_, found := ph.peerIDs[peerID]
	if !found {
		return
	}

	shard, index, _ := ph.getShardAndIndexForPeer(peerID)
	ph.removePeerFromMapAtIndex(shard, index)

	delete(ph.peerIDs, peerID)

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

	ph.peerIDs[peerID] = &peerIDPosition{
		shardID: shardID,
		index:   len(ph.peersIDsPerShard[shardID]) - 1,
	}
}

// this function must be called under mutex protection
func (ph *peersHolder) updatePeerInMaps(pubKey string, peerID core.PeerID, shardID uint32) {
	ph.pubKeysToPeerIDs[pubKey].pid = peerID
	ph.pubKeysToPeerIDs[pubKey].shardID = shardID

	index, added := ph.updatePeerInShardedMap(peerID, shardID)
	if added {
		ph.peerIDs[peerID] = &peerIDPosition{
			shardID: shardID,
			index:   index,
		}
	}
}

// this function must be called under mutex protection
func (ph *peersHolder) updatePeerInShardedMap(peerID core.PeerID, newShardID uint32) (int, bool) {
	shardID, index, found := ph.getShardAndIndexForPeer(peerID)
	if found {
		ph.removePeerFromMapAtIndex(shardID, index)
	}

	ph.peersIDsPerShard[newShardID] = append(ph.peersIDsPerShard[newShardID], peerID)

	return len(ph.peersIDsPerShard[newShardID]), true
}

func (ph *peersHolder) removePeerFromMapAtIndex(shardID uint32, index int) {
	ph.peersIDsPerShard[shardID] = append(ph.peersIDsPerShard[shardID][:index], ph.peersIDsPerShard[shardID][index+1:]...)
	if len(ph.peersIDsPerShard[shardID]) == 0 {
		delete(ph.peersIDsPerShard, shardID)
	}
}

// this function must be called under mutex protection
func (ph *peersHolder) getShardAndIndexForPeer(peerID core.PeerID) (uint32, int, bool) {
	peerIDPos, ok := ph.peerIDs[peerID]
	if !ok {
		return 0, 0, false
	}

	return peerIDPos.shardID, peerIDPos.index, true
}

// Clear will delete all the entries from the inner map
func (ph *peersHolder) Clear() {
	ph.Lock()
	ph.pubKeysToPeerIDs = make(map[string]*peerInfo)
	ph.peersIDsPerShard = make(map[uint32][]core.PeerID)
	ph.peerIDs = make(map[core.PeerID]*peerIDPosition)
	ph.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ph *peersHolder) IsInterfaceNil() bool {
	return ph == nil
}
