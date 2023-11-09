package networksharding

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const maxNumPidsPerPk = 3
const uint32Size = 4
const defaultShardId = uint32(0)
const indexNotFound = -1

var log = logger.GetOrCreate("sharding/networksharding")
var peerLog = logger.GetOrCreate("sharding/networksharding/peerlog")

var _ p2p.NetworkShardingCollector = (*PeerShardMapper)(nil)
var _ p2p.PeerShardResolver = (*PeerShardMapper)(nil)

// PeerShardMapper stores the mappings between peer IDs and shard IDs
// Both public key and peer id are verified before they are appended in this cache. In time, the current node
// will learn a large majority (or even the whole network) of the nodes that make up the network. The public key provided
// by this map is then fed to the nodes coordinator that will output the shard id in which that public key resides.
// This component also have a reversed lookup map that will ensure that there won't be unlimited peer ids with the
// same public key. This will prevent eclipse attacks.
// The mapping between shard id and public key is done by the nodes coordinator implementation but the fallbackPkShard
// fallback map is only used whenever nodes coordinator has a wrong view about the peers in a shard.
type PeerShardMapper struct {
	peerIdPkCache            storage.Cacher
	pkPeerIdCache            storage.Cacher
	fallbackPkShardCache     storage.Cacher
	fallbackPidShardCache    storage.Cacher
	peerIdSubTypeCache       storage.Cacher
	mutUpdatePeerIdPublicKey sync.RWMutex

	nodesCoordinator     nodesCoordinator.NodesCoordinator
	preferredPeersHolder p2p.PreferredPeersHolderHandler
}

// ArgPeerShardMapper is the initialization structure for the PeerShardMapper implementation
type ArgPeerShardMapper struct {
	PeerIdPkCache         storage.Cacher
	FallbackPkShardCache  storage.Cacher
	FallbackPidShardCache storage.Cacher
	NodesCoordinator      nodesCoordinator.NodesCoordinator
	PreferredPeersHolder  p2p.PreferredPeersHolderHandler
}

// NewPeerShardMapper creates a new peerShardMapper instance
func NewPeerShardMapper(arg ArgPeerShardMapper) (*PeerShardMapper, error) {

	if check.IfNil(arg.NodesCoordinator) {
		return nil, nodesCoordinator.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.PeerIdPkCache) {
		return nil, fmt.Errorf("%w for PeerIdPkCache", nodesCoordinator.ErrNilCacher)
	}
	if check.IfNil(arg.FallbackPkShardCache) {
		return nil, fmt.Errorf("%w for FallbackPkShardCache", nodesCoordinator.ErrNilCacher)
	}
	if check.IfNil(arg.FallbackPidShardCache) {
		return nil, fmt.Errorf("%w for FallbackPidShardCache", nodesCoordinator.ErrNilCacher)
	}
	if check.IfNil(arg.PreferredPeersHolder) {
		return nil, p2p.ErrNilPreferredPeersHolder
	}

	pkPeerId, err := cache.NewLRUCache(arg.PeerIdPkCache.MaxSize())
	if err != nil {
		return nil, err
	}

	peerIdSubTypeCache, err := cache.NewLRUCache(arg.PeerIdPkCache.MaxSize())
	if err != nil {
		return nil, err
	}

	return &PeerShardMapper{
		peerIdPkCache:         arg.PeerIdPkCache,
		pkPeerIdCache:         pkPeerId,
		fallbackPkShardCache:  arg.FallbackPkShardCache,
		fallbackPidShardCache: arg.FallbackPidShardCache,
		peerIdSubTypeCache:    peerIdSubTypeCache,
		nodesCoordinator:      arg.NodesCoordinator,
		preferredPeersHolder:  arg.PreferredPeersHolder,
	}, nil
}

