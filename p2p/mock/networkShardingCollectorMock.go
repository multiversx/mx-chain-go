package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type networkShardingCollectorMock struct {
	mutPeerIdPkMap sync.RWMutex
	peerIdPkMap    map[core.PeerID][]byte

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
		peerIdSubType:       make(map[core.PeerID]uint32),
		fallbackPkShardMap:  make(map[string]uint32),
		fallbackPidShardMap: make(map[string]uint32),
	}
}

// UpdatePeerIdPublicKey -
func (nscm *networkShardingCollectorMock) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	nscm.mutPeerIdPkMap.Lock()
	nscm.peerIdPkMap[pid] = pk
	nscm.mutPeerIdPkMap.Unlock()

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

// GetPeerInfo -
func (nscm *networkShardingCollectorMock) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	nscm.mutPeerIdSubType.Lock()
	defer nscm.mutPeerIdSubType.Unlock()

	return core.P2PPeerInfo{
		PeerType:    core.ObserverPeer,
		PeerSubType: core.P2PPeerSubType(nscm.peerIdSubType[pid]),
	}
}

// IsInterfaceNil -
func (nscm *networkShardingCollectorMock) IsInterfaceNil() bool {
	return nscm == nil
}
