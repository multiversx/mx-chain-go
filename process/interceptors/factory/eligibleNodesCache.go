package factory

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("eligibleNodesCache")

// max number of epochs to keep in the local map
const epochsDelta = uint32(3)

type epochLimits struct {
	min uint32
	max uint32
}

func (limits *epochLimits) isHigherInRange(newEpoch uint32) bool {
	return newEpoch > limits.min && newEpoch < limits.min+epochsDelta
}

func (limits *epochLimits) isHigherOutOfRange(newEpoch uint32) bool {
	return newEpoch >= limits.min+epochsDelta
}

func (limits *epochLimits) isLowerInRange(newEpoch uint32) bool {
	return newEpoch < limits.max && newEpoch > limits.max-epochsDelta && limits.max >= epochsDelta
}

func (limits *epochLimits) isLowerOutOfRange(newEpoch uint32) bool {
	return newEpoch <= limits.max-epochsDelta && limits.max >= epochsDelta
}

type eligibleNodesCache struct {
	// TODO: consider moving eligibleNodesMap and epochsLimitsForShardMap into a new component
	// as suggested here: https://github.com/multiversx/mx-chain-go/pull/6928#discussion_r2031035523
	// eligibleNodesMap holds a map with key shard id and value
	//		a map with key epoch and value
	//		a map with key eligible node pk and value struct{}
	// this is needed because:
	// 	- a shard node will intercept proof for self shard and meta
	// 	- a meta node will intercept proofs from all shards
	eligibleNodesMap map[uint32]map[uint32]map[string]struct{}
	// epochsLimitsForShardMap holds map with key shard id and value the limits for that shard
	epochsLimitsForShardMap map[uint32]*epochLimits
	mutEligibleNodesMap     sync.RWMutex

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
		eligibleNodesMap:        map[uint32]map[uint32]map[string]struct{}{},
		epochsLimitsForShardMap: map[uint32]*epochLimits{},
		peerShardMapper:         peerShardMapper,
		nodesCoordinator:        nodesCoordinator,
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

	// get the eligible list for the new epoch from nodesCoordinator
	eligibleNodesForShardInEpoch, err := cache.getEligibleNodesForShardInEpoch(epoch, shard)
	if err != nil {
		log.Trace("IsPeerEligible.getEligibleNodesForShardInEpoch failed", "error", err)
		return false
	}

	cache.updateMapsIfNeeded(shard, epoch, eligibleNodesForShardInEpoch)

	// check the eligible list from the new proof's epoch and shard
	peerInfo := cache.peerShardMapper.GetPeerInfo(pid)
	_, isNodeEligible := eligibleNodesForShardInEpoch[string(peerInfo.PkBytes)]

	return isNodeEligible
}

func (cache *eligibleNodesCache) updateMapsIfNeeded(
	shard uint32,
	newEpoch uint32,
	eligibleNodesForShardInEpoch map[string]struct{},
) {
	// if the shard is new, create the map and store the received epoch
	limitEpochsForShard, isShardCached := cache.epochsLimitsForShardMap[shard]
	if !isShardCached {
		cache.epochsLimitsForShardMap[shard] = &epochLimits{
			min: newEpoch,
			max: newEpoch,
		}
		cache.eligibleNodesMap[shard] = map[uint32]map[string]struct{}{
			newEpoch: eligibleNodesForShardInEpoch,
		}

		return
	}

	// reaching this point means that the shard is cached, but we have a new epoch
	// there are different situations for the new epoch:
	//	1. the epoch is higher than the min cached epoch + epochs delta ->
	//		store the new + all cached ones that are still in delta limit
	//  2. the epoch is lower than the max cached epoch - epochs delta ->
	//		store the new + all cached ones that are still in delta limit
	// 	3. the new epoch is higher than the min cached but still in the delta limits ->
	//		cache it and update the max cached if needed
	//	4. the new epoch is lower than the max cached but still in the delta limits ->
	//		cache it and update the min cached if needed

	// case 1.
	if limitEpochsForShard.isHigherOutOfRange(newEpoch) {
		cache.handleHigherOutOfRange(newEpoch, shard, eligibleNodesForShardInEpoch)
		return
	}

	// case 2.
	if limitEpochsForShard.isLowerOutOfRange(newEpoch) {
		cache.handleLowerOutOfRange(newEpoch, shard, eligibleNodesForShardInEpoch)
		return
	}

	// case 3.
	if limitEpochsForShard.isHigherInRange(newEpoch) {
		cache.handleHigherInRange(newEpoch, shard, eligibleNodesForShardInEpoch)
		return
	}

	// case 4.
	if limitEpochsForShard.isLowerInRange(newEpoch) {
		cache.handleLowerInRange(newEpoch, shard, eligibleNodesForShardInEpoch)
	}
}

