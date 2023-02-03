package mock

import (
	"fmt"
)

type multipleShardsCoordinatorMock struct {
	ComputeIdCalled func(address []byte) uint32
	numShards       uint32
	CurrentShard    uint32
}

// NewMultiShardsCoordinatorMock -
func NewMultiShardsCoordinatorMock(numShard uint32) *multipleShardsCoordinatorMock {
	return &multipleShardsCoordinatorMock{
		numShards: numShard,
	}
}

// NumberOfShards -
func (scm *multipleShardsCoordinatorMock) NumberOfShards() uint32 {
	return scm.numShards
}

// ComputeId -
func (scm *multipleShardsCoordinatorMock) ComputeId(address []byte) uint32 {
	if scm.ComputeIdCalled != nil {
		return scm.ComputeIdCalled(address)
	}

	return uint32(0)
}

// SelfId -
func (scm *multipleShardsCoordinatorMock) SelfId() uint32 {
	return scm.CurrentShard
}

// SetSelfId -
func (scm *multipleShardsCoordinatorMock) SetSelfId(_ uint32) error {
	return nil
}

// SameShard -
func (scm *multipleShardsCoordinatorMock) SameShard(_, _ []byte) bool {
	return true
}

// SetNoShards -
func (scm *multipleShardsCoordinatorMock) SetNoShards(numShards uint32) {
	scm.numShards = numShards
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// identifier is generated such as the first shard from identifier is always smaller than the last
func (scm *multipleShardsCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == scm.CurrentShard {
		return fmt.Sprintf("_%d", scm.CurrentShard)
	}

	if destShardID < scm.CurrentShard {
		return fmt.Sprintf("_%d_%d", destShardID, scm.CurrentShard)
	}

	return fmt.Sprintf("_%d_%d", scm.CurrentShard, destShardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *multipleShardsCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
