package blockAPI

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/assert"
)

func createMockShardAPIProcessor(
	shardID uint32,
	blockHeaderHash []byte,
	storerMock *genericMocks.StorerMock,
	withHistory bool,
	withKey bool,
) *shardAPIBlockProcessor {
	return newShardApiBlockProcessor(&ArgAPIBlockProcessor{
		APITransactionHandler: &mock.TransactionAPIHandlerStub{},
		SelfShardID:           shardID,
		Marshalizer:           &mock.MarshalizerFake{},
		Store: &mock.ChainStorerMock{
			GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
				return storerMock
			},
			GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
				if withKey {
					return storerMock.Get(key)
				}
				return blockHeaderHash, nil
			},
		},
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		HistoryRepo: &dblookupext.HistoryRepositoryStub{
			GetEpochByHashCalled: func(hash []byte) (uint32, error) {
				return 1, nil
			},
			IsEnabledCalled: func() bool {
				return withHistory
			},
		},
		ReceiptsRepository: &testscommon.ReceiptsRepositoryStub{},
	}, nil)
}

func TestShardAPIBlockProcessor_GetBlockByHashInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	shardID := uint32(3)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		true,
		false,
	)

	blk, err := shardAPIBlockProcessor.GetBlockByHash([]byte("invalidHash"), api.BlockQueryOptions{})
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestShardAPIBlockProcessor_GetBlockByNonceInvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	shardID := uint32(3)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		true,
		false,
	)

	blk, err := shardAPIBlockProcessor.GetBlockByNonce(100, api.BlockQueryOptions{})
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestShardAPIBlockProcessor_GetBlockByRoundInvalidRoundShouldErr(t *testing.T) {
	t.Parallel()

	shardID := uint32(3)
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		true,
		true,
	)

	blk, err := shardAPIBlockProcessor.GetBlockByRound(100, api.BlockQueryOptions{})
	assert.Nil(t, blk)
	assert.Error(t, err)
}

func TestShardAPIBlockProcessor_GetBlockByHashFromNormalNode(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(3)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMock()
	uint64Converter := mock.NewNonceHashConverterMock()

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		false,
		true,
	)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader, TxCount: 1},
		},
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	expectedBlock := &api.Block{
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   hex.EncodeToString(headerHash),
		NumTxs: 1,
		MiniBlocks: []*api.MiniBlock{
			{
				Hash:                    hex.EncodeToString(miniblockHeader),
				Type:                    block.TxBlock.String(),
				ProcessingType:          block.Normal.String(),
				ConstructionState:       block.Final.String(),
				IndexOfFirstTxProcessed: 0,
				IndexOfLastTxProcessed:  0,
			},
		},
		AccumulatedFees: "0",
		DeveloperFees:   "0",
		Status:          BlockStatusOnChain,
	}

	blk, err := shardAPIBlockProcessor.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestShardAPIBlockProcessor_GetBlockByNonceFromHistoryNode(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(3)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		true,
		false,
	)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader, TxCount: 1},
		},
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(50),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	expectedBlock := &api.Block{
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   hex.EncodeToString(headerHash),
		NumTxs: 1,
		MiniBlocks: []*api.MiniBlock{
			{
				Hash:                    hex.EncodeToString(miniblockHeader),
				Type:                    block.TxBlock.String(),
				ProcessingType:          block.Normal.String(),
				ConstructionState:       block.Final.String(),
				IndexOfFirstTxProcessed: 0,
				IndexOfLastTxProcessed:  0,
			},
		},
		AccumulatedFees: "100",
		DeveloperFees:   "50",
		Status:          BlockStatusOnChain,
	}

	blk, err := shardAPIBlockProcessor.GetBlockByNonce(1, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestShardAPIBlockProcessor_GetBlockByRoundFromStorer(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(3)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		true,
		true,
	)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader, TxCount: 1},
		},
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(50),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	uint64Converter := shardAPIBlockProcessor.uint64ByteSliceConverter
	roundBytes := uint64Converter.ToByteSlice(round)
	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(roundBytes, headerHash)
	_ = storerMock.Put(nonceBytes, headerHash)

	expectedBlock := &api.Block{
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   hex.EncodeToString(headerHash),
		NumTxs: 1,
		MiniBlocks: []*api.MiniBlock{
			{
				Hash:                    hex.EncodeToString(miniblockHeader),
				Type:                    block.TxBlock.String(),
				ProcessingType:          block.Normal.String(),
				ConstructionState:       block.Final.String(),
				IndexOfFirstTxProcessed: 0,
				IndexOfLastTxProcessed:  0,
			},
		},
		AccumulatedFees: "100",
		DeveloperFees:   "50",
		Status:          BlockStatusOnChain,
	}

	blk, err := shardAPIBlockProcessor.GetBlockByRound(round, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestShardAPIBlockProcessor_GetBlockByHashFromHistoryNodeStatusReverted(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(3)
	miniblockHeader := []byte("miniBlockHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	storerMock := genericMocks.NewStorerMockWithEpoch(1)
	uint64Converter := mock.NewNonceHashConverterMock()

	shardAPIBlockProcessor := createMockShardAPIProcessor(
		shardID,
		headerHash,
		storerMock,
		true,
		true,
	)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader, TxCount: 1},
		},
		AccumulatedFees: big.NewInt(100),
		DeveloperFees:   big.NewInt(50),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	correctHash := []byte("correct-hash")
	_ = storerMock.Put(nonceBytes, correctHash)

	expectedBlock := &api.Block{
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   hex.EncodeToString(headerHash),
		NumTxs: 1,
		MiniBlocks: []*api.MiniBlock{
			{
				Hash:                    hex.EncodeToString(miniblockHeader),
				Type:                    block.TxBlock.String(),
				ProcessingType:          block.Normal.String(),
				ConstructionState:       block.Final.String(),
				IndexOfFirstTxProcessed: 0,
				IndexOfLastTxProcessed:  0,
			},
		},
		AccumulatedFees: "100",
		DeveloperFees:   "50",
		Status:          BlockStatusReverted,
	}

	blk, err := shardAPIBlockProcessor.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}