// GetPeerInfo returns the corresponding shard ID of a given peer ID.
// It also returns the type of provided peer
func (psm *PeerShardMapper) GetPeerInfo(pid core.PeerID) core.P2PPeerInfo {
	var pInfo *core.P2PPeerInfo
	var ok bool

	defer func() {
		if pInfo != nil {
			log.Trace("PeerShardMapper.GetPeerInfo",
				"peer type", pInfo.PeerType.String(),
				"peer subtype", pInfo.PeerSubType.String(),
				"pid", p2p.PeerIdToShortString(pid),
				"pk", hex.EncodeToString(pInfo.PkBytes),
				"shardID", pInfo.ShardID,
			)
		}
	}()

	pInfo, ok = psm.getPeerInfoWithNodesCoordinator(pid)
	if ok {
		return *pInfo
	}

	shardId, ok := psm.getShardIDSearchingPkInFallbackCache(pInfo.PkBytes)
	if ok {
		pInfo.PeerType = core.ObserverPeer
		pInfo.ShardID = shardId
		pInfo.PeerSubType = psm.getPeerSubType(pid)

		return *pInfo
	}
	pInfo = psm.getPeerInfoSearchingPidInFallbackCache(pid)

	return *pInfo
}

func (psm *PeerShardMapper) getPeerInfoWithNodesCoordinator(pid core.PeerID) (*core.P2PPeerInfo, bool) {
	pkObj, ok := psm.peerIdPkCache.Get([]byte(pid))
	if !ok {
		return &core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}, false
	}

	pkBuff, ok := pkObj.([]byte)
	if !ok {
		log.Warn("PeerShardMapper.getPeerInfoWithNodesCoordinator: the contained element should have been of type []byte")

		return &core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}, false
	}

	_, shardId, err := psm.nodesCoordinator.GetValidatorWithPublicKey(pkBuff)
	if err != nil {
		return &core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
			PkBytes:  pkBuff,
		}, false
	}

	return &core.P2PPeerInfo{
		PeerType: core.ValidatorPeer,
		ShardID:  shardId,
		PkBytes:  pkBuff,
	}, true
}

func (psm *PeerShardMapper) getShardIDSearchingPkInFallbackCache(pkBuff []byte) (shardId uint32, ok bool) {
	if len(pkBuff) == 0 {
		return defaultShardId, false
	}

	shardObj, ok := psm.fallbackPkShardCache.Get(pkBuff)
	if !ok {
		return defaultShardId, false
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		log.Warn("PeerShardMapper.getShardIDSearchingPkInFallbackCache: the contained element should have been of type uint32")

		return defaultShardId, false
	}

	return shard, true
}

func (psm *PeerShardMapper) getPeerSubType(pid core.PeerID) core.P2PPeerSubType {
	subTypeObj, ok := psm.peerIdSubTypeCache.Get([]byte(pid))
	if !ok {
		return core.RegularPeer
	}

	subType, ok := subTypeObj.(core.P2PPeerSubType)
	if !ok {
		log.Warn("PeerShardMapper.getPeerSubType: the contained element should have been of type core.P2PPeerSubType")
		return core.RegularPeer
	}

	return subType
}

func (psm *PeerShardMapper) getPeerInfoSearchingPidInFallbackCache(pid core.PeerID) *core.P2PPeerInfo {
	shardObj, ok := psm.fallbackPidShardCache.Get([]byte(pid))
	if !ok {
		return &core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}
	}

	shard, ok := shardObj.(uint32)
	if !ok {
		log.Warn("PeerShardMapper.getPeerInfoSearchingPidInFallbackCache: the contained element should have been of type uint32")

		return &core.P2PPeerInfo{
			PeerType: core.UnknownPeer,
			ShardID:  0,
		}
	}

	return &core.P2PPeerInfo{
		PeerType:    core.ObserverPeer,
		PeerSubType: psm.getPeerSubType(pid),
		ShardID:     shard,
	}
}

// UpdatePeerIDPublicKeyPair updates the public key - peer ID pair in the corresponding maps
// It also uses the intermediate pkPeerId cache that will prevent having thousands of peer ID's with
// the same MultiversX PK that will make the node prone to an eclipse attack
func (psm *PeerShardMapper) UpdatePeerIDPublicKeyPair(pid core.PeerID, pk []byte) {
	isNew := psm.updatePeerIDPublicKey(pid, pk)
	if isNew {
		peerLog.Trace("new peer mapping", "pid", pid.Pretty(), "pk", pk)
	}
}

// UpdatePeerIDInfo updates the public keys and the shard ID for the peer ID in the corresponding maps
// It also uses the intermediate pkPeerId cache that will prevent having thousands of peer ID's with
// the same MultiversX PK that will make the node prone to an eclipse attack
func (psm *PeerShardMapper) UpdatePeerIDInfo(pid core.PeerID, pk []byte, shardID uint32) {
	isNew := psm.updatePeerIDPublicKey(pid, pk)
	if isNew {
		peerLog.Trace("new peer mapping", "pid", pid.Pretty(), "pk", pk)
	}

	if shardID == core.AllShardId {
		return
	}
	psm.putPublicKeyShardId(pk, shardID)
	psm.PutPeerIdShardId(pid, shardID)
}

