package dblookupext

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/require"
)

func createMockHistoryRepoArgs(epoch uint32) HistoryRepositoryArguments {
	args := HistoryRepositoryArguments{
		SelfShardID:                 0,
		MiniblocksMetadataStorer:    genericmocks.NewStorerMock("MiniblocksMetadata", epoch),
		MiniblockHashByTxHashStorer: genericmocks.NewStorerMock("MiniblockHashByTxHash", epoch),
		EpochByHashStorer:           genericmocks.NewStorerMock("EpochByHash", epoch),
		Marshalizer:                 &mock.MarshalizerMock{},
		Hasher:                      &mock.HasherMock{},
	}

	return args
}

func TestNewHistoryRepository(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(0)
	args.MiniblocksMetadataStorer = nil
	repo, err := NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs(0)
	args.MiniblockHashByTxHashStorer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs(0)
	args.EpochByHashStorer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs(0)
	args.Hasher = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilHasher, err)

	args = createMockHistoryRepoArgs(0)
	args.Marshalizer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilMarshalizer, err)

	args = createMockHistoryRepoArgs(0)
	repo, err = NewHistoryRepository(args)
	require.Nil(t, err)
	require.NotNil(t, repo)
}

func TestHistoryRepository_RecordBlock(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(0)
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	headerHash := []byte("headerHash")
	blockHeader := &block.Header{
		Epoch: 0,
	}
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{[]byte("txA")},
				SenderShardID:   0,
				ReceiverShardID: 1,
			},
			{
				TxHashes:        [][]byte{[]byte("txB")},
				SenderShardID:   0,
				ReceiverShardID: 2,
			},
		},
	}

	err = repo.RecordBlock(headerHash, blockHeader, blockBody)
	require.Nil(t, err)
	// Two miniblocks
	require.Equal(t, 2, repo.miniblocksMetadataStorer.(*genericmocks.StorerMock).GetCurrentEpochData().Len())
	// One block, two miniblocks
	require.Equal(t, 3, repo.epochByHashIndex.storer.(*genericmocks.StorerMock).GetCurrentEpochData().Len())
	// Two transactions
	require.Equal(t, 2, repo.miniblockHashByTxHashIndex.(*genericmocks.StorerMock).GetCurrentEpochData().Len())
}

func TestHistoryRepository_GetMiniblockMetadata(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	miniblockA := &block.MiniBlock{
		Type:     block.TxBlock,
		TxHashes: [][]byte{[]byte("txA")},
	}
	miniblockB := &block.MiniBlock{
		Type:     block.InvalidBlock,
		TxHashes: [][]byte{[]byte("txB")},
	}

	repo.RecordBlock([]byte("fooblock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
				miniblockB,
			},
		},
	)

	metadata, err := repo.GetMiniblockMetadataByTxHash([]byte("txA"))
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4321, int(metadata.Round))

	metadata, err = repo.GetMiniblockMetadataByTxHash([]byte("txB"))
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4321, int(metadata.Round))

	metadata, err = repo.GetMiniblockMetadataByTxHash([]byte("foobar"))
	require.NotNil(t, err)
}

func TestHistoryRepository_GetEpochForHash(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	miniblockA := &block.MiniBlock{
		TxHashes: [][]byte{[]byte("txA")},
	}
	miniblockB := &block.MiniBlock{
		TxHashes: [][]byte{[]byte("txB")},
	}
	miniblockHashA, _ := repo.computeMiniblockHash(miniblockA)
	miniblockHashB, _ := repo.computeMiniblockHash(miniblockB)

	repo.RecordBlock([]byte("fooblock"), &block.Header{Epoch: 42}, &block.Body{
		MiniBlocks: []*block.MiniBlock{
			miniblockA,
			miniblockB,
		},
	})

	// Get epoch by block hash
	epoch, err := repo.GetEpochByHash([]byte("fooblock"))
	require.Nil(t, err)
	require.Equal(t, 42, int(epoch))

	// Get epoch by miniblock hash
	epoch, err = repo.GetEpochByHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 42, int(epoch))
	epoch, err = repo.GetEpochByHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 42, int(epoch))

	// Get epoch by transaction hash DOES NOT WORK (not needed)
	epoch, err = repo.GetEpochByHash([]byte("txA"))
	require.NotNil(t, err)
	epoch, err = repo.GetEpochByHash([]byte("txA"))
	require.NotNil(t, err)
}

