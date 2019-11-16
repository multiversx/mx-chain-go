package networkSharding

import "github.com/ElrondNetwork/elrond-go/p2p"

func (psm *PeerShardMapper) GetPkFromPidPk(pid p2p.PeerID) []byte {
	pk, _ := psm.peerIdPk.Get([]byte(pid))

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
