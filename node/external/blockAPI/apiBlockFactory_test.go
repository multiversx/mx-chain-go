package blockAPI

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/txstatus"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockArgsAPIBlockProc() *ArgAPIBlockProcessor {
	statusComputer, _ := txstatus.NewStatusComputer(0, mock.NewNonceHashConverterMock(), &storageMocks.ChainStorerStub{})

	return &ArgAPIBlockProcessor{
		Store:                        &storageMocks.ChainStorerStub{},
		Marshalizer:                  &mock.MarshalizerFake{},
		Uint64ByteSliceConverter:     mock.NewNonceHashConverterMock(),
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		APITransactionHandler:        &mock.TransactionAPIHandlerStub{},
		StatusComputer:               statusComputer,
		Hasher:                       &mock.HasherMock{},
		AddressPubkeyConverter:       &testscommon.PubkeyConverterMock{},
		LogsFacade:                   &testscommon.LogsFacadeStub{},
		ReceiptsRepository:           &testscommon.ReceiptsRepositoryStub{},
		AlteredAccountsProvider:      &testscommon.AlteredAccountsProviderStub{},
		AccountsRepository:           &state.AccountsRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
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

	t.Run("NilTxHandler", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.APITransactionHandler = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilTransactionHandler, err)
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

	t.Run("NilLogsFacade", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.LogsFacade = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilLogsFacade, err)
	})

	t.Run("NilReceiptsRepository", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.ReceiptsRepository = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilReceiptsRepository, err)
	})

	t.Run("NilAlteredAccountsProvider", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.AlteredAccountsProvider = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilAlteredAccountsProvider, err)
	})

	t.Run("NilAccountsRepository", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.AccountsRepository = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilAccountsRepository, err)
	})

	t.Run("NilScheduledTxsExecutionHandler", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.ScheduledTxsExecutionHandler = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilScheduledTxsExecutionHandler, err)
	})

	t.Run("NilEnableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgsAPIBlockProc()
		arguments.EnableEpochsHandler = nil

		_, err := CreateAPIBlockProcessor(arguments)
		assert.Equal(t, errNilEnableEpochsHandler, err)
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
	storerMock := genericMocks.NewStorerMockWithEpoch(1)

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = historyProc
	args.Store = &storageMocks.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
		},
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return headerHash, nil
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	blk, err := apiBlockProc.GetBlockByHash(headerHash, api.BlockQueryOptions{})
	assert.Equal(t, "StorerMock: not found; key = 64303830383966326162373339353230353938666437616565643038633432373436306665393466323836333833303437663366363139353161666334653030, epoch = 1", err.Error())
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
	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	args := createMockArgsAPIBlockProc()
	args.Store = &storageMocks.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
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
			{Hash: miniblockHeader, TxCount: 1},
		},
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
	blockBytes, _ := json.Marshal(header)
	_ = storerMock.Put(headerHash, blockBytes)

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

	blk, err := apiBlockProc.GetBlockByHash(headerHash, api.BlockQueryOptions{})
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
	storerMock := genericMocks.NewStorerMock()

	args := createMockArgsAPIBlockProc()
	args.SelfShardID = core.MetachainShardId
	args.Store = &storageMocks.ChainStorerStub{
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

	blk, err := apiBlockProc.GetBlockByHash(headerHash, api.BlockQueryOptions{})
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
	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)
	headerHash := "d08089f2ab739520598fd7aeed08c427460fe94f286383047f3f61951afc4e00"

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = historyProc
	args.Store = &storageMocks.ChainStorerStub{
		GetCalled: func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
			return hex.DecodeString(headerHash)
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
		},
	}
	apiBlockProc, _ := CreateAPIBlockProcessor(args)

	header := &block.Header{
		Nonce:   nonce,
		Round:   round,
		ShardID: shardID,
		Epoch:   epoch,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: miniblockHeader, TxCount: 1},
		},
		AccumulatedFees: big.NewInt(1000),
		DeveloperFees:   big.NewInt(50),
	}
	headerBytes, _ := json.Marshal(header)
	_ = storerMock.Put(func() []byte { hashBytes, _ := hex.DecodeString(headerHash); return hashBytes }(), headerBytes)

	expectedBlock := &api.Block{
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   headerHash,
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
		AccumulatedFees: "1000",
		DeveloperFees:   "50",
		Status:          BlockStatusOnChain,
	}

	blk, err := apiBlockProc.GetBlockByNonce(1, api.BlockQueryOptions{})
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
	args.Store = &storageMocks.ChainStorerStub{
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
					{Hash: miniblockHeader, TxCount: 1},
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
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   headerHash,
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
		AccumulatedFees: "1000",
		DeveloperFees:   "50",
		Status:          BlockStatusOnChain,
	}

	blk, err := apiBlockProc.GetBlockByNonce(nonce, api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)

	blk, err = apiBlockProc.GetBlockByRound(round, api.BlockQueryOptions{})
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
	storerMock := genericMocks.NewStorerMockWithEpoch(epoch)

	args := createMockArgsAPIBlockProc()
	args.HistoryRepo = historyProc
	args.Store = &storageMocks.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
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
			{Hash: miniblockHeader, TxCount: 1},
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
		Nonce:  nonce,
		Round:  round,
		Shard:  shardID,
		Epoch:  epoch,
		Hash:   hex.EncodeToString([]byte(headerHash)),
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
		AccumulatedFees: "500",
		DeveloperFees:   "55",
		Status:          BlockStatusReverted,
	}

	blk, err := apiBlockProc.GetBlockByHash([]byte(headerHash), api.BlockQueryOptions{})
	assert.Nil(t, err)
	assert.Equal(t, expectedBlock, blk)
}
