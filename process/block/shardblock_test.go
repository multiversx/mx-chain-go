package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

const MaxGasLimitPerBlock = uint64(100000)

func createTestShardDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1})

	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1})
	rewardsTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 300, Type: storageUnit.LRUCache, Shards: 1})
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache, Shards: 1}
	hdrPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	hdrNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	peerChangeBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache, Shards: 1}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	currTxs, _ := dataPool.NewCurrentBlockPool()

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		rewardsTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
		currTxs,
	)

	return dPool
}

//------- NewShardProcessor

func initAccountsMock() *mock.AccountsStub {
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}
	return &mock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
}

func initBasicTestData() (*mock.PoolsHolderMock, *blockchain.BlockChain, []byte, block.Body, [][]byte, *mock.HasherMock, *mock.MarshalizerMock, error, []byte) {
	tdp := mock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	randSeed := []byte("rand seed")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Round:    1,
			Nonce:    1,
			RandSeed: randSeed,
		},
	}
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	return tdp, blkc, rootHash, body, txHashes, hasher, marshalizer, nil, mbHash
}

func initBlockHeader(prevHash []byte, prevRandSeed []byte, rootHash []byte, mbHdrs []block.MiniBlockHeader) block.Header {
	hdr := block.Header{
		Nonce:            2,
		Round:            2,
		PrevHash:         prevHash,
		PrevRandSeed:     prevRandSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	return hdr
}

func CreateMockArgumentsMultiShard() blproc.ArgShardProcessor {
	arguments := CreateMockArguments()
	arguments.DataPool = initDataPool([]byte("tx_hash1"))
	arguments.Accounts = initAccountsMock()
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(3)
	arguments.StartHeaders = createGenesisBlocks(arguments.ShardCoordinator)

	return arguments
}

//------- NewBlockProcessor

func TestNewBlockProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.DataPool = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.Store = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilStorage, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.Hasher = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilMarshalizerShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.Marshalizer = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.Accounts = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.ShardCoordinator = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.ForkDetector = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilRequestTransactionHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.RequestHandler = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilTransactionPool, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilTxCoordinator(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.TxCoordinator = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilUint64Converter(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments()
	arguments.Uint64Converter = nil
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Equal(t, process.ErrNilUint64Converter, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, err := blproc.NewShardProcessor(arguments)

	assert.Nil(t, err)
	assert.NotNil(t, sp)
}

//------- ProcessBlock

func TestShardProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	blk := make(block.Body, 0)
	err := sp.ProcessBlock(nil, &block.Header{}, blk, haveTime)

	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestShardProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	body := make(block.Body, 0)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, nil, body, haveTime)

	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestShardProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, nil, haveTime)

	assert.Equal(t, process.ErrNilBlockBody, err)
}

func TestShardProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	blk := make(block.Body, 0)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, blk, nil)

	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestShardProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	blkc := &blockchain.BlockChain{}
	hdr := block.Header{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		PrevRandSeed:  []byte("rand seed"),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
	}

	body := make(block.Body, 0)
	arguments := CreateMockArgumentsMultiShard()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}
	sp, _ := blproc.NewShardProcessor(arguments)
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)

	assert.NotNil(t, err)
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestShardProcessor_ProcessBlockHeaderBodyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{}
	hdr := block.Header{
		Nonce:         1,
		PrevHash:      []byte(""),
		PrevRandSeed:  []byte("rand seed"),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("00110"),
		ShardId:       0,
		RootHash:      []byte("rootHash"),
	}
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)
	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error { return nil }
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	arguments.ForkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{}

	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	hdr := block.Header{
		Nonce:            1,
		PrevHash:         []byte(""),
		PrevRandSeed:     []byte("rand seed"),
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         []byte("rootHash"),
		MiniBlockHeaders: mbHdrs,
	}
	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error { return nil }
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}

	accounts := &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(5),
		initStore(),
		marshalizer,
		hasher,
		tdp,
		&mock.AddressConverterMock{},
		accounts,
		&mock.RequestHandlerMock{},
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) error {
				return process.ErrHigherNonceInTransaction
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 0
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.MiniBlocksCompacterMock{
			ExpandCalled: func(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
				return miniBlocks, nil
			},
		},
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		accounts,
		tdp.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{
			InitCalled: func() {
			},
		},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err = sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestShardProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()
	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{
		Nonce:         0,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		PrevRandSeed:  []byte("rand seed"),
		Signature:     []byte("signature"),
		RootHash:      []byte("root hash"),
	}
	body := make(block.Body, 0)
	blkc := &blockchain.BlockChain{}
	err := sp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestShardProcessor_ProcessWithHeaderNotCorrectNonceShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{
		Nonce:         0,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		PrevRandSeed:  []byte("rand seed"),
		Signature:     []byte("signature"),
		RootHash:      []byte("root hash"),
	}
	body := make(block.Body, 0)
	blkc := &blockchain.BlockChain{}
	err := sp.ProcessBlock(blkc, hdr, body, haveTime)

	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestShardProcessor_ProcessWithHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	arguments.ForkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}

	randSeed := []byte("rand seed")
	sp, _ := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		PrevRandSeed:  randSeed,
		Signature:     []byte("signature"),
		RootHash:      []byte("root hash"),
	}
	body := make(block.Body, 0)
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		},
	}
	err := sp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)
}

func TestShardProcessor_ProcessBlockWithErrOnProcessBlockTransactionsCallShouldRevertState(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	randSeed := []byte("rand seed")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		},
	}
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	hdr := block.Header{
		Round:            1,
		Nonce:            1,
		PrevHash:         []byte(""),
		PrevRandSeed:     randSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         []byte("rootHash"),
		MiniBlockHeaders: mbHdrs,
	}

	// set accounts not dirty
	journalLen := func() int { return 0 }
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}

	err := errors.New("process block transaction error")
	txProcess := func(transaction *transaction.Transaction) error {
		return err
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	tpm := &mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	store := &mock.ChainStorerMock{}
	accounts := &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		store,
		marshalizer,
		hasher,
		tdp,
		&mock.AddressConverterMock{},
		accounts,
		&mock.RequestHandlerMock{},
		tpm,
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 0
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.MiniBlocksCompacterMock{
			ExpandCalled: func(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
				return miniBlocks, nil
			},
		},
		&mock.GasHandlerMock{
			ComputeGasConsumedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasConsumedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
	)
	container, _ := factory.Create()

	totalGasConsumed := uint64(0)
	tc, _ := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		accounts,
		tdp.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{
			InitCalled: func() {
				totalGasConsumed = 0
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
		},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Accounts = accounts
	arguments.ShardCoordinator = shardCoordinator
	arguments.ForkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err2 := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, err, err2)
	assert.True(t, wasCalled)
}

func TestShardProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	randSeed := []byte("rand seed")
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		},
	}
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	hdr := block.Header{
		Round:            1,
		Nonce:            1,
		PrevHash:         []byte(""),
		PrevRandSeed:     randSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         []byte("rootHash"),
		MiniBlockHeaders: mbHdrs,
	}

	// set accounts not dirty
	journalLen := func() int { return 0 }
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHashX"), nil
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	arguments.Hasher = &mock.HasherStub{}
	arguments.ForkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)
	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrRootStateDoesNotMatch, err)
	assert.True(t, wasCalled)
}

