package peersHolder

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/peersHolder/connectionStringValidator"
)

type peerInfo struct {
	pid     core.PeerID
	shardID uint32
}

type peerIDData struct {
	connectionAddress string
	shardID           uint32
	index             int
}

type peersHolder struct {
	preferredConnAddresses     []string
	connAddrToPeersInfo        map[string][]*peerInfo
	tempPeerIDsWaitingForShard map[core.PeerID]string
	peerIDsPerShard            map[uint32][]core.PeerID
	peerIDs                    map[core.PeerID]*peerIDData
	mut                        sync.RWMutex
}

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder(preferredConnectionAddresses []string) (*peersHolder, error) {
	preferredConnections := make([]string, 0)
	connAddrToPeerIDs := make(map[string][]*peerInfo)

	connectionValidator := connectionStringValidator.NewConnectionStringValidator()

	for _, connAddr := range preferredConnectionAddresses {
		if !connectionValidator.IsValid(connAddr) {
			return nil, fmt.Errorf("%w for preferred connection address %s", p2p.ErrInvalidValue, connAddr)
		}

		preferredConnections = append(preferredConnections, connAddr)
		connAddrToPeerIDs[connAddr] = nil
	}

	return &peersHolder{
		preferredConnAddresses:     preferredConnections,
		connAddrToPeersInfo:        connAddrToPeerIDs,
		tempPeerIDsWaitingForShard: make(map[core.PeerID]string),
		peerIDsPerShard:            make(map[uint32][]core.PeerID),
		peerIDs:                    make(map[core.PeerID]*peerIDData),
	}, nil
}

// PutConnectionAddress will perform the insert or the upgrade operation if the provided peerID is inside the preferred peers list
func (ph *peersHolder) PutConnectionAddress(peerID core.PeerID, connectionAddress string) {
	ph.mut.Lock()
	defer ph.mut.Unlock()

	knownConnection := ph.getKnownConnection(connectionAddress)
	if len(knownConnection) == 0 {
		return
	}

	peersInfo := ph.connAddrToPeersInfo[knownConnection]
	if peersInfo == nil {
		ph.addNewPeerInfoToMaps(peerID, knownConnection)
		return
	}

	// if we have new peer for same connection, add it to maps
	pInfo := ph.getPeerInfoForPeerID(peerID, peersInfo)
	if pInfo == nil {
		ph.addNewPeerInfoToMaps(peerID, knownConnection)
	}
}

func (ph *peersHolder) addNewPeerInfoToMaps(peerID core.PeerID, knownConnection string) {
	ph.tempPeerIDsWaitingForShard[peerID] = knownConnection

	newPeerInfo := &peerInfo{
		pid:     peerID,
		shardID: core.AllShardId, // this will be overwritten once shard is available
	}

	ph.connAddrToPeersInfo[knownConnection] = append(ph.connAddrToPeersInfo[knownConnection], newPeerInfo)
}

func (ph *peersHolder) getPeerInfoForPeerID(peerID core.PeerID, peersInfo []*peerInfo) *peerInfo {
	for _, pInfo := range peersInfo {
		if peerID == pInfo.pid {
			return pInfo
		}
	}

	return nil
}

// PutShardID will perform the insert or the upgrade operation if the provided peerID is inside the preferred peers list
func (ph *peersHolder) PutShardID(peerID core.PeerID, shardID uint32) {
	ph.mut.Lock()
	defer ph.mut.Unlock()

	knownConnection, isWaitingForShardID := ph.tempPeerIDsWaitingForShard[peerID]
	if !isWaitingForShardID {
		return
	}

	peersInfo, ok := ph.connAddrToPeersInfo[knownConnection]
	if !ok || peersInfo == nil {
		return
	}

	pInfo := ph.getPeerInfoForPeerID(peerID, peersInfo)
	if pInfo == nil {
		return
	}

	pInfo.shardID = shardID

	ph.peerIDsPerShard[shardID] = append(ph.peerIDsPerShard[shardID], peerID)

	ph.peerIDs[peerID] = &peerIDData{
		connectionAddress: knownConnection,
		shardID:           shardID,
		index:             len(ph.peerIDsPerShard[shardID]) - 1,
	}

	delete(ph.tempPeerIDsWaitingForShard, peerID)
}

// Get will return a map containing the preferred peer IDs, split by shard ID
func (ph *peersHolder) Get() map[uint32][]core.PeerID {
	var peerIDsPerShardCopy map[uint32][]core.PeerID

	ph.mut.RLock()
	peerIDsPerShardCopy = ph.peerIDsPerShard
	ph.mut.RUnlock()

	return peerIDsPerShardCopy
}

// Contains returns true if the provided peer id is a preferred connection
func (ph *peersHolder) Contains(peerID core.PeerID) bool {
	ph.mut.RLock()
	defer ph.mut.RUnlock()

	_, found := ph.peerIDs[peerID]
	return found
}

// Remove will remove the provided peer ID from the inner members
func (ph *peersHolder) Remove(peerID core.PeerID) {
	ph.mut.Lock()
	defer ph.mut.Unlock()

	pidData, found := ph.peerIDs[peerID]
	if !found {
		return
	}

	shard, index, _ := ph.getShardAndIndexForPeer(peerID)
	ph.removePeerFromMapAtIndex(shard, index)

	connAddress := pidData.connectionAddress

	delete(ph.peerIDs, peerID)

	ph.removePeerInfoAtConnectionAddress(peerID, connAddress)

	_, isWaitingForShardID := ph.tempPeerIDsWaitingForShard[peerID]
	if isWaitingForShardID {
		delete(ph.tempPeerIDsWaitingForShard, peerID)
	}
}

// removePeerInfoAtConnectionAddress removes the entry associated with the provided pid from connAddrToPeersInfo map
// it never removes the map key as it may be reused on a further reconnection
func (ph *peersHolder) removePeerInfoAtConnectionAddress(peerID core.PeerID, connAddr string) {
	peersInfo := ph.connAddrToPeersInfo[connAddr]
	if peersInfo == nil {
		return
	}

	var index int
	var pInfo *peerInfo
	for index, pInfo = range peersInfo {
		if peerID == pInfo.pid {
			ph.removePeerFromPeersInfoAtIndex(peersInfo, index, connAddr)
			return
		}
	}

}

func (ph *peersHolder) removePeerFromPeersInfoAtIndex(peersInfo []*peerInfo, index int, connAddr string) {
	peersInfo = append(peersInfo[:index], peersInfo[index+1:]...)
	if len(peersInfo) == 0 {
		peersInfo = nil
	}

	ph.connAddrToPeersInfo[connAddr] = peersInfo
}

// getKnownConnection checks if the connection address string contains any of the initial preferred connection address
// if true, it returns it
// this function must be called under mutex protection
func (ph *peersHolder) getKnownConnection(connectionAddressStr string) string {
	for _, preferredConnAddr := range ph.preferredConnAddresses {
		if strings.Contains(connectionAddressStr, preferredConnAddr) {
			return preferredConnAddr
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
	ph.mut.Lock()
	defer ph.mut.Unlock()

	ph.tempPeerIDsWaitingForShard = make(map[core.PeerID]string)
	ph.peerIDsPerShard = make(map[uint32][]core.PeerID)
	ph.peerIDs = make(map[core.PeerID]*peerIDData)
	ph.connAddrToPeersInfo = make(map[string][]*peerInfo)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ph *peersHolder) IsInterfaceNil() bool {
	return ph == nil
}
