package networksharding

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const MaxNumPidsPerPk = maxNumPidsPerPk

func (psm *PeerShardMapper) GetPkFromPidPk(pid core.PeerID) []byte {
	pk, ok := psm.peerIdPkCache.Get([]byte(pid))
	if !ok {
		return nil
	}

	return pk.([]byte)
}

func (psm *PeerShardMapper) GetShardIdFromPkShardId(pk []byte) uint32 {
	shard, _ := psm.fallbackPkShardCache.Get(pk)

	return shard.(uint32)
}

func (psm *PeerShardMapper) GetShardIdFromPidShardId(pid core.PeerID) uint32 {
	shard, _ := psm.fallbackPidShardCache.Get([]byte(pid))

	return shard.(uint32)
}

func (psm *PeerShardMapper) GetFromPkPeerId(pk []byte) []core.PeerID {
	objsPidsQueue, found := psm.pkPeerIdCache.Get(pk)
	if !found {
		return nil
	}

	return objsPidsQueue.(*pidQueue).data
}

func (psm *PeerShardMapper) PeerIdPk() storage.Cacher {
	return psm.peerIdPkCache
}

func (psm *PeerShardMapper) PkPeerId() storage.Cacher {
	return psm.pkPeerIdCache
}

func (psm *PeerShardMapper) FallbackPkShard() storage.Cacher {
	return psm.fallbackPkShardCache
}

func (psm *PeerShardMapper) FallbackPidShard() storage.Cacher {
	return psm.fallbackPidShardCache
}

func (psm *PeerShardMapper) Epoch() uint32 {
	psm.mutEpoch.RLock()
	defer psm.mutEpoch.RUnlock()

	return psm.epoch
}
