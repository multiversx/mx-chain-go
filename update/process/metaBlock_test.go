package process

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/assert"
)

func createMockBlockCreatorAfterHardFork() ArgsNewMetaBlockCreatorAfterHardfork {
	return ArgsNewMetaBlockCreatorAfterHardfork{
		ImportHandler:    &mock.ImportHandlerStub{},
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &mock.HasherMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
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

	rootHash1 := []byte("rootHash1")
	metaBlock := &block.MetaBlock{}
	args := createMockBlockCreatorAfterHardFork()
	args.ImportHandler = &mock.ImportHandlerStub{
		GetAccountsDBForShardCalled: func(shardID uint32) state.AccountsAdapter {
			return &mock.AccountsStub{
				CommitCalled: func() ([]byte, error) {
					return rootHash1, nil
				},
			}
		},
		GetHardForkMetaBlockCalled: func() *block.MetaBlock {
			return metaBlock
		},
	}

	blockCreator, _ := NewMetaBlockCreatorAfterHardfork(args)

	chainID, round, nonce, epoch := "id", uint64(10), uint64(12), uint32(1)
	body, _, err := blockCreator.CreateBody()
	assert.NoError(t, err)
	header, err := blockCreator.CreateBlock(body, chainID, round, nonce, epoch)
	assert.NoError(t, err)

	blockBody := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}
	validatorRootHash, _ := args.ValidatorAccounts.Commit()
	metaHdr := &block.MetaBlock{
		Nonce:                  nonce,
		Round:                  round,
		PrevRandSeed:           rootHash1,
		RandSeed:               rootHash1,
		RootHash:               rootHash1,
		ValidatorStatsRootHash: validatorRootHash,
		EpochStart:             block.EpochStart{},
		ChainID:                []byte(chainID),
		SoftwareVersion:        []byte(""),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		PubKeysBitmap:          []byte{1},
		Epoch:                  epoch,
	}
	assert.Equal(t, blockBody, body)
	assert.Equal(t, metaHdr, header)
}
