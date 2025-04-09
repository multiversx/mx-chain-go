package factory

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("eligibleNodesCache")

type eligibleNodesCache struct {
	// eligibleNodesMap holds a map with key epoch and value
	//		a map with key shard id and value
	//		a map with key eligible node pk and value struct{}
	// this is needed because:
	// 	- a shard node will intercept proof for self shard and meta
	// 	- a meta node will intercept proofs from all shards
	eligibleNodesMap    map[uint32]map[uint32]map[string]struct{}
	mutEligibleNodesMap sync.RWMutex
	peerShardMapper     process.PeerShardMapper
	nodesCoordinator    nodesCoordinator.NodesCoordinator
}

func newEligibleNodesCache(
	peerShardMapper process.PeerShardMapper,
	nodesCoordinator nodesCoordinator.NodesCoordinator,
) (*eligibleNodesCache, error) {
	if check.IfNil(peerShardMapper) {
		return nil, process.ErrNilPeerShardMapper
	}
	if check.IfNil(nodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}

	return &eligibleNodesCache{
		eligibleNodesMap: map[uint32]map[uint32]map[string]struct{}{},
		peerShardMapper:  peerShardMapper,
		nodesCoordinator: nodesCoordinator,
	}, nil
}

// IsPeerEligible returns true if the provided peer is eligible
func (cache *eligibleNodesCache) IsPeerEligible(pid core.PeerID, shard uint32, epoch uint32) bool {
	isPeerEligible, shouldUpdateCache := cache.isPeerEligible(pid, shard, epoch)
	if !shouldUpdateCache {
		return isPeerEligible
	}

	cache.mutEligibleNodesMap.Lock()
	defer cache.mutEligibleNodesMap.Unlock()

	// refresh the cache with data from nodesCoordinator
	eligibleNodesForShardInEpoch, err := cache.refreshCache(epoch, shard)
	if err != nil {
		log.Trace("IsPeerEligible.refreshCache failed", "error", err)
		return false
	}

	// check the eligible list from the new proof's epoch and shard
	peerInfo := cache.peerShardMapper.GetPeerInfo(pid)
	_, isNodeEligible := eligibleNodesForShardInEpoch[string(peerInfo.PkBytes)]

	return isNodeEligible
}

func (cache *eligibleNodesCache) isPeerEligible(pid core.PeerID, shard uint32, epoch uint32) (bool, bool) {
	cache.mutEligibleNodesMap.RLock()
	defer cache.mutEligibleNodesMap.RUnlock()

	eligibleNodesForEpoch, hasEpochCached := cache.eligibleNodesMap[epoch]
	if !hasEpochCached {
		// if the epoch not cached yet, peer is not eligible and cache should be updated from nodesCoordinator
		return false, true
	}

	eligibleNodesForShardInEpoch, hasEpochCachedForShard := eligibleNodesForEpoch[shard]
	if !hasEpochCachedForShard {
		// if the shard is not cached yet for the epoch, peer is not eligible and cache should be updated from nodesCoordinator
		return false, true
	}

	// check the eligible list from the proof's epoch and shard
	peerInfo := cache.peerShardMapper.GetPeerInfo(pid)
	_, isNodeEligible := eligibleNodesForShardInEpoch[string(peerInfo.PkBytes)]

	return isNodeEligible, false

}

func (cache *eligibleNodesCache) refreshCache(epoch uint32, shardID uint32) (map[string]struct{}, error) {
	eligibleList, err := cache.nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(epoch, shardID)
	if err != nil {
		return nil, err
	}

	cachedEpochs := cache.nodesCoordinator.GetCachedEpochs()
	epochsToBeDeleted := make([]uint32, 0)
	for cachedEpoch := range cache.eligibleNodesMap {
		_, isEpochCachedByNodesCoordinator := cachedEpochs[cachedEpoch]
		if isEpochCachedByNodesCoordinator {
			continue
		}

		epochsToBeDeleted = append(epochsToBeDeleted, cachedEpoch)
	}

	for _, epochToDelete := range epochsToBeDeleted {
		delete(cache.eligibleNodesMap, epochToDelete)
	}

	eligibleNodesForShardInEpoch := make(map[string]struct{}, len(eligibleList))
	for _, eligibleNode := range eligibleList {
		eligibleNodesForShardInEpoch[eligibleNode] = struct{}{}
	}

	_, hasEpochCached := cache.eligibleNodesMap[epoch]
	if !hasEpochCached {
		// new epoch
		cache.eligibleNodesMap[epoch] = map[uint32]map[string]struct{}{
			shardID: eligibleNodesForShardInEpoch,
		}

		return eligibleNodesForShardInEpoch, nil
	}

	// new shard
	cache.eligibleNodesMap[epoch][shardID] = eligibleNodesForShardInEpoch

	return eligibleNodesForShardInEpoch, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *eligibleNodesCache) IsInterfaceNil() bool {
	return cache == nil
}