func TestHistoryRepository_OnNotarizedBlocks(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	args.SelfShardID = 13
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	miniblockA := &block.MiniBlock{
		SenderShardID:   13,
		ReceiverShardID: 13,
		TxHashes:        [][]byte{[]byte("txA")},
	}
	miniblockB := &block.MiniBlock{
		SenderShardID:   13,
		ReceiverShardID: 14,
		TxHashes:        [][]byte{[]byte("txB")},
	}
	miniblockC := &block.MiniBlock{
		SenderShardID:   15,
		ReceiverShardID: 13,
		TxHashes:        [][]byte{[]byte("txC")},
	}
	miniblockHashA, _ := repo.computeMiniblockHash(miniblockA)
	miniblockHashB, _ := repo.computeMiniblockHash(miniblockB)
	miniblockHashC, _ := repo.computeMiniblockHash(miniblockC)

	// Let's have a block committed
	repo.RecordBlock([]byte("fooblock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
				miniblockB,
				miniblockC,
			},
		},
	)

	// Check "notarization coordinates"
	metadata, err := repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 0, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 0, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 0, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata.NotarizedAtDestinationInMetaNonce))

	// Now, let's receive a metablock that notarized a self-shard block, with miniblocks A and B
	metablock := &block.MetaBlock{
		Nonce: 4000,
		ShardInfo: []block.ShardData{
			{
				ShardID: 13,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   13,
						ReceiverShardID: 13,
						Hash:            miniblockHashA,
					},
					{
						SenderShardID:   13,
						ReceiverShardID: 14,
						Hash:            miniblockHashB,
					},
				}},
		},
	}

	repo.onNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockX")})

	// Check "notarization coordinates" again
	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 4000, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4000, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 4000, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 0, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata.NotarizedAtDestinationInMetaNonce))

	// Let's receive a metablock that notarized two shard blocks, with miniblocks B (at destination) and C (at source)
	metablock = &block.MetaBlock{
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				ShardID: 14,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   13,
						ReceiverShardID: 14,
						Hash:            miniblockHashB,
					},
				},
			},
			{
				ShardID: 15,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   15,
						ReceiverShardID: 13,
						Hash:            miniblockHashC,
					},
				},
			},
		},
	}

	repo.onNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockY")})

	// Check "notarization coordinates" again
	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 4000, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4000, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 4000, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4001, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata.NotarizedAtDestinationInMetaNonce))

	// Let's receive a metablock that notarized one shard block, with miniblock C (at destination)
	metablock = &block.MetaBlock{
		Nonce: 4002,
		ShardInfo: []block.ShardData{
			{
				ShardID: 13,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   15,
						ReceiverShardID: 13,
						Hash:            miniblockHashC,
					},
				},
			},
		},
	}

	repo.onNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockZ")})

	// Check "notarization coordinates" again
	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 4000, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4000, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 4000, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4001, int(metadata.NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4002, int(metadata.NotarizedAtDestinationInMetaNonce))
}

func TestHistoryRepository_OnNotarizedBlocksAtSourceBeforeCommittingAtDestination(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	args.SelfShardID = 14
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	// Assumming these miniblocks of transactions,
	miniblockA := &block.MiniBlock{
		SenderShardID:   12,
		ReceiverShardID: 14,
		TxHashes:        [][]byte{[]byte("txA")},
	}
	miniblockB := &block.MiniBlock{
		SenderShardID:   13,
		ReceiverShardID: 14,
		TxHashes:        [][]byte{[]byte("txB")},
	}
	miniblockHashA, _ := repo.computeMiniblockHash(miniblockA)
	miniblockHashB, _ := repo.computeMiniblockHash(miniblockB)

	// Let's receive a metablock and the notifications of "notarization at source"
	metablock := &block.MetaBlock{
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				ShardID: 12,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   12,
						ReceiverShardID: 14,
						Hash:            miniblockHashA,
					},
				},
			},
			{
				ShardID: 13,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   13,
						ReceiverShardID: 14,
						Hash:            miniblockHashB,
					},
				},
			},
		},
	}

	repo.onNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockFoo")})

	// Notifications have been queued
	require.Equal(t, 2, repo.pendingNotarizedAtSourceNotifications.Len())

	// Let's commit the blocks at destination
	repo.RecordBlock([]byte("fooBlock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
			},
		},
	)
	repo.RecordBlock([]byte("barBlock"),
		&block.Header{Epoch: 42, Round: 4322},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockB,
			},
		},
	)

	// Notifications have been cleared
	require.Equal(t, 0, repo.pendingNotarizedAtSourceNotifications.Len())

	// Check "notarization coordinates"
	metadata, err := repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtSourceInMetaHash)

	metadata, err = repo.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtSourceInMetaHash)
}

func TestHistoryRepository_OnNotarizedBlocksCrossEpoch(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	args.SelfShardID = 14
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	// Assumming one miniblock of transactions,
	miniblockA := &block.MiniBlock{
		SenderShardID:   14,
		ReceiverShardID: 14,
		TxHashes:        [][]byte{[]byte("txA")},
	}
	miniblockHashA, _ := repo.computeMiniblockHash(miniblockA)

	// Let's commit the intrashard block in epoch 42
	repo.RecordBlock([]byte("fooBlock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
			},
		},
	)

	// Now let's receive a metablock and the "notarized" notification, in the next epoch
	args.MiniblocksMetadataStorer.(*genericmocks.StorerMock).SetCurrentEpoch(43)
	metablock := &block.MetaBlock{
		Epoch: 43,
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				ShardID: 14,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   14,
						ReceiverShardID: 14,
						Hash:            miniblockHashA,
					},
				},
			},
		},
	}

	repo.onNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockFoo")})

	// Check "notarization coordinates"
	metadata, err := repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtSourceInMetaHash)
	require.Equal(t, 4001, int(metadata.NotarizedAtDestinationInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtDestinationInMetaHash)
}
