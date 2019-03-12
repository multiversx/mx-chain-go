package sharding

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// OneShardCoordinator creates a shard coordinator object
type OneShardCoordinator struct{}

// NumberOfShards gets number of shards
func (osc *OneShardCoordinator) NumberOfShards() uint32 {
	return 1
}

// ComputeId gets shard for the given address
func (osc *OneShardCoordinator) ComputeId(address state.AddressContainer) uint32 {
	return 0
}

// SelfId gets shard of the current node
func (osc *OneShardCoordinator) SelfId() uint32 {
	return 0
}

// SetSelfId gets shard of the current node
func (osc *OneShardCoordinator) SetSelfId(shardId uint32) error {
	return nil
}

// SameShard returns weather two addresses belong to the same shard
func (osc *OneShardCoordinator) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

