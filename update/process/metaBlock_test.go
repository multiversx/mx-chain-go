package process

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
)

func createMockBlockCreatorAfterHardFork() ArgsNewMetaBlockCreatorAfterHardFork {
	return ArgsNewMetaBlockCreatorAfterHardFork{
		Hasher:             &mock.HasherMock{},
		ImportHandler:      &mock.ImportHandlerStub{},
		Marshalizer:        &mock.MarshalizerMock{},
		PendingTxProcessor: &mock.PendingTransactionProcessorStub{},
		ShardCoordinator:   mock.NewOneShardCoordinatorMock(),
		Storage:            initStore(),
		TxCoordinator:      &mock.TransactionCoordinatorMock{},
		ValidatorAccounts: &mock.AccountsStub{
			CommitCalled: func() ([]byte, error) {
				return []byte("roothash"), nil
			},
		},
	}
}
func TestNewMetaBlockCreatorAfterHardfork_NilImport(t *testing.T) {
	t.Parallel()

	args := createMockBlockCreatorAfterHardFork()
	args.ImportHandler = nil

	blockCreator, err := NewMetaBlockCreatorAfterHardfork(args)
	assert.Nil(t, blockCreator)
	assert.Equal(t, update.ErrNilImportHandler, err)
}

func TestNewMetaBlockCreatorAfterHardfork_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockBlockCreatorAfterHardFork()
	args.Marshalizer = nil

	blockCreator, err := NewMetaBlockCreatorAfterHardfork(args)
	assert.Nil(t, blockCreator)
	assert.Equal(t, update.ErrNilMarshalizer, err)
}

func TestNewMetaBlockCreatorAfterHardfork_NilHasher(t *testing.T) {
	t.Parallel()

	args := createMockBlockCreatorAfterHardFork()
	args.Hasher = nil

	blockCreator, err := NewMetaBlockCreatorAfterHardfork(args)
	assert.Nil(t, blockCreator)
	assert.Equal(t, update.ErrNilHasher, err)
}

func TestNewMetaBlockCreatorAfterHardfork_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := createMockBlockCreatorAfterHardFork()
	args.ShardCoordinator = nil

	blockCreator, err := NewMetaBlockCreatorAfterHardfork(args)
	assert.Nil(t, blockCreator)
	assert.Equal(t, update.ErrNilShardCoordinator, err)
}

func TestNewMetaBlockCreatorAfterHardforkShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockBlockCreatorAfterHardFork()

	blockCreator, err := NewMetaBlockCreatorAfterHardfork(args)
	assert.NoError(t, err)
	assert.False(t, check.IfNil(blockCreator))
}

func TestMetaBlockCreator_CreateBlock(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	hashTx1, hashTx2 := []byte("hash1"), []byte("hash2")
	mb1 := &block.MiniBlock{SenderShardID: 1, ReceiverShardID: 2, TxHashes: [][]byte{hashTx1}}
	mb2 := &block.MiniBlock{SenderShardID: 3, ReceiverShardID: 4, TxHashes: [][]byte{hashTx2}}
	args := createMockBlockCreatorAfterHardFork()

	args.PendingTxProcessor = &mock.PendingTransactionProcessorStub{
		ProcessTransactionsDstMeCalled: func(mbInfo *update.MbInfo) (*block.MiniBlock, error) {
			return mb1, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	args.TxCoordinator = &mock.TransactionCoordinatorMock{
		CreatePostProcessMiniBlocksCalled: func() block.MiniBlockSlice {
			return block.MiniBlockSlice{mb2}
		},
		GetAllCurrentUsedTxsCalled: func(blockType block.Type) map[string]data.TransactionHandler {
			return map[string]data.TransactionHandler{
				string(hashTx2): &transaction.Transaction{},
			}
		},
	}

	epochStart := block.EpochStart{
		LastFinalizedHeaders: []block.EpochStartShardData{
			{
				FirstPendingMetaBlock: []byte("metaBlock_hash"),
				PendingMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash: []byte("miniBlock_hash"),
					},
				},
			},
		},
	}

	metaBlock := &block.MetaBlock{
		Round:      2,
		EpochStart: epochStart,
	}
	unFinishedMetaBlocks := map[string]*block.MetaBlock{
		"metaBlock_hash": {Round: 1},
	}
	args.ImportHandler = &mock.ImportHandlerStub{
		GetAccountsDBForShardCalled: func(shardID uint32) state.AccountsAdapter {
			return &mock.AccountsStub{
				CommitCalled: func() ([]byte, error) {
					return rootHash, nil
				},
			}
		},
		GetHardForkMetaBlockCalled: func() *block.MetaBlock {
			return metaBlock
		},
		GetUnFinishedMetaBlocksCalled: func() map[string]*block.MetaBlock {
			return unFinishedMetaBlocks
		},
		GetMiniBlocksCalled: func() map[string]*block.MiniBlock {
			return map[string]*block.MiniBlock{
				"miniBlock_hash": {},
			}
		},
	}

	blockCreator, _ := NewMetaBlockCreatorAfterHardfork(args)

	chainID, round, nonce, epoch := "id", uint64(10), uint64(12), uint32(1)
	body, _, err := blockCreator.CreateBody()
	assert.NoError(t, err)
	header, err := blockCreator.CreateBlock(body, chainID, round, nonce, epoch)
	assert.NoError(t, err)

	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{mb1, mb2},
	}
	mb1Hash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb1)
	mb2Hash, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb2)

	expectedMbHeaders := []block.MiniBlockHeader{
		{
			Hash:            mb1Hash,
			TxCount:         1,
			Type:            0,
			SenderShardID:   1,
			ReceiverShardID: 2,
		},
		{
			Hash:            mb2Hash,
			TxCount:         1,
			Type:            0,
			SenderShardID:   3,
			ReceiverShardID: 4,
		},
	}
	validatorRootHash, _ := args.ValidatorAccounts.Commit()
	expectedHeader := &block.MetaBlock{
		Nonce:                  nonce,
		Round:                  round,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: validatorRootHash,
		EpochStart:             epochStart,
		ChainID:                []byte(chainID),
		SoftwareVersion:        []byte(""),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		PubKeysBitmap:          []byte{1},
		Epoch:                  epoch,
		MiniBlockHeaders:       expectedMbHeaders,
		ReceiptsHash:           []byte("receiptHash"),
		TxCount:                2,
	}
	assert.Equal(t, expectedBody, body)
	assert.Equal(t, expectedHeader, header)
}
