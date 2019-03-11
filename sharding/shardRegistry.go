package sharding

import (
	"encoding/binary"
	"math"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// shardRegistry struct defines the functionality for handling transaction dispatching to
// the corresponding shards. The number of shards is currently passed as a constructor
// parameter and later it should be calculated by this structure
type shardRegistry struct {
	mask1 uint32
	mask2 uint32
	currentShardId uint32
	currentNumberOfShards uint32
}

// NewShardRegistry returns a new shardRegistry and initializes the masks
func NewShardRegistry(numberOfShards uint32) (*shardRegistry, error) {
	if numberOfShards < 1 {
		return nil, ErrInvalidNumberOfShards
	}
	sr := &shardRegistry{}
	sr.currentNumberOfShards = numberOfShards
	sr.mask1, sr.mask2 = sr.calculateMasks()

	return sr, nil
}

func (sr *shardRegistry) calculateMasks() (uint32, uint32) {
	n := math.Ceil(math.Log2(float64(sr.currentNumberOfShards)))
	return (1 << uint(n)) - 1, (1 << uint(n -1)) - 1
}

// ComputeShardForAddress calculates the shard for a given address used for transaction dispatching
func (sr *shardRegistry) ComputeShardForAddress(address state.AddressContainer) uint32 {
	addr := binary.BigEndian.Uint32(address.Bytes())
	shard := addr & sr.mask1
	if shard > sr.currentNumberOfShards -1 {
		shard = addr & sr.mask2
	}
	return shard
}

// CurrentNumberOfShards retruns the number of shards
func (sr *shardRegistry) CurrentNumberOfShards() uint32 {
	return sr.currentNumberOfShards
}

// CurrentShardId gets the shard id of the current node
func (sr *shardRegistry) CurrentShardId() uint32 {
	return sr.currentShardId
}

// SetCurrentShardId sets the shard id for the current node
func (sr *shardRegistry) SetCurrentShardId(shardId uint32) error {
	if shardId >= sr.currentNumberOfShards {
		return ErrInvalidShardId
	}
	sr.currentShardId = shardId
	return nil
}