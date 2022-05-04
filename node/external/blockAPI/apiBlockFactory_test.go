package blockAPI

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/txstatus"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/stretchr/testify/assert"
)

func createMockArgsAPIBlockProc() *ArgAPIBlockProcessor {
	statusComputer, _ := txstatus.NewStatusComputer(0, mock.NewNonceHashConverterMock(), &mock.ChainStorerStub{})
	return &ArgAPIBlockProcessor{
		Store:                    &mock.ChainStorerStub{},
		Marshalizer:              &mock.MarshalizerFake{},
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		HistoryRepo:              &dblookupext.HistoryRepositoryStub{},
		TxUnmarshaller:           &mock.TransactionAPIHandlerStub{},
		StatusComputer:           statusComputer,
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
	}
}

func TestCreateAPIBlockProcessorNilArgs(t *testing.T) {
	t.Parallel()

	t.Run("NilArgShouldErr", func(t *testing.T) {
		t.Parallel()

		_, err := CreateAPIBlockProcessor(nil)
		assert.Equal(t, errNilArgAPIBlockProcessor, err)
	})

	t.Run("NilUint64Converter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.Uint64ByteSliceConverter = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, process.ErrNilUint64Converter, err)
	})

	t.Run("NilStore", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.Store = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, process.ErrNilStorage, err)
	})

	t.Run("NilMarshalizer", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.Marshalizer = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})

	t.Run("NilHistoryRepo", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.HistoryRepo = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, process.ErrNilHistoryRepository, err)
	})

	t.Run("NilUnmarshalTxHandler", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.TxUnmarshaller = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilTransactionUnmarshaler, err)
	})

	t.Run("NilHasher", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.Hasher = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, process.ErrNilHasher, err)
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.AddressPubkeyConverter = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, process.ErrNilPubkeyConverter, err)
	})
}

func TestGetBlockByHash_KeyNotFound(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return 1, nil
		},
	}
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")
	storerMock := mock.NewStorerMock()

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = historyProc
	args.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return headerHash, nil
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	blk, err := apiBlockProc.GetBlockByHash(headerHash, false)
	assert.Equal(t, "key: ZDA4MDg5ZjJhYjczOTUyMDU5OGZkN2FlZWQwOGM0Mjc0NjBmZTk0ZjI4NjM4MzA0N2YzZjYxOTUxYWZjNGUwMA== not found", err.Error())
	assert.Nil(t, blk)
}

