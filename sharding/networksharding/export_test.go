package networksharding

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/storage"
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

	return objsPidsQueue.(*pidQueue).data
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

// Epoch -
func (psm *PeerShardMapper) Epoch() uint32 {
	psm.mutEpoch.RLock()
	defer psm.mutEpoch.RUnlock()

	return psm.epoch
}

// UpdatePeerIDPublicKey -
func (psm *PeerShardMapper) UpdatePeerIDPublicKey(pid core.PeerID, pk []byte) bool {
	return psm.updatePeerIDPublicKey(pid, pk)
}
