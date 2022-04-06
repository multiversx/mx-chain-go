package peersHolder

import (
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type peerInfo struct {
	pid     core.PeerID
	shardID uint32
}

type peerIDData struct {
	connectionAddressSlice string
	shardID                uint32
	index                  int
}

type peersHolder struct {
	preferredConnAddrSlices    []string
	connAddrSliceToPeerInfo    map[string]*peerInfo
	tempPeerIDsWaitingForShard map[core.PeerID]string
	peerIDsPerShard            map[uint32][]core.PeerID
	peerIDs                    map[core.PeerID]*peerIDData
	sync.RWMutex
}

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder(preferredConnectionAddressSlices [][]byte) *peersHolder {
	preferredConnections := make([]string, 0)
	connAddrSliceToPeerIDs := make(map[string]*peerInfo)

	for _, connAddrSlice := range preferredConnectionAddressSlices {
		preferredConnections = append(preferredConnections, string(connAddrSlice))
		connAddrSliceToPeerIDs[string(connAddrSlice)] = nil
	}

	return &peersHolder{
		preferredConnAddrSlices:    preferredConnections,
		connAddrSliceToPeerInfo:    connAddrSliceToPeerIDs,
		tempPeerIDsWaitingForShard: make(map[core.PeerID]string),
		peerIDsPerShard:            make(map[uint32][]core.PeerID),
		peerIDs:                    make(map[core.PeerID]*peerIDData),
	}
}

// PutConnectionAddress will perform the insert or the upgrade operation if the provided peerID is inside the preferred peers list
func (ph *peersHolder) PutConnectionAddress(peerID core.PeerID, connectionAddress []byte) {
	ph.Lock()
	defer ph.Unlock()

	knownSlice := ph.getKnownSlice(string(connectionAddress))
	if len(knownSlice) == 0 {
		return
	}

	pInfo := ph.connAddrSliceToPeerInfo[knownSlice]
	if pInfo == nil {
		ph.tempPeerIDsWaitingForShard[peerID] = knownSlice
		ph.connAddrSliceToPeerInfo[knownSlice] = &peerInfo{
			pid:     peerID,
			shardID: 0, // this will be overwritten once shard is available
		}

		return
	}

	isOldData := peerID == pInfo.pid
	if isOldData {
		return
	}

	pInfo.pid = peerID
}

// PutShardID will perform the insert or the upgrade operation if the provided peerID is inside the preferred peers list
func (ph *peersHolder) PutShardID(peerID core.PeerID, shardID uint32) {
	ph.Lock()
	defer ph.Unlock()

	knownSlice, isWaitingForShardID := ph.tempPeerIDsWaitingForShard[peerID]
	if !isWaitingForShardID {
		return
	}

	pInfo, ok := ph.connAddrSliceToPeerInfo[knownSlice]
	if !ok || pInfo == nil {
		return
	}

	pInfo.shardID = shardID

	ph.peerIDsPerShard[shardID] = append(ph.peerIDsPerShard[shardID], peerID)

	ph.peerIDs[peerID] = &peerIDData{
		connectionAddressSlice: knownSlice,
		shardID:                shardID,
		index:                  len(ph.peerIDsPerShard[shardID]) - 1,
	}

	delete(ph.tempPeerIDsWaitingForShard, peerID)
}

// Get will return a map containing the preferred peer IDs, split by shard ID
func (ph *peersHolder) Get() map[uint32][]core.PeerID {
	ph.RLock()
	peerIDsPerShardCopy := ph.peerIDsPerShard
	ph.RUnlock()

	return peerIDsPerShardCopy
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

	pidData, found := ph.peerIDs[peerID]
	if !found {
		return
	}

	shard, index, _ := ph.getShardAndIndexForPeer(peerID)
	ph.removePeerFromMapAtIndex(shard, index)

	connAddrSlice := pidData.connectionAddressSlice

	delete(ph.peerIDs, peerID)

	_, isPreferredPubKey := ph.connAddrSliceToPeerInfo[connAddrSlice]
	if isPreferredPubKey {
		// don't remove the entry because all the keys in this map refer to preferred connections and a reconnection might
		// be done later
		ph.connAddrSliceToPeerInfo[connAddrSlice] = nil
	}

	_, isWaitingForShardID := ph.tempPeerIDsWaitingForShard[peerID]
	if isWaitingForShardID {
		delete(ph.tempPeerIDsWaitingForShard, peerID)
	}
}

// getKnownSlice checks if the connection address contains any of the initial preferred connection address slices
// if true, it returns it
// this function must be called under mutex protection
func (ph *peersHolder) getKnownSlice(connectionAddressStr string) string {
	for _, preferredConnAddrSlice := range ph.preferredConnAddrSlices {
		if strings.Contains(connectionAddressStr, preferredConnAddrSlice) {
			return preferredConnAddrSlice
		}
	}

	return ""
}

// this function must be called under mutex protection
func (ph *peersHolder) removePeerFromMapAtIndex(shardID uint32, index int) {
	ph.peerIDsPerShard[shardID] = append(ph.peerIDsPerShard[shardID][:index], ph.peerIDsPerShard[shardID][index+1:]...)
	if len(ph.peerIDsPerShard[shardID]) == 0 {
		delete(ph.peerIDsPerShard, shardID)
	}
}

// this function must be called under mutex protection
func (ph *peersHolder) getShardAndIndexForPeer(peerID core.PeerID) (uint32, int, bool) {
	pidData, ok := ph.peerIDs[peerID]
	if !ok {
		return 0, 0, false
	}

	return pidData.shardID, pidData.index, true
}

// Clear will delete all the entries from the inner map
func (ph *peersHolder) Clear() {
	ph.Lock()
	defer ph.Unlock()

	ph.tempPeerIDsWaitingForShard = make(map[core.PeerID]string)
	ph.peerIDsPerShard = make(map[uint32][]core.PeerID)
	ph.peerIDs = make(map[core.PeerID]*peerIDData)
	ph.connAddrSliceToPeerInfo = make(map[string]*peerInfo)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ph *peersHolder) IsInterfaceNil() bool {
	return ph == nil
}
