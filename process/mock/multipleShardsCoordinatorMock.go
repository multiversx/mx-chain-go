package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type multipleShardsCoordinatorMock struct {
	noShards                     uint32
	ComputeShardForAddressCalled func(address state.AddressContainer) uint32
	CurrentShard                 uint32
}

func NewMultipleShardsCoordinatorMock() *multipleShardsCoordinatorMock {
	return &multipleShardsCoordinatorMock{noShards: 1}
}

func (scm *multipleShardsCoordinatorMock) CurrentNumberOfShards() uint32 {
	return scm.noShards
}

func (scm *multipleShardsCoordinatorMock) ComputeShardForAddress(address state.AddressContainer) uint32 {

	return scm.ComputeShardForAddressCalled(address)
}

func (scm *multipleShardsCoordinatorMock) CurrentShardId() uint32 {
	return scm.CurrentShard
}

func (scm *multipleShardsCoordinatorMock) SetCurrentShardId(shardId uint32) error {
	return nil
}
