package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// ShardCoordinatorMock -
type ShardCoordinatorMock struct {
	SelfShardId     uint32
	ComputeIdCalled func([]byte) uint32
}

// NumberOfShards -
func (scm *ShardCoordinatorMock) NumberOfShards() uint32 {
	return 1
}

// ComputeId -
func (scm *ShardCoordinatorMock) ComputeId(addr []byte) uint32 {
	if scm.ComputeIdCalled != nil {
		return scm.ComputeIdCalled(addr)
	}
	return 0
}

// SetSelfShardId -
func (scm *ShardCoordinatorMock) SetSelfShardId(shardId uint32) error {
	scm.SelfShardId = shardId
	return nil
}

// SelfId -
func (scm *ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfShardId
}

// SameShard -
func (scm *ShardCoordinatorMock) SameShard(_, _ []byte) bool {
	return true
}

// CommunicationIdentifier -
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
