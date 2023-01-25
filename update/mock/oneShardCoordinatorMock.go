package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

type oneShardCoordinatorMock struct {
	shardID         uint32
	noShards        uint32
	ComputeIdCalled func(address []byte) uint32
}

// NewOneShardCoordinatorMock -
func NewOneShardCoordinatorMock() *oneShardCoordinatorMock {
	return &oneShardCoordinatorMock{noShards: 1}
}

// NumberOfShards -
func (scm *oneShardCoordinatorMock) NumberOfShards() uint32 {
	return scm.noShards
}

// ComputeId -
func (scm *oneShardCoordinatorMock) ComputeId(address []byte) uint32 {
	if scm.ComputeIdCalled != nil {
		return scm.ComputeIdCalled(address)
	}

	return uint32(0)
}

// SelfId -
func (scm *oneShardCoordinatorMock) SelfId() uint32 {
	return scm.shardID
}

// SetSelfId -
func (scm *oneShardCoordinatorMock) SetSelfId(shardID uint32) error {
	scm.shardID = shardID
	return nil
}

// SameShard -
func (scm *oneShardCoordinatorMock) SameShard(_, _ []byte) bool {
	return true
}

// CommunicationIdentifier -
func (scm *oneShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == core.MetachainShardId {
		return "_0_META"
	}

	return "_0"
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *oneShardCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
