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
		psm.peerIdPk.Remove([]byte(pid))
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
		psm.fallbackPkShard.Remove(pkBuff)
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
		psm.fallbackPidShard.Remove([]byte(pid))
		return core.UnknownShardId
	}

	return shard
}

// UpdatePeerIdPublicKey updates the peer ID - public key pair in the corresponding map
// It also uses the intermediate pkPeerId cache that will prevent having thousands of peer ID's with
// the same Elrond PK that will make the node prone to an eclipse attack
func (psm *PeerShardMapper) UpdatePeerIdPublicKey(pid p2p.PeerID, pk []byte) {
	//mutUpdatePeerIdPublicKey is used as to consider this function a critical section
	psm.mutUpdatePeerIdPublicKey.Lock()
	defer psm.mutUpdatePeerIdPublicKey.Unlock()

	psm.removePidAssociation(pid)

	objPidsQueue, found := psm.pkPeerId.Get(pk)
	if !found {
		psm.peerIdPk.Put([]byte(pid), pk)
		pq := newPidQueue()
		pq.push(pid)
		psm.pkPeerId.Put(pk, pq)
		return
	}

	pq, ok := objPidsQueue.(*pidQueue)
	if !ok {
		psm.pkPeerId.Remove(pk)
		return
	}

	idxPid := pq.indexOf(pid)
	if idxPid > -1 {
		pq.promote(idxPid)
		psm.peerIdPk.Put([]byte(pid), pk)
		return
	}

	pq.push(pid)
	for len(pq.data) > maxNumPidsPerPk {
		evictedPid := pq.pop()

		psm.peerIdPk.Remove([]byte(evictedPid))
		psm.fallbackPidShard.Remove([]byte(evictedPid))
	}
	psm.pkPeerId.Put(pk, pq)
	psm.peerIdPk.Put([]byte(pid), pk)
}

func (psm *PeerShardMapper) removePidAssociation(pid p2p.PeerID) {
	oldPk, found := psm.peerIdPk.Get([]byte(pid))
	if !found {
		return
	}

	oldPkBuff, ok := oldPk.([]byte)
	if !ok {
		psm.peerIdPk.Remove([]byte(pid))
		return
	}

	objPidsQueue, found := psm.pkPeerId.Get(oldPkBuff)
	if !found {
		return
	}

	pq, ok := objPidsQueue.(*pidQueue)
	if !ok {
		psm.pkPeerId.Remove(oldPkBuff)
		return
	}

	pq.remove(pid)
	if len(pq.data) == 0 {
		psm.pkPeerId.Remove(oldPkBuff)
		return
	}

	psm.pkPeerId.Put(oldPkBuff, pq)
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
