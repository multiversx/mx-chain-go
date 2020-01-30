package track

import (
	"sync"
)

type blockBalancer struct {
	mutShardNumPendingMiniBlocks sync.RWMutex
	mapShardNumPendingMiniBlocks map[uint32]uint32
}

// NewBlockBalancer creates a block balancer object which implements blockBalancerHandler interface
func NewBlockBalancer() (*blockBalancer, error) {
	bb := blockBalancer{}
	bb.mapShardNumPendingMiniBlocks = make(map[uint32]uint32)
	return &bb, nil
}

// GetNumPendingMiniBlocks gets the number of pending miniblocks for a given shard
func (bb *blockBalancer) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	bb.mutShardNumPendingMiniBlocks.RLock()
	numPendingMiniBlocks := bb.mapShardNumPendingMiniBlocks[shardID]
	bb.mutShardNumPendingMiniBlocks.RUnlock()

	return numPendingMiniBlocks
}

// SetNumPendingMiniBlocks sets the number of pending miniblocks for a given shard
func (bb *blockBalancer) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	bb.mutShardNumPendingMiniBlocks.Lock()
	bb.mapShardNumPendingMiniBlocks[shardID] = numPendingMiniBlocks
	bb.mutShardNumPendingMiniBlocks.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (bb *blockBalancer) IsInterfaceNil() bool {
	return bb == nil
}
