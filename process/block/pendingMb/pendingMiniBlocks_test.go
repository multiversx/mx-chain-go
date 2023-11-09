package pendingMb_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/pendingMb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
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
	t.Run("same sender and receiver shard, empty pendingMb", func(t *testing.T) {
		miniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		}

		shardMiniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMiniblocks := pmb.GetPendingMiniBlocks(0)
		assert.Equal(t, 0, len(pendingMiniblocks))
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
		shardMiniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMiniblocks := pmb.GetPendingMiniBlocks(receivedShardId)
		assert.Equal(t, 0, len(pendingMiniblocks))
	})
	t.Run("metachain sender shard, empty pendingMb", func(t *testing.T) {
		miniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: core.AllShardId,
			},
		}
		shardMiniBlockHeaders := []block.MiniBlockHeader{
			{
				Hash:            []byte("mb header hash"),
				SenderShardID:   0,
				ReceiverShardID: 0,
			},
		}
		header := createMockHeader(shardMiniBlockHeaders, miniBlockHeaders)
		err := pmb.AddProcessedHeader(header)
		assert.Nil(t, err)

		pendingMiniblocks := pmb.GetPendingMiniBlocks(core.MetachainShardId)
		assert.Equal(t, 0, len(pendingMiniblocks))
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

		pendingMiniblocks := pmb.GetPendingMiniBlocks(receivedShardId)
		assert.Equal(t, 2, len(pendingMiniblocks))
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

		pendingMiniblocks := pmb.GetPendingMiniBlocks(receivedShardId)
		assert.Equal(t, 2, len(pendingMiniblocks))
	})
	t.Run("epoch start should reprocess everything", func(t *testing.T) {
		t.Parallel()

		pendingMiniblocks, _ := pendingMb.NewPendingMiniBlocks()
		pendingMiniblocks.SetInMapPendingMbShard("hash 1", 1)
		pendingMiniblocks.SetInMapPendingMbShard("hash 2", 2)
		pendingMiniblocks.SetInMapPendingMbShard("hash 3", 4)

		pendingMiniblocks.SetInBeforeRevertPendingMbShard("hash 5", 1)
		pendingMiniblocks.SetInBeforeRevertPendingMbShard("hash 6", 2)
		pendingMiniblocks.SetInBeforeRevertPendingMbShard("hash 7", 4)

		startOfEpochMetaBlock := &block.MetaBlock{
			ShardInfo: []block.ShardData{
				{
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            []byte("hash 1"),
							ReceiverShardID: 1,
						},
						{
							Hash:            []byte("hash 2"),
							ReceiverShardID: 2,
						},
					},
				},
			},
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{
						PendingMiniBlockHeaders: []block.MiniBlockHeader{
							{
								Hash:            []byte("hash 8"),
								ReceiverShardID: 3,
							},
							{
								Hash:            []byte("hash 9"),
								ReceiverShardID: 2,
							},
						},
					},
					{
						PendingMiniBlockHeaders: []block.MiniBlockHeader{
							{
								Hash:            []byte("hash 10"),
								ReceiverShardID: 4,
							},
							{
								Hash:            []byte("hash 11"),
								ReceiverShardID: 5,
							},
							{
								Hash:            []byte("hash 11"),
								ReceiverShardID: 6,
								SenderShardID:   6,
							},
						},
					},
				},
			},
		}

		err := pendingMiniblocks.AddProcessedHeader(startOfEpochMetaBlock)
		require.Nil(t, err)

		expectedMap := map[string]uint32{
			"hash 8":  3,
			"hash 9":  2,
			"hash 10": 4,
			"hash 11": 5,
		}
		expectedBefore := map[string]uint32{
			"hash 1": 1,
			"hash 2": 2,
			"hash 3": 4,
		}

		assert.Equal(t, expectedMap, pendingMiniblocks.GetMapPendingMbShard())
		assert.Equal(t, expectedBefore, pendingMiniblocks.GetBeforeRevertPendingMbShard())
	})
}