func TestShardProcessor_ProcessBlockOnlyIntraShardShouldPass(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	randSeed := []byte("rand seed")
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		},
	}
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	hdr := block.Header{
		Round:            1,
		Nonce:            1,
		PrevHash:         []byte(""),
		PrevRandSeed:     randSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	// set accounts not dirty
	journalLen := func() int { return 0 }
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockCrossShardWithoutMetaShouldFail(t *testing.T) {
	t.Parallel()

	randSeed := []byte("rand seed")
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		},
	}
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	tx := &transaction.Transaction{}
	tdp.Transactions().AddData(txHash, tx, shardCoordinator.CommunicationIdentifier(0))

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	hdr := block.Header{
		Round:            1,
		Nonce:            1,
		PrevHash:         []byte(""),
		PrevRandSeed:     randSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	// set accounts not dirty
	journalLen := func() int { return 0 }
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrCrossShardMBWithoutConfirmationFromMeta, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockCrossShardWithMetaShouldPass(t *testing.T) {
	t.Parallel()

	tdp, blkc, rootHash, body, txHashes, hasher, marshalizer, _, mbHash := initBasicTestData()
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	randSeed := []byte("rand seed")
	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	meta := block.MetaBlock{
		Nonce:     1,
		ShardInfo: shardHdrs,
		RandSeed:  randSeed,
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)

	meta = block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)

	// set accounts not dirty
	journalLen := func() int { return 0 }
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrCrossShardMBWithoutConfirmationFromMeta, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockHaveTimeLessThanZeroShouldErr(t *testing.T) {
	t.Parallel()
	txHash := []byte("tx_hash1")
	tdp := initDataPool(txHash)

	randSeed := []byte("rand seed")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    1,
			RandSeed: randSeed,
		},
	}
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		SenderShardID:   0,
		ReceiverShardID: 0,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	currHdr := blkc.GetCurrentBlockHeader()
	preHash, _ := core.CalculateHash(marshalizer, hasher, currHdr)
	hdr := block.Header{
		Round:            2,
		Nonce:            2,
		PrevHash:         preHash,
		PrevRandSeed:     randSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(tdp, genesisBlocks)
	haveTimeLessThanZero := func() time.Duration {
		return -1 * time.Millisecond
	}

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTimeLessThanZero)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestShardProcessor_ProcessBlockWithMissingMetaHdrShouldErr(t *testing.T) {
	t.Parallel()

	tdp, blkc, rootHash, body, txHashes, hasher, marshalizer, _, mbHash := initBasicTestData()
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	randSeed := []byte("rand seed")
	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	meta := block.MetaBlock{
		Nonce:     1,
		ShardInfo: shardHdrs,
		RandSeed:  randSeed,
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	meta = block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error {
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestShardProcessor_ProcessBlockWithWrongMiniBlockHeaderShouldErr(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	tdp := initDataPool(txHash)
	randSeed := []byte("rand seed")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    1,
			RandSeed: randSeed,
		},
	}
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		SenderShardID:   1,
		ReceiverShardID: 0,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

//------- checkAndRequestIfMetaHeadersMissing
func TestShardProcessor_CheckAndRequestIfMetaHeadersMissingShouldErr(t *testing.T) {
	t.Parallel()

	hdrNoncesRequestCalled := int32(0)
	tdp, blkc, rootHash, body, txHashes, hasher, marshalizer, _, mbHash := initBasicTestData()
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	randSeed := []byte("rand seed")

	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	meta := &block.MetaBlock{
		Nonce:     1,
		ShardInfo: shardHdrs,
		Round:     1,
		RandSeed:  randSeed,
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error {
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.RequestHandler = &mock.RequestHandlerMock{
		RequestHeaderHandlerByNonceCalled: func(destShardID uint32, nonce uint64) {
			atomic.AddInt32(&hdrNoncesRequestCalled, 1)
		},
	}
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)

	meta = &block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        2,
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)

	sp.CheckAndRequestIfMetaHeadersMissing(2)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&hdrNoncesRequestCalled))
	assert.Equal(t, err, process.ErrTimeIsOut)
}

//-------- isMetaHeaderFinal
func TestShardProcessor_IsMetaHeaderFinalShouldPass(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	randSeed := []byte("rand seed")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce:    1,
			RandSeed: randSeed,
		},
	}
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	meta := &block.MetaBlock{
		Nonce:     1,
		ShardInfo: shardHdrs,
		Round:     1,
		RandSeed:  randSeed,
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(tdp, genesisBlocks)

	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrTimeIsOut, err)
	res := sp.IsMetaHeaderFinal(&hdr, nil, 0)
	assert.False(t, res)
	res = sp.IsMetaHeaderFinal(nil, nil, 0)
	assert.False(t, res)

	meta = &block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        2,
		PrevHash:     metaHash,
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))
	tdp.MetaBlocks().Put(metaHash, meta)

	meta = &block.MetaBlock{
		Nonce:     1,
		ShardInfo: shardHdrs,
		Round:     1,
		RandSeed:  randSeed,
	}
	ordered, _ := sp.GetOrderedMetaBlocks(3)
	res = sp.IsMetaHeaderFinal(meta, ordered, 0)
	assert.True(t, res)
}

//-------- requestMissingFinalityAttestingHeaders
func TestShardProcessor_RequestMissingFinalityAttestingHeaders(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderMock()
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	sp, _ := blproc.NewShardProcessor(arguments)

	sp.SetHighestHdrNonceForCurrentBlock(sharding.MetachainShardId, 1)
	res := sp.RequestMissingFinalityAttestingHeaders()
	assert.Equal(t, res > 0, true)
}

//--------- verifyIncludedMetaBlocksFinality
func TestShardProcessor_CheckMetaHeadersValidityAndFinalityShouldPass(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	lastHdr := genesisBlocks[0]
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	randSeed := []byte("rand seed")
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	prevMeta := genesisBlocks[sharding.MetachainShardId]
	prevHash, _ = core.CalculateHash(marshalizer, hasher, prevMeta)
	meta1 := &block.MetaBlock{
		Nonce:        1,
		ShardInfo:    shardHdrs,
		Round:        1,
		PrevHash:     prevHash,
		PrevRandSeed: prevMeta.GetRandSeed(),
	}
	metaBytes, _ := marshalizer.Marshal(meta1)
	metaHash1 := hasher.Compute(string(metaBytes))
	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash1)

	tdp.MetaBlocks().Put(metaHash1, meta1)

	prevHash, _ = core.CalculateHash(marshalizer, hasher, meta1)
	meta2 := &block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
		Round:     2,
		PrevHash:  prevHash,
	}
	metaBytes, _ = marshalizer.Marshal(meta2)
	metaHash2 := hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash2, meta2)
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.StartHeaders = genesisBlocks
	sp, _ := blproc.NewShardProcessor(arguments)
	hdr.Round = 4

	sp.SetHdrForCurrentBlock(metaHash1, meta1, true)
	sp.SetHdrForCurrentBlock(metaHash2, meta2, false)

	err := sp.CheckMetaHeadersValidityAndFinality()
	assert.Nil(t, err)
}

func TestShardProcessor_CheckMetaHeadersValidityAndFinalityShouldReturnNilWhenNoMetaBlocksAreUsed(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderMock()
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(tdp, genesisBlocks)

	err := sp.CheckMetaHeadersValidityAndFinality()
	assert.Nil(t, err)
}

//------- CommitBlock

func TestShardProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = accounts
	sp, _ := blproc.NewShardProcessor(arguments)
	blk := make(block.Body, 0)

	err := sp.CommitBlock(nil, &block.Header{}, blk)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestShardProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	rootHash := []byte("root hash to be tested")
	accounts := &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	errMarshalizer := errors.New("failure")
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	body := make(block.Body, 0)
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			if reflect.DeepEqual(obj, hdr) {
				return nil, errMarshalizer
			}

			return []byte("obj"), nil
		},
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizer
	arguments.Accounts = accounts
	sp, _ := blproc.NewShardProcessor(arguments)
	blkc := createTestBlockchain()

	err := sp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errMarshalizer, err)
}

func TestShardProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	errPersister := errors.New("failure")
	wasCalled := false
	rootHash := []byte("root hash to be tested")
	marshalizer := &mock.MarshalizerMock{}
	accounts := &mock.AccountsStub{
		CommitCalled: func() ([]byte, error) {
			return nil, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	body := make(block.Body, 0)
	hdrUnit := &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			hdr, _ := marshalizer.Marshal(&block.Header{})
			return hdr, nil
		},
		PutCalled: func(key, data []byte) error {
			wasCalled = true
			return errPersister
		},
		HasCalled: func(key []byte) error {
			return nil
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.BlockHeaderUnit, hdrUnit)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Accounts = accounts
	arguments.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadereHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
	)

	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})

	err := sp.CommitBlock(blkc, hdr, body)
	assert.True(t, wasCalled)
	assert.Nil(t, err)
}

func TestShardProcessor_CommitBlockStorageFailsForBodyShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	wasCalled := false
	errPersister := errors.New("failure")
	rootHash := []byte("root hash to be tested")
	accounts := &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		CommitCalled: func() (i []byte, e error) {
			return nil, nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	mb := block.MiniBlock{}
	body := make(block.Body, 0)
	body = append(body, &mb)

	miniBlockUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			wasCalled = true
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Accounts = accounts
	arguments.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	sp, err := blproc.NewShardProcessor(arguments)
	assert.Nil(t, err)

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
	)

	_ = blkc.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})

	err = sp.CommitBlock(blkc, hdr, body)

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestShardProcessor_CommitBlockNilNoncesDataPoolShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	rootHash := []byte("root hash to be tested")
	accounts := &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	body := make(block.Body, 0)
	store := initStore()

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Accounts = accounts
	sp, _ := blproc.NewShardProcessor(arguments)

	tdp.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return nil
	}
	blkc := createTestBlockchain()
	err := sp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrNilHeadersNoncesDataPool, err)
}

