package dblookupext

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
	epochStartMocks "github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockHistoryRepoArgs(epoch uint32) HistoryRepositoryArguments {
	sp, _ := esdtSupply.NewSuppliesProcessor(&mock.MarshalizerMock{}, &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, storage.ErrKeyNotFound
		},
	}, &storageStubs.StorerStub{})

	args := HistoryRepositoryArguments{
		SelfShardID:                 0,
		MiniblocksMetadataStorer:    genericMocks.NewStorerMockWithEpoch(epoch),
		MiniblockHashByTxHashStorer: genericMocks.NewStorerMockWithEpoch(epoch),
		EpochByHashStorer:           genericMocks.NewStorerMockWithEpoch(epoch),
		EventsHashesByTxHashStorer:  genericMocks.NewStorerMockWithEpoch(epoch),
		BlockHashByRound:            genericMocks.NewStorerMockWithEpoch(epoch),
		Marshalizer:                 &mock.MarshalizerMock{},
		Hasher:                      &hashingMocks.HasherMock{},
		ESDTSuppliesHandler:         sp,
		Uint64ByteSliceConverter:    &epochStartMocks.Uint64ByteSliceConverterMock{},
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
	args.Uint64ByteSliceConverter = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, process.ErrNilUint64Converter, err)

	args = createMockHistoryRepoArgs(0)
	repo, err = NewHistoryRepository(args)
	require.Nil(t, err)
	require.NotNil(t, repo)
}

func TestHistoryRepository_RecordBlockInvalidBlockRoundByHashStorerExpectError(t *testing.T) {
	t.Parallel()

	errPut := errors.New("error put")
	args := createMockHistoryRepoArgs(0)
	args.BlockHashByRound = &storageStubs.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPut
		},
	}

	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	err = repo.RecordBlock([]byte("headerHash"), &block.Header{}, &block.Body{}, nil, nil, nil, nil)
	require.Equal(t, err, errPut)
}

