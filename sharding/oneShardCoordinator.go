package sharding

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// OneShardCoordinator creates a shard coordinator object
type OneShardCoordinator struct{}

// CurrentNumberOfShards gets number of shards
func (osc *OneShardCoordinator) CurrentNumberOfShards() uint32 {
	return 1
}

// ComputeShardForAddress gets shard for the given address
func (osc *OneShardCoordinator) ComputeShardForAddress(address state.AddressContainer) uint32 {
	return 0
}

// CurrentShardId gets shard of the current node
func (osc *OneShardCoordinator) CurrentShardId() uint32 {
	return 0
}

// SetCurrentShardId gets shard of the current node
func (osc *OneShardCoordinator) SetCurrentShardId(shardId uint32) {

}

