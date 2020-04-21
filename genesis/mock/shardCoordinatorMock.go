package mock

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// ShardCoordinatorMock -
type ShardCoordinatorMock struct {
	NumOfShards uint32
	SelfShardId uint32
}

// NumberOfShards -
func (scm *ShardCoordinatorMock) NumberOfShards() uint32 {
	return scm.NumOfShards
}

// ComputeId -
func (scm *ShardCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	maskHigh, maskLow := scm.calculateMasks()

	bytesNeed := int(scm.NumOfShards/256) + 1
	startingIndex := 0
	if len(address.Bytes()) > bytesNeed {
		startingIndex = len(address.Bytes()) - bytesNeed
	}

	buffNeeded := address.Bytes()[startingIndex:]
	if core.IsSmartContractOnMetachain(buffNeeded, address.Bytes()) {
		return core.MetachainShardId
	}

	addr := uint32(0)
	for i := 0; i < len(buffNeeded); i++ {
		addr = addr<<8 + uint32(buffNeeded[i])
	}

	shard := addr & maskHigh
	if shard > scm.NumOfShards-1 {
		shard = addr & maskLow
	}

	return shard
}

func (scm *ShardCoordinatorMock) calculateMasks() (uint32, uint32) {
	n := math.Ceil(math.Log2(float64(scm.NumOfShards)))
	return (1 << uint(n)) - 1, (1 << uint(n-1)) - 1
}

// SelfId -
func (scm *ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfShardId
}

// SameShard -
func (scm *ShardCoordinatorMock) SameShard(_, _ state.AddressContainer) bool {
	return false
}

// CommunicationIdentifier -
func (scm *ShardCoordinatorMock) CommunicationIdentifier(_ uint32) string {
	return ""
}

// IsInterfaceNil -
func (scm *ShardCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
