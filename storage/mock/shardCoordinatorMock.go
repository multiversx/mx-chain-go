package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// ShardCoordinatorMock -
type ShardCoordinatorMock struct {
	SelfShardId uint32
	NumShards   uint32
}

// NewShardCoordinatorMock -
func NewShardCoordinatorMock(selfShardID uint32, numShards uint32) *ShardCoordinatorMock {
	return &ShardCoordinatorMock{
		SelfShardId: selfShardID,
		NumShards:   numShards,
	}
}

// NumberOfShards -
func (scm *ShardCoordinatorMock) NumberOfShards() uint32 {
	return scm.NumShards
}

// ComputeId -
func (scm *ShardCoordinatorMock) ComputeId(_ []byte) uint32 {
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
