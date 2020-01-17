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

func (bb *blockBalancer) getNumPendingMiniBlocks(shardID uint32) uint32 {
	bb.mutShardNumPendingMiniBlocks.RLock()
	numPendingMiniBlocks := bb.mapShardNumPendingMiniBlocks[shardID]
	bb.mutShardNumPendingMiniBlocks.RUnlock()

	return numPendingMiniBlocks
}

func (bb *blockBalancer) setNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	bb.mutShardNumPendingMiniBlocks.Lock()
	bb.mapShardNumPendingMiniBlocks[shardID] = numPendingMiniBlocks
	bb.mutShardNumPendingMiniBlocks.Unlock()
}

func (bb *blockBalancer) restoreNumPendingMiniBlocksToGenesis() {
	bb.mutShardNumPendingMiniBlocks.Lock()
	bb.mapShardNumPendingMiniBlocks = make(map[uint32]uint32)
	bb.mutShardNumPendingMiniBlocks.Unlock()
}