func TestShardProcessor_CommitBlockNoTxInPoolShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	txCache := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		LenCalled: func() int {
			return 0
		},
	}
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return txCache
			},
			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {
			},
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if reflect.DeepEqual(key, []byte("tx1_hash")) {
					return &transaction.Transaction{Nonce: 10}, true
				}
				return nil, false
			},
			RegisterHandlerCalled: func(i func(key []byte)) {

			},
		}
	}

	txHash := []byte("txHash")
	rootHash := []byte("root hash")
	hdrHash := []byte("header hash")
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{txHash},
	}
	body := block.Body{&mb}
	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		initStore(),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		tdp,
		&mock.AddressConverterMock{},
		initAccountsMock(),
		&mock.RequestHandlerMock{},
		&mock.TxProcessorMock{},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		tdp.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.ForkDetector = fd
	arguments.TxCoordinator = tc
	arguments.Accounts = accounts
	sp, _ := blproc.NewShardProcessor(arguments)

	blkc := createTestBlockchain()

	err = sp.CommitBlock(blkc, hdr, body)
	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestShardProcessor_CommitBlockOkValsShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")

	rootHash := []byte("root hash")
	hdrHash := []byte("header hash")
	randSeed := []byte("rand seed")

	prevHdr := &block.Header{
		Nonce:         0,
		Round:         0,
		PubKeysBitmap: rootHash,
		PrevHash:      hdrHash,
		Signature:     rootHash,
		RootHash:      rootHash,
		RandSeed:      randSeed,
	}

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: rootHash,
		PrevHash:      hdrHash,
		Signature:     rootHash,
		RootHash:      rootHash,
		PrevRandSeed:  randSeed,
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{txHash},
	}
	body := block.Body{&mb}

	mbHdr := block.MiniBlockHeader{
		TxCount: uint32(len(mb.TxHashes)),
		Hash:    hdrHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)
	hdr.MiniBlockHeaders = mbHdrs

	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	forkDetectorAddCalled := false
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			if header == hdr {
				forkDetectorAddCalled = true
				return nil
			}

			return errors.New("should have not got here")
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Accounts = accounts
	arguments.ForkDetector = fd
	sp, _ := blproc.NewShardProcessor(arguments)

	blkc := createTestBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return prevHdr
	}
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	err := sp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Nil(t, err)
	err = sp.CommitBlock(blkc, hdr, body)
	assert.Nil(t, err)
	assert.True(t, forkDetectorAddCalled)
	assert.Equal(t, hdrHash, blkc.GetCurrentBlockHeaderHash())
	//this should sleep as there is an async call to display current hdr and block in CommitBlock
	time.Sleep(time.Second)
}

func TestShardProcessor_CommitBlockCallsIndexerMethods(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")

	rootHash := []byte("root hash")
	hdrHash := []byte("header hash")
	randSeed := []byte("rand seed")

	prevHdr := &block.Header{
		Nonce:         0,
		Round:         0,
		PubKeysBitmap: rootHash,
		PrevHash:      hdrHash,
		Signature:     rootHash,
		RootHash:      rootHash,
		RandSeed:      randSeed,
	}

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: rootHash,
		PrevHash:      hdrHash,
		Signature:     rootHash,
		RootHash:      rootHash,
		PrevRandSeed:  randSeed,
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{txHash},
	}
	body := block.Body{&mb}

	mbHdr := block.MiniBlockHeader{
		TxCount: uint32(len(mb.TxHashes)),
		Hash:    hdrHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)
	hdr.MiniBlockHeaders = mbHdrs

	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	var saveBlockCalled map[string]data.TransactionHandler
	saveBlockCalledMutex := sync.Mutex{}

	arguments := CreateMockArgumentsMultiShard()
	arguments.Core = &mock.ServiceContainerMock{
		IndexerCalled: func() indexer.Indexer {
			return &mock.IndexerMock{
				SaveBlockCalled: func(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler) {
					saveBlockCalledMutex.Lock()
					saveBlockCalled = txPool
					saveBlockCalledMutex.Unlock()
				},
			}
		},
	}
	arguments.DataPool = tdp
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Accounts = accounts
	arguments.ForkDetector = fd
	arguments.TxCoordinator = &mock.TransactionCoordinatorMock{
		GetAllCurrentUsedTxsCalled: func(blockType block.Type) map[string]data.TransactionHandler {
			switch blockType {
			case block.TxBlock:
				return map[string]data.TransactionHandler{
					"tx_1": &transaction.Transaction{Nonce: 1},
					"tx_2": &transaction.Transaction{Nonce: 2},
				}
			case block.SmartContractResultBlock:
				return map[string]data.TransactionHandler{
					"utx_1": &smartContractResult.SmartContractResult{Nonce: 1},
					"utx_2": &smartContractResult.SmartContractResult{Nonce: 2},
				}
			default:
				return nil
			}
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	blkc := createTestBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return prevHdr
	}
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	err := sp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Nil(t, err)
	err = sp.CommitBlock(blkc, hdr, body)
	assert.Nil(t, err)

	// Wait for the index block go routine to start
	time.Sleep(time.Second * 2)

	saveBlockCalledMutex.Lock()
	wasCalled := saveBlockCalled
	saveBlockCalledMutex.Unlock()

	assert.Equal(t, 4, len(wasCalled))
}

func TestShardProcessor_CreateTxBlockBodyWithDirtyAccStateShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	bl, err := sp.CreateBlockBody(&block.Header{PrevRandSeed: []byte("randSeed")}, func() bool { return true })
	// nil block
	assert.Nil(t, bl)
	// error
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestShardProcessor_CreateTxBlockBodyWithNoTimeShouldEmptyBlock(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 0 }
	rootHashfunc := func() ([]byte, error) {
		return []byte("roothash"), nil
	}
	revToSnapshot := func(snapshot int) error { return nil }
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled:       journalLen,
		RootHashCalled:         rootHashfunc,
		RevertToSnapshotCalled: revToSnapshot,
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	haveTime := func() bool {
		return false
	}
	bl, err := sp.CreateBlockBody(&block.Header{PrevRandSeed: []byte("randSeed")}, haveTime)
	// no error
	assert.Equal(t, process.ErrTimeIsOut, err)
	// no miniblocks
	assert.Nil(t, bl)
}

func TestShardProcessor_CreateTxBlockBodyOK(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 0 }
	rootHashfunc := func() ([]byte, error) {
		return []byte("roothash"), nil
	}
	haveTime := func() bool {
		return true
	}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Accounts = &mock.AccountsStub{
		JournalLenCalled: journalLen,
		RootHashCalled:   rootHashfunc,
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	blk, err := sp.CreateBlockBody(&block.Header{PrevRandSeed: []byte("randSeed")}, haveTime)
	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

//------- ComputeNewNoncePrevHash

func TestNode_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}
	arguments := CreateMockArgumentsMultiShard()
	arguments.Store = initStore()
	arguments.DataPool = tdp
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer

	be, _ := blproc.NewShardProcessor(arguments)
	hdr, txBlock := createTestHdrTxBlockBody()
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return []byte("hdrHeaderMarshalized"), nil
		}
		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	hasher.ComputeCalled = func(s string) []byte {
		if s == "hdrHeaderMarshalized" {
			return []byte("header hash")
		}
		if s == "txBlockBodyMarshalized" {
			return []byte("tx block body hash")
		}
		return nil
	}
	_, err := be.ComputeHeaderHash(hdr)
	assert.Nil(t, err)
}

