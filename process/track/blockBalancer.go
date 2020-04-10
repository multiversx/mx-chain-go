package track

import (
	"sync"
)

type blockBalancer struct {
	mutBalancerData              sync.RWMutex
	mapShardNumPendingMiniBlocks map[uint32]uint32
	lastProcessedMetaNonce       map[uint32]uint64
}

// NewBlockBalancer creates a block balancer object which implements blockBalancerHandler interface
func NewBlockBalancer() (*blockBalancer, error) {
	bb := blockBalancer{}
	bb.mapShardNumPendingMiniBlocks = make(map[uint32]uint32)
	bb.lastProcessedMetaNonce = make(map[uint32]uint64)
	return &bb, nil
}

// GetNumPendingMiniBlocks gets the number of pending miniblocks for a given shard
func (bb *blockBalancer) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	bb.mutBalancerData.RLock()
	numPendingMiniBlocks := bb.mapShardNumPendingMiniBlocks[shardID]
	bb.mutBalancerData.RUnlock()

	return numPendingMiniBlocks
}

// SetNumPendingMiniBlocks sets the number of pending miniblocks for a given shard
func (bb *blockBalancer) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	bb.mutBalancerData.Lock()
	bb.mapShardNumPendingMiniBlocks[shardID] = numPendingMiniBlocks
	bb.mutBalancerData.Unlock()
}

// GetLastShardProcessedMetaNonce returns the last processed meta nonce for given shard
func (bb *blockBalancer) GetLastShardProcessedMetaNonce(shardID uint32) uint64 {
	bb.mutBalancerData.RLock()
	lastProcessedMetaNonce := bb.lastProcessedMetaNonce[shardID]
	bb.mutBalancerData.RUnlock()

	return lastProcessedMetaNonce
}

// SetLastShardProcessedMetaNonce sets the last processed meta nonce for given shard
func (bb *blockBalancer) SetLastShardProcessedMetaNonce(shardID uint32, nonce uint64) {
	bb.mutBalancerData.Lock()
	bb.lastProcessedMetaNonce[shardID] = nonce
	bb.mutBalancerData.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (bb *blockBalancer) IsInterfaceNil() bool {
	return bb == nil
}
