package mock

import (
	"math"

	"github.com/multiversx/mx-chain-core-go/core"
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
func (scm *ShardCoordinatorMock) ComputeId(address []byte) uint32 {
	maskHigh, maskLow := scm.calculateMasks()

	bytesNeed := int(scm.NumOfShards/256) + 1
	startingIndex := 0
	if len(address) > bytesNeed {
		startingIndex = len(address) - bytesNeed
	}

	buffNeeded := address[startingIndex:]
	if core.IsSmartContractOnMetachain(buffNeeded, address) {
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
func (scm *ShardCoordinatorMock) SameShard(address1, address2 []byte) bool {
	if len(address1) == 0 || len(address2) == 0 {
		return false
	}

	return scm.ComputeId(address1) == scm.ComputeId(address2)
}

// CommunicationIdentifier -
func (scm *ShardCoordinatorMock) CommunicationIdentifier(_ uint32) string {
	return ""
}

// IsInterfaceNil -
func (scm *ShardCoordinatorMock) IsInterfaceNil() bool {
	return scm == nil
}