func createTestHdrTxBlockBody() (*block.Header, block.Body) {
	hasher := mock.HasherMock{}
	hdr := &block.Header{
		Nonce:         1,
		ShardId:       2,
		Epoch:         3,
		Round:         4,
		TimeStamp:     uint64(11223344),
		PrevHash:      hasher.Compute("prev hash"),
		PubKeysBitmap: []byte{255, 0, 128},
		Signature:     hasher.Compute("signature"),
		RootHash:      hasher.Compute("root hash"),
	}
	txBlock := block.Body{
		{
			ReceiverShardID: 0,
			SenderShardID:   0,
			TxHashes: [][]byte{
				hasher.Compute("txHash_0_1"),
				hasher.Compute("txHash_0_2"),
			},
		},
		{
			ReceiverShardID: 1,
			SenderShardID:   0,
			TxHashes: [][]byte{
				hasher.Compute("txHash_1_1"),
				hasher.Compute("txHash_1_2"),
			},
		},
		{
			ReceiverShardID: 2,
			SenderShardID:   0,
			TxHashes: [][]byte{
				hasher.Compute("txHash_2_1"),
			},
		},
		{
			ReceiverShardID: 3,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
	}
	return hdr, txBlock
}

//------- ComputeNewNoncePrevHash

func TestShardProcessor_DisplayLogInfo(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	hasher := mock.HasherMock{}
	hdr, txBlock := createTestHdrTxBlockBody()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	statusHandler := &mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {

		},
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp

	sp, _ := blproc.NewShardProcessor(arguments)
	assert.NotNil(t, sp)
	hdr.PrevHash = hasher.Compute("prev hash")
	sp.DisplayLogInfo(hdr, txBlock, []byte("tx_hash1"), shardCoordinator.NumberOfShards(), shardCoordinator.SelfId(), tdp, statusHandler)
}

func TestBlockProcessor_ApplyBodyToHeaderShouldNotReturnNil(t *testing.T) {
	t.Parallel()
	arguments := CreateMockArgumentsMultiShard()

	bp, _ := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{}
	err := bp.ApplyBodyToHeader(hdr, nil)
	assert.Nil(t, err)
	assert.NotNil(t, hdr)
}

func TestShardProcessor_ApplyBodyToHeaderShouldErrWhenMarshalizerErrors(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	bp, _ := blproc.NewShardProcessor(arguments)
	body := block.Body{
		{
			ReceiverShardID: 1,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
		{
			ReceiverShardID: 2,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
		{
			ReceiverShardID: 3,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
	}
	hdr := &block.Header{}
	err := bp.ApplyBodyToHeader(hdr, body)
	assert.NotNil(t, err)
}

func TestShardProcessor_ApplyBodyToHeaderReturnsOK(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	bp, _ := blproc.NewShardProcessor(arguments)
	body := block.Body{
		{
			ReceiverShardID: 1,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
		{
			ReceiverShardID: 2,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
		{
			ReceiverShardID: 3,
			SenderShardID:   0,
			TxHashes:        make([][]byte, 0),
		},
	}
	hdr := &block.Header{}
	err := bp.ApplyBodyToHeader(hdr, body)
	assert.Nil(t, err)
	assert.Equal(t, len(body), len(hdr.MiniBlockHeaders))
}

func TestShardProcessor_CommitBlockShouldRevertAccountStateWhenErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalEntries := 3
	revToSnapshot := func(snapshot int) error {
		journalEntries = 0
		return nil
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.Accounts = &mock.AccountsStub{
		RevertToSnapshotCalled: revToSnapshot,
	}
	bp, _ := blproc.NewShardProcessor(arguments)
	err := bp.CommitBlock(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, journalEntries)
}

func TestShardProcessor_MarshalizedDataToBroadcastShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash0 := []byte("txHash0")
	mb0 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        [][]byte{txHash0},
	}
	txHash1 := []byte("txHash1")
	mb1 := block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxHashes:        [][]byte{txHash1},
	}
	body := make(block.Body, 0)
	body = append(body, &mb0)
	body = append(body, &mb1)
	body = append(body, &mb0)
	body = append(body, &mb1)
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		initStore(),
		marshalizer,
		&mock.HasherMock{},
		tdp,
		&mock.AddressConverterMock{},
		initAccountsMock(),
		&mock.RequestHandlerMock{},
		&mock.TxProcessorMock{},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		tdp.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizer
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)
	msh, mstx, err := sp.MarshalizedDataToBroadcast(&block.Header{}, body)
	assert.Nil(t, err)
	assert.NotNil(t, msh)
	assert.NotNil(t, mstx)
	_, found := msh[0]
	assert.False(t, found)

	expectedBody := make(block.Body, 0)
	err = marshalizer.Unmarshal(&expectedBody, msh[1])
	assert.Nil(t, err)
	assert.Equal(t, len(expectedBody), 2)
	assert.Equal(t, &mb1, expectedBody[0])
	assert.Equal(t, &mb1, expectedBody[1])
}

func TestShardProcessor_MarshalizedDataWrongType(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizer

	sp, _ := blproc.NewShardProcessor(arguments)
	wr := &wrongBody{}
	msh, mstx, err := sp.MarshalizedDataToBroadcast(&block.Header{}, wr)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
	assert.Nil(t, msh)
	assert.Nil(t, mstx)
}

func TestShardProcessor_MarshalizedDataNilInput(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizer

	sp, _ := blproc.NewShardProcessor(arguments)
	msh, mstx, err := sp.MarshalizedDataToBroadcast(nil, nil)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
	assert.Nil(t, msh)
	assert.Nil(t, mstx)
}

func TestShardProcessor_MarshalizedDataMarshalWithoutSuccess(t *testing.T) {
	t.Parallel()
	wasCalled := false
	tdp := initDataPool([]byte("tx_hash1"))
	txHash0 := []byte("txHash0")
	mb0 := block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxHashes:        [][]byte{txHash0},
	}
	body := make(block.Body, 0)
	body = append(body, &mb0)
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			wasCalled = true
			return nil, process.ErrMarshalWithoutSuccess
		},
	}

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		initStore(),
		marshalizer,
		&mock.HasherMock{},
		tdp,
		&mock.AddressConverterMock{},
		initAccountsMock(),
		&mock.RequestHandlerMock{},
		&mock.TxProcessorMock{},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		tdp.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizer
	arguments.TxCoordinator = tc

	sp, _ := blproc.NewShardProcessor(arguments)

	msh, mstx, err := sp.MarshalizedDataToBroadcast(&block.Header{}, body)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, len(msh))
	assert.Equal(t, 0, len(mstx))
}

//------- receivedMetaBlock

func TestShardProcessor_ReceivedMetaBlockShouldRequestMissingMiniBlocks(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()

	//we will have a metablock that will return 3 miniblock hashes
	//1 miniblock hash will be in cache
	//2 will be requested on network

	miniBlockHash1 := []byte("miniblock hash 1 found in cache")
	miniBlockHash2 := []byte("miniblock hash 2")
	miniBlockHash3 := []byte("miniblock hash 3")

	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
					{Hash: miniBlockHash1, SenderShardID: 1, ReceiverShardID: 0},
					{Hash: miniBlockHash2, SenderShardID: 1, ReceiverShardID: 0},
					{Hash: miniBlockHash3, SenderShardID: 1, ReceiverShardID: 0},
				}},
		}}

	//put this metaBlock inside datapool
	metaBlockHash := []byte("metablock hash")
	datapool.MetaBlocks().Put(metaBlockHash, metaBlock)
	//put the existing miniblock inside datapool
	datapool.MiniBlocks().Put(miniBlockHash1, &block.MiniBlock{})

	miniBlockHash1Requested := int32(0)
	miniBlockHash2Requested := int32(0)
	miniBlockHash3Requested := int32(0)

	requestHandler := &mock.RequestHandlerMock{RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
		if bytes.Equal(miniBlockHash1, miniblockHash) {
			atomic.AddInt32(&miniBlockHash1Requested, 1)
		}
		if bytes.Equal(miniBlockHash2, miniblockHash) {
			atomic.AddInt32(&miniBlockHash2Requested, 1)
		}
		if bytes.Equal(miniBlockHash3, miniblockHash) {
			atomic.AddInt32(&miniBlockHash3Requested, 1)
		}
	}}

	tc, _ := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		datapool.MiniBlocks(),
		requestHandler,
		&mock.PreProcessorContainerMock{},
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.RequestHandler = requestHandler
	arguments.TxCoordinator = tc

	bp, _ := blproc.NewShardProcessor(arguments)
	bp.ReceivedMetaBlock(metaBlockHash)

	//we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&miniBlockHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
}

//--------- receivedMetaBlockNoMissingMiniBlocks
func TestShardProcessor_ReceivedMetaBlockNoMissingMiniBlocksShouldPass(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()

	//we will have a metablock that will return 3 miniblock hashes
	//1 miniblock hash will be in cache
	//2 will be requested on network

	miniBlockHash1 := []byte("miniblock hash 1 found in cache")

	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{
					{
						Hash:            miniBlockHash1,
						SenderShardID:   1,
						ReceiverShardID: 0,
					},
				},
			},
		}}

	//put this metaBlock inside datapool
	metaBlockHash := []byte("metablock hash")
	datapool.MetaBlocks().Put(metaBlockHash, &metaBlock)
	//put the existing miniblock inside datapool
	datapool.MiniBlocks().Put(miniBlockHash1, &block.MiniBlock{})

	noOfMissingMiniBlocks := int32(0)

	requestHandler := &mock.RequestHandlerMock{RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
		atomic.AddInt32(&noOfMissingMiniBlocks, 1)
	}}

	tc, _ := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		datapool.MiniBlocks(),
		requestHandler,
		&mock.PreProcessorContainerMock{},
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.RequestHandler = requestHandler
	arguments.TxCoordinator = tc

	sp, _ := blproc.NewShardProcessor(arguments)
	sp.ReceivedMetaBlock(metaBlockHash)
	assert.Equal(t, int32(0), atomic.LoadInt32(&noOfMissingMiniBlocks))
}

