package sharding_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func getAddressFromUint32(address uint32) state.AddressContainer {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, address)
	return &mock.AddressMock{Bts: buff}
}

func TestMultiShardCoordinator_NewMultiShardCoordinator(t *testing.T) {
	nrOfShards := uint32(10)
	sr, _ := sharding.NewMultiShardCoordinator(nrOfShards, 0)
	assert.Equal(t, nrOfShards, sr.NumberOfShards())
	expectedMask1, expectedMask2 := sr.CalculateMasks()
	actualMask1, actualMask2 := sr.Masks()
	assert.Equal(t, expectedMask1, actualMask1)
	assert.Equal(t, expectedMask2, actualMask2)
}

func TestMultiShardCoordinator_NewMultiShardCoordinatorInvalidNumberOfShards(t *testing.T) {
	sr, err := sharding.NewMultiShardCoordinator(0, 0)
	assert.Nil(t, sr)
	assert.Equal(t, sharding.ErrInvalidNumberOfShards, err)
}

func TestMultiShardCoordinator_NewMultiShardCoordinatorSelfIdGraterThanNrOfShardsShouldError(t *testing.T) {
	_, err := sharding.NewMultiShardCoordinator(1, 2)
	assert.Equal(t, sharding.ErrInvalidShardId, err)
}

func TestMultiShardCoordinator_NewMultiShardCoordinatorCorrectSelfId(t *testing.T) {
	currentShardId := uint32(0)
	sr, _ := sharding.NewMultiShardCoordinator(1, currentShardId)
	assert.Equal(t, currentShardId, sr.SelfId())
}

func TestMultiShardCoordinator_ComputeIdDoesNotGenerateInvalidShards(t *testing.T) {
	nrOfShards := uint32(10)
	selfId := uint32(0)
	sr, _ := sharding.NewMultiShardCoordinator(nrOfShards, selfId)

	for i := 0; i < 200; i++ {
		addr := getAddressFromUint32(uint32(i))
		shardId := sr.ComputeId(addr)
		assert.True(t, shardId < sr.NumberOfShards())
	}
}

func TestMultiShardCoordinator_ComputeId10ShardsShouldWork(t *testing.T) {
	nrOfShards := uint32(10)
	selfId := uint32(0)
	sr, _ := sharding.NewMultiShardCoordinator(nrOfShards, selfId)

	dataSet := []struct {
		address, shardId uint32
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
		{6, 6},
		{7, 7},
		{8, 8},
		{9, 9},
		{10, 2},
		{11, 3},
		{12, 4},
		{13, 5},
		{14, 6},
		{15, 7},
	}
	for _, data := range dataSet {
		addr := getAddressFromUint32(data.address)
		shardId := sr.ComputeId(addr)

		assert.Equal(t, data.shardId, shardId)
	}
}

func TestMultiShardCoordinator_ComputeIdSameSuffixHasSameShard(t *testing.T) {
	nrOfShards := uint32(2)
	selfId := uint32(0)
	sr, _ := sharding.NewMultiShardCoordinator(nrOfShards, selfId)

	dataSet := []struct {
		address, shardId uint32
	}{
		{0, 0},
		{1, 1},
		{2, 0},
		{3, 1},
		{4, 0},
		{5, 1},
		{6, 0},
		{7, 1},
		{8, 0},
		{9, 1},
	}
	for _, data := range dataSet {
		addr := getAddressFromUint32(data.address)
		shardId := sr.ComputeId(addr)

		assert.Equal(t, data.shardId, shardId)
	}
}

func TestMultiShardCoordinator_SameShardSameAddress(t *testing.T) {
	shard, _ := sharding.NewMultiShardCoordinator(1, 0)
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(1))

	assert.True(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_SameShardSameAddressMultipleShards(t *testing.T) {
	shard, _ := sharding.NewMultiShardCoordinator(11, 0)
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(1))

	assert.True(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_SameShardDifferentAddress(t *testing.T) {
	shard, _ := sharding.NewMultiShardCoordinator(1, 0)
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(2))

	assert.True(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_SameShardDifferentAddressMultipleShards(t *testing.T) {
	shard, _ := sharding.NewMultiShardCoordinator(2, 0)

	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(2))

	assert.False(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_CommunicationIdentifierSameShard(t *testing.T) {
	destId := uint32(1)
	selfId := uint32(1)
	shard, _ := sharding.NewMultiShardCoordinator(2, selfId)
	assert.Equal(t, fmt.Sprintf("_%d", selfId), shard.CommunicationIdentifier(destId))
}

func TestMultiShardCoordinator_CommunicationIdentifierSmallerDestination(t *testing.T) {
	destId := uint32(0)
	selfId := uint32(1)
	shard, _ := sharding.NewMultiShardCoordinator(2, selfId)
	assert.Equal(t, fmt.Sprintf("_%d_%d", destId, selfId), shard.CommunicationIdentifier(destId))
}

func TestMultiShardCoordinator_CommunicationIdentifier(t *testing.T) {
	destId := uint32(1)
	selfId := uint32(0)
	shard, _ := sharding.NewMultiShardCoordinator(2, selfId)
	assert.Equal(t, fmt.Sprintf("_%d_%d", selfId, destId), shard.CommunicationIdentifier(destId))
}
