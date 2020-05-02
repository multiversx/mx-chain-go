package sharding

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getAddressFromUint32(address uint32) []byte {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, address)

	return buff
}

func TestMultiShardCoordinator_NewMultiShardCoordinator(t *testing.T) {
	nrOfShards := uint32(10)
	sr, _ := NewMultiShardCoordinator(nrOfShards, 0)
	assert.Equal(t, nrOfShards, sr.NumberOfShards())
	expectedMask1, expectedMask2 := sr.calculateMasks()
	actualMask1 := sr.maskHigh
	actualMask2 := sr.maskLow
	assert.Equal(t, expectedMask1, actualMask1)
	assert.Equal(t, expectedMask2, actualMask2)
}

func TestMultiShardCoordinator_NewMultiShardCoordinatorInvalidNumberOfShards(t *testing.T) {
	sr, err := NewMultiShardCoordinator(0, 0)
	assert.Nil(t, sr)
	assert.Equal(t, ErrInvalidNumberOfShards, err)
}

func TestMultiShardCoordinator_NewMultiShardCoordinatorSelfIdGraterThanNrOfShardsShouldError(t *testing.T) {
	_, err := NewMultiShardCoordinator(1, 2)
	assert.Equal(t, ErrInvalidShardId, err)
}

func TestMultiShardCoordinator_NewMultiShardCoordinatorCorrectSelfId(t *testing.T) {
	currentShardId := uint32(0)
	sr, _ := NewMultiShardCoordinator(1, currentShardId)
	assert.Equal(t, currentShardId, sr.SelfId())
}

func TestMultiShardCoordinator_ComputeIdDoesNotGenerateInvalidShards(t *testing.T) {
	nrOfShards := uint32(10)
	selfId := uint32(0)
	sr, _ := NewMultiShardCoordinator(nrOfShards, selfId)

	for i := 0; i < 200; i++ {
		addr := getAddressFromUint32(uint32(i))
		shardId := sr.ComputeId(addr)
		assert.True(t, shardId < sr.NumberOfShards())
	}
}

func TestMultiShardCoordinator_ComputeId10ShardsShouldWork(t *testing.T) {
	nrOfShards := uint32(10)
	selfId := uint32(0)
	sr, _ := NewMultiShardCoordinator(nrOfShards, selfId)

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

func TestMultiShardCoordinator_ComputeId10ShardsBigNumbersShouldWork(t *testing.T) {
	nrOfShards := uint32(10)
	selfId := uint32(0)
	sr, _ := NewMultiShardCoordinator(nrOfShards, selfId)

	dataSet := []struct {
		address string
		shardId uint32
	}{
		{"2ca2ed0a1c77b5ddeabf99e3f17074f1c77b5ddea37dbaacf501ef1752950c50", 0},
		{"5f7e73b922883bf97b5ddeab9e3f103b8ddea37dbaaca24b5ddea37dbaac1061", 1},
		{"65b9926097345bf7b5ddeab99e3f1cc7c71c01bfd3e1efacc1d0df1ba8f96172", 2},
		{"c1c77b5ddea5c71c4160c861c01ba8f9617a65010d2baac7827a501bbf29aad3", 3},
		{"22c2e1facc1d1c77b5d16160c861c01ba8f96170c86783b8deabaac782733b84", 4},
		{"4cc88bdac668dc1878271e79a67b5ddeaddea37dbaacbbf99e3f1f4e5a4d7085", 5},
		{"b533facc1daa3a617466f4160c861c01ba8f9617a65010d160c86783b82a5836", 6},
		{"b1487283ad280316baa160c861c01ba8f9617c78270c8668dc18b5ddea37dba7", 7},
		{"acfba138faed1c7b5d668dc1878b5ddea37dbaac271edeab7160c861c01ba8f8", 8},
		{"cc3757647aebf9d160c86e9f9eb5dde0c86783b89160a37dbaac3f15c8f48999", 9},
		{"1a30b33104a94a65010d2a5e87285eb0ea37dbaac60c86c3facc1d4d1ba8f96a", 2},
		{"fcca8da9ba5160c86783b89160ea37dbaac60c86c86783b8c868be1ba8f9617b", 3},
		{"8f9b094668dc1878271ed1b1b5ddea37dbaac60c86760f2e4c71c01bf6a913cc", 4},
		{"a2d768be59a607d160c86eb5ddea37dbaafacc1dc9fa0e0c86783b8916092cbd", 5},
		{"365865f21b2e0d668dc18ea37dbaac60c8678271e160c86e9160c86783b8fe6e", 6},
		{"16cc745884a65ba160c861c01ba8f9617ac7827d160c86e9f010d2a592b3a52f", 7},
	}
	for _, data := range dataSet {
		buff, err := hex.DecodeString(data.address)
		assert.Nil(t, err)

		shardId := sr.ComputeId(buff)

		assert.Equal(t, data.shardId, shardId)
	}
}

func TestMultiShardCoordinator_ComputeIdSameSuffixHasSameShard(t *testing.T) {
	nrOfShards := uint32(2)
	selfId := uint32(0)
	sr, _ := NewMultiShardCoordinator(nrOfShards, selfId)

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
	shard, _ := NewMultiShardCoordinator(1, 0)
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(1))

	assert.True(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_SameShardSameAddressMultipleShards(t *testing.T) {
	shard, _ := NewMultiShardCoordinator(11, 0)
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(1))

	assert.True(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_SameShardDifferentAddress(t *testing.T) {
	shard, _ := NewMultiShardCoordinator(1, 0)
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(2))

	assert.True(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_SameShardDifferentAddressMultipleShards(t *testing.T) {
	shard, _ := NewMultiShardCoordinator(2, 0)

	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(2))

	assert.False(t, shard.SameShard(addr1, addr2))
}

func TestMultiShardCoordinator_CommunicationIdentifierSameShard(t *testing.T) {
	destId := uint32(1)
	selfId := uint32(1)
	shard, _ := NewMultiShardCoordinator(2, selfId)
	assert.Equal(t, fmt.Sprintf("_%d", selfId), shard.CommunicationIdentifier(destId))
}

func TestMultiShardCoordinator_CommunicationIdentifierSmallerDestination(t *testing.T) {
	destId := uint32(0)
	selfId := uint32(1)
	shard, _ := NewMultiShardCoordinator(2, selfId)
	assert.Equal(t, fmt.Sprintf("_%d_%d", destId, selfId), shard.CommunicationIdentifier(destId))
}

func TestMultiShardCoordinator_CommunicationIdentifier(t *testing.T) {
	destId := uint32(1)
	selfId := uint32(0)
	shard, _ := NewMultiShardCoordinator(2, selfId)
	assert.Equal(t, fmt.Sprintf("_%d_%d", selfId, destId), shard.CommunicationIdentifier(destId))
}