//--------- createAndProcessCrossMiniBlocksDstMe
func TestShardProcessor_CreateAndProcessCrossMiniBlocksDstMe(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	meta := &block.MetaBlock{
		Nonce:        1,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        1,
		PrevRandSeed: []byte("roothash"),
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)

	haveTimeTrue := func() bool {
		return true
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	sp, _ := blproc.NewShardProcessor(arguments)
	miniBlockSlice, usedMetaHdrsHashes, noOfTxs, err := sp.CreateAndProcessCrossMiniBlocksDstMe(2, 2, haveTimeTrue)
	assert.Equal(t, err == nil, true)
	assert.Equal(t, len(miniBlockSlice) == 0, true)
	assert.Equal(t, usedMetaHdrsHashes, uint32(0))
	assert.Equal(t, noOfTxs, uint32(0))
}

func TestShardProcessor_NewShardProcessorWrongTypeOfStartHeaderShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderMock()
	txHash := []byte(nil)
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))

	startHeaders := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	startHeaders[sharding.MetachainShardId] = &block.Header{}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.StartHeaders = startHeaders

	sp, err := blproc.NewShardProcessor(arguments)

	assert.Nil(t, sp)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestShardProcessor_CreateAndProcessCrossMiniBlocksDstMeProcessPartOfMiniBlocksInMetaBlock(t *testing.T) {
	t.Parallel()

	haveTimeTrue := func() bool {
		return true
	}
	tdp := mock.NewPoolsHolderMock()
	destShardId := uint32(2)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}
	miniblocks := make([]*block.MiniBlock, 6)

	txHash := []byte("txhash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock1 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	miniblock2 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   2,
		TxHashes:        txHashes,
	}

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)

	destShards := []uint32{1, 3, 4}
	for i := 0; i < 6; i++ {
		miniblocks[i], _ = createDummyMiniBlock(fmt.Sprintf("tx hash %d", i), marshalizer, hasher, destShardId, destShards[i/2])
	}

	//put 2 metablocks in pool
	meta := &block.MetaBlock{
		Nonce:        1,
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks),
		Round:        1,
		PrevRandSeed: []byte("roothash"),
	}

	mb1Hash := []byte("meta block 1")
	tdp.MetaBlocks().Put(
		mb1Hash,
		meta,
	)

	meta = &block.MetaBlock{
		Nonce:     2,
		ShardInfo: createShardData(hasher, marshalizer, miniBlocks),
		Round:     2,
	}

	mb2Hash := []byte("meta block 2")
	tdp.MetaBlocks().Put(
		mb2Hash,
		meta,
	)

	meta = &block.MetaBlock{
		Nonce:        3,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        3,
		PrevRandSeed: []byte("roothash"),
	}

	mb3Hash := []byte("meta block 3")
	tdp.MetaBlocks().Put(
		mb3Hash,
		meta,
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	sp, _ := blproc.NewShardProcessor(arguments)

	miniBlocksReturned, usedMetaHdrsHashes, nrTxAdded, err := sp.CreateAndProcessCrossMiniBlocksDstMe(2, 2, haveTimeTrue)

	assert.Equal(t, 0, len(miniBlocksReturned))
	assert.Equal(t, uint32(0), usedMetaHdrsHashes)
	assert.Equal(t, uint32(0), nrTxAdded)
	assert.Nil(t, err)
}

//------- createMiniBlocks

func TestShardProcessor_CreateMiniBlocksShouldWorkWithIntraShardTxs(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()

	//we will have a 3 txs in pool

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(0)
	receiverShardId := uint32(0)

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	datapool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  string(txHash1),
	}, cacheId)
	datapool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  string(txHash2),
	}, cacheId)
	datapool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  string(txHash3),
	}, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	txProcessorMock := &mock.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) error {
			//execution, in this context, means moving the tx nonce to itx corresponding execution result variable
			if transaction.Data == string(txHash1) {
				tx1ExecutionResult = transaction.Nonce
			}
			if transaction.Data == string(txHash2) {
				tx2ExecutionResult = transaction.Nonce
			}
			if transaction.Data == string(txHash3) {
				tx3ExecutionResult = transaction.Nonce
			}

			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	accntAdapter := &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}

	totalGasConsumed := uint64(0)
	factory, _ := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		initStore(),
		marshalizer,
		hasher,
		datapool,
		&mock.AddressConverterMock{},
		accntAdapter,
		&mock.RequestHandlerMock{},
		txProcessorMock,
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 0
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&mock.MiniBlocksCompacterMock{
			CompactCalled: func(miniBlocks block.MiniBlockSlice, mapHashesAndTxs map[string]data.TransactionHandler) block.MiniBlockSlice {
				return miniBlocks
			},
		},
		&mock.GasHandlerMock{
			SetGasConsumedCalled: func(gasConsumed uint64, hash []byte) {
				totalGasConsumed += gasConsumed
			},
			TotalGasConsumedCalled: func() uint64 {
				return totalGasConsumed
			},
			ComputeGasConsumedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		accntAdapter,
		datapool.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.Accounts = accntAdapter
	arguments.TxCoordinator = tc
	bp, _ := blproc.NewShardProcessor(arguments)

	blockBody, err := bp.CreateMiniBlocks(15000, 0, func() bool { return true })

	assert.Nil(t, err)
	//testing execution
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
	//one miniblock output
	assert.Equal(t, 1, len(blockBody))
	//miniblock should have 3 txs
	assert.Equal(t, 3, len(blockBody[0].TxHashes))
	//testing all 3 hashes are present in block body
	assert.True(t, isInTxHashes(txHash1, blockBody[0].TxHashes))
	assert.True(t, isInTxHashes(txHash2, blockBody[0].TxHashes))
	assert.True(t, isInTxHashes(txHash3, blockBody[0].TxHashes))
}

