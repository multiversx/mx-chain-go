package dblookupext

import (
	"errors"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
	epochStartMocks "github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
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
		ExecutionResultsStorer:      genericMocks.NewStorerMockWithEpoch(epoch),
		Marshalizer:                 &mock.MarshalizerMock{},
		Hasher:                      &hashingMocks.HasherMock{},
		ESDTSuppliesHandler:         sp,
		Uint64ByteSliceConverter:    &epochStartMocks.Uint64ByteSliceConverterMock{},
		DataPool:                    &dataRetrieverMock.PoolsHolderMock{},
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
	args.DataPool = nil
	repo, err = NewHistoryRepository(args)
	require.Nil(t, repo)
	require.Equal(t, process.ErrNilDataPoolHolder, err)

	args = createMockHistoryRepoArgs(0)
	repo, err = NewHistoryRepository(args)
	require.Nil(t, err)
	require.NotNil(t, repo)
}

func TestHistoryRepository_RecordBlockErrCannotCastToBlockBody(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(0)

	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	err = repo.RecordBlock([]byte("headerHash"), &block.Header{}, nil, nil, nil, nil, nil)
	require.Equal(t, errCannotCastToBlockBody, err)
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

	err = repo.RecordBlock(headerHash, blockHeader, blockBody, nil, nil, nil, nil)
	require.Nil(t, err)
	// Two miniblocks
	require.Equal(t, 2, repo.miniblocksMetadataStorer.(*genericMocks.StorerMock).GetCurrentEpochData().Len())
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

	_ = repo.RecordBlock([]byte("fooblock"),
		&block.Header{Epoch: 42, Round: 4321},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockA,
				miniblockB,
			},
		},
		nil, nil, nil, nil,
	)

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
	miniblockHashA, _ := repo.computeMiniblockHash(miniblockA)
	miniblockHashB, _ := repo.computeMiniblockHash(miniblockB)

	_ = repo.RecordBlock([]byte("fooblock"), &block.Header{Epoch: 42}, &block.Body{
		MiniBlocks: []*block.MiniBlock{
			miniblockA,
			miniblockB,
		},
	}, nil, nil, nil, nil)

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
		}, nil, nil, nil, nil,
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
		}, nil, nil, nil, nil,
	)
	_ = repo.RecordBlock([]byte("barBlock"),
		&block.Header{Epoch: 42, Round: 4322},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblockB,
			},
		}, nil, nil, nil, nil,
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
		}, nil, nil, nil, nil,
	)

	// Now let's receive a metablock and the "notarized" notification, in the next epoch
	args.MiniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(43)
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
		}, nil, nil, nil, nil,
	)

	// Let's go to next epoch
	args.MiniblocksMetadataStorer.(*genericMocks.StorerMock).SetCurrentEpoch(43)

	// Let's commit a block with the same miniblock, in the next epoch, on the canonical chain
	_ = repo.RecordBlock([]byte("fooOnChain"),
		&block.Header{Epoch: 43, Round: 4350},
		&block.Body{
			MiniBlocks: []*block.MiniBlock{
				miniblock,
			},
		}, nil, nil, nil, nil,
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
	_ = args.MiniblocksMetadataStorer.(*genericMocks.StorerMock).GetFromEpochWithMarshalizer(miniblockHash, 42, orphanedMetadata, args.Marshalizer)

	require.Equal(t, 42, int(orphanedMetadata.Epoch))
	require.Equal(t, 0, int(orphanedMetadata.NotarizedAtSourceInMetaNonce))
	require.Nil(t, orphanedMetadata.NotarizedAtSourceInMetaHash)
	require.Equal(t, 0, int(orphanedMetadata.NotarizedAtDestinationInMetaNonce))
	require.Nil(t, orphanedMetadata.NotarizedAtDestinationInMetaHash)
}

