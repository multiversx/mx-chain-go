package sharding

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type OneShardCoordinator struct{}

func (osc *OneShardCoordinator) NoShards() uint32 {
	return 1
}

func (osc *OneShardCoordinator) SetNoShards(uint32) {
}

func (osc *OneShardCoordinator) ComputeShardForAddress(address state.AddressContainer, addressConverter state.AddressConverter) uint32 {
	return 0
}

func (osc *OneShardCoordinator) ShardForCurrentNode() uint32 {
	return 0
}
