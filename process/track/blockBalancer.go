package track

import (
	"sync"
)

type blockBalancer struct {
	mutShardPendingMiniBlockHeaders sync.RWMutex
	mapShardPendingMiniBlockHeaders map[uint32]uint32
}

// NewBlockBalancer creates a block balancer object which implements blockBalancerHandler interface
func NewBlockBalancer() (*blockBalancer, error) {
	bn := blockBalancer{}
	bn.mapShardPendingMiniBlockHeaders = make(map[uint32]uint32)
	return &bn, nil
}

func (bb *blockBalancer) pendingMiniBlockHeaders(shardID uint32) uint32 {
	bb.mutShardPendingMiniBlockHeaders.RLock()
	nbPendingMiniBlockHeaders := bb.mapShardPendingMiniBlockHeaders[shardID]
	bb.mutShardPendingMiniBlockHeaders.RUnlock()

	return nbPendingMiniBlockHeaders
}

func (bb *blockBalancer) setPendingMiniBlockHeaders(shardID uint32, nbPendingMiniBlockHeaders uint32) {
	bb.mutShardPendingMiniBlockHeaders.Lock()
	bb.mapShardPendingMiniBlockHeaders[shardID] = nbPendingMiniBlockHeaders
	bb.mutShardPendingMiniBlockHeaders.Unlock()
}