func TestPendingMiniBlockHeaders_Revert(t *testing.T) {
	t.Parallel()

	t.Run("nil header should error", func(t *testing.T) {
		t.Parallel()

		pmb, _ := pendingMb.NewPendingMiniBlocks()
		err := pmb.RevertHeader(nil)

		assert.Equal(t, process.ErrNilHeaderHandler, err)
	})
	t.Run("wrong header should error", func(t *testing.T) {
		t.Parallel()

		pmb, _ := pendingMb.NewPendingMiniBlocks()
		header := &block.Header{}
		err := pmb.RevertHeader(header)

		assert.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
	})
	t.Run("not an epoch start should revert", func(t *testing.T) {
		t.Parallel()

		pmb, _ := pendingMb.NewPendingMiniBlocks()
		pmb.SetInMapPendingMbShard("hash 1", 1)
		pmb.SetInMapPendingMbShard("hash 3", 1)
		pmb.SetInMapPendingMbShard("hash 4", 2)

		header := &block.MetaBlock{
			ShardInfo: []block.ShardData{
				{
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            []byte("hash 1"),
							ReceiverShardID: 1,
						},
						{
							Hash:            []byte("hash 2"),
							ReceiverShardID: 2,
						},
					},
				},
			},
		}
		err := pmb.RevertHeader(header)
		require.Nil(t, err)

		expectedMap := map[string]uint32{
			"hash 2": 2,
			"hash 3": 1,
			"hash 4": 2,
		}
		assert.Equal(t, expectedMap, pmb.GetMapPendingMbShard())
		assert.Equal(t, make(map[string]uint32), pmb.GetBeforeRevertPendingMbShard())
	})
	t.Run("epoch start should revert completely", func(t *testing.T) {
		t.Parallel()

		pmb, _ := pendingMb.NewPendingMiniBlocks()
		pmb.SetInMapPendingMbShard("hash 1", 1)
		pmb.SetInMapPendingMbShard("hash 2", 2)
		pmb.SetInMapPendingMbShard("hash 3", 4)

		pmb.SetInBeforeRevertPendingMbShard("hash 5", 1)
		pmb.SetInBeforeRevertPendingMbShard("hash 6", 2)
		pmb.SetInBeforeRevertPendingMbShard("hash 7", 4)

		startOfEpochMetaBlock := &block.MetaBlock{
			ShardInfo: []block.ShardData{
				{
					ShardMiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash:            []byte("hash 1"),
							ReceiverShardID: 1,
						},
						{
							Hash:            []byte("hash 2"),
							ReceiverShardID: 2,
						},
					},
				},
			},
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{},
				},
			},
		}
		err := pmb.RevertHeader(startOfEpochMetaBlock)
		require.Nil(t, err)

		expectedMap := map[string]uint32{
			"hash 5": 1,
			"hash 6": 2,
			"hash 7": 4,
		}
		assert.Equal(t, expectedMap, pmb.GetMapPendingMbShard())
		assert.Equal(t, expectedMap, pmb.GetBeforeRevertPendingMbShard())
	})
}

func TestPendingMiniBlockHeaders_SetPendingMiniBlocks(t *testing.T) {
	t.Parallel()

	pmb, _ := pendingMb.NewPendingMiniBlocks()

	mbHashes := make([][]byte, 0)
	mbHashes = append(mbHashes, []byte("mbHash1"))
	mbHashes = append(mbHashes, []byte("mbHash2"))

	pmb.SetPendingMiniBlocks(1, mbHashes)

	pendingMiniblocks := pmb.GetPendingMiniBlocks(1)
	assert.Equal(t, 2, len(pendingMiniblocks))

	pendingMiniblocks = pmb.GetPendingMiniBlocks(0)
	assert.Equal(t, 0, len(pendingMiniblocks))
}
