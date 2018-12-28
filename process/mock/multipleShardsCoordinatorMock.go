package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type multipleShardsCoordinatorMock struct {
	noShards                     uint32
	ComputeShardForAddressCalled func(address state.AddressContainer, addressConverter state.AddressConverter) uint32
	CurrentShard                 uint32
}

func NewMultipleShardsCoordinatorMock() *multipleShardsCoordinatorMock {
	return &multipleShardsCoordinatorMock{noShards: 1}
}

func (scm *multipleShardsCoordinatorMock) NoShards() uint32 {
	return scm.noShards
}

func (scm *multipleShardsCoordinatorMock) SetNoShards(shards uint32) {
	scm.noShards = shards
}

func (scm *multipleShardsCoordinatorMock) ComputeShardForAddress(
	address state.AddressContainer,
	addressConverter state.AddressConverter) uint32 {

	return scm.ComputeShardForAddressCalled(address, addressConverter)
}

func (scm *multipleShardsCoordinatorMock) ShardForCurrentNode() uint32 {
	return scm.CurrentShard
}