func TestShardProcessor_GetProcessedMetaBlockFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	//we have 3 metablocks in pool each containing 2 miniblocks.
	//blockbody will have 2 + 1 miniblocks from 2 out of the 3 metablocks
	//The test should remove only one metablock

	destShardId := uint32(2)

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()

	miniblockHashes := make([][]byte, 6)

	destShards := []uint32{1, 3, 4}
	for i := 0; i < 6; i++ {
		_, hash := createDummyMiniBlock(fmt.Sprintf("tx hash %d", i), marshalizer, hasher, destShardId, destShards[i/2])
		miniblockHashes[i] = hash
	}

	//put 3 metablocks in pool
	metaBlockHash1 := []byte("meta block 1")
	metaBlock1 := createDummyMetaBlock(destShardId, destShards[0], miniblockHashes[0], miniblockHashes[1])
	datapool.MetaBlocks().Put(
		metaBlockHash1,
		metaBlock1,
	)
	metaBlockHash2 := []byte("meta block 2")
	metaBlock2 := createDummyMetaBlock(destShardId, destShards[1], miniblockHashes[2], miniblockHashes[3])
	datapool.MetaBlocks().Put(
		metaBlockHash2,
		metaBlock2,
	)
	metaBlockHash3 := []byte("meta block 3")
	metaBlock3 := createDummyMetaBlock(destShardId, destShards[2], miniblockHashes[4], miniblockHashes[5])
	datapool.MetaBlocks().Put(
		metaBlockHash3,
		metaBlock3,
	)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = destShardId
	shardCoordinator.SetNoShards(destShardId + 1)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.ShardCoordinator = shardCoordinator
	arguments.ForkDetector = &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	arguments.StartHeaders = createGenesisBlocks(shardCoordinator)
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.SetHdrForCurrentBlock(metaBlockHash1, metaBlock1, true)
	bp.SetHdrForCurrentBlock(metaBlockHash2, metaBlock2, true)
	bp.SetHdrForCurrentBlock(metaBlockHash3, metaBlock3, true)

	//create mini block headers with first 3 miniblocks from miniblocks var
	mbHeaders := []block.MiniBlockHeader{
		{Hash: miniblockHashes[0]},
		{Hash: miniblockHashes[1]},
		{Hash: miniblockHashes[2]},
	}

	hashes := [][]byte{
		metaBlockHash1,
		metaBlockHash2,
		metaBlockHash3,
	}

	blockHeader := &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	err := bp.AddProcessedCrossMiniBlocksFromHeader(blockHeader)

	assert.Nil(t, err)
	//check WasMiniBlockProcessed for remaining metablocks
	assert.True(t, bp.IsMiniBlockProcessed(metaBlockHash2, miniblockHashes[2]))
	assert.False(t, bp.IsMiniBlockProcessed(metaBlockHash2, miniblockHashes[3]))

	assert.False(t, bp.IsMiniBlockProcessed(metaBlockHash3, miniblockHashes[4]))
	assert.False(t, bp.IsMiniBlockProcessed(metaBlockHash3, miniblockHashes[5]))
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilBlockHeader(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	be, _ := blproc.NewShardProcessor(arguments)
	err := be.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilTxBlockBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.RestoreBlockIntoPools(&block.Header{}, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestShardProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx hash 1")

	datapool := mock.NewPoolsHolderMock()
	marshalizerMock := &mock.MarshalizerMock{}
	hasherMock := &mock.HasherStub{}

	body := make(block.Body, 0)
	tx := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}
	buffTx, _ := marshalizerMock.Marshal(tx)

	store := &mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			m := make(map[string][]byte, 0)
			m[string(txHash)] = buffTx
			return m, nil
		},
	}

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		store,
		marshalizerMock,
		hasherMock,
		datapool,
		&mock.AddressConverterMock{},
		initAccountsMock(),
		&mock.RequestHandlerMock{},
		&mock.TxProcessorMock{},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
		&mock.RewardTxProcessorMock{},
		&mock.IntermediateTransactionHandlerMock{},
		&mock.FeeHandlerStub{},
		&mock.MiniBlocksCompacterMock{},
		&mock.GasHandlerMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		datapool.MiniBlocks(),
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
		&mock.GasHandlerMock{},
	)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasherMock
	arguments.Marshalizer = marshalizerMock
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	miniblockHash := []byte("mini block hash 1")
	hasherMock.ComputeCalled = func(s string) []byte {
		return miniblockHash
	}

	metablockHash := []byte("meta block hash 1")
	metablockHeader := createDummyMetaBlock(0, 1, miniblockHash)
	datapool.MetaBlocks().Put(
		metablockHash,
		metablockHeader,
	)

	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		return &mock.StorerStub{
			RemoveCalled: func(key []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return marshalizerMock.Marshal(metablockHeader)
			},
		}
	}

	miniBlockHeader := block.MiniBlockHeader{
		Hash:            miniblockHash,
		SenderShardID:   miniblock.SenderShardID,
		ReceiverShardID: miniblock.ReceiverShardID,
	}

	metaBlockHashes := make([][]byte, 0)
	metaBlockHashes = append(metaBlockHashes, metablockHash)

	err = sp.RestoreBlockIntoPools(&block.Header{MetaBlockHashes: [][]byte{metablockHash}, MiniBlockHeaders: []block.MiniBlockHeader{miniBlockHeader}}, body)

	miniblockFromPool, _ := datapool.MiniBlocks().Get(miniblockHash)
	txFromPool, _ := datapool.Transactions().SearchFirstData(txHash)
	assert.Nil(t, err)
	assert.Equal(t, &miniblock, miniblockFromPool)
	assert.Equal(t, tx, txFromPool)
	assert.Equal(t, false, sp.IsMiniBlockProcessed(metablockHash, miniblockHash))
}

func TestShardProcessor_DecodeBlockBody(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	marshalizerMock := &mock.MarshalizerMock{}
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizerMock
	sp, err := blproc.NewShardProcessor(arguments)
	body := make(block.Body, 0)
	body = append(body, &block.MiniBlock{ReceiverShardID: 69})
	message, err := marshalizerMock.Marshal(body)
	assert.Nil(t, err)

	dcdBlk := sp.DecodeBlockBody(nil)
	assert.Nil(t, dcdBlk)

	dcdBlk = sp.DecodeBlockBody(message)
	assert.Equal(t, body, dcdBlk)
	assert.Equal(t, uint32(69), body[0].ReceiverShardID)
}

func TestShardProcessor_DecodeBlockHeader(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizerMock := &mock.MarshalizerMock{}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = tdp
	arguments.Marshalizer = marshalizerMock
	sp, err := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(0)
	hdr.Signature = []byte("A")
	message, err := marshalizerMock.Marshal(hdr)
	assert.Nil(t, err)

	message, err = marshalizerMock.Marshal(hdr)
	assert.Nil(t, err)

	dcdHdr := sp.DecodeBlockHeader(nil)
	assert.Nil(t, dcdHdr)

	dcdHdr = sp.DecodeBlockHeader(message)
	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte("A"), dcdHdr.GetSignature())
}

func TestShardProcessor_IsHdrConstructionValid(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := initDataPool([]byte("tx_hash1"))

	shardNr := uint32(5)
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.StartHeaders = createGenesisBlocks(arguments.ShardCoordinator)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    44,
		RandSeed: prevRandSeed}
	notarizedHdrs[sharding.MetachainShardId] = append(notarizedHdrs[sharding.MetachainShardId], lastHdr)

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(sharding.MetachainShardId).(*block.MetaBlock))
	prevHdr := &block.MetaBlock{
		Round:        10,
		Nonce:        45,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = sp.ComputeHeaderHash(prevHdr)
	currHdr := &block.MetaBlock{
		Round:        11,
		Nonce:        46,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}

	err := sp.IsHdrConstructionValid(nil, prevHdr)
	assert.Equal(t, err, process.ErrNilBlockHeader)

	err = sp.IsHdrConstructionValid(currHdr, nil)
	assert.Equal(t, err, process.ErrNilBlockHeader)

	currHdr.Nonce = 0
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrWrongNonceInBlock)

	currHdr.Nonce = 0
	prevHdr.Nonce = 0
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrRootStateDoesNotMatch)

	currHdr.Nonce = 0
	prevHdr.Nonce = 0
	prevHdr.RootHash = nil
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)

	currHdr.Nonce = 46
	prevHdr.Nonce = 45
	prevHdr.Round = currHdr.Round + 1
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrLowerRoundInBlock)

	prevHdr.Round = currHdr.Round - 1
	currHdr.Nonce = prevHdr.Nonce + 2
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrWrongNonceInBlock)

	currHdr.Nonce = prevHdr.Nonce + 1
	currHdr.PrevHash = []byte("wronghash")
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrBlockHashDoesNotMatch)

	prevHdr.RandSeed = []byte("randomwrong")
	currHdr.PrevHash, _ = sp.ComputeHeaderHash(prevHdr)
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrRandSeedDoesNotMatch)

	currHdr.PrevHash = prevHash
	prevHdr.RandSeed = currRandSeed
	prevHdr.RootHash = []byte("prevRootHash")
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)
}

func TestShardProcessor_RemoveAndSaveLastNotarizedMetaHdrNoDstMB(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()
	forkDetector := &mock.ForkDetectorMock{}
	highNonce := uint64(500)
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return highNonce
	}

	putCalledNr := 0
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			putCalledNr++
			return nil
		},
	}

	shardNr := uint32(5)
	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.ForkDetector = forkDetector
	arguments.StartHeaders = createGenesisBlocks(arguments.ShardCoordinator)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	notarizedHdrs[sharding.MetachainShardId] = append(notarizedHdrs[sharding.MetachainShardId], lastHdr)

	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(sharding.MetachainShardId).(*block.MetaBlock))
	prevHdr := &block.MetaBlock{
		Round:        10,
		Nonce:        45,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash")}

	prevHash, _ = sp.ComputeHeaderHash(prevHdr)
	currHdr := &block.MetaBlock{
		Round:        11,
		Nonce:        46,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash")}
	currHash, _ := sp.ComputeHeaderHash(currHdr)
	prevHash, _ = sp.ComputeHeaderHash(prevHdr)

	shardHdr := &block.Header{Round: 15}
	mbHeaders := make([]block.MiniBlockHeader, 0)
	blockHeader := &block.Header{}

	// test header not in pool and defer called
	processedMetaHdrs, err := sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetaBlocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 0, putCalledNr)

	notarizedHdrs = sp.NotarizedHdrs()
	assert.Equal(t, firstNonce, sp.LastNotarizedHdrForShard(sharding.MetachainShardId).GetNonce())
	assert.Equal(t, 0, len(processedMetaHdrs))

	// wrong header type in pool and defer called
	datapool.MetaBlocks().Put(currHash, shardHdr)
	sp.SetHdrForCurrentBlock(currHash, shardHdr, true)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	blockHeader = &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err = sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetaBlocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 0, putCalledNr)

	notarizedHdrs = sp.NotarizedHdrs()
	assert.Equal(t, firstNonce, sp.LastNotarizedHdrForShard(sharding.MetachainShardId).GetNonce())

	// put headers in pool
	datapool.MetaBlocks().Put(currHash, currHdr)
	datapool.MetaBlocks().Put(prevHash, prevHdr)

	sp.CreateBlockStarted()
	sp.SetHdrForCurrentBlock(currHash, currHdr, true)
	sp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	hashes = make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader = &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err = sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetaBlocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 4, putCalledNr)

	assert.Equal(t, currHdr, sp.LastNotarizedHdrForShard(sharding.MetachainShardId))
}

