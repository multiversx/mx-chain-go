package sharding

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// ShardCoordinator defines what a shard state coordinator should hold
type ShardCoordinator interface {
	NoShards() uint32
	SetNoShards(uint32)
	ComputeShardForAddress(address state.AddressContainer, addressConverter state.AddressConverter) uint32
	ShardForCurrentNode() uint32
	CrossShardIdentifier(crossShardID uint32) string
}
