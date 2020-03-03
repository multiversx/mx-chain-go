package disabled

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

type shardCoordinator struct {
	numShards uint32
}

// NewShardCoordinator -
func NewShardCoordinator() *shardCoordinator {
	return &shardCoordinator{numShards: 1}
}

// NumberOfShards -
func (scm *shardCoordinator) NumberOfShards() uint32 {
	return scm.numShards
}

// SetNoShards -
func (scm *shardCoordinator) SetNoShards(shards uint32) {
	scm.numShards = shards
}

// ComputeId -
func (scm *shardCoordinator) ComputeId(address state.AddressContainer) uint32 {

	return uint32(0)
}

// SelfId -
func (scm *shardCoordinator) SelfId() uint32 {
	return 0
}

// SetSelfId -
func (scm *shardCoordinator) SetSelfId(shardId uint32) error {
	return nil
}

// SameShard -
func (scm *shardCoordinator) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

// CommunicationIdentifier -
func (scm *shardCoordinator) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == core.MetachainShardId {
		return "_0_META"
	}

	return "_0"
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardCoordinator) IsInterfaceNil() bool {
	return scm == nil
}
