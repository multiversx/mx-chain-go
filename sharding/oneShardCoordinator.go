package sharding

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// OneShardCoordinator creates a shard coordinator object
type OneShardCoordinator struct{}

// NoShards gets number of shards
func (osc *OneShardCoordinator) NoShards() uint32 {
	return 1
}

// SetNoShards sets number of shards
func (osc *OneShardCoordinator) SetNoShards(uint32) {
}

// ComputeShardForAddress gets shard for the given address
func (osc *OneShardCoordinator) ComputeShardForAddress(address state.AddressContainer, addressConverter state.AddressConverter) uint32 {
	return 0
}

// ShardForCurrentNode gets shard of the current node
func (osc *OneShardCoordinator) ShardForCurrentNode() uint32 {
	return 0
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// for this implementation, it will always return "_0" as there is a single shard
func (osc *OneShardCoordinator) CommunicationIdentifier(destShardID uint32) string {
	return "_0"
}
