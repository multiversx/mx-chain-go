package networkSharding

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// PeerShardMapper stores the mappings between peer IDs and shard IDs
type PeerShardMapper struct {
	peerIdPk         storage.Cacher
	fallbackPkShard  storage.Cacher
	fallbackPidShard storage.Cacher

	nodesCoordinator sharding.NodesCoordinator
}

// NewPeerShardMapper creates a new peerShardMapper instance
func NewPeerShardMapper(
	peerIdPk storage.Cacher,
	fallbackPkShard storage.Cacher,
	fallbackPidShard storage.Cacher,
	nodesCoordinator sharding.NodesCoordinator,
) (*PeerShardMapper, error) {

	if check.IfNil(nodesCoordinator) {
		return nil, sharding.ErrNilNodesCoordinator
	}
	if check.IfNil(peerIdPk) {
		return nil, sharding.ErrNilCacher
	}
	if check.IfNil(fallbackPkShard) {
		return nil, sharding.ErrNilCacher
	}
	if check.IfNil(fallbackPidShard) {
		return nil, sharding.ErrNilCacher
	}

	return &PeerShardMapper{
		peerIdPk:         peerIdPk,
		fallbackPkShard:  fallbackPkShard,
		fallbackPidShard: fallbackPidShard,
		nodesCoordinator: nodesCoordinator,
	}, nil
}

// ByID returns the corresponding shard ID of a given peer ID.
// If the pid is unknown, it will return sharding.UnknownShardId
func (psm *PeerShardMapper) ByID(pid p2p.PeerID) (shardId uint32) {
	shardId, pk, ok := psm.byIDWithNodesCoordinator(pid)
	if ok {
		return shardId
	}

	shardId, ok = psm.byIDSearchingPkInFallbackCache(pk)
	if ok {
		return shardId
	}

	return psm.byIDSearchingPidInFallbackCache(pid)
}

func (psm *PeerShardMapper) byIDWithNodesCoordinator(pid p2p.PeerID) (shardId uint32, pk []byte, ok bool) {
	pkObj, ok := psm.peerIdPk.Get([]byte(pid))
	if !ok {
		return sharding.UnknownShardId, nil, false
	}

	pkBuff, ok := pkObj.([]byte)
	if !ok {
		return sharding.UnknownShardId, nil, false
	}

	_, shardId, err := psm.nodesCoordinator.GetValidatorWithPublicKey(pkBuff)
	if err == nil {
		return shardId, pkBuff, true
	}

	return sharding.UnknownShardId, pkBuff, false
}

func (psm *PeerShardMapper) byIDSearchingPkInFallbackCache(pkBuff []byte) (shardId uint32, ok bool) {
	if len(pkBuff) == 0 {
		return sharding.UnknownShardId, false
	}

	shardObj, ok := psm.fallbackPkShard.Get(pkBuff)
	if !ok {
		return sharding.UnknownShardId, false
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		return sharding.UnknownShardId, false
	}

	return shard, true
}

func (psm *PeerShardMapper) byIDSearchingPidInFallbackCache(pid p2p.PeerID) (shardId uint32) {
	shardObj, ok := psm.peerIdPk.Get([]byte(pid))
	if !ok {
		return sharding.UnknownShardId
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		return sharding.UnknownShardId
	}

	return shard
}

// UpdatePeerIdPublicKey updates the peer ID - public key pair in the corresponding map
func (psm *PeerShardMapper) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	psm.peerIdPk.HasOrAdd([]byte(pid), pk)
}

// UpdatePublicKeyShardId updates the fallback search map containing public key and shard IDs
func (psm *PeerShardMapper) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	psm.fallbackPkShard.HasOrAdd(pk, shardId)
}

// UpdatePeerIdShardId updates the fallback search map containing peer IDs and shard IDs
func (psm *PeerShardMapper) UpdatePeerIdShardId(pid p2p.PeerID, shardId uint32) {
	psm.fallbackPidShard.HasOrAdd([]byte(pid), shardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (psm *PeerShardMapper) IsInterfaceNil() bool {
	return psm == nil
}
