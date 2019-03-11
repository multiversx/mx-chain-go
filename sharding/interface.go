package sharding

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// Sharder defines what a shard state coordinator should hold
type Sharder interface {
	CurrentNumberOfShards() uint32
	ComputeShardForAddress(address state.AddressContainer) uint32
	CurrentShardId() uint32
	SetCurrentShardId(shardId uint32)
}