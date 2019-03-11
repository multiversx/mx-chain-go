package sharding_test

import (
	"encoding/binary"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func TestShardRegistry_NewShardRegistry(t *testing.T) {
	nrOfShards := uint32(10)
	sr, _ := sharding.NewShardRegistry(nrOfShards)
	assert.Equal(t, nrOfShards, sr.CurrentNumberOfShards())
	expectedMask1, expectedMask2 := sr.CalculateMasks()
	actualMask1, actualMask2 := sr.Masks()
	assert.Equal(t, expectedMask1, actualMask1)
	assert.Equal(t, expectedMask2, actualMask2)
}

func TestShardRegistry_NewShardRegistryInvalidNumberOfShards(t *testing.T) {
	sr, err := sharding.NewShardRegistry(0)
	assert.Nil(t, sr)
	assert.Equal(t, sharding.ErrInvalidNumberOfShards, err)
}

func TestShardRegistry_SetCurrentShardIdGraterThanNrOfShardsShouldError(t *testing.T) {
	sr, _ := sharding.NewShardRegistry(1)
	err := sr.SetCurrentShardId(2)
	assert.Equal(t, sharding.ErrInvalidShardId, err)
}

func TestShardRegistry_SetCurrentShardId(t *testing.T) {
	currentShardId := uint32(0)
	sr, _ := sharding.NewShardRegistry(1)
	_ = sr.SetCurrentShardId(currentShardId)
	assert.Equal(t, currentShardId, sr.CurrentShardId())
}

func TestShardRegistry_ComputeShardForAddressDoesNotGenerateInvalidShards(t *testing.T) {
	nrOfShards := uint32(10)
	sr, _ := sharding.NewShardRegistry(nrOfShards)

	for i := 0; i < 200; i++ {
		buff := make([]byte, 4)
		binary.BigEndian.PutUint32(buff, uint32(i))
		addr := &mock.AddressMock{Bts: buff}
		shardId := sr.ComputeShardForAddress(addr)
		assert.True(t, shardId < sr.CurrentNumberOfShards())
	}
}

func TestShardRegistry_ComputeShardForAddressSameSuffixHasSameShard(t *testing.T) {
	nrOfShards := uint32(2)
	sr, _ := sharding.NewShardRegistry(nrOfShards)

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
		buff := make([]byte, 4)
		binary.BigEndian.PutUint32(buff, data.address)
		addr := &mock.AddressMock{Bts: buff}
		shardId := sr.ComputeShardForAddress(addr)

		assert.Equal(t, data.shardId, shardId)
	}
}
