package poolsCleaner_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/poolsCleaner"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNewHeaderPoolsCleaner_ShouldErrNilForkDetector(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		nil,
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldErrNilHeadersNoncesDataPool(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		nil,
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldErrNilHeadersDataPool(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		nil,
		&mock.CacherMock{},
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldErrNilNotarizedHeadersDataPool(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilNotarizedHeadersDataPool, err)
	assert.Nil(t, hpc)
}

func TestNewHeaderPoolsCleaner_ShouldWork(t *testing.T) {
	hpc, err := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, hpc)
}

func TestClean_ShouldNotCleanBehindNonceOne(t *testing.T) {
	hpc, _ := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherMock{},
		&mock.CacherMock{},
	)

	finalNonceInSelfShard := uint64(1)
	finalNoncesInNotarizedShards := make(map[uint32]uint64)
	finalNoncesInNotarizedShards[sharding.MetachainShardId] = 1

	hpc.Clean(finalNonceInSelfShard, finalNoncesInNotarizedShards)
	assert.Equal(t, uint64(0), hpc.NumRemovedHeaders())

}

func TestClean_ShouldNotCleanWhenKeyIsNotFound(t *testing.T) {
	headerHash := []byte("headerHash")
	metaBlockHash := []byte("metaBlockHash")

	hpc, _ := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{},
		&mock.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{headerHash}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		},
		&mock.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{metaBlockHash}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		},
	)

	finalNonceInSelfShard := uint64(2)
	finalNoncesInNotarizedShards := make(map[uint32]uint64)
	finalNoncesInNotarizedShards[sharding.MetachainShardId] = 2

	hpc.Clean(finalNonceInSelfShard, finalNoncesInNotarizedShards)
	assert.Equal(t, uint64(0), hpc.NumRemovedHeaders())

}

func TestClean_ShouldWork(t *testing.T) {
	header := &block.Header{Nonce: 2}
	headerHash := []byte("headerHash")

	metaBlock := &block.MetaBlock{Nonce: 2}
	metaBlockHash := []byte("metaBlockHash")

	hpc, _ := poolsCleaner.NewHeaderPoolsCleaner(
		mock.NewMultipleShardsCoordinatorMock(),
		&mock.Uint64SyncMapCacherStub{
			RemoveCalled: func(nonce uint64, shardId uint32) {},
		},
		&mock.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{headerHash}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, headerHash) {
					return header, true
				}
				return nil, false
			},
			RemoveCalled: func(key []byte) {},
		},
		&mock.CacherStub{
			KeysCalled: func() [][]byte {
				return [][]byte{metaBlockHash}
			},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, metaBlockHash) {
					return metaBlock, true
				}
				return nil, false
			},
			RemoveCalled: func(key []byte) {},
		},
	)

	finalNonceInSelfShard := uint64(2)
	finalNoncesInNotarizedShards := make(map[uint32]uint64)
	finalNoncesInNotarizedShards[sharding.MetachainShardId] = 2

	hpc.Clean(finalNonceInSelfShard, finalNoncesInNotarizedShards)
	assert.Equal(t, uint64(0), hpc.NumRemovedHeaders())

	finalNonceInSelfShard = uint64(3)
	finalNoncesInNotarizedShards = make(map[uint32]uint64)
	finalNoncesInNotarizedShards[sharding.MetachainShardId] = 3

	hpc.Clean(finalNonceInSelfShard, finalNoncesInNotarizedShards)
	assert.Equal(t, uint64(2), hpc.NumRemovedHeaders())
}
