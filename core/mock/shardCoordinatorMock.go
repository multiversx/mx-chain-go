package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ShardCoordinatorMock -
type ShardCoordinatorMock struct {
}

// NumberOfShards -
func (scm ShardCoordinatorMock) NumberOfShards() uint32 {
	panic("implement me")
}

// ComputeId -
func (scm ShardCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	panic("implement me")
}

// SetSelfId -
func (scm ShardCoordinatorMock) SetSelfId(shardId uint32) error {
	panic("implement me")
}

// SelfId -
func (scm ShardCoordinatorMock) SelfId() uint32 {
	return 0
}

// SameShard -
func (scm ShardCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

// CommunicationIdentifier -
func (scm ShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == sharding.MetachainShardId {
		return "_0_META"
	}

	return "_0"
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm ShardCoordinatorMock) IsInterfaceNil() bool {
	if &scm == nil {
		return true
	}
	return false
}
