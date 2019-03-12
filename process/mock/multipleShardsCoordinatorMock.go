package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type multipleShardsCoordinatorMock struct {
	noShards                     uint32
	ComputeIdCalled func(address state.AddressContainer) uint32
	CurrentShard                 uint32
}

func NewMultipleShardsCoordinatorMock() *multipleShardsCoordinatorMock {
	return &multipleShardsCoordinatorMock{noShards: 1}
}

func (scm *multipleShardsCoordinatorMock) NumberOfShards() uint32 {
	return scm.noShards
}

func (scm *multipleShardsCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {

	return scm.ComputeIdCalled(address)
}

func (scm *multipleShardsCoordinatorMock) SelfId() uint32 {
	return scm.CurrentShard
}

func (scm *multipleShardsCoordinatorMock) SetSelfId(shardId uint32) error {
	return nil
}

func (scm *multipleShardsCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}
