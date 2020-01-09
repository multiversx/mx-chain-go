package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type ShardCoordinatorMock struct {
	SelfShardId uint32
	NumShards   uint32
}

func NewShardCoordinatorMock(selfShardID uint32, numShards uint32) *ShardCoordinatorMock {
	return &ShardCoordinatorMock{
		SelfShardId: selfShardID,
		NumShards:   numShards,
	}
}

func (scm *ShardCoordinatorMock) NumberOfShards() uint32 {
	return scm.NumShards
}

func (scm *ShardCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	return 0
}

func (scm *ShardCoordinatorMock) SetSelfShardId(shardId uint32) error {
	scm.SelfShardId = shardId
	return nil
}

func (scm *ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfShardId
}

func (scm *ShardCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

func (scm *ShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == core.MetachainShardId {
		return "_0_META"
	}

	return "_0"
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *ShardCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
