package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
)

// ShardCoordinatorMock -
type ShardCoordinatorMock struct {
	ShardID     uint32
	NumOfShards uint32
}

// NumberOfShards -
func (scm ShardCoordinatorMock) NumberOfShards() uint32 {
	if scm.NumOfShards == 0 {
		return 2
	}

	return scm.NumOfShards
}

// ComputeId -
func (scm ShardCoordinatorMock) ComputeId(_ []byte) uint32 {
	panic("implement me")
}

// SetSelfId -
func (scm ShardCoordinatorMock) SetSelfId(_ uint32) error {
	panic("implement me")
}

// SelfId -
func (scm ShardCoordinatorMock) SelfId() uint32 {
	return scm.ShardID
}

// SameShard -
func (scm ShardCoordinatorMock) SameShard(_, _ []byte) bool {
	return true
}

// CommunicationIdentifier -
func (scm ShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == core.MetachainShardId {
		return "_0_META"
	}

	if scm.SelfId() < destShardID {
		return fmt.Sprintf("_%d_%d", scm.SelfId(), destShardID)
	} else if scm.SelfId() > destShardID {
		return fmt.Sprintf("_%d_%d", destShardID, scm.SelfId())
	}

	return fmt.Sprintf("_%d", scm.SelfId())
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm ShardCoordinatorMock) IsInterfaceNil() bool {
	return false
}