func (psm *PeerShardMapper) putPublicKeyShardId(pk []byte, shardId uint32) {
	psm.fallbackPkShardCache.Put(pk, shardId, uint32Size)
}

// PutPeerIdShardId puts the peer ID and shard ID into fallback cache in case it does not exist
func (psm *PeerShardMapper) PutPeerIdShardId(pid core.PeerID, shardId uint32) {
	psm.fallbackPidShardCache.Put([]byte(pid), shardId, uint32Size)
	psm.preferredPeersHolder.PutShardID(pid, shardId)
}

// updatePeerIDPublicKey will update the pid <-> pk mapping, returning true if the pair is a new known pair
func (psm *PeerShardMapper) updatePeerIDPublicKey(pid core.PeerID, pk []byte) bool {
	// mutUpdatePeerIdPublicKey is used as to consider this function a critical section
	psm.mutUpdatePeerIdPublicKey.Lock()
	defer psm.mutUpdatePeerIdPublicKey.Unlock()

	isNew := !bytes.Equal(pk, psm.removePidAssociation(pid))

	objPidsQueue, found := psm.pkPeerIdCache.Get(pk)
	if !found {
		psm.peerIdPkCache.Put([]byte(pid), pk, len(pk))
		pq := common.NewPidQueue()
		pq.Push(pid)
		psm.pkPeerIdCache.Put(pk, pq, len(pk))

		return isNew
	}

	pq, ok := objPidsQueue.(common.PidQueueHandler)
	if !ok {
		log.Warn("PeerShardMapper.UpdatePeerIdPublicKey: the contained element should have been of type pidQueue")

		return isNew
	}

	idxPid := pq.IndexOf(pid)
	if idxPid != indexNotFound {
		pq.Promote(idxPid)
		psm.peerIdPkCache.Put([]byte(pid), pk, len(pk))

		return isNew
	}

	pq.Push(pid)
	for pq.Len() > maxNumPidsPerPk {
		evictedPid := pq.Pop()

		psm.peerIdPkCache.Remove([]byte(evictedPid))
		psm.fallbackPidShardCache.Remove([]byte(evictedPid))
	}
	psm.pkPeerIdCache.Put(pk, pq, pq.DataSizeInBytes())
	psm.peerIdPkCache.Put([]byte(pid), pk, len(pk))

	return isNew
}

// removePidAssociation removes the pid association between the pid and public key, returning old public key stored, if existing
func (psm *PeerShardMapper) removePidAssociation(pid core.PeerID) []byte {
	oldPk, found := psm.peerIdPkCache.Get([]byte(pid))
	if !found {
		return nil
	}

	oldPkBuff, ok := oldPk.([]byte)
	if !ok {
		psm.peerIdPkCache.Remove([]byte(pid))
		return nil
	}

	objPidsQueue, found := psm.pkPeerIdCache.Get(oldPkBuff)
	if !found {
		return oldPkBuff
	}

	pq, ok := objPidsQueue.(common.PidQueueHandler)
	if !ok {
		psm.pkPeerIdCache.Remove(oldPkBuff)
		return oldPkBuff
	}

	pq.Remove(pid)
	if pq.Len() == 0 {
		psm.pkPeerIdCache.Remove(oldPkBuff)
		return oldPkBuff
	}

	psm.pkPeerIdCache.Put(oldPkBuff, pq, pq.DataSizeInBytes())
	return oldPkBuff
}

// PutPeerIdSubType puts the peerIdSubType search map containing peer IDs and peer subtypes
func (psm *PeerShardMapper) PutPeerIdSubType(pid core.PeerID, peerSubType core.P2PPeerSubType) {
	psm.peerIdSubTypeCache.Put([]byte(pid), peerSubType, uint32Size)
}

// NotifyOrder returns the notification order of this component
func (psm *PeerShardMapper) NotifyOrder() uint32 {
	return common.NetworkShardingOrder
}

// IsInterfaceNil returns true if there is no value under the interface
func (psm *PeerShardMapper) IsInterfaceNil() bool {
	return psm == nil
}
