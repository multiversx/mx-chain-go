package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
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
		fallbackPkShardMap:  make(map[string]uint32),
		fallbackPidShardMap: make(map[string]uint32),
	}
}

// UpdatePeerIdPublicKey -
func (nscm *networkShardingCollectorMock) UpdatePeerIdPublicKey(pid core.PeerID, pk []byte) {
	nscm.mutPeerIdPkMap.Lock()
	nscm.peerIdPkMap[pid] = pk
	nscm.mutPeerIdPkMap.Unlock()
}

// UpdatePublicKeyShardId -
func (nscm *networkShardingCollectorMock) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nscm.mutFallbackPkShardMap.Lock()
	nscm.fallbackPkShardMap[string(pk)] = shardId
	nscm.mutFallbackPkShardMap.Unlock()
}

// UpdatePeerIdShardId -
func (nscm *networkShardingCollectorMock) UpdatePeerIdShardId(pid core.PeerID, shardId uint32) {
	nscm.mutFallbackPidShardMap.Lock()
	nscm.fallbackPidShardMap[string(pid)] = shardId
	nscm.mutFallbackPidShardMap.Unlock()
}

func (nscm *networkShardingCollectorMock) UpdatePeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	nscm.mutPeerIdSubType.Lock()
	nscm.peerIdSubType[pid] = uint32(peerSubType)
	nscm.mutPeerIdSubType.Unlock()
}

// GetPeerInfo -
func (nscm *networkShardingCollectorMock) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (nscm *networkShardingCollectorMock) IsInterfaceNil() bool {
	return nscm == nil
}