func TestGetBlockByHashFromHistoryNode(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return 1, nil
		},
	}
	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(5)
	miniblockHeader := []byte("mbHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")

	uint64Converter := mock.NewNonceHashConverterMock()
	storerMock := mock.NewStorerMock()

	args := createMockArgsAPIBlockProc()
	args.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return headerHash, nil
		},
	}
	args.HistoryRepo = historyProc
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader},
		},
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
	blockBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, blockBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	expectedBlock := &api.Block{
		Nonce: nonce,
		Round: round,
		Shard: shardID,
		Epoch: epoch,
		Hash:  hex.EncodeToString(headerHash),
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees: "0",
		DeveloperFees:   "0",
		Status:          BlockStatusOnChain,
	}

	blk, err := apiBlockProc.GetBlockByHash(headerHash, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestGetBlockByHashFromNormalNode(t *testing.T) {
	t.Parallel()

	uint64Converter := mock.NewNonceHashConverterMock()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	miniblockHeader := []byte("mbHash")
	headerHash := []byte("d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00")
	storerMock := mock.NewStorerMock()

	args := createMockArgsAPIBlockProc()
	args.SelfShardID = core.MetachainShardId
	args.Store = &mock.ChainStorerMock{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return storerMock.Get(key)
		},
	}
	args.HistoryRepo = &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return false
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	header := &block.MetaBlock{
		Nonce: nonce,
		Round: round,
		Epoch: epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader},
		},
		AccumulatedFees:        big.NewInt(100),
		DeveloperFees:          big.NewInt(10),
		AccumulatedFeesInEpoch: big.NewInt(2000),
		DevFeesInEpoch:         big.NewInt(49),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, headerBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	_ = storerMock.Put(nonceBytes, headerHash)

	expectedBlock := &api.Block{
		Nonce: nonce,
		Round: round,
		Shard: core.MetachainShardId,
		Epoch: epoch,
		Hash:  hex.EncodeToString(headerHash),
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		NotarizedBlocks:        []*api.NotarizedBlock{},
		Status:                 BlockStatusOnChain,
		AccumulatedFees:        "100",
		DeveloperFees:          "10",
		AccumulatedFeesInEpoch: "2000",
		DeveloperFeesInEpoch:   "49",
	}

	blk, err := apiBlockProc.GetBlockByHash(headerHash, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestGetBlockByNonceFromHistoryNode(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return 1, nil
		},
	}

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(5)
	miniblockHeader := []byte("mbHash")
	storerMock := mock.NewStorerMock()
	headerHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = historyProc
	args.Store = &mock.ChainStorerMock{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return hex.DecodeString(headerHash)
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader},
		},
		AccumulatedFees: big.NewInt(1000),
		DeveloperFees:   big.NewInt(50),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(func() []byte { hashBytes, _ := hex.DecodeString(headerHash); return hashBytes }(), headerBytes)

	expectedBlock := &api.Block{
		Nonce: nonce,
		Round: round,
		Shard: shardID,
		Epoch: epoch,
		Hash:  headerHash,
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees: "1000",
		DeveloperFees:   "50",
		Status:          BlockStatusOnChain,
	}

	blk, err := apiBlockProc.GetBlockByNonce(1, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestGetBlockByNonce_GetBlockByRound_FromNormalNode(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(5)
	miniblockHeader := []byte("mbHash")
	headerHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return false
		},
	}
	args.Store = &mock.ChainStorerMock{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			if unitType == dataRetriever.ShardHdrNonceHashDataUnit ||
				unitType == dataRetriever.RoundHdrHashDataUnit {
				return hex.DecodeString(headerHash)
			}
			blk := &block.Header{
				Nonce:   nonce,
				Round:   round,
				ShardID: shardID,
				Epoch:   epoch,
				MiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: miniblockHeader},
				},
				AccumulatedFees: big.NewInt(1000),
				DeveloperFees:   big.NewInt(50),
			}
			blockBytes, _ := json.Marshal(blk)
			return blockBytes, nil
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	expectedBlock := &api.Block{
		Nonce: nonce,
		Round: round,
		Shard: shardID,
		Epoch: epoch,
		Hash:  headerHash,
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees: "1000",
		DeveloperFees:   "50",
		Status:          BlockStatusOnChain,
	}

	blk, err := apiBlockProc.GetBlockByNonce(nonce, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)

	blk, err = apiBlockProc.GetBlockByRound(round, false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}

func TestGetBlockByHashFromHistoryNode_StatusReverted(t *testing.T) {
	t.Parallel()

	historyProc := &dblookupext.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetEpochByHashCalled: func(hash []byte) (uint32, error) {
			return 1, nil
		},
	}
	nonce := uint64(1)
	round := uint64(2)
	epoch := uint32(1)
	shardID := uint32(5)
	miniblockHeader := []byte("mbHash")
	headerHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"
	uint64Converter := mock.NewNonceHashConverterMock()
	storerMock := mock.NewStorerMock()

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = historyProc
	args.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return storerMock.Get(key)
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader},
		},
		AccumulatedFees: big.NewInt(500),
		DeveloperFees:   big.NewInt(55),
	}
	blockBytes, _ := json.Marshal(header)
	_ = storerMock.Put([]byte(headerHash), blockBytes)

	nonceBytes := uint64Converter.ToByteSlice(nonce)
	correctHash := []byte("correct-hash")
	_ = storerMock.Put(nonceBytes, correctHash)

	expectedBlock := &api.Block{
		Nonce: nonce,
		Round: round,
		Shard: shardID,
		Epoch: epoch,
		Hash:  hex.EncodeToString([]byte(headerHash)),
		MiniBlocks: []*api.MiniBlock{
			{
				Hash: hex.EncodeToString(miniblockHeader),
				Type: block.TxBlock.String(),
			},
		},
		AccumulatedFees: "500",
		DeveloperFees:   "55",
		Status:          BlockStatusReverted,
	}

	blk, err := apiBlockProc.GetBlockByHash([]byte(headerHash), false)
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}