func createShardData(hasher hashing.Hasher, marshalizer marshal.Marshalizer, miniBlocks []block.MiniBlock) []block.ShardData {
	shardData := make([]block.ShardData, len(miniBlocks))
	for i := 0; i < len(miniBlocks); i++ {
		marshaled, _ := marshalizer.Marshal(miniBlocks[i])
		hashed, _ := core.CalculateHash(marshalizer, hasher, string(marshaled))

		shardMBHeader := block.ShardMiniBlockHeader{
			ReceiverShardID: miniBlocks[i].ReceiverShardID,
			SenderShardID:   miniBlocks[i].SenderShardID,
			TxCount:         uint32(len(miniBlocks[i].TxHashes)),
			Hash:            hashed,
		}
		shardMBHeaders := make([]block.ShardMiniBlockHeader, 0)
		shardMBHeaders = append(shardMBHeaders, shardMBHeader)

		shardData[0].ShardID = miniBlocks[i].SenderShardID
		shardData[0].TxCount = 10
		shardData[0].HeaderHash = []byte("headerHash")
		shardData[0].ShardMiniBlockHeaders = shardMBHeaders
	}

	return shardData
}

func TestShardProcessor_RemoveAndSaveLastNotarizedMetaHdrNotAllMBFinished(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()
	forkDetector := &mock.ForkDetectorMock{}
	highNonce := uint64(500)
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return highNonce
	}

	putCalledNr := 0
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			putCalledNr++
			return nil
		},
	}

	shardNr := uint32(5)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.ForkDetector = forkDetector
	arguments.StartHeaders = createGenesisBlocks(arguments.ShardCoordinator)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	notarizedHdrs[sharding.MetachainShardId] = append(notarizedHdrs[sharding.MetachainShardId], lastHdr)

	txHash := []byte("txhash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock1 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	miniblock2 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   2,
		TxHashes:        txHashes,
	}
	miniblock3 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   3,
		TxHashes:        txHashes,
	}
	miniblock4 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   4,
		TxHashes:        txHashes,
	}
	mbHeaders := make([]block.MiniBlockHeader, 0)

	marshaled, _ := marshalizer.Marshal(miniblock1)
	hashed, _ := core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	marshaled, _ = marshalizer.Marshal(miniblock2)
	hashed, _ = core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	marshaled, _ = marshalizer.Marshal(miniblock3)
	hashed, _ = core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)
	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(sharding.MetachainShardId).(*block.MetaBlock))
	prevHdr := &block.MetaBlock{
		Round:        10,
		Nonce:        45,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash"),
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks)}

	miniBlocks = make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock3, miniblock4)
	prevHash, _ = sp.ComputeHeaderHash(prevHdr)
	currHdr := &block.MetaBlock{
		Round:        11,
		Nonce:        46,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash"),
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks)}
	currHash, _ := sp.ComputeHeaderHash(currHdr)
	prevHash, _ = sp.ComputeHeaderHash(prevHdr)

	// put headers in pool
	datapool.MetaBlocks().Put(currHash, currHdr)
	datapool.MetaBlocks().Put(prevHash, prevHdr)

	sp.SetHdrForCurrentBlock(currHash, currHdr, true)
	sp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader := &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err := sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetaBlocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 2, putCalledNr)

	assert.Equal(t, prevHdr, sp.LastNotarizedHdrForShard(sharding.MetachainShardId))
}

func TestShardProcessor_RemoveAndSaveLastNotarizedMetaHdrAllMBFinished(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := mock.NewPoolsHolderMock()
	forkDetector := &mock.ForkDetectorMock{}
	highNonce := uint64(500)
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return highNonce
	}
	putCalledNr := 0
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			putCalledNr++
			return nil
		},
	}

	shardNr := uint32(5)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.ShardCoordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.ForkDetector = forkDetector
	arguments.StartHeaders = createGenesisBlocks(arguments.ShardCoordinator)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	notarizedHdrs[sharding.MetachainShardId] = append(notarizedHdrs[sharding.MetachainShardId], lastHdr)

	txHash := []byte("txhash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock1 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	miniblock2 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   2,
		TxHashes:        txHashes,
	}
	miniblock3 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   3,
		TxHashes:        txHashes,
	}
	miniblock4 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   4,
		TxHashes:        txHashes,
	}

	mbHeaders := make([]block.MiniBlockHeader, 0)

	marshaled, _ := marshalizer.Marshal(miniblock1)
	hashed, _ := core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	marshaled, _ = marshalizer.Marshal(miniblock2)
	hashed, _ = core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	marshaled, _ = marshalizer.Marshal(miniblock3)
	hashed, _ = core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	marshaled, _ = marshalizer.Marshal(miniblock4)
	hashed, _ = core.CalculateHash(marshalizer, hasher, string(marshaled))
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)
	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(sharding.MetachainShardId).(*block.MetaBlock))
	prevHdr := &block.MetaBlock{
		Round:        10,
		Nonce:        45,
		PrevRandSeed: prevRandSeed,
		RandSeed:     currRandSeed,
		PrevHash:     prevHash,
		RootHash:     []byte("prevRootHash"),
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks)}

	miniBlocks = make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock3, miniblock4)
	prevHash, _ = sp.ComputeHeaderHash(prevHdr)
	currHdr := &block.MetaBlock{
		Round:        11,
		Nonce:        46,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash"),
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks)}
	currHash, _ := sp.ComputeHeaderHash(currHdr)
	prevHash, _ = sp.ComputeHeaderHash(prevHdr)

	// put headers in pool
	datapool.MetaBlocks().Put(currHash, currHdr)
	datapool.MetaBlocks().Put(prevHash, prevHdr)
	datapool.MetaBlocks().Put([]byte("shouldNotRemove"), &block.MetaBlock{
		Round:        12,
		PrevRandSeed: []byte("nextrand"),
		PrevHash:     currHash,
		Nonce:        47})

	sp.SetHdrForCurrentBlock(currHash, currHdr, true)
	sp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader := &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err := sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(processedMetaHdrs))

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetaBlocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 4, putCalledNr)

	assert.Equal(t, currHdr, sp.LastNotarizedHdrForShard(sharding.MetachainShardId))
}

func createOneHeaderOneBody() (*block.Header, block.Body) {
	txHash := []byte("tx_hash1")
	rootHash := []byte("rootHash")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))
	mbHdr := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	hdr := &block.Header{
		Nonce:            1,
		PrevHash:         []byte(""),
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}

	return hdr, body
}

func TestShardProcessor_CheckHeaderBodyCorrelationReceiverMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].ReceiverShardID = body[0].ReceiverShardID + 1
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationSenderMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].SenderShardID = body[0].SenderShardID + 1
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationTxCountMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].TxCount = uint32(len(body[0].TxHashes) + 1)
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationHashMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].Hash = []byte("wrongHash")
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationShouldPass(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Nil(t, err)
}

func TestShardProcessor_RestoreMetaBlockIntoPoolShouldPass(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	poolFake := mock.NewPoolsHolderMock()

	metaBlock := block.MetaBlock{
		Nonce:     1,
		ShardInfo: make([]block.ShardData, 0),
	}

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = poolFake
	arguments.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
				GetCalled: func(key []byte) ([]byte, error) {
					return marshalizer.Marshal(&metaBlock)
				},
			}
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	miniblockHashes := make(map[string]uint32, 0)

	meta := block.MetaBlock{
		Nonce:     1,
		ShardInfo: make([]block.ShardData, 0),
	}
	hasher := &mock.HasherStub{}

	metaBytes, _ := marshalizer.Marshal(meta)
	hasher.ComputeCalled = func(s string) []byte {
		return []byte("cool")
	}
	metaHash := hasher.Compute(string(metaBytes))
	metablockHashes := make([][]byte, 0)
	metablockHashes = append(metablockHashes, metaHash)

	metaBlockRestored, ok := poolFake.MetaBlocks().Get(metaHash)

	assert.Equal(t, nil, metaBlockRestored)
	assert.False(t, ok)

	err := sp.RestoreMetaBlockIntoPool(miniblockHashes, metablockHashes)

	metaBlockRestored, _ = poolFake.MetaBlocks().Get(metaHash)

	assert.Equal(t, &metaBlock, metaBlockRestored)
	assert.Nil(t, err)
}