func (cache *eligibleNodesCache) handleHigherOutOfRange(
	newEpoch uint32,
	shard uint32,
	eligibleNodesForShardInEpoch map[string]struct{},
) {
	tmpMin := newEpoch
	for cachedEpoch := range cache.eligibleNodesMap[shard] {
		shouldDeleteEpoch := cachedEpoch <= newEpoch-epochsDelta
		if shouldDeleteEpoch {
			delete(cache.eligibleNodesMap[shard], cachedEpoch)
			continue
		}

		if cachedEpoch < tmpMin {
			tmpMin = cachedEpoch
		}
	}

	// store the limits
	// if nothing was saved from previous cached ones, they would be the new ones
	cache.epochsLimitsForShardMap[shard] = &epochLimits{
		min: tmpMin,
		max: newEpoch,
	}
	cache.eligibleNodesMap[shard][newEpoch] = eligibleNodesForShardInEpoch
}

func (cache *eligibleNodesCache) handleLowerOutOfRange(
	newEpoch uint32,
	shard uint32,
	eligibleNodesForShardInEpoch map[string]struct{},
) {
	tmpMax := newEpoch
	for cachedEpoch := range cache.eligibleNodesMap[shard] {
		shouldDeleteEpoch := cachedEpoch >= newEpoch+epochsDelta
		if shouldDeleteEpoch {
			delete(cache.eligibleNodesMap[shard], cachedEpoch)
			continue
		}

		if cachedEpoch > tmpMax {
			tmpMax = cachedEpoch
		}
	}

	// store the limits
	// if nothing was saved from previous cached ones, they would be the new ones
	cache.epochsLimitsForShardMap[shard] = &epochLimits{
		min: newEpoch,
		max: tmpMax,
	}
	cache.eligibleNodesMap[shard][newEpoch] = eligibleNodesForShardInEpoch
}

func (cache *eligibleNodesCache) handleHigherInRange(
	newEpoch uint32,
	shard uint32,
	eligibleNodesForShardInEpoch map[string]struct{},
) {
	// append the new epoch and store it as the max limit
	cache.eligibleNodesMap[shard][newEpoch] = eligibleNodesForShardInEpoch
	if cache.epochsLimitsForShardMap[shard].max < newEpoch {
		cache.epochsLimitsForShardMap[shard].max = newEpoch
	}
}

func (cache *eligibleNodesCache) handleLowerInRange(
	newEpoch uint32,
	shard uint32,
	eligibleNodesForShardInEpoch map[string]struct{},
) {
	// append the new epoch and store it as the min limit
	cache.eligibleNodesMap[shard][newEpoch] = eligibleNodesForShardInEpoch

	if cache.epochsLimitsForShardMap[shard].min > newEpoch {
		cache.epochsLimitsForShardMap[shard].min = newEpoch
	}
}

func (cache *eligibleNodesCache) isPeerEligible(pid core.PeerID, shard uint32, epoch uint32) (bool, bool) {
	cache.mutEligibleNodesMap.RLock()
	defer cache.mutEligibleNodesMap.RUnlock()

	eligibleNodesForShard, hasShardCached := cache.eligibleNodesMap[shard]
	if !hasShardCached {
		// if the shard is not cached yet, peer is not eligible and cache should be updated from nodesCoordinator
		return false, true
	}

	eligibleNodesForShardInEpoch, hasEpochCachedForShard := eligibleNodesForShard[epoch]
	if !hasEpochCachedForShard {
		// if the epoch is not cached yet for the shard, peer is not eligible and cache should be updated from nodesCoordinator
		return false, true
	}

	// check the eligible list from the proof's epoch and shard
	peerInfo := cache.peerShardMapper.GetPeerInfo(pid)
	_, isNodeEligible := eligibleNodesForShardInEpoch[string(peerInfo.PkBytes)]

	return isNodeEligible, false

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
