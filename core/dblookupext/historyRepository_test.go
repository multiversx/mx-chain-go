package dblookupext

import (
	"sync"
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
		EventsHashesByTxHashStorer:  genericmocks.NewStorerMock("EventsHashesByTxHash", epoch),
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
	args.EventsHashesByTxHashStorer = nil
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

	err = repo.RecordBlock(headerHash, blockHeader, blockBody, nil, nil)
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

	_ = repo.RecordBlock([]byte("fooblock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
				miniblockB,
			},
		},
		nil, nil,
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

	_ = repo.RecordBlock([]byte("fooblock"), &block.Header{Epoch: 42}, &block.Body{
		MiniBlocks: []*block.MiniBlock{
			miniblockA,
			miniblockB,
		},
	}, nil, nil)

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
	_, err = repo.GetEpochByHash([]byte("txA"))
	require.NotNil(t, err)
	_, err = repo.GetEpochByHash([]byte("txA"))
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
	_ = repo.RecordBlock([]byte("fooblock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
				miniblockB,
				miniblockC,
			},
		}, nil, nil,
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

	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockX")})

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

	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockY")})

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

	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockZ")})

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

	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockFoo")})

	// Notifications have been queued
	require.Equal(t, 2, repo.pendingNotarizedAtSourceNotifications.Len())

	// Let's commit the blocks at destination
	_ = repo.RecordBlock([]byte("fooBlock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
			},
		}, nil, nil,
	)
	_ = repo.RecordBlock([]byte("barBlock"),
		&block.Header{Epoch: 42, Round: 4322},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockB,
			},
		}, nil, nil,
	)

	// Notifications have not been cleared after record block
	require.Equal(t, 2, repo.pendingNotarizedAtSourceNotifications.Len())

	// Now receive any new notarization notification
	repo.OnNotarizedBlocks(42, []data.HeaderHandler{}, [][]byte{[]byte("nothing")})

	// Notifications have been processed & cleared
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
	_ = repo.RecordBlock([]byte("fooBlock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
			},
		}, nil, nil,
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

	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockFoo")})

	// Check "notarization coordinates"
	metadata, err := repo.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtSourceInMetaHash)
	require.Equal(t, 4001, int(metadata.NotarizedAtDestinationInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtDestinationInMetaHash)
}

func TestHistoryRepository_CommitOnForkThenNewEpochThenCommit(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	args.SelfShardID = 14
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	// Assumming one miniblock of transactions,
	miniblock := &block.MiniBlock{
		SenderShardID:   14,
		ReceiverShardID: 14,
		TxHashes:        [][]byte{[]byte("txA")},
	}
	miniblockHash, _ := repo.computeMiniblockHash(miniblock)

	// Let's commit the intrashard block in epoch 42, on a fork
	_ = repo.RecordBlock([]byte("fooOnFork"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblock,
			},
		}, nil, nil,
	)

	// Let's go to next epoch
	args.MiniblocksMetadataStorer.(*genericmocks.StorerMock).SetCurrentEpoch(43)

	// Let's commit a block with the same miniblock, in the next epoch, on the canonical chain
	_ = repo.RecordBlock([]byte("fooOnChain"),
		&block.Header{Epoch: 43, Round: 4350},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblock,
			},
		}, nil, nil,
	)

	// Now let's receive a metablock and the "notarized" notification
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
						Hash:            miniblockHash,
					},
				},
			},
		},
	}
	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockFoo")})

	// Check "notarization coordinates"
	metadata, err := repo.getMiniblockMetadataByMiniblockHash(miniblockHash)
	require.Nil(t, err)
	require.Equal(t, 43, int(metadata.Epoch))
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtSourceInMetaHash)
	require.Equal(t, 4001, int(metadata.NotarizedAtDestinationInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtDestinationInMetaHash)

	// Epoch 42 will contain an orphaned record
	orphanedMetadata := &MiniblockMetadata{}
	_ = args.MiniblocksMetadataStorer.(*genericmocks.StorerMock).GetFromEpochWithMarshalizer(miniblockHash, 42, orphanedMetadata, args.Marshalizer)

	require.Equal(t, 42, int(orphanedMetadata.Epoch))
	require.Equal(t, 0, int(orphanedMetadata.NotarizedAtSourceInMetaNonce))
	require.Nil(t, orphanedMetadata.NotarizedAtSourceInMetaHash)
	require.Equal(t, 0, int(orphanedMetadata.NotarizedAtDestinationInMetaNonce))
	require.Nil(t, orphanedMetadata.NotarizedAtDestinationInMetaHash)
}

func TestHistoryRepository_ConcurrentlyRecordAndNotarizeSameBlockMultipleTimes_Loop(t *testing.T) {
	for i := 0; i < 100; i++ {
		TestHistoryRepository_ConcurrentlyRecordAndNotarizeSameBlockMultipleTimes(t)
	}
}

func TestHistoryRepository_ConcurrentlyRecordAndNotarizeSameBlockMultipleTimes(t *testing.T) {
	args := createMockHistoryRepoArgs(42)
	args.SelfShardID = 14
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	miniblock := &block.MiniBlock{
		SenderShardID:   14,
		ReceiverShardID: 14,
		TxHashes:        [][]byte{[]byte("txA")},
	}
	miniblockHash, _ := repo.computeMiniblockHash(miniblock)

	metablock := &block.MetaBlock{
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				ShardID: 14,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   14,
						ReceiverShardID: 14,
						Hash:            miniblockHash,
					},
				},
			},
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		// Simulate commit & rollback
		for i := 0; i < 500; i++ {
			_ = repo.RecordBlock([]byte("fooBlock"),
				&block.Header{Epoch: 42, Round: 4321},
				&block.Body{
					MiniBlocks: []*block.MiniBlock{
						miniblock,
					},
				}, nil, nil,
			)
		}

		wg.Done()
	}()

	go func() {
		// Receive less notifications (to test more aggressively)
		for i := 0; i < 50; i++ {
			repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{[]byte("metablockFoo")})
		}

		wg.Done()
	}()

	wg.Wait()

	// Simulate continuation of the blockchain (so that any pending notifications are consumed)
	repo.OnNotarizedBlocks(42, []data.HeaderHandler{}, [][]byte{[]byte("nothing")})

	metadata, err := repo.getMiniblockMetadataByMiniblockHash(miniblockHash)
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4321, int(metadata.Round))
	require.Equal(t, 4001, int(metadata.NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtSourceInMetaHash)
	require.Equal(t, 4001, int(metadata.NotarizedAtDestinationInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata.NotarizedAtDestinationInMetaHash)
}
