package factory

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type eligibleNodesCache struct {
	// eligibleNodesMap holds a map with key shard id and value
	//		a map with key epoch and value
	//		a map with key eligible node pk and value struct{}
	// this is needed because:
	// 	- a shard node will intercept proof for self shard and meta
	// 	- a meta node will intercept proofs from all shards
	eligibleNodesMap    map[uint32]map[uint32]map[string]struct{}
	mutEligibleNodesMap sync.RWMutex

	peerShardMapper  process.PeerShardMapper
	nodesCoordinator process.NodesCoordinator
}

func newEligibleNodesCache(
	peerShardMapper process.PeerShardMapper,
	nodesCoordinator process.NodesCoordinator,
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
	cache.mutEligibleNodesMap.Lock()
	defer cache.mutEligibleNodesMap.Unlock()

	var err error
	eligibleNodesForShard, hasShardCached := cache.eligibleNodesMap[shard]
	if !hasShardCached {
		// if no eligible list is cached for the current shard, get it from nodesCoordinator
		eligibleNodesForShardInEpoch, err := cache.getEligibleNodesForShardInEpoch(epoch, shard)
		if err != nil {
			return false
		}

		eligibleNodesForShard = map[uint32]map[string]struct{}{
			epoch: eligibleNodesForShardInEpoch,
		}
		cache.eligibleNodesMap[shard] = eligibleNodesForShard
	}

	eligibleNodesForShardInEpoch, hasEpochCachedForShard := eligibleNodesForShard[epoch]
	if !hasEpochCachedForShard {
		// get the eligible list for the new epoch from nodesCoordinator
		eligibleNodesForShardInEpoch, err = cache.getEligibleNodesForShardInEpoch(epoch, shard)
		if err != nil {
			return false
		}

		// if no eligible list was cached for this epoch, it means that it is a new one
		// it is safe to clean the cached eligible list for other epochs for this shard
		cache.eligibleNodesMap[shard] = map[uint32]map[string]struct{}{
			epoch: eligibleNodesForShardInEpoch,
		}
	}

	// check the eligible list from the proof's epoch and shard
	peerInfo := cache.peerShardMapper.GetPeerInfo(pid)
	_, isNodeEligible := eligibleNodesForShardInEpoch[string(peerInfo.PkBytes)]

	return isNodeEligible
}

func (cache *eligibleNodesCache) getEligibleNodesForShardInEpoch(epoch uint32, shardID uint32) (map[string]struct{}, error) {
	eligibleList, err := cache.nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(epoch, shardID)
	if err != nil {
		return nil, err
	}

	eligibleNodesForShardInEpoch := make(map[string]struct{}, len(eligibleList))
	for _, eligibleNode := range eligibleList {
		eligibleNodesForShardInEpoch[eligibleNode] = struct{}{}
	}

	return eligibleNodesForShardInEpoch, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *eligibleNodesCache) IsInterfaceNil() bool {
	return cache == nil
}
