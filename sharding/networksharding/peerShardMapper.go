package networksharding

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

const maxNumPidsPerPk = 3

var log = logger.GetOrCreate("sharding/networksharding")

// PeerShardMapper stores the mappings between peer IDs and shard IDs
// Both public key and peer id are verified before they are appended in this cache. In time, the current node
// will learn a large majority (or even the whole network) of the nodes that make up the network. The public key provided
// by this map is then fed to the nodes coordinator that will output the shard id in which that public key resides.
// This component also have a reversed lookup map that will ensure that there won't be unlimited peer ids with the
// same public key. This will prevent eclipse attacks.
// The mapping between shard id and public key is done by the nodes coordinator implementation but the fallbackPkShard
// fallback map is only used whenever nodes coordinator has a wrong view about the peers in a shard.
type PeerShardMapper struct {
	peerIdPk         storage.Cacher
	pkPeerId         storage.Cacher
	fallbackPkShard  storage.Cacher
	fallbackPidShard storage.Cacher

	mutUpdatePeerIdPublicKey sync.Mutex

	mutEpoch         sync.RWMutex
	epoch            uint32
	nodesCoordinator sharding.NodesCoordinator
}

// NewPeerShardMapper creates a new peerShardMapper instance
func NewPeerShardMapper(
	peerIdPk storage.Cacher,
	fallbackPkShard storage.Cacher,
	fallbackPidShard storage.Cacher,
	nodesCoordinator sharding.NodesCoordinator,
	epochStart uint32,
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
		epoch:            epochStart,
	}, nil
}

// GetPeerInfo returns the corresponding shard ID of a given peer ID.
// It also returns the type of provided peer
func (psm *PeerShardMapper) GetPeerInfo(pid p2p.PeerID) core.P2PPeerInfo {
	pInfo, pk, ok := psm.getPeerInfoWithNodesCoordinator(pid)
	if ok {
		return pInfo
	}

	shardId, ok := psm.getShardIDSearchingPkInFallbackCache(pk)
	if ok {
		return core.P2PPeerInfo{
			PeerType: core.ObserverdPeer,
			ShardID:  shardId,
		}
	}

	return psm.getPeerInfoSearchingPidInFallbackCache(pid)
}

func (psm *PeerShardMapper) getPeerInfoWithNodesCoordinator(pid p2p.PeerID) (core.P2PPeerInfo, []byte, bool) {
	pkObj, ok := psm.peerIdPk.Get([]byte(pid))
	if !ok {
		log.Trace("unknown peer",
			"pid", p2p.PeerIdToShortString(pid),
		)

		return core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}, nil, false
	}

	pkBuff, ok := pkObj.([]byte)
	if !ok {
		log.Warn("PeerShardMapper.getShardIDWithNodesCoordinator: the contained element should have been of type []byte")

		return core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}, nil, false
	}

	psm.mutEpoch.RLock()
	epoch := psm.epoch
	psm.mutEpoch.RUnlock()

	_, shardId, err := psm.nodesCoordinator.GetValidatorWithPublicKey(pkBuff, epoch)
	if err != nil {
		log.Trace("unknown peer",
			"pid", p2p.PeerIdToShortString(pid),
			"pk", pkBuff,
			"error", err,
		)

		return core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}, pkBuff, false
	}

	log.Trace("peer",
		"pid", p2p.PeerIdToShortString(pid),
		"pk", pkBuff,
		"shard ID", shardId,
	)

	return core.P2PPeerInfo{
		PeerType: core.ValidatorPeer,
		ShardID:  shardId,
	}, pkBuff, true
}

func (psm *PeerShardMapper) getShardIDSearchingPkInFallbackCache(pkBuff []byte) (shardId uint32, ok bool) {
	defaultShardId := uint32(0)

	if len(pkBuff) == 0 {
		log.Trace("unknown peer pk is nil")

		return defaultShardId, false
	}

	shardObj, ok := psm.fallbackPkShard.Get(pkBuff)
	if !ok {
		log.Trace("unknown peer",
			"pk", pkBuff,
		)

		return defaultShardId, false
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		log.Warn("PeerShardMapper.getShardIDSearchingPkInFallbackCache: the contained element should have been of type uint32")

		return defaultShardId, false
	}

	log.Trace("peer",
		"pk", pkBuff,
		"shard ID", shardId,
	)

	return shard, true
}

func (psm *PeerShardMapper) getPeerInfoSearchingPidInFallbackCache(pid p2p.PeerID) core.P2PPeerInfo {
	shardObj, ok := psm.fallbackPidShard.Get([]byte(pid))
	if !ok {
		log.Trace("unknown peer",
			"pid", pid,
		)

		return core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		log.Warn("PeerShardMapper.getShardIDSearchingPidInFallbackCache: the contained element should have been of type uint32")

		return core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}
	}

	log.Trace("peer",
		"pid", pid,
		"shard ID", shard,
	)

	return core.P2PPeerInfo{
		PeerType: core.ObserverdPeer,
		ShardID:  shard,
	}
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
		log.Warn("PeerShardMapper.UpdatePeerIdPublicKey: the contained element should have been of type pidQueue")
		return
	}

	idxPid := pq.indexOf(pid)
	if idxPid != indexNotFound {
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

// EpochStartAction is the method called whenever an action needs to be undertaken in respect to the epoch change
func (psm *PeerShardMapper) EpochStartAction(hdr data.HeaderHandler) {
	if check.IfNil(hdr) {
		log.Warn("nil header on PeerShardMapper.EpochStartAction")
		return
	}
	log.Trace("PeerShardMapper.EpochStartAction event", "epoch", hdr.GetEpoch())

	psm.mutEpoch.Lock()
	psm.epoch = hdr.GetEpoch()
	psm.mutEpoch.Unlock()
}

// EpochStartPrepare is the method called whenever an action needs to be undertaken in respect to the epoch preparation change
func (psm *PeerShardMapper) EpochStartPrepare(metaHdr data.HeaderHandler, _ data.BodyHandler) {
	if check.IfNil(metaHdr) {
		log.Warn("nil header on PeerShardMapper.EpochStartPrepare")
		return
	}

	log.Trace("PeerShardMapper.EpochStartPrepare event", "epoch", metaHdr.GetEpoch())
}

// NotifyOrder returns the notification order of this component
func (psm *PeerShardMapper) NotifyOrder() uint32 {
	return core.NetworkShardingOrder
}

// IsInterfaceNil returns true if there is no value under the interface
func (psm *PeerShardMapper) IsInterfaceNil() bool {
	return psm == nil
}
