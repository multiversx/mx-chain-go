package metachain

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() *ArgsPendingMiniBlocks {
	return &ArgsPendingMiniBlocks{
		Marshalizer: &mock.MarshalizerMock{},
		Storage:     &mock.StorerStub{},
		MetaBlockStorage: &mock.StorerStub{
			GetCalled: func(key []byte) (i []byte, e error) {
				return nil, epochStart.ErrMetaHdrNotFound
			},
		},
		MetaBlockPool: &mock.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		},
	}
}

func isMbInSlice(hash []byte, shdMbHdrs []block.ShardMiniBlockHeader) bool {
	for _, shdMbHdr := range shdMbHdrs {
		if bytes.Equal(shdMbHdr.Hash, hash) {
			return true
		}
	}
	return false
}

func TestNewPendingMiniBlocks_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	pmb, err := NewPendingMiniBlocks(nil)

	assert.Nil(t, pmb)
	assert.Equal(t, epochStart.ErrNilArgsPendingMiniblocks, err)
}

func TestNewPendingMiniBlocks_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.Marshalizer = nil

	pmb, err := NewPendingMiniBlocks(arguments)
	assert.Nil(t, pmb)
	assert.Equal(t, epochStart.ErrNilMarshalizer, err)
}

func TestNewPendingMiniBlocks_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.Storage = nil

	pmb, err := NewPendingMiniBlocks(arguments)
	assert.Nil(t, pmb)
	assert.Equal(t, epochStart.ErrNilStorage, err)
}

func TestNewPendingMiniBlocks_ShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()

	pmb, err := NewPendingMiniBlocks(arguments)
	assert.NotNil(t, pmb)
	assert.Nil(t, err)
}

func TestPendingMiniBlockHeaders_AddCommittedHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)

	err := pmb.AddProcessedHeader(nil)
	assert.Equal(t, epochStart.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderWrongHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.Header{}

	err := pmb.AddProcessedHeader(header)
	assert.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeader(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")

	arguments := createMockArguments()
	arguments.Storage = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}

	header := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
				{Hash: hash1, SenderShardID: 1},
				{Hash: hash2, SenderShardID: 1},
			}},
		},
	}
	arguments.MetaBlockPool = &mock.CacherStub{PeekCalled: func(key []byte) (value interface{}, ok bool) {
		return header, true
	}}
	shardHeader := &block.Header{MetaBlockHashes: [][]byte{[]byte("metaHash")}}

	pmb, _ := NewPendingMiniBlocks(arguments)
	err := pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	//Check miniblocks headers are returned
	shdMbHdrs, err := pmb.PendingMiniBlockHeaders([]data.HeaderHandler{shardHeader})
	assert.True(t, isMbInSlice(hash1, shdMbHdrs))
	assert.True(t, isMbInSlice(hash2, shdMbHdrs))

	err = pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	//Check miniblocks headers are removed from pending list
	shdMbHdrs, err = pmb.PendingMiniBlockHeaders([]data.HeaderHandler{shardHeader})
	assert.False(t, isMbInSlice(hash1, shdMbHdrs))
	assert.False(t, isMbInSlice(hash2, shdMbHdrs))
}

func TestPendingMiniBlockHeaders_PendingMiniBlockHeaders(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hash3 := []byte("hash3")
	hash4 := []byte("hash4")
	hash5 := []byte("hash5")

	arguments := createMockArguments()
	arguments.Storage = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	header := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
				{Hash: hash1, SenderShardID: 1, TxCount: 5},
				{Hash: hash2, SenderShardID: 1, TxCount: 5},
				{Hash: hash3, SenderShardID: 1, TxCount: 5},
				{Hash: hash4, SenderShardID: 1, TxCount: 5},
				{Hash: hash5, SenderShardID: 1, TxCount: 5},
			}},
		},
	}
	arguments.MetaBlockPool = &mock.CacherStub{PeekCalled: func(key []byte) (value interface{}, ok bool) {
		return header, true
	}}
	shardHeader := &block.Header{MetaBlockHashes: [][]byte{[]byte("metaHash")}}

	pmb, _ := NewPendingMiniBlocks(arguments)
	err := pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	//Check miniblocks headers are returned
	shdMbHdrs, _ := pmb.PendingMiniBlockHeaders([]data.HeaderHandler{shardHeader})
	assert.Equal(t, len(shdMbHdrs), len(header.ShardInfo[0].ShardMiniBlockHeaders))
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderCannotMarshalShouldRevert(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")

	arguments := createMockArguments()
	arguments.Storage = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
		RemoveCalled: func(key []byte) error {
			return nil
		},
	}
	arguments.Marshalizer = &mock.MarshalizerMock{
		Fail: true,
	}

	header := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
				{Hash: hash1, SenderShardID: 1},
				{Hash: hash2, SenderShardID: 1},
			}},
		},
	}
	arguments.MetaBlockPool = &mock.CacherStub{PeekCalled: func(key []byte) (value interface{}, ok bool) {
		return header, true
	}}
	shardHeader := &block.Header{MetaBlockHashes: [][]byte{[]byte("metaHash")}}

	pmb, _ := NewPendingMiniBlocks(arguments)
	err := pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	shdMbHdrs, _ := pmb.PendingMiniBlockHeaders([]data.HeaderHandler{shardHeader})
	assert.True(t, isMbInSlice(hash1, shdMbHdrs))
	assert.True(t, isMbInSlice(hash2, shdMbHdrs))

	err = pmb.AddProcessedHeader(header)
	assert.NotNil(t, err)

	//Check miniblocks headers are not removed from pending list
	shdMbHdrs, _ = pmb.PendingMiniBlockHeaders([]data.HeaderHandler{shardHeader})
	assert.True(t, isMbInSlice(hash1, shdMbHdrs))
	assert.False(t, isMbInSlice(hash2, shdMbHdrs))
}

func TestPendingMiniBlockHeaders_RevertHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)

	err := pmb.RevertHeader(nil)
	assert.Equal(t, epochStart.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderWrongHeaderTypeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.Header{}

	err := pmb.RevertHeader(header)
	assert.Equal(t, epochStart.ErrWrongTypeAssertion, err)
}
