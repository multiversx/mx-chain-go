package networksharding

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const MaxNumPidsPerPk = maxNumPidsPerPk

func (psm *PeerShardMapper) GetPkFromPidPk(pid p2p.PeerID) []byte {
	pk, ok := psm.peerIdPk.Get([]byte(pid))
	if !ok {
		return nil
	}

	return pk.([]byte)
}

func (psm *PeerShardMapper) GetShardIdFromPkShardId(pk []byte) uint32 {
	shard, _ := psm.fallbackPkShard.Get(pk)

	return shard.(uint32)
}

func (psm *PeerShardMapper) GetShardIdFromPidShardId(pid p2p.PeerID) uint32 {
	shard, _ := psm.fallbackPidShard.Get([]byte(pid))

	return shard.(uint32)
}

func (psm *PeerShardMapper) GetFromPkPeerId(pk []byte) []p2p.PeerID {
	objsPidsQueue, found := psm.pkPeerId.Get(pk)
	if !found {
		return nil
	}

	return objsPidsQueue.(*pidQueue).data
}

func (psm *PeerShardMapper) PeerIdPk() storage.Cacher {
	return psm.peerIdPk
}

func (psm *PeerShardMapper) PkPeerId() storage.Cacher {
	return psm.pkPeerId
}

func (psm *PeerShardMapper) FallbackPkShard() storage.Cacher {
	return psm.fallbackPkShard
}

func (psm *PeerShardMapper) FallbackPidShard() storage.Cacher {
	return psm.fallbackPidShard
}

func (psm *PeerShardMapper) Epoch() uint32 {
	psm.mutEpoch.RLock()
	defer psm.mutEpoch.RUnlock()

	return psm.epoch
}
