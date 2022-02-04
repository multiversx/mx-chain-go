package pendingMb_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/pendingMb"
	"github.com/stretchr/testify/assert"
)

func TestNewPendingMiniBlocks_ShouldWork(t *testing.T) {
	t.Parallel()

	pmb, err := pendingMb.NewPendingMiniBlocks()

	assert.False(t, check.IfNil(pmb))
	assert.Nil(t, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	err := pmb.AddProcessedHeader(nil)

	assert.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderWrongHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	header := &block.Header{}
	err := pmb.AddProcessedHeader(header)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func createMockHeader(
	shardMiniBlockHeaders []block.MiniBlockHeader,
	miniBlockHeaders []block.MiniBlockHeader,
) *block.MetaBlock {
	return &block.MetaBlock{
		Nonce: 1,
		ShardInfo: []block.ShardData{
			{
				ShardMiniBlockHeaders: shardMiniBlockHeaders,
			},
		},
		MiniBlockHeaders: miniBlockHeaders,
	}
}

func TestPendingMiniBlockHeaders_AddProcessedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	shardMiniBlockHeaders := []block.MiniBlockHeader{
		{
			Hash:            []byte("mb header hash"),
			SenderShardID:   0,
			ReceiverShardID: 0,
		},
	}
	miniBlockHeaders := []block.MiniBlockHeader{
		{
			Hash:            []byte("mb header hash"),
			SenderShardID:   0,
			ReceiverShardID: 0,
		},
	}

	t.Run("same sender and receiver shard, empty pendingMb", func(t *testing.T) {
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMb := pmb.GetPendingMiniBlocks(0)
		assert.Equal(t, 0, len(pendingMb))
	})
	t.Run("metachain receiver shard, empty pendingMb", func(t *testing.T) {
		receivedShardId := core.MetachainShardId
		miniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: receivedShardId,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMb := pmb.GetPendingMiniBlocks(receivedShardId)
		assert.Equal(t, 0, len(pendingMb))
	})
	t.Run("metachain sender shard, empty pendingMb", func(t *testing.T) {
		miniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: core.AllShardId,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMb := pmb.GetPendingMiniBlocks(core.MetachainShardId)
		assert.Equal(t, 0, len(pendingMb))
	})
	t.Run("different sender and receiver shard, non empty pendingMb", func(t *testing.T) {
		receivedShardId := uint32(1)
		miniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: receivedShardId,
			},
		}
		shardMiniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("shard mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: receivedShardId,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMb := pmb.GetPendingMiniBlocks(receivedShardId)
		assert.Equal(t, 2, len(pendingMb))
	})
	t.Run("different sender and receiver shard, delete already added, non empty pendingMb", func(t *testing.T) {
		receivedShardId := uint32(1)
		miniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: receivedShardId,
			},
		}
		shardMiniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("shard mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: receivedShardId,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)

		mbHashes := make([][]byte, 0)
		mbHashes = append(mbHashes, []byte("mbHash1"))
		mbHashes = append(mbHashes, []byte("mbHash2"))

		pmb.SetPendingMiniBlocks(receivedShardId, mbHashes)

		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMb := pmb.GetPendingMiniBlocks(receivedShardId)
		assert.Equal(t, 2, len(pendingMb))
	})
}

func TestPendingMiniBlockHeaders_RevertHeaderNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	err := pmb.RevertHeader(nil)

	assert.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestPendingMiniBlockHeaders_RevertHeaderWrongHeaderTypeShouldErr(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()
	header := &block.Header{}
	err := pmb.RevertHeader(header)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestPendingMiniBlockHeaders_SetPendingMiniBlocks(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()

	mbHashes := make([][]byte, 0)
	mbHashes = append(mbHashes, []byte("mbHash1"))
	mbHashes = append(mbHashes, []byte("mbHash2"))

	pmb.SetPendingMiniBlocks(1, mbHashes)

	pendingMb := pmb.GetPendingMiniBlocks(1)
	assert.Equal(t, 2, len(pendingMb))

	pendingMb = pmb.GetPendingMiniBlocks(0)
	assert.Equal(t, 0, len(pendingMb))
}
