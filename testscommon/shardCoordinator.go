package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// ShardsCoordinatorMock -
type ShardsCoordinatorMock struct {
	NoShards        uint32
	CurrentShard    uint32
	ComputeIdCalled func(address []byte) uint32
	SelfIDCalled    func() uint32
}

// NewMultiShardsCoordinatorMock -
func NewMultiShardsCoordinatorMock(numShards uint32) *ShardsCoordinatorMock {
	return &ShardsCoordinatorMock{NoShards: numShards}
}

// NumberOfShards -
func (scm *ShardsCoordinatorMock) NumberOfShards() uint32 {
	return scm.NoShards
}

// ComputeId -
func (scm *ShardsCoordinatorMock) ComputeId(address []byte) uint32 {
	if scm.ComputeIdCalled == nil {
		return scm.SelfId()
	}
	return scm.ComputeIdCalled(address)
}

// SelfId -
func (scm *ShardsCoordinatorMock) SelfId() uint32 {
	if scm.SelfIDCalled != nil {
		return scm.SelfIDCalled()
	}

	return scm.CurrentShard
}

// SetSelfId -
func (scm *ShardsCoordinatorMock) SetSelfId(_ uint32) error {
	return nil
}

// SameShard -
func (scm *ShardsCoordinatorMock) SameShard(_, _ []byte) bool {
	return true
}

// SetNoShards -
func (scm *ShardsCoordinatorMock) SetNoShards(noShards uint32) {
	scm.NoShards = noShards
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// identifier is generated such as the first shard from identifier is always smaller than the last
func (scm *ShardsCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	return core.CommunicationIdentifierBetweenShards(scm.CurrentShard, destShardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *ShardsCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
