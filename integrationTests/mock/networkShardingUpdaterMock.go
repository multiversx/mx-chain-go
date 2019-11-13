package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
)

type networkShardingUpdaterMock struct {
	mutPeerIdPkMap sync.RWMutex
	peerIdPkMap    map[p2p.PeerID][]byte

	mutFallbackPkShardMap sync.RWMutex
	fallbackPkShardMap    map[string]uint32
}

func NewNetworkShardingUpdaterMock() *networkShardingUpdaterMock {
	return &networkShardingUpdaterMock{
		peerIdPkMap:        make(map[p2p.PeerID][]byte),
		fallbackPkShardMap: make(map[string]uint32),
	}
}

func (nsum *networkShardingUpdaterMock) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	nsum.mutPeerIdPkMap.Lock()
	nsum.peerIdPkMap[pid] = pk
	nsum.mutPeerIdPkMap.Unlock()
}

func (nsum *networkShardingUpdaterMock) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nsum.mutFallbackPkShardMap.Lock()
	nsum.fallbackPkShardMap[string(pk)] = shardId
	nsum.mutFallbackPkShardMap.Unlock()
}

func (nsum *networkShardingUpdaterMock) IsInterfaceNil() bool {
	return nsum == nil
}
