package networkSharding

import "github.com/ElrondNetwork/elrond-go/p2p"

func (psm *peerShardMapper) GetPkFromMap(pid p2p.PeerID) []byte {
	psm.mutPeerIdPkMap.RLock()
	defer psm.mutPeerIdPkMap.RUnlock()

	return psm.peerIdPkMap[pid]
}

func (psm *peerShardMapper) GetShardIdFromMap(pk []byte) uint32 {
	psm.mutFallbackPkShardMap.RLock()
	defer psm.mutFallbackPkShardMap.RUnlock()

	return psm.fallbackPkShardMap[string(pk)]
}
