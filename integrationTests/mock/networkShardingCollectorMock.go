package mock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

type networkShardingCollectorMock struct {
	mutMaps     sync.RWMutex
	peerIdPkMap map[core.PeerID][]byte
	pkPeerIdMap map[string]core.PeerID

	mutFallbackPkShardMap sync.RWMutex
	fallbackPkShardMap    map[string]uint32

	mutFallbackPidShardMap sync.RWMutex
	fallbackPidShardMap    map[string]uint32

	mutPeerIdSubType sync.RWMutex
	peerIdSubType    map[core.PeerID]uint32
}

// NewNetworkShardingCollectorMock -
func NewNetworkShardingCollectorMock() *networkShardingCollectorMock {
	return &networkShardingCollectorMock{
		peerIdPkMap:         make(map[core.PeerID][]byte),
		pkPeerIdMap:         make(map[string]core.PeerID),
		peerIdSubType:       make(map[core.PeerID]uint32),
		fallbackPkShardMap:  make(map[string]uint32),
		fallbackPidShardMap: make(map[string]uint32),
	}
}

// UpdatePeerIDPublicKeyPair -
func (nscm *networkShardingCollectorMock) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	nscm.mutMaps.Lock()
	nscm.peerIdPkMap[pid] = pk
	nscm.pkPeerIdMap[string(pk)] = pid
	nscm.mutMaps.Unlock()
}

// UpdatePeerIDInfo -
func (nscm *networkShardingCollectorMock) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	nscm.mutMaps.Lock()
	nscm.peerIdPkMap[pid] = pk
	nscm.pkPeerIdMap[string(pk)] = pid
	nscm.mutMaps.Unlock()

	if shardID == core.AllShardId {
		return
	}

	nscm.mutFallbackPkShardMap.Lock()
	nscm.fallbackPkShardMap[string(pk)] = shardID
	nscm.mutFallbackPkShardMap.Unlock()

	nscm.mutFallbackPidShardMap.Lock()
	nscm.fallbackPidShardMap[string(pid)] = shardID
	nscm.mutFallbackPidShardMap.Unlock()
}

// PutPeerIdSubType -
func (nscm *networkShardingCollectorMock) PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	nscm.mutPeerIdSubType.Lock()
	nscm.peerIdSubType[pid] = uint32(peerSubType)
	nscm.mutPeerIdSubType.Unlock()
}

// PutPeerIdShardId -
func (nscm *networkShardingCollectorMock) PutPeerIdShardId(pid core.PeerID, shardID uint32) {
	nscm.mutFallbackPidShardMap.Lock()
	nscm.fallbackPidShardMap[string(pid)] = shardID
	nscm.mutFallbackPidShardMap.Unlock()
}

// GetPeerInfo -
func (nscm *networkShardingCollectorMock) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	nscm.mutPeerIdSubType.Lock()
	defer nscm.mutPeerIdSubType.Unlock()

	return core.P2PPeerInfo{
		PeerType:    core.ObserverPeer,
		PeerSubType: core.P2PPeerSubType(nscm.peerIdSubType[pid]),
		PkBytes:     nscm.peerIdPkMap[pid],
	}
}

// GetLastKnownPeerID -
func (nscm *networkShardingCollectorMock) GetLastKnownPeerID(pk []byte) (core.PeerID, bool) {
	nscm.mutMaps.RLock()
	defer nscm.mutMaps.RUnlock()

	pid, ok := nscm.pkPeerIdMap[string(pk)]

	return pid, ok
}

// IsInterfaceNil -
func (nscm *networkShardingCollectorMock) IsInterfaceNil() bool {
	return nscm == nil
}