func TestHistoryRepository_RecordBlock(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(0)
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	headerHash := []byte("headerHash")
	blockHeader := &block.Header{
		Nonce: 4,
		Round: 5,
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
	blockHeader.MiniBlockHeaders = []block.MiniBlockHeader{
		{
			Hash:            repo.computeMiniblockHash(blockBody.MiniBlocks[0]),
			SenderShardID:   0,
			ReceiverShardID: 1,
			TxCount:         uint32(len(blockBody.MiniBlocks[0].TxHashes)),
		},
		{
			Hash:            repo.computeMiniblockHash(blockBody.MiniBlocks[1]),
			SenderShardID:   0,
			ReceiverShardID: 2,
			TxCount:         uint32(len(blockBody.MiniBlocks[1].TxHashes)),
		},
	}

	err = repo.RecordBlock(headerHash, blockHeader, blockBody, nil, nil, nil, nil)
	require.Nil(t, err)
	// Two miniblocks
	require.Equal(t, 2, repo.miniblocksHandler.miniblocksMetadataStorer.(*genericMocks.StorerMock).GetCurrentEpochData().Len())
	// One block, two miniblocks
	require.Equal(t, 3, repo.epochByHashIndex.storer.(*genericMocks.StorerMock).GetCurrentEpochData().Len())
	// Two transactions
	require.Equal(t, 2, repo.miniblockHashByTxHashIndex.(*genericMocks.StorerMock).GetCurrentEpochData().Len())
	// One block, one block header
	require.Equal(t, 1, repo.blockHashByRound.(*genericMocks.StorerMock).GetCurrentEpochData().Len())
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

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			miniblockA,
			miniblockB,
		},
	}
	header := &block.Header{
		Epoch: 42,
		Round: 4321,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:    repo.computeMiniblockHash(miniblockA),
				TxCount: uint32(len(miniblockA.TxHashes)),
			},
			{
				Hash:    repo.computeMiniblockHash(miniblockB),
				TxCount: uint32(len(miniblockB.TxHashes)),
			},
		},
	}

	_ = repo.RecordBlock([]byte("fooblock"), header, body, nil, nil, nil, nil)

	metadata, err := repo.GetMiniblockMetadataByTxHash([]byte("txA"))
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4321, int(metadata.Round))

	metadata, err = repo.GetMiniblockMetadataByTxHash([]byte("txB"))
	require.Nil(t, err)
	require.Equal(t, 42, int(metadata.Epoch))
	require.Equal(t, 4321, int(metadata.Round))

	_, err = repo.GetMiniblockMetadataByTxHash([]byte("foobar"))
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
	miniblockHashA := repo.computeMiniblockHash(miniblockA)
	miniblockHashB := repo.computeMiniblockHash(miniblockB)

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			miniblockA,
			miniblockB,
		},
	}
	header := &block.Header{
		Epoch: 42,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:    repo.computeMiniblockHash(miniblockA),
				TxCount: uint32(len(miniblockA.TxHashes)),
			},
			{
				Hash:    repo.computeMiniblockHash(miniblockB),
				TxCount: uint32(len(miniblockB.TxHashes)),
			},
		},
	}
	_ = repo.RecordBlock([]byte("fooblock"), header, body, nil, nil, nil, nil)

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
	miniblockHashA := repo.computeMiniblockHash(miniblockA)
	miniblockHashB := repo.computeMiniblockHash(miniblockB)
	miniblockHashC := repo.computeMiniblockHash(miniblockC)

	header := &block.Header{
		Epoch: 42,
		Round: 4321,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:            miniblockHashA,
				SenderShardID:   miniblockA.SenderShardID,
				ReceiverShardID: miniblockA.ReceiverShardID,
				TxCount:         uint32(len(miniblockA.TxHashes)),
			},
			{
				Hash:            miniblockHashB,
				SenderShardID:   miniblockB.SenderShardID,
				ReceiverShardID: miniblockB.ReceiverShardID,
				TxCount:         uint32(len(miniblockB.TxHashes)),
			},
			{
				Hash:            miniblockHashC,
				SenderShardID:   miniblockC.SenderShardID,
				ReceiverShardID: miniblockC.ReceiverShardID,
				TxCount:         uint32(len(miniblockC.TxHashes)),
			},
		},
	}
	headerHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header)
	// Let's have a block committed
	_ = repo.RecordBlock(headerHash,
		header,
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
				miniblockB,
				miniblockC,
			},
		}, nil, nil, nil, nil,
	)

	// Check "notarization coordinates"
	metadata, err := repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 0, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 0, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 0, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	// Now, let's receive a metablock that notarized a self-shard block, with miniblocks A and B
	metablock := &block.MetaBlock{
		Nonce: 4000,
		ShardInfo: []block.ShardData{
			{
				HeaderHash: headerHash,
				ShardID:    13,
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
	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 0, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	// Let's receive a metablock that notarized two shard blocks, with miniblocks B (at destination) and C (at source)
	metablock = &block.MetaBlock{
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				HeaderHash: headerHash,
				ShardID:    14,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						SenderShardID:   13,
						ReceiverShardID: 14,
						Hash:            miniblockHashB,
					},
				},
			},
			{
				HeaderHash: headerHash,
				ShardID:    15,
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
	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4001, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4001, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 0, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	// Let's receive a metablock that notarized one shard block, with miniblock C (at destination)
	metablock = &block.MetaBlock{
		Nonce: 4002,
		ShardInfo: []block.ShardData{
			{
				HeaderHash: headerHash,
				ShardID:    13,
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
	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashB)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4000, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4001, int(metadata[0].NotarizedAtDestinationInMetaNonce))

	metadata, err = repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashC)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 4001, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, 4002, int(metadata[0].NotarizedAtDestinationInMetaNonce))
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
	miniblockHashA := repo.computeMiniblockHash(miniblockA)

	header := &block.Header{
		Epoch: 42,
		Round: 4321,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				SenderShardID:   miniblockA.SenderShardID,
				ReceiverShardID: miniblockA.ReceiverShardID,
				Hash:            miniblockHashA,
				TxCount:         uint32(len(miniblockA.TxHashes)),
			},
		},
	}
	headerHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, header)

	// Let's commit the intrashard block in epoch 42
	_ = repo.RecordBlock(
		headerHash,
		header,
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
			},
		}, nil, nil, nil, nil,
	)

	// Now let's receive a metablock and the "notarized" notification, in the next epoch
	args.MiniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(43)
	metablock := &block.MetaBlock{
		Epoch: 43,
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				HeaderHash: headerHash,
				ShardID:    14,
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
	metadata, err := repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHashA)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 42, int(metadata[0].Epoch))
	require.Equal(t, 4001, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata[0].NotarizedAtSourceInMetaHash)
	require.Equal(t, 4001, int(metadata[0].NotarizedAtDestinationInMetaNonce))
	require.Equal(t, []byte("metablockFoo"), metadata[0].NotarizedAtDestinationInMetaHash)
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
	miniblockHash := repo.computeMiniblockHash(miniblock)

	forkedHeader := &block.Header{
		Epoch: 42,
		Round: 4321,
		Nonce: 4300,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				SenderShardID:   miniblock.SenderShardID,
				ReceiverShardID: miniblock.ReceiverShardID,
				Hash:            miniblockHash,
				TxCount:         uint32(len(miniblock.TxHashes)),
			},
		},
	}
	forkedHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, forkedHeader)

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			miniblock,
		},
	}

	// Let's commit the intrashard block in epoch 42, on a fork
	_ = repo.RecordBlock(forkedHeaderHash,
		forkedHeader,
		body, nil, nil, nil, nil,
	)

	_ = repo.RevertBlock(forkedHeader, body)

	// Let's go to next epoch
	args.MiniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(43)

	canonicalHeader := &block.Header{
		Epoch: 43,
		Round: 4400,
		Nonce: 4300,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				SenderShardID:   miniblock.SenderShardID,
				ReceiverShardID: miniblock.ReceiverShardID,
				Hash:            miniblockHash,
				TxCount:         uint32(len(miniblock.TxHashes)),
			},
		},
	}
	canonicalHeaderHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, canonicalHeader)

	// Let's commit a block with the same miniblock, in the next epoch, on the canonical chain
	_ = repo.RecordBlock(canonicalHeaderHash,
		canonicalHeader,
		body, nil, nil, nil, nil,
	)

	// Now let's receive a metablock and the "notarized" notification
	metablock := &block.MetaBlock{
		Epoch: 43,
		Nonce: 4001,
		ShardInfo: []block.ShardData{
			{
				ShardID:    14,
				HeaderHash: canonicalHeaderHash,
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
	metablockHash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, metablock)
	repo.OnNotarizedBlocks(core.MetachainShardId, []data.HeaderHandler{metablock}, [][]byte{metablockHash})

	// Check "notarization coordinates"
	metadata, err := repo.miniblocksHandler.getMiniblockMetadataByMiniblockHash(miniblockHash)
	require.Nil(t, err)
	require.Equal(t, 1, len(metadata))
	require.Equal(t, 43, int(metadata[0].Epoch))
	require.Equal(t, 4001, int(metadata[0].NotarizedAtSourceInMetaNonce))
	require.Equal(t, metablockHash, metadata[0].NotarizedAtSourceInMetaHash)
	require.Equal(t, 4001, int(metadata[0].NotarizedAtDestinationInMetaNonce))
	require.Equal(t, metablockHash, metadata[0].NotarizedAtDestinationInMetaHash)

	// Epoch 42 will not contain data because it was reverted
	orphanedMetadata := &MiniblockMetadata{}
	err = args.MiniblocksMetadataStorer.(*genericMocks.StorerMock).GetFromEpochWithMarshalizer(miniblockHash, 42, orphanedMetadata, args.Marshalizer)
	require.NotNil(t, err)
}

func TestHistoryRepository_getMiniblockMetadataByMiniblockHashGetFromEpochErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	epochIndex := newHashToEpochIndex(
		&storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return []byte("{}"), nil
			},
		},
		&mock.MarshalizerMock{},
	)
	hr := &historyRepository{
		epochByHashIndex: epochIndex,
		miniblocksHandler: &miniblocksHandler{
			miniblocksMetadataStorer: &storageStubs.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					return nil, expectedErr
				},
			},
			epochIndex: epochIndex,
		},
	}

	mbMetadata, err := hr.miniblocksHandler.getMiniblockMetadataByMiniblockHash([]byte("hash"))
	assert.Nil(t, mbMetadata)
	assert.Equal(t, err, expectedErr)
}
