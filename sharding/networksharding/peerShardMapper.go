package networksharding

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

const maxNumPidsPerPk = 3

// PeerShardMapper stores the mappings between peer IDs and shard IDs
type PeerShardMapper struct {
	peerIdPk         storage.Cacher
	pkPeerId         storage.Cacher
	fallbackPkShard  storage.Cacher
	fallbackPidShard storage.Cacher

	mutUpdatePeerIdPublicKey sync.Mutex

	nodesCoordinator sharding.NodesCoordinator
	epochHandler     sharding.EpochHandler
}

// NewPeerShardMapper creates a new peerShardMapper instance
func NewPeerShardMapper(
	peerIdPk storage.Cacher,
	fallbackPkShard storage.Cacher,
	fallbackPidShard storage.Cacher,
	nodesCoordinator sharding.NodesCoordinator,
	epochHandler sharding.EpochHandler,
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
	if check.IfNil(epochHandler) {
		return nil, sharding.ErrNilEpochHandler
	}

	pkPeerId, err := lrucache.NewCache(peerIdPk.MaxSize())
	if err != nil {
		return nil, err
	}

	return &PeerShardMapper{
		peerIdPk:         peerIdPk,
		pkPeerId:         pkPeerId,
		fallbackPkShard:  fallbackPkShard,
		fallbackPidShard: fallbackPidShard,
		nodesCoordinator: nodesCoordinator,
		epochHandler:     epochHandler,
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
		return core.UnknownShardId, nil, false
	}

	pkBuff, ok := pkObj.([]byte)
	if !ok {
		return core.UnknownShardId, nil, false
	}

	_, shardId, err := psm.nodesCoordinator.GetValidatorWithPublicKey(pkBuff, psm.epochHandler.Epoch())
	if err != nil {
		return core.UnknownShardId, pkBuff, false
	}

	return shardId, pkBuff, true
}

func (psm *PeerShardMapper) byIDSearchingPkInFallbackCache(pkBuff []byte) (shardId uint32, ok bool) {
	if len(pkBuff) == 0 {
		return core.UnknownShardId, false
	}

	shardObj, ok := psm.fallbackPkShard.Get(pkBuff)
	if !ok {
		return core.UnknownShardId, false
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		return core.UnknownShardId, false
	}

	return shard, true
}

func (psm *PeerShardMapper) byIDSearchingPidInFallbackCache(pid p2p.PeerID) (shardId uint32) {
	shardObj, ok := psm.fallbackPidShard.Get([]byte(pid))
	if !ok {
		return core.UnknownShardId
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		return core.UnknownShardId
	}

	return shard
}

// UpdatePeerIdPublicKey updates the peer ID - public key pair in the corresponding map
func (psm *PeerShardMapper) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	psm.mutUpdatePeerIdPublicKey.Lock()
	defer psm.mutUpdatePeerIdPublicKey.Unlock()

	objPids, found := psm.pkPeerId.Get(pk)
	if !found {
		psm.peerIdPk.HasOrAdd([]byte(pid), pk)
		psm.pkPeerId.HasOrAdd(pk, []p2p.PeerID{pid})
		return
	}

	pids, ok := objPids.([]p2p.PeerID)
	if !ok {
		psm.pkPeerId.Remove(pk)
		return
	}

	found = searchPid(pid, pids)
	if found {
		psm.peerIdPk.HasOrAdd([]byte(pid), pk)
		return
	}

	for len(pids) > maxNumPidsPerPk {
		evictedPid := pids[0]
		pids = pids[1:]

		psm.peerIdPk.Remove([]byte(evictedPid))
		psm.fallbackPidShard.Remove([]byte(evictedPid))
	}
	psm.pkPeerId.Put(pk, pids)
	psm.peerIdPk.HasOrAdd([]byte(pid), pk)
}

func searchPid(pid p2p.PeerID, pids []p2p.PeerID) bool {
	for _, p := range pids {
		if p == pid {
			return true
		}
	}

	return false
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