func TestHistoryRepository_getMiniblockMetadataByMiniblockHashGetFromEpochErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	hr := &historyRepository{
		epochByHashIndex: newHashToEpochIndex(
			&storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return []byte("{}"), nil
				},
			},
			&mock.MarshalizerMock{},
		),
		miniblocksMetadataStorer: &storageStubs.StorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		},
	}

	mbMetadata, err := hr.getMiniblockMetadataByMiniblockHash([]byte("hash"))
	assert.Nil(t, mbMetadata)
	assert.Equal(t, err, expectedErr)
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
				}, nil, nil, nil, nil,
			)
		}

		wg.Done()
	}()

	go func() {
		// Receive fewer notifications (to test more aggressively)
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

func TestRecordHeaderV3(t *testing.T) {
	t.Parallel()

	t.Run("record block v3 should work no execution results", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)
		repo, err := NewHistoryRepository(args)
		require.Nil(t, err)

		header := &block.HeaderV3{}
		body := &block.Body{}

		headerHash := []byte("headerHash")
		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.Nil(t, err)
	})

	t.Run("record block v3 should work", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)
		args.DataPool = dataRetrieverMock.NewPoolsHolderMock()
		repo, err := NewHistoryRepository(args)
		require.Nil(t, err)

		executionResultHeaderHash := []byte("executionResultHeaderHash")
		mb := &block.MiniBlock{SenderShardID: 0}
		mbHash1, _ := repo.computeMiniblockHash(mb)
		header := &block.HeaderV3{
			Nonce: 100,
			Round: 101,
			Epoch: 42,
			ExecutionResults: []*block.ExecutionResult{
				{

					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash:  executionResultHeaderHash,
						HeaderNonce: 99,
						HeaderRound: 100,
						HeaderEpoch: 42,
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: mbHash1,
						},
					},
				},
			},
		}

		mbBytes, _ := repo.marshalizer.Marshal(mb)
		args.DataPool.ExecutedMiniBlocks().Put(mbHash1, mbBytes, 1)

		cachedIntermediateTxsMap := map[block.Type]map[string]data.TransactionHandler{}
		cachedIntermediateTxsMap[block.SmartContractResultBlock] = map[string]data.TransactionHandler{
			"h1": &smartContractResult.SmartContractResult{},
		}
		cachedIntermediateTxsMap[block.ReceiptBlock] = map[string]data.TransactionHandler{
			"r1": &receipt.Receipt{},
		}
		args.DataPool.PostProcessTransactions().Put(executionResultHeaderHash, cachedIntermediateTxsMap, 1)

		expectedLogs := []*data.LogData{
			{
				LogHandler: &transaction.Log{},
				TxHash:     "t1",
			},
		}
		logsKey := common.PrepareLogEventsKey(executionResultHeaderHash)
		args.DataPool.PostProcessTransactions().Put(logsKey, expectedLogs, 1)

		expectedMbs := []*block.MiniBlock{
			{SenderShardID: 0},
		}
		intraMbsBytes, _ := repo.marshalizer.Marshal(expectedMbs)

		args.DataPool.ExecutedMiniBlocks().Put(executionResultHeaderHash, intraMbsBytes, 0)

		body := &block.Body{}

		headerHash := []byte("headerHash")
		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.Nil(t, err)

		epoch, err := repo.GetEpochByHash(executionResultHeaderHash)
		require.Nil(t, err)
		require.Equal(t, 42, int(epoch))

		epoch, err = repo.GetEpochByHash(mbHash1)
		require.Nil(t, err)
		require.Equal(t, 42, int(epoch))
	})

	t.Run("record block v3 should error because the headerHash is not found in cache", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)
		args.DataPool = dataRetrieverMock.NewPoolsHolderMock()
		repo, err := NewHistoryRepository(args)

		executionResultHeaderHash := []byte("executionResultHeaderHash")
		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: executionResultHeaderHash,
					},
				},
			},
		}

		body := &block.Body{}
		headerHash := []byte("headerHash")
		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.NotNil(t, err)
		require.ErrorContains(t, err, process.ErrMissingHeader.Error())
	})

	t.Run("record block v3 should error because logs were not found in dataPool", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)
		args.DataPool = dataRetrieverMock.NewPoolsHolderMock()
		repo, err := NewHistoryRepository(args)

		executionResultHeaderHash := []byte("executionResultHeaderHash")
		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: executionResultHeaderHash,
					},
				},
			},
		}

		cachedIntermediateTxsMap := map[block.Type]map[string]data.TransactionHandler{}
		args.DataPool.PostProcessTransactions().Put(executionResultHeaderHash, cachedIntermediateTxsMap, 1)

		body := &block.Body{}
		headerHash := []byte("headerHash")
		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.NotNil(t, err)
		require.ErrorContains(t, err, process.ErrMissingHeader.Error())
	})

	t.Run("record block v3 should error because mini blocks were not cached", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)
		args.DataPool = dataRetrieverMock.NewPoolsHolderMock()
		repo, err := NewHistoryRepository(args)

		executionResultHeaderHash := []byte("executionResultHeaderHash")
		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: executionResultHeaderHash,
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: []byte("mbHash"),
						},
					},
				},
			},
		}

		cachedIntermediateTxsMap := map[block.Type]map[string]data.TransactionHandler{}
		args.DataPool.PostProcessTransactions().Put(executionResultHeaderHash, cachedIntermediateTxsMap, 1)

		expectedLogs := []*data.LogData{
			{
				LogHandler: &transaction.Log{},
				TxHash:     "t1",
			},
		}
		logsKey := common.PrepareLogEventsKey(executionResultHeaderHash)
		args.DataPool.PostProcessTransactions().Put(logsKey, expectedLogs, 1)

		body := &block.Body{}
		headerHash := []byte("headerHash")
		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.NotNil(t, err)
		require.ErrorContains(t, err, process.ErrMissingMiniBlock.Error())
	})

	t.Run("record block v3 should error because intra mini blocks were not found", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)
		args.DataPool = dataRetrieverMock.NewPoolsHolderMock()
		repo, err := NewHistoryRepository(args)

		executionResultHeaderHash := []byte("executionResultHeaderHash")
		mb := &block.MiniBlock{SenderShardID: 0}
		mbHash1, _ := repo.computeMiniblockHash(mb)
		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: executionResultHeaderHash,
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{
							Hash: mbHash1,
						},
					},
				},
			},
		}

		mbBytes, _ := repo.marshalizer.Marshal(mb)
		args.DataPool.ExecutedMiniBlocks().Put(mbHash1, mbBytes, 1)

		cachedIntermediateTxsMap := map[block.Type]map[string]data.TransactionHandler{}
		args.DataPool.PostProcessTransactions().Put(executionResultHeaderHash, cachedIntermediateTxsMap, 1)

		expectedLogs := []*data.LogData{
			{
				LogHandler: &transaction.Log{},
				TxHash:     "t1",
			},
		}
		logsKey := common.PrepareLogEventsKey(executionResultHeaderHash)
		args.DataPool.PostProcessTransactions().Put(logsKey, expectedLogs, 1)

		body := &block.Body{}
		headerHash := []byte("headerHash")
		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.NotNil(t, err)
		require.ErrorContains(t, err, process.ErrMissingHeader.Error())
	})

	t.Run("record block v3 should error because of Marshal on recordBlock", func(t *testing.T) {
		args := createMockHistoryRepoArgs(42)

		expectedError := errors.New("expected error")
		args.Marshalizer = &mock.MarshalizerMock{MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedError
		}}

		args.DataPool = dataRetrieverMock.NewPoolsHolderMock()
		repo, err := NewHistoryRepository(args)

		header := &block.HeaderV3{}
		body := &block.Body{}

		headerHash := []byte("headerHash")

		err = repo.RecordBlock(headerHash, header, body, nil, nil, nil, nil)
		require.ErrorContains(t, err, expectedError.Error())
	})
}

func TestHistoryRepository_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createMockHistoryRepoArgs(42)
	repo, err := NewHistoryRepository(args)
	require.Nil(t, err)

	require.False(t, repo.IsInterfaceNil())

	var nilRepo *historyRepository
	require.True(t, nilRepo.IsInterfaceNil())
}
