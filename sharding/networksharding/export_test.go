package networksharding

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

// MaxNumPidsPerPk -
const MaxNumPidsPerPk = maxNumPidsPerPk

// GetPkFromPidPk -
func (psm *PeerShardMapper) GetPkFromPidPk(pid core.PeerID) []byte {
	pk, ok := psm.peerIdPkCache.Get([]byte(pid))
	if !ok {
		return nil
	}

	return pk.([]byte)
}

// GetShardIdFromPkShardId -
func (psm *PeerShardMapper) GetShardIdFromPkShardId(pk []byte) uint32 {
	shard, _ := psm.fallbackPkShardCache.Get(pk)

	return shard.(uint32)
}

// GetShardIdFromPidShardId -
func (psm *PeerShardMapper) GetShardIdFromPidShardId(pid core.PeerID) uint32 {
	shard, _ := psm.fallbackPidShardCache.Get([]byte(pid))

	return shard.(uint32)
}

// GetFromPkPeerId -
func (psm *PeerShardMapper) GetFromPkPeerId(pk []byte) []core.PeerID {
	objsPidsQueue, found := psm.pkPeerIdCache.Get(pk)
	if !found {
		return nil
	}

	pq := objsPidsQueue.(common.PidQueueHandler)
	pidsQueue := make([]core.PeerID, pq.Len())

	for i := 0; i < pq.Len(); i++ {
		pidsQueue[i] = pq.Get(i)
	}

	return pidsQueue
}

// PeerIdPk -
func (psm *PeerShardMapper) PeerIdPk() storage.Cacher {
	return psm.peerIdPkCache
}

// PkPeerId -
func (psm *PeerShardMapper) PkPeerId() storage.Cacher {
	return psm.pkPeerIdCache
}

// FallbackPkShard -
func (psm *PeerShardMapper) FallbackPkShard() storage.Cacher {
	return psm.fallbackPkShardCache
}

// FallbackPidShard -
func (psm *PeerShardMapper) FallbackPidShard() storage.Cacher {
	return psm.fallbackPidShardCache
}

// UpdatePeerIDPublicKey -
func (psm *PeerShardMapper) UpdatePeerIDPublicKey(pid core.PeerID, pk []byte) bool {
	return psm.updatePeerIDPublicKey(pid, pk)
}
