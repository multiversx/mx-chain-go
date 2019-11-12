package metachain

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() *ArgsPendingMiniBlocks {
	storer := &mock.StorerStub{}
	marshalizer := &mock.MarshalizerMock{}

	return &ArgsPendingMiniBlocks{
		Marshalizer: marshalizer,
		Storage:     storer,
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
	assert.Equal(t, endOfEpoch.ErrNilArgsPendingMiniblocks, err)
}

func TestNewPendingMiniBlocks_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.Marshalizer = nil

	pmb, err := NewPendingMiniBlocks(arguments)
	assert.Nil(t, pmb)
	assert.Equal(t, endOfEpoch.ErrNilMarshalizer, err)
}

func TestNewPendingMiniBlocks_NilStorageShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	arguments.Storage = nil

	pmb, err := NewPendingMiniBlocks(arguments)
	assert.Nil(t, pmb)
	assert.Equal(t, endOfEpoch.ErrNilStorage, err)
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
	assert.Equal(t, endOfEpoch.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderWrongHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.Header{}

	err := pmb.AddProcessedHeader(header)
	assert.Equal(t, endOfEpoch.ErrWrongTypeAssertion, err)
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
	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
				{Hash: hash1, SenderShardId: 1},
				{Hash: hash2, SenderShardId: 1},
			}},
		},
	}

	err := pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	//Check miniblocks headers are returned
	shdMbHdrs := pmb.PendingMiniBlockHeaders()
	assert.True(t, isMbInSlice(hash1, shdMbHdrs))
	assert.True(t, isMbInSlice(hash2, shdMbHdrs))

	err = pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	//Check miniblocks headers are removed from pending list
	shdMbHdrs = pmb.PendingMiniBlockHeaders()
	assert.False(t, isMbInSlice(hash1, shdMbHdrs))
	assert.False(t, isMbInSlice(hash2, shdMbHdrs))
}

func TestPendingMiniBlockHeaders_PendingMiniBlockHeadersSliceIsSorted(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")
	hash3 := []byte("hash3")
	hash4 := []byte("hash4")
	hash5 := []byte("hash5")

	sortedTxCount := []uint32{1, 2, 3, 4, 5}
	numMiniBlocks := 5
	arguments := createMockArguments()
	arguments.Storage = &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
				{Hash: hash1, SenderShardId: 1, TxCount: sortedTxCount[2]},
				{Hash: hash2, SenderShardId: 1, TxCount: sortedTxCount[4]},
				{Hash: hash3, SenderShardId: 1, TxCount: sortedTxCount[1]},
				{Hash: hash4, SenderShardId: 1, TxCount: sortedTxCount[0]},
				{Hash: hash5, SenderShardId: 1, TxCount: sortedTxCount[3]},
			}},
		},
	}

	err := pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	//Check miniblocks headers are returned
	shdMbHdrs := pmb.PendingMiniBlockHeaders()
	for i := 0; i < numMiniBlocks; i++ {
		assert.Equal(t, shdMbHdrs[i].TxCount, sortedTxCount[i])
	}
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

	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
				{Hash: hash1, SenderShardId: 1},
				{Hash: hash2, SenderShardId: 1},
			}},
		},
	}

	err := pmb.AddProcessedHeader(header)
	assert.Nil(t, err)

	shdMbHdrs := pmb.PendingMiniBlockHeaders()
	assert.True(t, isMbInSlice(hash1, shdMbHdrs))
	assert.True(t, isMbInSlice(hash2, shdMbHdrs))

	err = pmb.AddProcessedHeader(header)
	assert.NotNil(t, err)

	//Check miniblocks headers are not removed from pending list
	shdMbHdrs = pmb.PendingMiniBlockHeaders()
	assert.True(t, isMbInSlice(hash1, shdMbHdrs))
	assert.False(t, isMbInSlice(hash2, shdMbHdrs))
}

func TestPendingMiniBlockHeaders_RevertHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)

	err := pmb.RevertHeader(nil)
	assert.Equal(t, endOfEpoch.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderWrongHeaderTypeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockArguments()
	pmb, _ := NewPendingMiniBlocks(arguments)
	header := &block.Header{}

	err := pmb.RevertHeader(header)
	assert.Equal(t, endOfEpoch.ErrWrongTypeAssertion, err)
}
