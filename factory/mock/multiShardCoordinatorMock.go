package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

// MultipleShardsCoordinatorMock -
type MultipleShardsCoordinatorMock struct {
	NoShards        uint32
	CurrentShard    uint32
	ComputeIdCalled func(address []byte) uint32
	SelfIDCalled    func() uint32
}

// NewMultiShardsCoordinatorMock -
func NewMultiShardsCoordinatorMock(nrShard uint32) *MultipleShardsCoordinatorMock {
	return &MultipleShardsCoordinatorMock{NoShards: nrShard}
}

// NumberOfShards -
func (scm *MultipleShardsCoordinatorMock) NumberOfShards() uint32 {
	return scm.NoShards
}

// ComputeId -
func (scm *MultipleShardsCoordinatorMock) ComputeId(address []byte) uint32 {
	if scm.ComputeIdCalled == nil {
		return scm.SelfId()
	}
	return scm.ComputeIdCalled(address)
}

// SelfId -
func (scm *MultipleShardsCoordinatorMock) SelfId() uint32 {
	if scm.SelfIDCalled != nil {
		return scm.SelfIDCalled()
	}

	return scm.CurrentShard
}

// SetSelfId -
func (scm *MultipleShardsCoordinatorMock) SetSelfId(_ uint32) error {
	return nil
}

// SameShard -
func (scm *MultipleShardsCoordinatorMock) SameShard(_, _ []byte) bool {
	return true
}

// SetNoShards -
func (scm *MultipleShardsCoordinatorMock) SetNoShards(noShards uint32) {
	scm.NoShards = noShards
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// identifier is generated such as the first shard from identifier is always smaller than the last
func (scm *MultipleShardsCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == scm.CurrentShard {
		return ShardIdToString(destShardID)
	}

	if destShardID < scm.CurrentShard {
		return ShardIdToString(destShardID) + ShardIdToString(scm.CurrentShard)
	}

	if destShardID == core.AllShardId {
		return ShardIdToString(core.AllShardId)
	}

	return ShardIdToString(scm.CurrentShard) + ShardIdToString(destShardID)
}

// ShardIdToString returns the string according to the shard id
func ShardIdToString(shardId uint32) string {
	if shardId == core.MetachainShardId {
		return "_META"
	}
	if shardId == core.AllShardId {
		return "_ALL"
	}
	return fmt.Sprintf("_%d", shardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *MultipleShardsCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
