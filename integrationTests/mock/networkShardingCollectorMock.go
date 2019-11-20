package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

type networkShardingCollectorMock struct {
	mutPeerIdPkMap sync.RWMutex
	peerIdPkMap    map[p2p.PeerID][]byte

	mutFallbackPkShardMap sync.RWMutex
	fallbackPkShardMap    map[string]uint32
}

func NewNetworkShardingCollectorMock() *networkShardingCollectorMock {
	return &networkShardingCollectorMock{
		peerIdPkMap:        make(map[p2p.PeerID][]byte),
		fallbackPkShardMap: make(map[string]uint32),
	}
}

func (nscm *networkShardingCollectorMock) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	nscm.mutPeerIdPkMap.Lock()
	nscm.peerIdPkMap[pid] = pk
	nscm.mutPeerIdPkMap.Unlock()
}

func (nscm *networkShardingCollectorMock) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nscm.mutFallbackPkShardMap.Lock()
	nscm.fallbackPkShardMap[string(pk)] = shardId
	nscm.mutFallbackPkShardMap.Unlock()
}

func (nscm *networkShardingCollectorMock) IsInterfaceNil() bool {
	return nscm == nil
}
