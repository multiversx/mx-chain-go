package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// ShardCoordinatorMock -
type ShardCoordinatorMock struct {
	SelfID uint32
}

// NumberOfShards -
func (scm *ShardCoordinatorMock) NumberOfShards() uint32 {
	panic("implement me")
}

// ComputeId -
func (scm *ShardCoordinatorMock) ComputeId(_ []byte) uint32 {
	return 0
}

// SetSelfId -
func (scm *ShardCoordinatorMock) SetSelfId(shardID uint32) {
	scm.SelfID = shardID
}

// SelfId -
func (scm *ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfID
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
