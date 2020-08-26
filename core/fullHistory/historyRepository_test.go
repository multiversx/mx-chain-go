package fullHistory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/require"
)

func createMockHistoryRepoArgs() HistoryRepositoryArguments {
	return HistoryRepositoryArguments{
		SelfShardID:                 0,
		MiniblocksMetadataStorer:    genericmocks.NewStorerMock(),
		MiniblockHashByTxHashStorer: genericmocks.NewStorerMock(),
		EpochByHashStorer:           genericmocks.NewStorerMock(),
		Marshalizer:                 &mock.MarshalizerMock{},
		Hasher:                      &mock.HasherMock{},
	}
}

func TestNewHistoryRepository(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
	args.MiniblocksMetadataStorer = nil
	repo, err := NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs()
	args.MiniblockHashByTxHashStorer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs()
	args.EpochByHashStorer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilStore, err)

	args = createMockHistoryRepoArgs()
	args.Hasher = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilHasher, err)

	args = createMockHistoryRepoArgs()
	args.Marshalizer = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, core.ErrNilMarshalizer, err)

	args = createMockHistoryRepoArgs()
	repo, err = NewHistoryRepository(args)
	require.Nil(t, err)
	require.NotNil(t, repo)
}

func TestHistoryRepository_RecordBlock(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
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
	require.Equal(t, 2, len(repo.miniblocksMetadataStorer.(*genericmocks.StorerMock).Data))
	// One block, two miniblocks, two transactions
	require.Equal(t, 5, len(repo.epochByHashIndex.storer.(*genericmocks.StorerMock).Data))
	// Two transactions
	require.Equal(t, 2, len(repo.miniblockHashByTxHashIndex.(*genericmocks.StorerMock).Data))
}

func TestHistoryRepository_GetMiniblockMetadata(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
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

	args := createMockHistoryRepoArgs()
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

	// Get epoch by transaction hash
	epoch, err = repo.GetEpochByHash([]byte("txA"))
	require.Nil(t, err)
	require.Equal(t, 42, int(epoch))
	epoch, err = repo.GetEpochByHash([]byte("txA"))
	require.Nil(t, err)
	require.Equal(t, 42, int(epoch))
}

func TestHistoryRepository_OnNotarizedBlocks(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs()
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

	args := createMockHistoryRepoArgs()
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
	require.Equal(t, 2, repo.notarizedAtSourceNotifications.Len())

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
	require.Equal(t, 0, repo.notarizedAtSourceNotifications.Len())

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