func TestShardPreprocessor_getAllMiniBlockDstMeFromMetaShouldPass(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txHash := []byte("tx_hash1")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	hasher := &mock.HasherStub{}

	mbbytes, _ := marshalizer.Marshal(miniblock)
	mbHash := hasher.Compute(string(mbbytes))

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   2,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash,
	}
	shardMiniblockHdrs := make([]block.ShardMiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardID:               1,
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)
	metaBlock := &block.MetaBlock{Nonce: 1, Round: 1, ShardInfo: shardHdrs}

	idp := initDataPool([]byte("tx_hash1"))

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = idp
	sp, _ := blproc.NewShardProcessor(arguments)

	metaBytes, _ := marshalizer.Marshal(metaBlock)
	hasher.ComputeCalled = func(s string) []byte {
		return []byte("cool")
	}
	metaHash := hasher.Compute(string(metaBytes))
	sp.SetHdrForCurrentBlock(metaHash, metaBlock, true)

	metablockHashes := make([][]byte, 0)
	metablockHashes = append(metablockHashes, metaHash)
	header := &block.Header{Nonce: 1, Round: 1, MetaBlockHashes: metablockHashes}

	orderedMetaBlocks, err := sp.GetAllMiniBlockDstMeFromMeta(header)

	assert.Equal(t, 1, len(orderedMetaBlocks))
	assert.Equal(t, orderedMetaBlocks[""], metaHash)
	assert.Nil(t, err)
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachainNothingToProcess(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArgumentsMultiShard()
	sp, _ := blproc.NewShardProcessor(arguments)
	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(nil)

	assert.NotNil(t, hdrs)
	assert.Equal(t, 0, len(hdrs))
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachaiMetaHdrsWithoutOwnHdr(t *testing.T) {
	t.Parallel()

	processedHdrs := make([]data.HeaderHandler, 0)
	datapool := createTestShardDataPool()
	store := initStore()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.StartHeaders = genesisBlocks

	sp, _ := blproc.NewShardProcessor(arguments)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: []byte("hash"), ShardID: 1})
	_ = datapool.Headers().Put([]byte("hash"), &block.Header{ShardId: 0, Nonce: 1})

	prevMetaHdr := genesisBlocks[sharding.MetachainShardId]
	prevHash, _ := core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr := &block.MetaBlock{
		Nonce:        1,
		Epoch:        0,
		Round:        1,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
		ShardInfo:    shardInfo,
	}
	currHash, _ := core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	prevMetaHdr = currMetaHdr
	prevHash, _ = core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr = &block.MetaBlock{
		Nonce:        2,
		Epoch:        0,
		Round:        2,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
		ShardInfo:    shardInfo,
	}
	currHash, _ = core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(processedHdrs)

	assert.NotNil(t, hdrs)
	assert.Equal(t, 0, len(hdrs))
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachaiMetaHdrsWithOwnHdrButNotStored(t *testing.T) {
	t.Parallel()

	processedHdrs := make([]data.HeaderHandler, 0)
	datapool := createTestShardDataPool()
	store := initStore()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.StartHeaders = genesisBlocks

	sp, _ := blproc.NewShardProcessor(arguments)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: []byte("hash"), ShardID: 0})

	prevMetaHdr := genesisBlocks[sharding.MetachainShardId]
	prevHash, _ := core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr := &block.MetaBlock{
		Nonce:        1,
		Epoch:        0,
		Round:        1,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
		ShardInfo:    shardInfo,
	}
	currHash, _ := core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	prevMetaHdr = currMetaHdr
	prevHash, _ = core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr = &block.MetaBlock{
		Nonce:        2,
		Epoch:        0,
		Round:        2,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
		ShardInfo:    shardInfo,
	}
	currHash, _ = core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(processedHdrs)

	assert.Nil(t, hdrs)
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachaiMetaHdrsWithOwnHdrStored(t *testing.T) {
	t.Parallel()

	processedHdrs := make([]data.HeaderHandler, 0)
	datapool := createTestShardDataPool()
	store := initStore()
	hasher := &mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = datapool
	arguments.Store = store
	arguments.Hasher = hasher
	arguments.Marshalizer = marshalizer
	arguments.StartHeaders = genesisBlocks

	sp, _ := blproc.NewShardProcessor(arguments)

	ownHdr := &block.Header{
		Nonce: 1,
		Round: 1,
	}
	ownHash, _ := core.CalculateHash(marshalizer, hasher, ownHdr)
	_ = datapool.Headers().Put(ownHash, ownHdr)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: ownHash, ShardID: 0})

	prevMetaHdr := genesisBlocks[sharding.MetachainShardId]
	prevHash, _ := core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr := &block.MetaBlock{
		Nonce:        1,
		Epoch:        0,
		Round:        1,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
		ShardInfo:    shardInfo,
	}
	currHash, _ := core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)

	ownHdr = &block.Header{
		Nonce: 2,
		Round: 2,
	}
	ownHash, _ = core.CalculateHash(marshalizer, hasher, ownHdr)
	mrsOwnHdr, _ := marshalizer.Marshal(ownHdr)
	_ = store.Put(dataRetriever.BlockHeaderUnit, ownHash, mrsOwnHdr)

	shardInfo = make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: ownHash, ShardID: 0})

	prevMetaHdr = currMetaHdr
	prevHash, _ = core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr = &block.MetaBlock{
		Nonce:        2,
		Epoch:        0,
		Round:        2,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
		ShardInfo:    shardInfo,
	}
	currHash, _ = core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	prevMetaHdr = currMetaHdr
	prevHash, _ = core.CalculateHash(marshalizer, hasher, prevMetaHdr)
	currMetaHdr = &block.MetaBlock{
		Nonce:        3,
		Epoch:        0,
		Round:        3,
		PrevHash:     prevHash,
		PrevRandSeed: prevMetaHdr.GetRandSeed(),
		RandSeed:     prevMetaHdr.GetRandSeed(),
	}
	currHash, _ = core.CalculateHash(marshalizer, hasher, currMetaHdr)
	_ = datapool.MetaBlocks().Put(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(processedHdrs)

	assert.NotNil(t, hdrs)
	assert.Equal(t, ownHdr.GetNonce(), hdrs[0].GetNonce())
}

func TestShardProcessor_RestoreMetaBlockIntoPoolVerifyMiniblocks(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	poolMock := mock.NewPoolsHolderMock()

	storer := &mock.ChainStorerMock{}
	shardC := mock.NewMultiShardsCoordinatorMock(3)

	arguments := CreateMockArgumentsMultiShard()
	arguments.DataPool = poolMock
	arguments.Store = storer
	arguments.ShardCoordinator = shardC
	arguments.StartHeaders = createGenesisBlocks(shardC)
	arguments.Rounder = &mock.RounderMock{}
	sp, _ := blproc.NewShardProcessor(arguments)

	miniblockHashes := make(map[string]uint32, 0)

	testMBHash := []byte("hash")
	shardMBHdr := block.ShardMiniBlockHeader{
		Hash:            testMBHash,
		SenderShardID:   shardC.SelfId() + 1,
		ReceiverShardID: shardC.SelfId(),
	}
	shardMBHeaders := make([]block.ShardMiniBlockHeader, 0)
	shardMBHeaders = append(shardMBHeaders, shardMBHdr)

	shardHdr := block.ShardData{ShardMiniBlockHeaders: shardMBHeaders, ShardID: shardC.SelfId() + 1}

	shardInfos := make([]block.ShardData, 0)
	shardInfos = append(shardInfos, shardHdr)

	meta := &block.MetaBlock{
		Nonce:     1,
		ShardInfo: shardInfos,
	}

	hasher := &mock.HasherStub{}

	metaBytes, _ := marshalizer.Marshal(meta)
	hasher.ComputeCalled = func(s string) []byte {
		return []byte("cool")
	}
	metaHash := hasher.Compute(string(metaBytes))
	sp.AddProcessedMiniBlock(metaHash, testMBHash)
	metablockHashes := make([][]byte, 0)
	metablockHashes = append(metablockHashes, metaHash)

	metaBlockRestored, ok := poolMock.MetaBlocks().Get(metaHash)

	assert.Equal(t, nil, metaBlockRestored)
	assert.False(t, ok)

	storer.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		return metaBytes, nil
	}
	storer.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		return &mock.StorerStub{
			RemoveCalled: func(key []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return metaBytes, nil
			},
		}
	}

	err := sp.RestoreMetaBlockIntoPool(miniblockHashes, metablockHashes)

	metaBlockRestored, _ = poolMock.MetaBlocks().Get(metaHash)

	assert.Equal(t, meta, metaBlockRestored)
	assert.Nil(t, err)
	assert.True(t, sp.IsMiniBlockProcessed(metaHash, testMBHash))
}
