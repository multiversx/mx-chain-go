package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

type MultipleShardsCoordinatorMock struct {
	NoShards        uint32
	ComputeIdCalled func(address state.AddressContainer) uint32
	CurrentShard    uint32
}

func NewMultipleShardsCoordinatorMock() *MultipleShardsCoordinatorMock {
	return &MultipleShardsCoordinatorMock{NoShards: 1}
}

func NewMultiShardsCoordinatorMock(nrShard uint32) *MultipleShardsCoordinatorMock {
	return &MultipleShardsCoordinatorMock{NoShards: nrShard}
}

func (scm *MultipleShardsCoordinatorMock) NumberOfShards() uint32 {
	return scm.NoShards
}

func (scm *MultipleShardsCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	if scm.ComputeIdCalled == nil {
		return scm.SelfId()
	}
	return scm.ComputeIdCalled(address)
}

func (scm *MultipleShardsCoordinatorMock) SelfId() uint32 {
	return scm.CurrentShard
}

func (scm *MultipleShardsCoordinatorMock) SetSelfId(shardId uint32) error {
	return nil
}

func (scm *MultipleShardsCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

func (scm *MultipleShardsCoordinatorMock) SetNoShards(noShards uint32) {
	scm.NoShards = noShards
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// identifier is generated such as the first shard from identifier is always smaller than the last
func (scm *MultipleShardsCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == scm.CurrentShard {
		return fmt.Sprintf("_%d", scm.CurrentShard)
	}

	if destShardID < scm.CurrentShard {
		return fmt.Sprintf("_%d_%d", destShardID, scm.CurrentShard)
	}

	return fmt.Sprintf("_%d_%d", scm.CurrentShard, destShardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *MultipleShardsCoordinatorMock) IsInterfaceNil() bool {
	if scm == nil {
		return true
	}
	return false
}
