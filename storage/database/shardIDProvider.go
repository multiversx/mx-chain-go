package database

import (
	"errors"
	"math"
)

// TODO: add unit tests

const bitsPerByte = 8

// ErrInvalidNumberOfShards signals that an invalid number of shards was passed to the sharding registry
var ErrInvalidNumberOfShards = errors.New("the number of shards must be greater than zero")

type shardIDProvider struct {
	maskHigh    uint32
	maskLow     uint32
	numOfShards uint32
	bytesNeeded int
}

// NewShardIDProvider will create a new shard ID provider component
func NewShardIDProvider(numOfShards uint32) (*shardIDProvider, error) {
	if numOfShards == 0 {
		return nil, ErrInvalidNumberOfShards
	}

	sp := &shardIDProvider{
		numOfShards: numOfShards,
	}
	sp.calculateMasks()
	sp.calculateBytesNeeded()

	return sp, nil
}

// ComputeId computes the shard id for a given key
func (sp *shardIDProvider) ComputeId(key []byte) uint32 {
	startingIndex := 0
	if len(key) > sp.bytesNeeded {
		startingIndex = len(key) - sp.bytesNeeded
	}

	buffNeeded := key[startingIndex:]
	addr := uint32(0)
	for i := 0; i < len(buffNeeded); i++ {
		addr = addr<<8 + uint32(buffNeeded[i])
	}

	shardIndex := addr & sp.maskHigh
	if shardIndex > sp.numOfShards-1 {
		shardIndex = addr & sp.maskLow
	}

	return shardIndex
}

// NumberOfShards returns the number of shards
func (sp *shardIDProvider) NumberOfShards() uint32 {
	return sp.numOfShards
}

// GetShardIDs will return a list of all shard ids
func (sp *shardIDProvider) GetShardIDs() []uint32 {
	shardIDs := make([]uint32, sp.NumberOfShards())

	for i := uint32(0); i < sp.NumberOfShards(); i++ {
		shardIDs[i] = i
	}

	return shardIDs
}

// calculateMasks will create two numbers whose binary form is composed of as many
// ones needed to be taken into consideration for the shard assignment. The result
// of a bitwise AND operation of an key with this mask will result in the
// shard id for the key
func (sp *shardIDProvider) calculateMasks() {
	n := math.Ceil(math.Log2(float64(sp.numOfShards)))
	sp.maskHigh = (1 << uint(n)) - 1
	sp.maskLow = (1 << uint(n-1)) - 1
}

// calculateBytesNeeded will calculate the number of bytes needed from an address for index calculation
func (sp *shardIDProvider) calculateBytesNeeded() {
	n := sp.numOfShards
	if n == 1 {
		sp.bytesNeeded = 1
		return
	}
	sp.bytesNeeded = int(math.Floor(math.Log2(float64(n-1))))/bitsPerByte + 1
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *shardIDProvider) IsInterfaceNil() bool {
	return sp == nil
}
