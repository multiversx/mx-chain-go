package mock

import (
	"fmt"
	"math"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type multipleShardsCoordinatorFake struct {
	numOfShards  uint32
	CurrentShard uint32
	maskHigh     uint32
	maskLow      uint32
}

func NewMultipleShardsCoordinatorFake(numOfShards uint32, currentShard uint32) *multipleShardsCoordinatorFake {
	mscf := &multipleShardsCoordinatorFake{
		numOfShards:  numOfShards,
		CurrentShard: currentShard,
	}
	mscf.maskHigh, mscf.maskLow = mscf.calculateMasks()
	return mscf
}

func (mscf *multipleShardsCoordinatorFake) calculateMasks() (uint32, uint32) {
	n := math.Ceil(math.Log2(float64(mscf.numOfShards)))
	return (1 << uint(n)) - 1, (1 << uint(n-1)) - 1
}

func (mscf *multipleShardsCoordinatorFake) NumberOfShards() uint32 {
	return mscf.numOfShards
}

func (mscf *multipleShardsCoordinatorFake) ComputeId(address state.AddressContainer) uint32 {
	bytesNeed := int(mscf.numOfShards/256) + 1
	startingIndex := 0
	if len(address.Bytes()) > bytesNeed {
		startingIndex = len(address.Bytes()) - bytesNeed
	}

	buffNeeded := address.Bytes()[startingIndex:]

	addr := uint32(0)
	for i := 0; i < len(buffNeeded); i++ {
		addr = addr<<8 + uint32(buffNeeded[i])
	}

	shard := addr & mscf.maskHigh
	if shard > mscf.numOfShards-1 {
		shard = addr & mscf.maskLow
	}
	return shard
}

func (mscf *multipleShardsCoordinatorFake) SelfId() uint32 {
	return mscf.CurrentShard
}

func (mscf *multipleShardsCoordinatorFake) SetSelfId(shardId uint32) error {
	return nil
}

func (mscf *multipleShardsCoordinatorFake) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

func (mscf *multipleShardsCoordinatorFake) SetNoShards(numOfShards uint32) {
	mscf.numOfShards = numOfShards
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// identifier is generated such as the first shard from identifier is always smaller than the last
func (mscf *multipleShardsCoordinatorFake) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == mscf.CurrentShard {
		return fmt.Sprintf("_%d", mscf.CurrentShard)
	}

	if destShardID < mscf.CurrentShard {
		return fmt.Sprintf("_%d_%d", destShardID, mscf.CurrentShard)
	}

	return fmt.Sprintf("_%d_%d", mscf.CurrentShard, destShardID)
}
