package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

type pidSig struct {
	pid []byte
	sig []byte
}

type networkShardingCollectorMock struct {
	mutPeerIdPkMap sync.RWMutex
	peerIdPkMap    map[core.PeerID][]byte

	mutFallbackPkShardMap sync.RWMutex
	fallbackPkShardMap    map[string]uint32

	mutFallbackPidShardMap sync.RWMutex
	fallbackPidShardMap    map[string]uint32

	mutPkPidSignature sync.RWMutex
	pkPidSignature    map[string]*pidSig
}

// NewNetworkShardingCollectorMock -
func NewNetworkShardingCollectorMock() *networkShardingCollectorMock {
	return &networkShardingCollectorMock{
		peerIdPkMap:         make(map[core.PeerID][]byte),
		fallbackPkShardMap:  make(map[string]uint32),
		fallbackPidShardMap: make(map[string]uint32),
		pkPidSignature:      make(map[string]*pidSig),
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

// UpdatePublicKeyPIDSignature -
func (nscm *networkShardingCollectorMock) UpdatePublicKeyPIDSignature(pk []byte, pid []byte, sig []byte) {
	nscm.mutPkPidSignature.Lock()
	nscm.pkPidSignature[string(pk)] = &pidSig{pid: pid, sig: sig}
	nscm.mutPkPidSignature.Unlock()
}

// GetPidAndSignatureFromPk -
func (nscm *networkShardingCollectorMock) GetPidAndSignatureFromPk(pk []byte) (pid []byte, signature []byte) {
	nscm.mutPkPidSignature.Lock()
	entry := nscm.pkPidSignature[string(pk)]
	nscm.mutPkPidSignature.Unlock()

	if entry != nil {
		return entry.pid, entry.sig
	}

	return nil, nil
}

// GetPeerInfo -
func (nscm *networkShardingCollectorMock) GetPeerInfo(_ core.PeerID) core.P2PPeerInfo {
	return core.P2PPeerInfo{}
}

// IsInterfaceNil -
func (nscm *networkShardingCollectorMock) IsInterfaceNil() bool {
	return nscm == nil
}
