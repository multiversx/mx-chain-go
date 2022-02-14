package block_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	atomicCore "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/blockchain"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MaxGasLimitPerBlock = uint64(100000)

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

// ------- NewShardProcessor

func initAccountsMock() *stateMock.AccountsStub {
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}
	return &stateMock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
}

func initBasicTestData() (*dataRetrieverMock.PoolsHolderMock, data.ChainHandler, []byte, *block.Body, [][]byte, *hashingMocks.HasherMock, *mock.MarshalizerMock, error, []byte) {
	tdp := dataRetrieverMock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	randSeed := []byte("rand seed")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, 0, process.ShardCacherIdentifier(1, 0))
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Round:    1,
			Nonce:    1,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	rootHash := []byte("rootHash")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	return hdr
}

func CreateCoreComponentsMultiShard() (
	*mock.CoreComponentsMock,
	*mock.DataComponentsMock,
	*mock.BootstrapComponentsMock,
	*mock.StatusComponentsMock,
) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.BlockChain, _ = blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = dataComponents.BlockChain.SetGenesisHeader(&block.Header{Nonce: 0})
	dataComponents.DataPool = initDataPool([]byte("tx_hash1"))
	bootstrapComponents.Coordinator = mock.NewMultiShardsCoordinatorMock(3)

	return coreComponents, dataComponents, bootstrapComponents, statusComponents
}

func CreateMockArgumentsMultiShard(
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
) blproc.ArgShardProcessor {

	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = initAccountsMock()

	return arguments
}

// ------- NewBlockProcessor

func TestNewShardProcessor(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

	tests := []struct {
		args        func() blproc.ArgShardProcessor
		expectedErr error
	}{
		{
			args: func() blproc.ArgShardProcessor {
				return CreateMockArgumentsMultiShard(coreComponents, nil, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilDataComponentsHolder,
		},
		{
			args: func() blproc.ArgShardProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.DataPool = nil

				return CreateMockArgumentsMultiShard(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilDataPoolHolder,
		},
		{
			args: func() blproc.ArgShardProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.DataPool = &dataRetrieverMock.PoolsHolderStub{
					HeadersCalled: func() dataRetriever.HeadersPool {
						return nil
					},
				}
				return CreateMockArgumentsMultiShard(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilHeadersDataPool,
		},
		{
			args: func() blproc.ArgShardProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.DataPool = &dataRetrieverMock.PoolsHolderStub{
					HeadersCalled: func() dataRetriever.HeadersPool {
						return &mock.HeadersCacherStub{}
					},
					TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
						return nil
					},
				}
				return CreateMockArgumentsMultiShard(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilTransactionPool,
		},
		{
			args: func() blproc.ArgShardProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.BlockChain = &testscommon.ChainHandlerStub{
					GetGenesisHeaderCalled: func() data.HeaderHandler {
						return nil
					},
				}
				return CreateMockArgumentsMultiShard(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilHeaderHandler,
		},
		{
			args: func() blproc.ArgShardProcessor {
				return CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		sp, err := blproc.NewShardProcessor(test.args())
		if test.expectedErr != nil {
			require.Nil(t, sp)
			require.Error(t, err)
			require.True(t, strings.Contains(err.Error(), test.expectedErr.Error()))
		} else {
			require.NotNil(t, sp)
			require.Nil(t, err)
		}
	}
}

// ------- ProcessBlock

func TestShardProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	body := &block.Body{}
	err := sp.ProcessBlock(nil, body, haveTime)

	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestShardProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	err := sp.ProcessBlock(&block.Header{}, nil, haveTime)

	assert.Equal(t, process.ErrNilBlockBody, err)
}

func TestShardProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	blk := &block.Body{}
	err := sp.ProcessBlock(&block.Header{}, blk, nil)

	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestShardProcess_CreateNewBlockHeaderProcessHeaderExpectCheckRoundCalled(t *testing.T) {
	t.Parallel()

	round := uint64(4)
	checkRoundCt := atomicCore.Counter{}

	roundNotifier := &mock.RoundNotifierStub{
		CheckRoundCalled: func(r uint64) {
			checkRoundCt.Increment()
			require.Equal(t, round, r)
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.RoundNotifier = roundNotifier

	shardProcessor, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: round}
	bodyHandler, _ := shardProcessor.CreateBlockBody(header, func() bool { return true })

	headerHandler, err := shardProcessor.CreateNewHeader(round, 1)
	require.Nil(t, err)
	require.Equal(t, int64(1), checkRoundCt.Get())

	err = shardProcessor.ProcessBlock(headerHandler, bodyHandler, func() time.Duration { return time.Second })
	require.Nil(t, err)
	require.Equal(t, int64(2), checkRoundCt.Get())
}

func TestShardProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	hdr := block.Header{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		PrevRandSeed:  []byte("rand seed"),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
	}

	body := &block.Body{}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}
	sp, _ := blproc.NewShardProcessor(arguments)
	err := sp.ProcessBlock(&hdr, body, haveTime)

	assert.NotNil(t, err)
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestShardProcessor_ProcessBlockHeaderBodyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	hdr := block.Header{
		Nonce:         1,
		PrevHash:      []byte(""),
		PrevRandSeed:  []byte("rand seed"),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("00110"),
		ShardID:       0,
		RootHash:      []byte("rootHash"),
	}
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error { return nil }
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
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
	err := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")

	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
		RootHash:         []byte("rootHash"),
		MiniBlockHeaders: mbHdrs,
	}
	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error { return nil }
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}

	accounts := &stateMock.AccountsStub{
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
		createMockPubkeyConverter(),
		accounts,
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return 0, process.ErrHigherNonceInTransaction
			},
		},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&testscommon.GasHandlerStub{
			ComputeGasProvidedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled:    func(gasRefunded uint64, hash []byte) {},
			RemoveGasRefundedCalled: func(hashes [][]byte) {},
			RemoveGasProvidedCalled: func(hashes [][]byte) {},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)
	container, _ := factory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(accounts, tdp, container)
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		InitCalled: func() {
		},
	}
	tc, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err = sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrReceiptsHashMissmatch, err)
}

func TestShardProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
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
	body := &block.Body{}
	err := sp.ProcessBlock(hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestShardProcessor_ProcessWithHeaderNotCorrectNonceShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
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
	body := &block.Body{}
	err := sp.ProcessBlock(hdr, body, haveTime)

	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestShardProcessor_ProcessWithHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	randSeed := []byte("rand seed")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.ForkDetector = &mock.ForkDetectorMock{
		ProbableHighestNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}

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
	body := &block.Body{}
	err := sp.ProcessBlock(hdr, body, haveTime)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)
}

func TestShardProcessor_ProcessBlockWithErrOnProcessBlockTransactionsCallShouldRevertState(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	randSeed := []byte("rand seed")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
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
	txProcess := func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
		return 0, err
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	tpm := &testscommon.TxProcessorMock{ProcessTransactionCalled: txProcess}
	store := &mock.ChainStorerMock{}
	accounts := &stateMock.AccountsStub{
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
		createMockPubkeyConverter(),
		accounts,
		&testscommon.RequestHandlerStub{},
		tpm,
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&testscommon.GasHandlerStub{
			ComputeGasProvidedByMiniBlockCalled: func(miniBlock *block.MiniBlock, mapHashTx map[string]data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			TotalGasProvidedCalled: func() uint64 {
				return 0
			},
			SetGasRefundedCalled:    func(gasRefunded uint64, hash []byte) {},
			RemoveGasRefundedCalled: func(hashes [][]byte) {},
			RemoveGasProvidedCalled: func(hashes [][]byte) {},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)
	container, _ := factory.Create()

	totalGasProvided := uint64(0)
	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(accounts, tdp, container)
	argsTransactionCoordinator.GasHandler = &testscommon.GasHandlerStub{
		InitCalled: func() {
			totalGasProvided = 0
		},
		TotalGasProvidedCalled: func() uint64 {
			return totalGasProvided
		},
	}
	tc, _ := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	dataComponents.BlockChain = blkc
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = accounts

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
	err2 := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrReceiptsHashMissmatch, err2)
	assert.True(t, wasCalled)
}

func TestShardProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	randSeed := []byte("rand seed")
	txHash := []byte("tx_hash1")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
		RootHash:         []byte("rootHash"),
		MiniBlockHeaders: mbHdrs,
		AccumulatedFees:  big.NewInt(0),
		DeveloperFees:    big.NewInt(0),
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

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &mock.HasherStub{}
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
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
	err := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrRootStateDoesNotMatch, err)
	assert.True(t, wasCalled)
}

func TestShardProcessor_ProcessBlockOnlyIntraShardShouldPass(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	randSeed := []byte("rand seed")
	txHash := []byte("tx_hash1")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	rootHash := []byte("rootHash")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
		AccumulatedFees:  big.NewInt(0),
		DeveloperFees:    big.NewInt(0),
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

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockCrossShardWithoutMetaShouldFail(t *testing.T) {
	t.Parallel()

	randSeed := []byte("rand seed")
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    0,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	rootHash := []byte("rootHash")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	tx := &transaction.Transaction{}
	tdp.Transactions().AddData(txHash, tx, tx.Size(), shardCoordinator.CommunicationIdentifier(0))

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
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
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(&hdr, body, haveTime)
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
	blkc.SetCurrentBlockHeaderHash(prevHash)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.MiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.MiniBlockHeader, 0)
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
	metaBytes, _ := marshalizer.Marshal(&meta)
	metaHash := hasher.Compute(string(metaBytes))

	tdp.Headers().AddHeader(metaHash, &meta)

	meta = block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(&meta)
	metaHash = hasher.Compute(string(metaBytes))

	tdp.Headers().AddHeader(metaHash, &meta)

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
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrCrossShardMBWithoutConfirmationFromMeta, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockHaveTimeLessThanZeroShouldErr(t *testing.T) {
	t.Parallel()
	txHash := []byte("tx_hash1")
	tdp := initDataPool(txHash)

	randSeed := []byte("rand seed")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    1,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	rootHash := []byte("rootHash")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
	blkc.SetCurrentBlockHeaderHash(preHash)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	hdr := block.Header{
		Round:            2,
		Nonce:            2,
		PrevHash:         preHash,
		PrevRandSeed:     randSeed,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardID:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(tdp, genesisBlocks, blkc)
	haveTimeLessThanZero := func() time.Duration {
		return -1 * time.Millisecond
	}

	// should return err
	err := sp.ProcessBlock(&hdr, body, haveTimeLessThanZero)
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
	blkc.SetCurrentBlockHeaderHash(prevHash)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.MiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.MiniBlockHeader, 0)
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
	metaBytes, _ := marshalizer.Marshal(&meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.Headers().AddHeader(metaHash, &meta)

	meta = block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(&meta)
	metaHash = hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.Headers().AddHeader(metaHash, &meta)

	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error {
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestShardProcessor_ProcessBlockWithWrongMiniBlockHeaderShouldErr(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	tdp := initDataPool(txHash)
	randSeed := []byte("rand seed")
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(
		&block.Header{
			Nonce:    1,
			RandSeed: randSeed,
		}, []byte("root hash"),
	)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	rootHash := []byte("rootHash")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	// should return err
	err := sp.ProcessBlock(&hdr, body, haveTime)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

// ------- requestEpochStartInfo
func TestShardProcessor_RequestEpochStartInfo(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

	headerHash := []byte("hash")
	headerEpoch := uint32(101)
	blockStartOfEpoch := &block.Header{
		Epoch:              headerEpoch,
		EpochStartMetaHash: headerHash,
	}
	errGetHeader := errors.New("error getting headers")

	t.Run("header is not start of epoch, should not request any header", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderCalled: func(hash []byte) {
				require.Fail(t, "should not try to request meta header")
			},
		}

		shardProc, _ := blproc.NewShardProcessor(args)
		header := &testscommon.HeaderHandlerStub{
			IsStartOfEpochBlockCalled: func() bool {
				return false
			},
		}
		err := shardProc.RequestEpochStartInfo(header, haveTime)
		time.Sleep(time.Second)
		require.Nil(t, err)
	})

	t.Run("meta epoch greater than header epoch, should not request any header", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderCalled: func(hash []byte) {
				require.Fail(t, "should not try to request meta header")
			},
		}
		args.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			MetaEpochCalled: func() uint32 {
				return headerEpoch + 1
			},
		}

		shardProc, _ := blproc.NewShardProcessor(args)

		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, haveTime)
		time.Sleep(time.Second)
		require.Nil(t, err)
	})

	t.Run("is epoch start, should not request meta header", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return true
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderCalled: func(hash []byte) {
				require.Fail(t, "should not try to request meta header")
			},
		}

		shardProc, _ := blproc.NewShardProcessor(args)
		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, haveTime)
		time.Sleep(time.Second)
		require.Nil(t, err)
	})

	t.Run("epoch start triggered while trying to request header from pool", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		flag := &atomicCore.Flag{}
		args.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				if flag.IsSet() {
					return true
				}

				flag.SetValue(true)
				return false
			},
		}
		requestsCt := &atomicCore.Counter{}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderCalled: func(hash []byte) {
				require.Equal(t, headerHash, hash)
				requestsCt.Increment()
			},
		}

		shardProc, _ := blproc.NewShardProcessor(args)
		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, haveTime)
		time.Sleep(time.Second)
		require.Nil(t, err)
		require.Equal(t, int64(1), requestsCt.Get())
	})

	t.Run("no time left", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		shardProc, _ := blproc.NewShardProcessor(args)
		noTimeFunc := func() time.Duration {
			return -1 * time.Millisecond
		}
		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, noTimeFunc)
		time.Sleep(time.Second)
		require.Equal(t, process.ErrTimeIsOut, err)
	})

	t.Run("cannot get meta header by hash first time, expect second time is returned", func(t *testing.T) {
		t.Parallel()

		metaHeaderNonce := uint64(1)
		requestsCt := &atomicCore.Counter{}
		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				require.Equal(t, headerHash, hash)

				switch requestsCt.Get() {
				case 1:
					return nil, errGetHeader
				case 2:
					return &block.Header{Nonce: metaHeaderNonce, ShardID: core.MetachainShardId}, nil
				default:
					require.Fail(t, "should not try to get header from pool more than 2 times")
					return nil, nil
				}
			},
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				require.Equal(t, metaHeaderNonce+1, hdrNonce)
				require.Equal(t, core.MetachainShardId, shardId)

				return nil, nil, nil
			},
		}
		poolsHolder := &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return headersPool
			},
			TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{}
			},
		}

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.DataComponents = &mock.DataComponentsMock{
			Storage:  &mock.ChainStorerMock{},
			DataPool: poolsHolder,
			BlockChain: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return &block.Header{}
				},
			},
		}

		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderByNonceCalled: func(nonce uint64) {
				require.Fail(t, "should not try to request meta header by nonce, expect we have it in pool")
			},
			RequestMetaHeaderCalled: func(hash []byte) {
				require.Equal(t, headerHash, hash)
				requestsCt.Increment()
			},
		}

		shardProc, _ := blproc.NewShardProcessor(args)
		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, haveTime)
		time.Sleep(time.Second)
		require.Nil(t, err)
		require.Equal(t, int64(2), requestsCt.Get())
	})

	t.Run("get headers by nonce and shard id from headers pool should work without any additional requests", func(t *testing.T) {
		t.Parallel()

		headerNonce := uint64(1)
		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				require.Equal(t, headerHash, hash)
				return &block.Header{Nonce: headerNonce, ShardID: core.MetachainShardId}, nil
			},
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				require.Equal(t, headerNonce+1, hdrNonce)
				require.Equal(t, core.MetachainShardId, shardId)
				return nil, nil, nil
			},
		}
		poolsHolder := &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return headersPool
			},
			TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{}
			},
		}

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.DataComponents = &mock.DataComponentsMock{
			Storage:  &mock.ChainStorerMock{},
			DataPool: poolsHolder,
			BlockChain: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return &block.Header{}
				},
			},
		}

		requestMetaHeaderCt := &atomicCore.Counter{}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderByNonceCalled: func(nonce uint64) {
				require.Fail(t, "should have this header in pool")
			},
			RequestMetaHeaderCalled: func(hash []byte) {
				requestMetaHeaderCt.Increment()
				require.Equal(t, headerHash, hash)
			},
		}

		shardProc, _ := blproc.NewShardProcessor(args)
		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, haveTime)
		time.Sleep(time.Second)
		require.Nil(t, err)
		require.Equal(t, int64(1), requestMetaHeaderCt.Get())
	})

	t.Run("cannot get header by nonce and shard id from headers pool from first try, request it and expect we get it", func(t *testing.T) {
		t.Parallel()

		metaHeaderNonce := uint64(1)
		wasMetaHeaderRequested := &atomicCore.Flag{}
		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				require.Equal(t, headerHash, hash)
				return &block.Header{Nonce: metaHeaderNonce, ShardID: core.MetachainShardId}, nil
			},
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				require.Equal(t, metaHeaderNonce+1, hdrNonce)
				require.Equal(t, core.MetachainShardId, shardId)
				if wasMetaHeaderRequested.IsSet() {
					return nil, nil, nil
				}

				return nil, nil, errGetHeader
			},
		}
		poolsHolder := &dataRetrieverMock.PoolsHolderStub{
			HeadersCalled: func() dataRetriever.HeadersPool {
				return headersPool
			},
			TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{}
			},
		}

		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.DataComponents = &mock.DataComponentsMock{
			Storage:  &mock.ChainStorerMock{},
			DataPool: poolsHolder,
			BlockChain: &testscommon.ChainHandlerStub{
				GetGenesisHeaderCalled: func() data.HeaderHandler {
					return &block.Header{}
				},
			},
		}
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMetaHeaderByNonceCalled: func(nonce uint64) {
				wasMetaHeaderRequested.SetValue(true)
				require.Equal(t, metaHeaderNonce+1, nonce)
			},
		}

		ct := &atomicCore.Counter{}
		timeLeft := func() time.Duration {
			ct.Increment()
			switch ct.Get() {
			case 1, 2:
				return time.Millisecond
			default:
				require.Fail(t, "should only try to get header twice")
				return -time.Millisecond
			}
		}

		shardProc, _ := blproc.NewShardProcessor(args)
		err := shardProc.RequestEpochStartInfo(blockStartOfEpoch, timeLeft)
		time.Sleep(time.Second)
		require.Nil(t, err)
		require.Equal(t, int64(2), ct.Get())
	})

}

// ------- checkAndRequestIfMetaHeadersMissing
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
	blkc.SetCurrentBlockHeaderHash(prevHash)
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	randSeed := []byte("rand seed")

	hdr := initBlockHeader(prevHash, randSeed, rootHash, mbHdrs)

	shardMiniBlock := block.MiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.MiniBlockHeader, 0)
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
	tdp.Headers().AddHeader(metaHash, meta)

	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error {
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	dataComponents.DataPool = tdp
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.RequestHandler = &testscommon.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			atomic.AddInt32(&hdrNoncesRequestCalled, 1)
		},
	}
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revertToSnapshot,
		RootHashCalled:         rootHashCalled,
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.ProcessBlock(&hdr, body, haveTime)

	meta = &block.MetaBlock{
		Nonce:        2,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        2,
		PrevRandSeed: randSeed,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))

	tdp.Headers().AddHeader(metaHash, meta)

	sp.CheckAndRequestIfMetaHeadersMissing()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&hdrNoncesRequestCalled))
	assert.Equal(t, err, process.ErrTimeIsOut)
}

// -------- requestMissingFinalityAttestingHeaders
func TestShardProcessor_RequestMissingFinalityAttestingHeaders(t *testing.T) {
	t.Parallel()

	tdp := dataRetrieverMock.NewPoolsHolderMock()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	sp.SetHighestHdrNonceForCurrentBlock(core.MetachainShardId, 1)
	res := sp.RequestMissingFinalityAttestingHeaders()
	assert.Equal(t, res > 0, true)
}

// --------- verifyIncludedMetaBlocksFinality
func TestShardProcessor_CheckMetaHeadersValidityAndFinalityShouldPass(t *testing.T) {
	t.Parallel()

	tdp := dataRetrieverMock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, 0, process.ShardCacherIdentifier(1, 0))
	rootHash := []byte("rootHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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

	shardMiniBlock := block.MiniBlockHeader{
		ReceiverShardID: mbHdr.ReceiverShardID,
		SenderShardID:   mbHdr.SenderShardID,
		TxCount:         mbHdr.TxCount,
		Hash:            mbHdr.Hash,
	}
	shardMiniblockHdrs := make([]block.MiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)

	prevMeta := genesisBlocks[core.MetachainShardId]
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

	tdp.Headers().AddHeader(metaHash1, meta1)

	prevHash, _ = core.CalculateHash(marshalizer, hasher, meta1)
	meta2 := &block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
		Round:     2,
		PrevHash:  prevHash,
	}
	metaBytes, _ = marshalizer.Marshal(meta2)
	metaHash2 := hasher.Compute(string(metaBytes))

	tdp.Headers().AddHeader(metaHash2, meta2)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)
	hdr.Round = 4

	sp.SetHdrForCurrentBlock(metaHash1, meta1, true)
	sp.SetHdrForCurrentBlock(metaHash2, meta2, false)

	err := sp.CheckMetaHeadersValidityAndFinality()
	assert.Nil(t, err)
}

func TestShardProcessor_CheckMetaHeadersValidityAndFinalityShouldReturnNilWhenNoMetaBlocksAreUsed(t *testing.T) {
	t.Parallel()

	tdp := dataRetrieverMock.NewPoolsHolderMock()
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(
		tdp,
		genesisBlocks,
		&testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{Nonce: 0}
			},
		},
	)

	err := sp.CheckMetaHeadersValidityAndFinality()
	assert.Nil(t, err)
}

// ------- CommitBlock

func TestShardProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	rootHash := []byte("root hash to be tested")
	accounts := &stateMock.AccountsStub{
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
	body := &block.Body{}
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			if reflect.DeepEqual(obj, hdr) {
				return nil, errMarshalizer
			}

			return []byte("obj"), nil
		},
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = accounts
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CommitBlock(hdr, body)
	assert.Equal(t, errMarshalizer, err)
}

func TestShardProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	errPersister := errors.New("failure")
	putCalledNr := uint32(0)
	rootHash := []byte("root hash to be tested")
	marshalizer := &mock.MarshalizerMock{}
	accounts := &stateMock.AccountsStub{
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
	body := &block.Body{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	hdrUnit := &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return marshalizer.Marshal(&block.Header{})
		},
		PutCalled: func(key, data []byte) error {
			atomic.AddUint32(&putCalledNr, 1)
			wg.Done()
			return errPersister
		},
		HasCalled: func(key []byte) error {
			return nil
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.BlockHeaderUnit, hdrUnit)

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.Storage = store
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = accounts
	arguments.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return nil
		},
	}

	blockTrackerMock := mock.NewBlockTrackerMock(mock.NewOneShardCoordinatorMock(), createGenesisBlocks(mock.NewOneShardCoordinatorMock()))
	blockTrackerMock.GetCrossNotarizedHeaderCalled = func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
		return &block.MetaBlock{}, []byte("hash"), nil
	}
	arguments.BlockTracker = blockTrackerMock

	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CommitBlock(hdr, body)
	wg.Wait()
	assert.True(t, atomic.LoadUint32(&putCalledNr) > 0)
	assert.Nil(t, err)
}

func TestShardProcessor_CommitBlockStorageFailsForBodyShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	putCalledNr := uint32(0)
	errPersister := errors.New("failure")
	rootHash := []byte("root hash to be tested")
	accounts := &stateMock.AccountsStub{
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
	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &mb)

	wg := sync.WaitGroup{}
	wg.Add(1)
	miniBlockUnit := &storageStubs.StorerStub{
		PutCalled: func(key, data []byte) error {
			atomic.AddUint32(&putCalledNr, 1)
			wg.Done()
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {},
	})
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.Storage = store
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = accounts
	arguments.ForkDetector = &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return nil
		},
	}
	blockTrackerMock := mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), createGenesisBlocks(bootstrapComponents.ShardCoordinator()))
	blockTrackerMock.GetCrossNotarizedHeaderCalled = func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
		return &block.MetaBlock{}, []byte("hash"), nil
	}
	arguments.BlockTracker = blockTrackerMock

	sp, err := blproc.NewShardProcessor(arguments)
	assert.Nil(t, err)

	err = sp.CommitBlock(hdr, body)
	wg.Wait()
	assert.Nil(t, err)
	assert.True(t, atomic.LoadUint32(&putCalledNr) > 0)
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
		Nonce:           1,
		Round:           1,
		PubKeysBitmap:   rootHash,
		PrevHash:        hdrHash,
		Signature:       rootHash,
		RootHash:        rootHash,
		PrevRandSeed:    randSeed,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{txHash},
	}
	body := &block.Body{MiniBlocks: []*block.MiniBlock{&mb}}

	mbHdr := block.MiniBlockHeader{
		TxCount: uint32(len(mb.TxHashes)),
		Hash:    hdrHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)
	hdr.MiniBlockHeaders = mbHdrs

	accounts := &stateMock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	forkDetectorAddCalled := false
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
			if header == hdr {
				forkDetectorAddCalled = true
				return nil
			}

			return errors.New("should have not got here")
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return nil
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	blkc := createTestBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return prevHdr
	}
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = hasher
	dataComponents.DataPool = tdp
	dataComponents.Storage = store
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = accounts
	arguments.ForkDetector = fd
	blockTrackerMock := mock.NewBlockTrackerMock(mock.NewOneShardCoordinatorMock(), createGenesisBlocks(mock.NewOneShardCoordinatorMock()))
	blockTrackerMock.GetCrossNotarizedHeaderCalled = func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
		return &block.MetaBlock{}, []byte("hash"), nil
	}
	arguments.BlockTracker = blockTrackerMock

	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.ProcessBlock(hdr, body, haveTime)
	assert.Nil(t, err)
	err = sp.CommitBlock(hdr, body)
	assert.Nil(t, err)
	assert.True(t, forkDetectorAddCalled)
	assert.Equal(t, hdrHash, blkc.GetCurrentBlockHeaderHash())
	// this should sleep as there is an async call to display current hdr and block in CommitBlock
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
		Nonce:           1,
		Round:           1,
		PubKeysBitmap:   rootHash,
		PrevHash:        hdrHash,
		Signature:       rootHash,
		RootHash:        rootHash,
		PrevRandSeed:    randSeed,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{txHash},
	}
	body := &block.Body{MiniBlocks: []*block.MiniBlock{&mb}}

	mbHdr := block.MiniBlockHeader{
		TxCount: uint32(len(mb.TxHashes)),
		Hash:    hdrHash,
	}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)
	hdr.MiniBlockHeaders = mbHdrs

	accounts := &stateMock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
			return nil
		},
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return nil
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	var txsPool *indexer.Pool
	saveBlockCalledMutex := sync.Mutex{}

	blkc := createTestBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return prevHdr
	}
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	dataComponents.BlockChain = blkc

	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	statusComponents.Outport = &testscommon.OutportStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) {
			saveBlockCalledMutex.Lock()
			txsPool = args.TransactionsPool
			saveBlockCalledMutex.Unlock()
		},
		HasDriversCalled: func() bool {
			return true
		},
	}

	arguments.AccountsDB[state.UserAccountsState] = accounts
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
	blockTrackerMock := mock.NewBlockTrackerMock(mock.NewOneShardCoordinatorMock(), createGenesisBlocks(mock.NewOneShardCoordinatorMock()))
	blockTrackerMock.GetCrossNotarizedHeaderCalled = func(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
		return &block.MetaBlock{}, []byte("hash"), nil
	}
	arguments.BlockTracker = blockTrackerMock

	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.ProcessBlock(hdr, body, haveTime)
	assert.Nil(t, err)
	err = sp.CommitBlock(hdr, body)
	assert.Nil(t, err)

	// Wait for the index block go routine to start
	time.Sleep(time.Second * 2)

	assert.Equal(t, 2, len(txsPool.Txs))
	assert.Equal(t, 2, len(txsPool.Scrs))
}

func TestShardProcessor_CreateTxBlockBodyWithDirtyAccStateShouldReturnEmptyBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RevertToSnapshotCalled: revToSnapshot,
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	bl, err := sp.CreateBlockBody(&block.Header{PrevRandSeed: []byte("randSeed")}, func() bool { return true })
	assert.Nil(t, err)
	assert.Equal(t, &block.Body{}, bl)
}

func TestShardProcessor_CreateTxBlockBodyWithNoTimeShouldReturnEmptyBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 0 }
	rootHashfunc := func() ([]byte, error) {
		return []byte("roothash"), nil
	}
	revToSnapshot := func(snapshot int) error { return nil }
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled:       journalLen,
		RootHashCalled:         rootHashfunc,
		RevertToSnapshotCalled: revToSnapshot,
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	haveTimeTrue := func() bool {
		return false
	}
	bl, err := sp.CreateBlockBody(&block.Header{PrevRandSeed: []byte("randSeed")}, haveTimeTrue)
	assert.Nil(t, err)
	assert.Equal(t, &block.Body{}, bl)
}

func TestShardProcessor_CreateTxBlockBodyOK(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 0 }
	rootHashfunc := func() ([]byte, error) {
		return []byte("roothash"), nil
	}
	haveTimeTrue := func() bool {
		return true
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		JournalLenCalled: journalLen,
		RootHashCalled:   rootHashfunc,
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	blk, err := sp.CreateBlockBody(&block.Header{PrevRandSeed: []byte("randSeed")}, haveTimeTrue)
	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

// ------- ComputeNewNoncePrevHash

func TestNode_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.Storage = initStore()
	dataComponents.DataPool = tdp
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

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

func createTestHdrTxBlockBody() (*block.Header, *block.Body) {
	hasher := hashingMocks.HasherMock{}
	hdr := &block.Header{
		Nonce:         1,
		ShardID:       2,
		Epoch:         3,
		Round:         4,
		TimeStamp:     uint64(11223344),
		PrevHash:      hasher.Compute("prev hash"),
		PubKeysBitmap: []byte{255, 0, 128},
		Signature:     hasher.Compute("signature"),
		RootHash:      hasher.Compute("root hash"),
	}
	txBlock := &block.Body{
		MiniBlocks: []*block.MiniBlock{
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
		},
	}
	return hdr, txBlock
}

// ------- ComputeNewNoncePrevHash

func TestShardProcessor_DisplayLogInfo(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	hasher := hashingMocks.HasherMock{}
	hdr, txBlock := createTestHdrTxBlockBody()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	statusHandler := &statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)
	assert.NotNil(t, sp)
	hdr.PrevHash = hasher.Compute("prev hash")
	sp.DisplayLogInfo(hdr, txBlock, []byte("tx_hash1"), shardCoordinator.NumberOfShards(), shardCoordinator.SelfId(), tdp, statusHandler, &mock.BlockTrackerMock{})
}

func TestBlockProcessor_ApplyBodyToHeaderNilBodyError(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	bp, _ := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{}
	_, err := bp.ApplyBodyToHeader(hdr, nil)
	assert.Equal(t, process.ErrNilBlockBody, err)
}

func TestBlockProcessor_ApplyBodyToHeaderShouldNotReturnNil(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	bp, _ := blproc.NewShardProcessor(arguments)
	hdr := &block.Header{}
	_, err := bp.ApplyBodyToHeader(hdr, &block.Body{})
	assert.Nil(t, err)
	assert.NotNil(t, hdr)
}

func TestShardProcessor_ApplyBodyToHeaderShouldErrWhenMarshalizerErrors(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.IntMarsh = &mock.MarshalizerMock{Fail: true}
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{{
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
		},
	}
	hdr := &block.Header{}
	_, err := bp.ApplyBodyToHeader(hdr, body)
	assert.NotNil(t, err)
}

func TestShardProcessor_ApplyBodyToHeaderReturnsOK(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
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
		},
	}
	hdr := &block.Header{}
	_, err := bp.ApplyBodyToHeader(hdr, body)
	assert.Nil(t, err)
	assert.Equal(t, len(body.MiniBlocks), len(hdr.MiniBlockHeaders))
}

func TestShardProcessor_CommitBlockShouldRevertCurrentBlockWhenErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalEntries := 3
	revToSnapshot := func(snapshot int) error {
		journalEntries = 0
		return nil
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		RevertToSnapshotCalled: revToSnapshot,
	}
	bp, _ := blproc.NewShardProcessor(arguments)
	err := bp.CommitBlock(nil, nil)
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
	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &mb0)
	body.MiniBlocks = append(body.MiniBlocks, &mb1)
	body.MiniBlocks = append(body.MiniBlocks, &mb0)
	body.MiniBlocks = append(body.MiniBlocks, &mb1)
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		initStore(),
		marshalizer,
		&hashingMocks.HasherMock{},
		tdp,
		createMockPubkeyConverter(),
		initAccountsMock(),
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&mock.FeeHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)
	container, _ := factory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(initAccountsMock(), tdp, container)
	tc, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)
	msh, mstx, err := sp.MarshalizedDataToBroadcast(&block.Header{}, body)
	assert.Nil(t, err)
	assert.NotNil(t, msh)
	assert.NotNil(t, mstx)
	_, found := msh[0]
	assert.False(t, found)

	bh := &block.Body{}
	err = marshalizer.Unmarshal(bh, msh[1])
	assert.Nil(t, err)
	assert.Equal(t, len(bh.MiniBlocks), 2)
	assert.Equal(t, &mb1, bh.MiniBlocks[0])
	assert.Equal(t, &mb1, bh.MiniBlocks[1])
}

func TestShardProcessor_MarshalizedDataWrongType(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

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

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

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
	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &mb0)
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			wasCalled = true
			return nil, process.ErrMarshalWithoutSuccess
		},
	}

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		initStore(),
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		tdp,
		createMockPubkeyConverter(),
		initAccountsMock(),
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&mock.FeeHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)
	container, _ := factory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(initAccountsMock(), tdp, container)
	tc, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = tc

	sp, _ := blproc.NewShardProcessor(arguments)

	msh, mstx, err := sp.MarshalizedDataToBroadcast(&block.Header{}, body)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, len(msh))
	assert.Equal(t, 0, len(mstx))
}

// ------- receivedMetaBlock

func TestShardProcessor_ReceivedMetaBlockShouldRequestMissingMiniBlocks(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()

	// we will have a metablock that will return 3 miniblock hashes
	// 1 miniblock hash will be in cache
	// 2 will be requested on network

	miniBlockHash1 := []byte("miniblock hash 1 found in cache")
	miniBlockHash2 := []byte("miniblock hash 2")
	miniBlockHash3 := []byte("miniblock hash 3")

	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: miniBlockHash1, SenderShardID: 1, ReceiverShardID: 0},
					{Hash: miniBlockHash2, SenderShardID: 1, ReceiverShardID: 0},
					{Hash: miniBlockHash3, SenderShardID: 1, ReceiverShardID: 0},
				}},
		}}

	// put this metaBlock inside datapool
	metaBlockHash := []byte("metablock hash")
	datapool.Headers().AddHeader(metaBlockHash, metaBlock)
	// put the existing miniblock inside datapool
	datapool.MiniBlocks().Put(miniBlockHash1, &block.MiniBlock{}, 0)

	miniBlockHash1Requested := int32(0)
	miniBlockHash2Requested := int32(0)
	miniBlockHash3Requested := int32(0)

	requestHandler := &testscommon.RequestHandlerStub{
		RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
			if bytes.Equal(miniBlockHash1, miniblockHash) {
				atomic.AddInt32(&miniBlockHash1Requested, 1)
			}
			if bytes.Equal(miniBlockHash2, miniblockHash) {
				atomic.AddInt32(&miniBlockHash2Requested, 1)
			}
			if bytes.Equal(miniBlockHash3, miniblockHash) {
				atomic.AddInt32(&miniBlockHash3Requested, 1)
			}
		},
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(initAccountsMock(), datapool, &mock.PreProcessorContainerMock{})
	argsTransactionCoordinator.RequestHandler = requestHandler
	tc, _ := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.RequestHandler = requestHandler
	arguments.TxCoordinator = tc

	bp, _ := blproc.NewShardProcessor(arguments)
	bp.ReceivedMetaBlock(metaBlock, metaBlockHash)

	// we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(common.ExtraDelayForRequestBlockInfo + time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&miniBlockHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
}

// --------- receivedMetaBlockNoMissingMiniBlocks
func TestShardProcessor_ReceivedMetaBlockNoMissingMiniBlocksShouldPass(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()

	// we will have a metablock that will return 3 miniblock hashes
	// 1 miniblock hash will be in cache
	// 2 will be requested on network

	miniBlockHash1 := []byte("miniblock hash 1 found in cache")

	metaBlock := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash:            miniBlockHash1,
						SenderShardID:   1,
						ReceiverShardID: 0,
					},
				},
			},
		}}

	// put this metaBlock inside datapool
	metaBlockHash := []byte("metablock hash")
	datapool.Headers().AddHeader(metaBlockHash, metaBlock)
	// put the existing miniblock inside datapool
	datapool.MiniBlocks().Put(miniBlockHash1, &block.MiniBlock{}, 0)

	noOfMissingMiniBlocks := int32(0)

	requestHandler := &testscommon.RequestHandlerStub{
		RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
			atomic.AddInt32(&noOfMissingMiniBlocks, 1)
		},
	}

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(initAccountsMock(), datapool, &mock.PreProcessorContainerMock{})
	argsTransactionCoordinator.RequestHandler = requestHandler
	tc, _ := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.RequestHandler = requestHandler
	arguments.TxCoordinator = tc

	sp, _ := blproc.NewShardProcessor(arguments)
	sp.ReceivedMetaBlock(metaBlock, metaBlockHash)

	// we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(common.ExtraDelayForRequestBlockInfo + time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&noOfMissingMiniBlocks))
}

// --------- createAndProcessCrossMiniBlocksDstMe
func TestShardProcessor_CreateAndProcessCrossMiniBlocksDstMe(t *testing.T) {
	t.Parallel()

	tdp := dataRetrieverMock.NewPoolsHolderMock()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, 0, process.ShardCacherIdentifier(1, 0))

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	meta := &block.MetaBlock{
		Nonce:        1,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        1,
		PrevRandSeed: []byte("roothash"),
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	tdp.Headers().AddHeader(metaHash, meta)

	haveTimeTrue := func() bool {
		return true
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	miniBlockSlice, usedMetaHdrsHashes, noOfTxs, err := sp.CreateAndProcessMiniBlocksDstMe(haveTimeTrue)
	assert.Equal(t, err == nil, true)
	assert.Equal(t, len(miniBlockSlice) == 0, true)
	assert.Equal(t, usedMetaHdrsHashes, uint32(0))
	assert.Equal(t, noOfTxs, uint32(0))
}

func TestShardProcessor_CreateAndProcessCrossMiniBlocksDstMeProcessPartOfMiniBlocksInMetaBlock(t *testing.T) {
	t.Parallel()

	haveTimeTrue := func() bool {
		return true
	}
	tdp := dataRetrieverMock.NewPoolsHolderMock()
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

	// put 2 metablocks in pool
	meta := &block.MetaBlock{
		Nonce:        1,
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks),
		Round:        1,
		PrevRandSeed: []byte("roothash"),
	}

	mb1Hash := []byte("meta block 1")
	tdp.Headers().AddHeader(mb1Hash, meta)

	meta = &block.MetaBlock{
		Nonce:     2,
		ShardInfo: createShardData(hasher, marshalizer, miniBlocks),
		Round:     2,
	}

	mb2Hash := []byte("meta block 2")
	tdp.Headers().AddHeader(mb2Hash, meta)

	meta = &block.MetaBlock{
		Nonce:        3,
		ShardInfo:    make([]block.ShardData, 0),
		Round:        3,
		PrevRandSeed: []byte("roothash"),
	}

	mb3Hash := []byte("meta block 3")
	tdp.Headers().AddHeader(mb3Hash, meta)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	miniBlocksReturned, usedMetaHdrsHashes, nrTxAdded, err := sp.CreateAndProcessMiniBlocksDstMe(haveTimeTrue)

	assert.Equal(t, 0, len(miniBlocksReturned))
	assert.Equal(t, uint32(0), usedMetaHdrsHashes)
	assert.Equal(t, uint32(0), nrTxAdded)
	assert.Nil(t, err)
}

// ------- createMiniBlocks

func TestShardProcessor_CreateMiniBlocksShouldWorkWithIntraShardTxs(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()

	// we will have a 3 txs in pool

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(0)
	receiverShardId := uint32(0)

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	// put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	datapool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, 0, cacheId)
	datapool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, 0, cacheId)
	datapool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, 0, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	txProcessorMock := &testscommon.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			// execution, in this context, means moving the tx nonce to itx corresponding execution result variable
			if bytes.Equal(transaction.Data, txHash1) {
				tx1ExecutionResult = transaction.Nonce
			}
			if bytes.Equal(transaction.Data, txHash2) {
				tx2ExecutionResult = transaction.Nonce
			}
			if bytes.Equal(transaction.Data, txHash3) {
				tx3ExecutionResult = transaction.Nonce
			}

			return 0, nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	accntAdapter := &stateMock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			assert.Fail(t, "revert should have not been called")
			return nil
		},
		JournalLenCalled: func() int {
			return 0
		},
	}

	totalGasProvided := uint64(0)
	factory, _ := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		initStore(),
		marshalizer,
		hasher,
		datapool,
		createMockPubkeyConverter(),
		accntAdapter,
		&testscommon.RequestHandlerStub{},
		txProcessorMock,
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
			MaxGasLimitPerBlockCalled: func() uint64 {
				return MaxGasLimitPerBlock
			},
		},
		&testscommon.GasHandlerStub{
			SetGasProvidedCalled: func(gasProvided uint64, hash []byte) {
				totalGasProvided += gasProvided
			},
			TotalGasProvidedCalled: func() uint64 {
				return totalGasProvided
			},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, nil
			},
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
			TotalGasRefundedCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)
	container, _ := factory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(accntAdapter, datapool, container)
	tc, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	require.Nil(t, err)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = accntAdapter
	arguments.TxCoordinator = tc
	bp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	blockBody, err := bp.CreateMiniBlocks(func() bool { return true })

	assert.Nil(t, err)
	// testing execution
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
	// one miniblock output
	assert.Equal(t, 1, len(blockBody.MiniBlocks))
	// miniblock should have 3 txs
	assert.Equal(t, 3, len(blockBody.MiniBlocks[0].TxHashes))
	// testing all 3 hashes are present in block body
	assert.True(t, isInTxHashes(txHash1, blockBody.MiniBlocks[0].TxHashes))
	assert.True(t, isInTxHashes(txHash2, blockBody.MiniBlocks[0].TxHashes))
	assert.True(t, isInTxHashes(txHash3, blockBody.MiniBlocks[0].TxHashes))
}

func TestShardProcessor_GetProcessedMetaBlockFromPoolShouldWork(t *testing.T) {
	t.Parallel()

	// we have 3 metablocks in pool each containing 2 miniblocks.
	// blockbody will have 2 + 1 miniblocks from 2 out of the 3 metablocks
	// The test should remove only one metablock

	destShardId := uint32(2)

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()

	miniblockHashes := make([][]byte, 6)

	destShards := []uint32{1, 3, 4}
	for i := 0; i < 6; i++ {
		_, hash := createDummyMiniBlock(fmt.Sprintf("tx hash %d", i), marshalizer, hasher, destShardId, destShards[i/2])
		miniblockHashes[i] = hash
	}

	// put 3 metablocks in pool
	metaBlockHash1 := []byte("meta block 1")
	metaBlock1 := createDummyMetaBlock(destShardId, destShards[0], miniblockHashes[0], miniblockHashes[1])
	datapool.Headers().AddHeader(metaBlockHash1, metaBlock1)

	metaBlockHash2 := []byte("meta block 2")
	metaBlock2 := createDummyMetaBlock(destShardId, destShards[1], miniblockHashes[2], miniblockHashes[3])
	datapool.Headers().AddHeader(metaBlockHash2, metaBlock2)

	metaBlockHash3 := []byte("meta block 3")
	metaBlock3 := createDummyMetaBlock(destShardId, destShards[2], miniblockHashes[4], miniblockHashes[5])
	datapool.Headers().AddHeader(metaBlockHash3, metaBlock3)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = destShardId
	shardCoordinator.SetNoShards(destShardId + 1)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.ForkDetector = &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.SetHdrForCurrentBlock(metaBlockHash1, metaBlock1, true)
	bp.SetHdrForCurrentBlock(metaBlockHash2, metaBlock2, true)
	bp.SetHdrForCurrentBlock(metaBlockHash3, metaBlock3, true)

	// create mini block headers with first 3 miniblocks from miniblocks var
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
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilBlockHeader(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	be, _ := blproc.NewShardProcessor(arguments)
	err := be.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldWorkNilTxBlockBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.RestoreBlockIntoPools(&block.Header{}, nil)
	assert.Nil(t, err)
}

func TestShardProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx hash 1")

	datapool := dataRetrieverMock.NewPoolsHolderMock()
	marshalizerMock := &mock.MarshalizerMock{}
	hasherMock := &mock.HasherStub{}

	body := &block.Body{}
	tx := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}
	buffTx, _ := marshalizerMock.Marshal(tx)

	store := &mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			m := make(map[string][]byte)
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
		createMockPubkeyConverter(),
		initAccountsMock(),
		&testscommon.RequestHandlerStub{},
		&testscommon.TxProcessorMock{},
		&testscommon.SCProcessorMock{},
		&testscommon.SmartContractResultsProcessorMock{},
		&testscommon.RewardTxProcessorMock{},
		&mock.FeeHandlerStub{},
		&testscommon.GasHandlerStub{},
		&mock.BlockTrackerMock{},
		&testscommon.BlockSizeComputationStub{},
		&testscommon.BalanceComputationStub{},
		&epochNotifier.EpochNotifierStub{},
		2,
		2,
		2,
		&testscommon.TxTypeHandlerMock{},
		&testscommon.ScheduledTxsExecutionStub{},
	)
	container, _ := factory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(initAccountsMock(), datapool, container)
	tc, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	assert.Nil(t, err)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasherMock
	coreComponents.IntMarsh = marshalizerMock
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = tc
	sp, _ := blproc.NewShardProcessor(arguments)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	miniblockHash := []byte("mini block hash 1")
	hasherMock.ComputeCalled = func(s string) []byte {
		return miniblockHash
	}

	metablockHash := []byte("meta block hash 1")
	metablockHeader := createDummyMetaBlock(0, 1, miniblockHash)
	datapool.Headers().AddHeader(metablockHash, metablockHeader)

	store.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		return &storageStubs.StorerStub{
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

	err = sp.RestoreBlockIntoPools(&block.Header{MetaBlockHashes: [][]byte{metablockHash}, MiniBlockHeaders: []block.MiniBlockHeader{miniBlockHeader}}, body)
	assert.Nil(t, err)

	miniblockFromPool, _ := datapool.MiniBlocks().Get(miniblockHash)
	txFromPool, _ := datapool.Transactions().SearchFirstData(txHash)
	assert.Nil(t, err)
	assert.Equal(t, &miniblock, miniblockFromPool)
	assert.Equal(t, tx, txFromPool)
}

func TestShardProcessor_DecodeBlockBody(t *testing.T) {
	t.Parallel()

	tdp := initDataPool([]byte("tx_hash1"))
	marshalizerMock := &mock.MarshalizerMock{}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizerMock
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &block.MiniBlock{ReceiverShardID: 69})
	message, err := marshalizerMock.Marshal(body)
	assert.Nil(t, err)

	bodyNil := &block.Body{}
	dcdBlk := sp.DecodeBlockBody(nil)
	assert.Equal(t, bodyNil, dcdBlk)

	dcdBlk = sp.DecodeBlockBody(message)
	assert.Equal(t, body, dcdBlk)
	assert.Equal(t, uint32(69), body.MiniBlocks[0].ReceiverShardID)
}

func TestShardProcessor_DecodeBlockHeader(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizerMock := &mock.MarshalizerMock{}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = tdp
	coreComponents.IntMarsh = marshalizerMock
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	assert.Nil(t, err)
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(0)
	hdr.Signature = []byte("A")
	hdr.AccumulatedFees = big.NewInt(0)
	hdr.DeveloperFees = big.NewInt(0)
	_, err = marshalizerMock.Marshal(hdr)
	assert.Nil(t, err)

	message, err := marshalizerMock.Marshal(hdr)
	assert.Nil(t, err)

	dcdHdr := sp.DecodeBlockHeader(nil)
	assert.Nil(t, dcdHdr)

	dcdHdr = sp.DecodeBlockHeader(message)
	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte("A"), dcdHdr.GetSignature())
}

func TestShardProcessor_IsHdrConstructionValid(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := initDataPool([]byte("tx_hash1"))

	shardNr := uint32(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bootstrapComponents.Coordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    44,
		RandSeed: prevRandSeed}
	notarizedHdrs[core.MetachainShardId] = append(notarizedHdrs[core.MetachainShardId], lastHdr)

	// put the existing headers inside datapool

	// header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(core.MetachainShardId).(*block.MetaBlock))
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

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	forkDetector := &mock.ForkDetectorMock{}
	highNonce := uint64(500)
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return highNonce
	}

	wg := sync.WaitGroup{}
	wg.Add(4)
	putCalledNr := uint32(0)
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			atomic.AddUint32(&putCalledNr, 1)
			wg.Done()
			return nil
		},
	}

	shardNr := uint32(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bootstrapComponents.Coordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.ForkDetector = forkDetector
	startHeaders := createGenesisBlocks(bootstrapComponents.ShardCoordinator())
	arguments.BlockTracker = mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}

	arguments.BlockTracker.AddCrossNotarizedHeader(core.MetachainShardId, lastHdr, nil)

	// header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(core.MetachainShardId).(*block.MetaBlock))
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

	err = sp.SaveLastNotarizedHeader(core.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.UpdateCrossShardInfo(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), atomic.LoadUint32(&putCalledNr))

	assert.Equal(t, firstNonce, sp.LastNotarizedHdrForShard(core.MetachainShardId).GetNonce())
	assert.Equal(t, 0, len(processedMetaHdrs))

	// wrong header type in pool and defer called
	datapool.Headers().AddHeader(currHash, shardHdr)
	sp.SetHdrForCurrentBlock(currHash, shardHdr, true)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	blockHeader = &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err = sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)

	err = sp.SaveLastNotarizedHeader(core.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.UpdateCrossShardInfo(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), atomic.LoadUint32(&putCalledNr))

	assert.Equal(t, firstNonce, sp.LastNotarizedHdrForShard(core.MetachainShardId).GetNonce())

	// put headers in pool
	datapool.Headers().AddHeader(currHash, currHdr)
	datapool.Headers().AddHeader(prevHash, prevHdr)

	_ = sp.CreateBlockStarted()
	sp.SetHdrForCurrentBlock(currHash, currHdr, true)
	sp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	hashes = make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader = &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err = sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(core.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.UpdateCrossShardInfo(processedMetaHdrs)
	wg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, uint32(4), atomic.LoadUint32(&putCalledNr))

	assert.Equal(t, currHdr, sp.LastNotarizedHdrForShard(core.MetachainShardId))
}

func createShardData(hasher hashing.Hasher, marshalizer marshal.Marshalizer, miniBlocks []block.MiniBlock) []block.ShardData {
	shardData := make([]block.ShardData, len(miniBlocks))
	for i := 0; i < len(miniBlocks); i++ {
		hashed, _ := core.CalculateHash(marshalizer, hasher, &miniBlocks[i])

		shardMBHeader := block.MiniBlockHeader{
			ReceiverShardID: miniBlocks[i].ReceiverShardID,
			SenderShardID:   miniBlocks[i].SenderShardID,
			TxCount:         uint32(len(miniBlocks[i].TxHashes)),
			Hash:            hashed,
		}
		shardMBHeaders := make([]block.MiniBlockHeader, 0)
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

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	forkDetector := &mock.ForkDetectorMock{}
	highNonce := uint64(500)
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return highNonce
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	putCalledNr := uint32(0)
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			atomic.AddUint32(&putCalledNr, 1)
			wg.Done()
			return nil
		},
	}

	shardNr := uint32(5)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bootstrapComponents.Coordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.ForkDetector = forkDetector
	startHeaders := createGenesisBlocks(bootstrapComponents.ShardCoordinator())
	arguments.BlockTracker = mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	notarizedHdrs[core.MetachainShardId] = append(notarizedHdrs[core.MetachainShardId], lastHdr)

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

	hashed, err := core.CalculateHash(marshalizer, hasher, &miniblock1)
	require.Nil(t, err)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	hashed, err = core.CalculateHash(marshalizer, hasher, &miniblock2)
	require.Nil(t, err)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	hashed, err = core.CalculateHash(marshalizer, hasher, &miniblock3)
	require.Nil(t, err)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)
	// header shard 0
	prevHash, err := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(core.MetachainShardId).(*block.MetaBlock))
	require.Nil(t, err)
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
	prevHash, err = sp.ComputeHeaderHash(prevHdr)
	require.Nil(t, err)
	currHdr := &block.MetaBlock{
		Round:        11,
		Nonce:        46,
		PrevRandSeed: currRandSeed,
		RandSeed:     []byte("nextrand"),
		PrevHash:     prevHash,
		RootHash:     []byte("currRootHash"),
		ShardInfo:    createShardData(hasher, marshalizer, miniBlocks)}
	currHash, err := sp.ComputeHeaderHash(currHdr)
	require.Nil(t, err)
	prevHash, err = sp.ComputeHeaderHash(prevHdr)
	require.Nil(t, err)

	// put headers in pool
	datapool.Headers().AddHeader(currHash, currHdr)
	datapool.Headers().AddHeader(prevHash, prevHdr)

	sp.SetHdrForCurrentBlock(currHash, currHdr, true)
	sp.SetHdrForCurrentBlock(prevHash, prevHdr, true)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader := &block.Header{MetaBlockHashes: hashes, MiniBlockHeaders: mbHeaders}

	processedMetaHdrs, err := sp.GetOrderedProcessedMetaBlocksFromHeader(blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(core.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.UpdateCrossShardInfo(processedMetaHdrs)
	wg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), atomic.LoadUint32(&putCalledNr))

	assert.Equal(t, prevHdr, sp.LastNotarizedHdrForShard(core.MetachainShardId))
}

func TestShardProcessor_RemoveAndSaveLastNotarizedMetaHdrAllMBFinished(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	forkDetector := &mock.ForkDetectorMock{}
	highNonce := uint64(500)
	forkDetector.GetHighestFinalBlockNonceCalled = func() uint64 {
		return highNonce
	}

	wg := sync.WaitGroup{}
	wg.Add(4)
	putCalledNr := uint32(0)
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			atomic.AddUint32(&putCalledNr, 1)
			wg.Done()
			return nil
		},
	}

	shardNr := uint32(5)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bootstrapComponents.Coordinator = mock.NewMultiShardsCoordinatorMock(shardNr)
	arguments.ForkDetector = forkDetector
	startHeaders := createGenesisBlocks(bootstrapComponents.ShardCoordinator())
	arguments.BlockTracker = mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders)
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	notarizedHdrs := sp.NotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	notarizedHdrs[core.MetachainShardId] = append(notarizedHdrs[core.MetachainShardId], lastHdr)

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

	mbHeaders := make([]block.MiniBlockHeader, 0, 4)

	hashed, _ := core.CalculateHash(marshalizer, hasher, &miniblock1)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	hashed, _ = core.CalculateHash(marshalizer, hasher, &miniblock2)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	hashed, _ = core.CalculateHash(marshalizer, hasher, &miniblock3)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	hashed, _ = core.CalculateHash(marshalizer, hasher, &miniblock4)
	mbHeaders = append(mbHeaders, block.MiniBlockHeader{Hash: hashed})

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)
	// header shard 0
	prevHash, _ := sp.ComputeHeaderHash(sp.LastNotarizedHdrForShard(core.MetachainShardId).(*block.MetaBlock))
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
	datapool.Headers().AddHeader(currHash, currHdr)
	datapool.Headers().AddHeader(prevHash, prevHdr)
	datapool.Headers().AddHeader([]byte("shouldNotRemove"), &block.MetaBlock{
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

	err = sp.SaveLastNotarizedHeader(core.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.UpdateCrossShardInfo(processedMetaHdrs)
	wg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, uint32(4), atomic.LoadUint32(&putCalledNr))

	assert.Equal(t, currHdr, sp.LastNotarizedHdrForShard(core.MetachainShardId))
}

func createOneHeaderOneBody() (*block.Header, *block.Body) {
	txHash := []byte("tx_hash1")
	rootHash := []byte("rootHash")
	body := &block.Body{}
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	hasher := &mock.HasherStub{}
	marshalizer := &mock.MarshalizerMock{}

	mbbytes, _ := marshalizer.Marshal(&miniblock)
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
		ShardID:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}

	return hdr, body
}

func TestShardProcessor_CheckHeaderBodyCorrelationReceiverMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].ReceiverShardID = body.MiniBlocks[0].ReceiverShardID + 1
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationSenderMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].SenderShardID = body.MiniBlocks[0].SenderShardID + 1
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationTxCountMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].TxCount = uint32(len(body.MiniBlocks[0].TxHashes) + 1)
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationHashMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	hdr.MiniBlockHeaders[0].Hash = []byte("wrongHash")
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationShouldPass(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Nil(t, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationNilMiniBlock(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	body.MiniBlocks[0] = nil

	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilMiniBlock, err)
}

func TestShardProcessor_RestoreMetaBlockIntoPoolShouldPass(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	poolFake := dataRetrieverMock.NewPoolsHolderMock()

	metaBlock := block.MetaBlock{
		Nonce:     1,
		ShardInfo: make([]block.ShardData, 0),
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &storageStubs.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
				GetCalled: func(key []byte) ([]byte, error) {
					return marshalizer.Marshal(&metaBlock)
				},
			}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolFake
	dataComponents.Storage = store
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	miniblockHashes := make(map[string]uint32)

	meta := &block.MetaBlock{
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

	metaBlockRestored, err := poolFake.Headers().GetHeaderByHash(metaHash)

	assert.Equal(t, nil, metaBlockRestored)
	assert.Error(t, err)

	err = sp.RestoreMetaBlockIntoPool(miniblockHashes, metablockHashes)

	metaBlockRestored, _ = poolFake.Headers().GetHeaderByHash(metaHash)

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

	mbbytes, _ := marshalizer.Marshal(&miniblock)
	mbHash := hasher.Compute(string(mbbytes))

	shardMiniBlock := block.MiniBlockHeader{
		ReceiverShardID: 0,
		SenderShardID:   2,
		TxCount:         uint32(len(txHashes)),
		Hash:            mbHash,
	}
	shardMiniblockHdrs := make([]block.MiniBlockHeader, 0)
	shardMiniblockHdrs = append(shardMiniblockHdrs, shardMiniBlock)
	shardHeader := block.ShardData{
		ShardID:               1,
		ShardMiniBlockHeaders: shardMiniblockHdrs,
	}
	shardHdrs := make([]block.ShardData, 0)
	shardHdrs = append(shardHdrs, shardHeader)
	metaBlock := &block.MetaBlock{Nonce: 1, Round: 1, ShardInfo: shardHdrs}

	idp := initDataPool([]byte("tx_hash1"))

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = idp
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
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

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(nil)

	assert.NotNil(t, hdrs)
	assert.Equal(t, 0, len(hdrs))
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachaiMetaHdrsWithoutOwnHdr(t *testing.T) {
	t.Parallel()

	processedHdrs := make([]data.HeaderHandler, 0)
	datapool := dataRetrieverMock.CreatePoolsHolder(1, 0)
	store := initStore()
	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.BlockTracker = &mock.BlockTrackerMock{}

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: []byte("hash"), ShardID: 1})
	datapool.Headers().AddHeader([]byte("hash"), &block.Header{ShardID: 0, Nonce: 1})

	prevMetaHdr := genesisBlocks[core.MetachainShardId]
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
	datapool.Headers().AddHeader(currHash, currMetaHdr)
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
	datapool.Headers().AddHeader(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(processedHdrs)

	assert.NotNil(t, hdrs)
	assert.Equal(t, 0, len(hdrs))
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachaiMetaHdrsWithOwnHdrButNotStored(t *testing.T) {
	t.Parallel()

	processedHdrs := make([]data.HeaderHandler, 0)
	datapool := dataRetrieverMock.CreatePoolsHolder(1, 0)
	store := initStore()
	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.BlockTracker = &mock.BlockTrackerMock{}

	sp, _ := blproc.NewShardProcessor(arguments)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: []byte("hash"), ShardID: 0})

	prevMetaHdr := genesisBlocks[core.MetachainShardId]
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
	datapool.Headers().AddHeader(currHash, currMetaHdr)
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
	datapool.Headers().AddHeader(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(processedHdrs)

	assert.Equal(t, 0, len(hdrs))
}

func TestShardProcessor_GetHighestHdrForOwnShardFromMetachaiMetaHdrsWithOwnHdrStored(t *testing.T) {
	t.Parallel()

	processedHdrs := make([]data.HeaderHandler, 0)
	datapool := dataRetrieverMock.CreatePoolsHolder(1, 0)
	store := initStore()
	hasher := &hashingMocks.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasher
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.BlockTracker = &mock.BlockTrackerMock{}

	sp, _ := blproc.NewShardProcessor(arguments)

	ownHdr := &block.Header{
		Nonce: 1,
		Round: 1,
	}
	ownHash, _ := core.CalculateHash(marshalizer, hasher, ownHdr)
	datapool.Headers().AddHeader(ownHash, ownHdr)

	shardInfo := make([]block.ShardData, 0)
	shardInfo = append(shardInfo, block.ShardData{HeaderHash: ownHash, ShardID: 0})

	prevMetaHdr := genesisBlocks[core.MetachainShardId]
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
	datapool.Headers().AddHeader(currHash, currMetaHdr)

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
	datapool.Headers().AddHeader(currHash, currMetaHdr)
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
	datapool.Headers().AddHeader(currHash, currMetaHdr)
	processedHdrs = append(processedHdrs, currMetaHdr)

	hdrs, _, _ := sp.GetHighestHdrForOwnShardFromMetachain(processedHdrs)

	assert.NotNil(t, hdrs)
	assert.Equal(t, ownHdr.GetNonce(), hdrs[0].GetNonce())
}

func TestShardProcessor_RestoreMetaBlockIntoPoolVerifyMiniblocks(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	poolMock := dataRetrieverMock.CreatePoolsHolder(1, 0)

	storer := &mock.ChainStorerMock{}
	shardC := mock.NewMultiShardsCoordinatorMock(3)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolMock
	dataComponents.Storage = storer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bootstrapComponents.Coordinator = shardC
	arguments.BlockTracker = &mock.BlockTrackerMock{}
	sp, _ := blproc.NewShardProcessor(arguments)

	miniblockHashes := make(map[string]uint32)

	testMBHash := []byte("hash")
	shardMBHdr := block.MiniBlockHeader{
		Hash:            testMBHash,
		SenderShardID:   shardC.SelfId() + 1,
		ReceiverShardID: shardC.SelfId(),
	}
	shardMBHeaders := make([]block.MiniBlockHeader, 0)
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
	metablockHashes := make([][]byte, 0)
	metablockHashes = append(metablockHashes, metaHash)

	metaBlockRestored, err := poolMock.Headers().GetHeaderByHash(metaHash)

	assert.Equal(t, nil, metaBlockRestored)
	assert.Error(t, err)

	storer.GetCalled = func(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
		return metaBytes, nil
	}
	storer.GetStorerCalled = func(unitType dataRetriever.UnitType) storage.Storer {
		return &storageStubs.StorerStub{
			RemoveCalled: func(key []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return metaBytes, nil
			},
		}
	}

	err = sp.RestoreMetaBlockIntoPool(miniblockHashes, metablockHashes)

	metaBlockRestored, _ = poolMock.Headers().GetHeaderByHash(metaHash)

	assert.Equal(t, meta, metaBlockRestored)
	assert.Nil(t, err)
}

// ------- updateStateStorage

func TestShardProcessor_updateStateStorage(t *testing.T) {
	t.Parallel()

	pruneTrieWasCalled := false
	cancelPruneWasCalled := false
	rootHash := []byte("root-hash")
	poolMock := dataRetrieverMock.NewPoolsHolderMock()

	hdrStore := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			hdr := block.Header{Nonce: 7, RootHash: rootHash}
			return json.Marshal(hdr)
		},
	}

	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			if unitType == dataRetriever.ScheduledSCRsUnit {
				return nil
			}

			return hdrStore
		},
	}

	shardC := mock.NewMultiShardsCoordinatorMock(3)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolMock
	dataComponents.Storage = storer
	bootstrapComponents.Coordinator = shardC
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.BlockTracker = &mock.BlockTrackerMock{}
	arguments.Config.StateTriesConfig.CheckpointRoundsModulus = 2
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		IsPruningEnabledCalled: func() bool {
			return true
		},
		PruneTrieCalled: func(rootHashParam []byte, identifier state.TriePruningIdentifier) {
			pruneTrieWasCalled = true
			assert.Equal(t, rootHash, rootHashParam)
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			cancelPruneWasCalled = true
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	finalHeaders := make([]data.HeaderHandler, 0)
	hdr1 := &block.Header{Nonce: 0, Round: 0}
	hdr2 := &block.Header{Nonce: 1, Round: 1}
	finalHeaders = append(finalHeaders, hdr1, hdr2)
	sp.UpdateStateStorage(finalHeaders, &block.Header{})

	assert.True(t, pruneTrieWasCalled)
	assert.True(t, cancelPruneWasCalled)
}

func TestShardProcessor_checkEpochCorrectnessCrossChainNilCurrentBlock(t *testing.T) {
	t.Parallel()

	chain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return nil
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.BlockChain = chain
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CheckEpochCorrectnessCrossChain()
	assert.Equal(t, nil, err)
}

func TestShardProcessor_checkEpochCorrectnessCrossChainCorrectEpoch(t *testing.T) {
	t.Parallel()

	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 10
		},
		EpochCalled: func() uint32 {
			return 1
		},
	}
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Epoch: 1}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = epochStartTrigger
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CheckEpochCorrectnessCrossChain()
	assert.Equal(t, nil, err)

	sp, _ = blproc.NewShardProcessor(arguments)

	err = sp.CheckEpochCorrectnessCrossChain()
	assert.Equal(t, nil, err)
}

func TestShardProcessor_checkEpochCorrectnessCrossChainInCorrectEpochStorageError(t *testing.T) {
	t.Parallel()

	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 10
		},
		EpochCalled: func() uint32 {
			return 1
		},
		MetaEpochCalled: func() uint32 {
			return 1
		},
	}

	header := &block.Header{Epoch: epochStartTrigger.Epoch() - 1, Round: epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod + 1}
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return header
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = epochStartTrigger

	sp, _ := blproc.NewShardProcessor(arguments)

	store := dataComponents.StorageService()
	val, _ := coreComponents.InternalMarshalizer().Marshal(header)
	key := coreComponents.Hasher().Compute(string(val))
	_ = store.Put(
		dataRetriever.ShardHdrNonceHashDataUnit,
		coreComponents.Uint64ByteSliceConverter().ToByteSlice(header.Nonce),
		key,
	)

	err := sp.CheckEpochCorrectnessCrossChain()
	assert.True(t, errors.Is(err, process.ErrMissingHeader))
}

func TestShardProcessor_checkEpochCorrectnessCrossChainInCorrectEpochRollback1Block(t *testing.T) {
	t.Parallel()

	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 10
		},
		EpochCalled: func() uint32 {
			return 1
		},
		MetaEpochCalled: func() uint32 {
			return 1
		},
	}
	store := initStore()
	nonceCalled := uint64(444444)
	forkDetector := &mock.ForkDetectorMock{SetRollBackNonceCalled: func(nonce uint64) {
		nonceCalled = nonce
	}}
	prevHash := []byte("prevHash")
	currHeader := &block.Header{
		Nonce:    10,
		Epoch:    epochStartTrigger.Epoch() - 1,
		Round:    epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod + 1,
		PrevHash: prevHash}

	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return currHeader
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.Storage = store
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = epochStartTrigger
	arguments.ForkDetector = forkDetector

	sp, _ := blproc.NewShardProcessor(arguments)

	prevHeader := &block.Header{
		Nonce: 8,
		Epoch: epochStartTrigger.Epoch() - 1,
		Round: epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod,
	}

	prevHeaderData, _ := coreComponents.InternalMarshalizer().Marshal(prevHeader)
	_ = store.Put(
		dataRetriever.ShardHdrNonceHashDataUnit,
		coreComponents.Uint64ByteSliceConverter().ToByteSlice(prevHeader.Nonce),
		prevHash,
	)
	_ = store.Put(dataRetriever.BlockHeaderUnit, prevHash, prevHeaderData)

	err := sp.CheckEpochCorrectnessCrossChain()
	assert.Equal(t, process.ErrEpochDoesNotMatch, err)
	assert.Equal(t, nonceCalled, currHeader.Nonce)
}

func TestShardProcessor_checkEpochCorrectnessCrossChainInCorrectEpochRollback2Blocks(t *testing.T) {
	t.Parallel()

	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 10
		},
		EpochCalled: func() uint32 {
			return 1
		},
		MetaEpochCalled: func() uint32 {
			return 1
		},
	}
	store := initStore()
	nonceCalled := uint64(444444)
	forkDetector := &mock.ForkDetectorMock{SetRollBackNonceCalled: func(nonce uint64) {
		nonceCalled = nonce
	}}
	prevHash := []byte("prevHash")
	header := &block.Header{
		Nonce:    10,
		Epoch:    epochStartTrigger.Epoch() - 1,
		Round:    epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod + 2,
		PrevHash: prevHash}

	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return header
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.Storage = store
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = epochStartTrigger
	arguments.ForkDetector = forkDetector

	sp, _ := blproc.NewShardProcessor(arguments)

	prevPrevHash := []byte("prevPrevHash")
	prevHeader := &block.Header{
		Nonce:    8,
		Epoch:    epochStartTrigger.Epoch() - 1,
		Round:    epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod + 1,
		PrevHash: prevPrevHash,
	}
	prevHeaderData, _ := coreComponents.InternalMarshalizer().Marshal(prevHeader)
	_ = store.Put(
		dataRetriever.ShardHdrNonceHashDataUnit,
		coreComponents.Uint64ByteSliceConverter().ToByteSlice(prevHeader.Nonce),
		prevHash,
	)
	_ = store.Put(dataRetriever.BlockHeaderUnit, prevHash, prevHeaderData)

	prevPrevHeader := &block.Header{
		Nonce:    7,
		Epoch:    epochStartTrigger.Epoch() - 1,
		Round:    epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod,
		PrevHash: prevPrevHash,
	}
	prevPrevHeaderData, _ := coreComponents.InternalMarshalizer().Marshal(prevPrevHeader)
	_ = store.Put(
		dataRetriever.ShardHdrNonceHashDataUnit,
		coreComponents.Uint64ByteSliceConverter().ToByteSlice(prevPrevHeader.Nonce),
		prevPrevHash,
	)
	_ = store.Put(dataRetriever.BlockHeaderUnit, prevPrevHash, prevPrevHeaderData)

	err := sp.CheckEpochCorrectnessCrossChain()
	assert.Equal(t, process.ErrEpochDoesNotMatch, err)
	assert.Equal(t, nonceCalled, prevHeader.Nonce)
}

func TestShardProcessor_GetBootstrapHeadersInfoShouldReturnNilWhenNoSelfNotarizedHeadersExists(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	bootstrapHeaderInfos := sp.GetBootstrapHeadersInfo(nil, nil)

	assert.Nil(t, bootstrapHeaderInfos)
}

func TestShardProcessor_GetBootstrapHeadersInfoShouldReturnOneItemWhenFinalNonceIsHigherThanGenesis(t *testing.T) {
	t.Parallel()

	finalNonce := uint64(1)
	finalHash := []byte("final hash")

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.ForkDetector = &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return finalNonce
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return finalHash
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	bootstrapHeaderInfos := sp.GetBootstrapHeadersInfo(nil, nil)

	require.Equal(t, 1, len(bootstrapHeaderInfos))
	assert.Equal(t, finalHash, bootstrapHeaderInfos[0].Hash)
}

func TestShardProcessor_GetBootstrapHeadersInfoShouldReturnOneItemWhenFinalNonceIsNotHigherThanSelfNotarizedNonce(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.ForkDetector = &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 0
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	hash := []byte("hash")
	header := &block.Header{}

	selfNotarizedHeaders := make([]data.HeaderHandler, 0)
	selfNotarizedHeadersHashes := make([][]byte, 0)

	selfNotarizedHeaders = append(selfNotarizedHeaders, header)
	selfNotarizedHeadersHashes = append(selfNotarizedHeadersHashes, hash)

	bootstrapHeaderInfos := sp.GetBootstrapHeadersInfo(selfNotarizedHeaders, selfNotarizedHeadersHashes)

	require.Equal(t, 1, len(bootstrapHeaderInfos))
	assert.Equal(t, hash, bootstrapHeaderInfos[0].Hash)
}

func TestShardProcessor_GetBootstrapHeadersInfoShouldReturnTwoItemsWhenFinalNonceIsHigherThanSelfNotarizedNonce(t *testing.T) {
	t.Parallel()

	finalNonce := uint64(2)
	finalHash := []byte("final hash")

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.ForkDetector = &mock.ForkDetectorMock{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return finalNonce
		},
		GetHighestFinalBlockHashCalled: func() []byte {
			return finalHash
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	hash := []byte("hash")
	header := &block.Header{Nonce: 1}

	selfNotarizedHeaders := make([]data.HeaderHandler, 0)
	selfNotarizedHeadersHashes := make([][]byte, 0)

	selfNotarizedHeaders = append(selfNotarizedHeaders, header)
	selfNotarizedHeadersHashes = append(selfNotarizedHeadersHashes, hash)

	bootstrapHeaderInfos := sp.GetBootstrapHeadersInfo(selfNotarizedHeaders, selfNotarizedHeadersHashes)

	require.Equal(t, 2, len(bootstrapHeaderInfos))
	assert.Equal(t, hash, bootstrapHeaderInfos[0].Hash)
	assert.Equal(t, finalHash, bootstrapHeaderInfos[1].Hash)
}

func TestShardProcessor_RequestMetaHeadersIfNeededShouldAddHeaderIntoTrackerPool(t *testing.T) {
	t.Parallel()

	var addedNonces []uint64
	poolsHolderStub := initDataPool([]byte(""))
	poolsHolderStub.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				addedNonces = append(addedNonces, hdrNonce)
				return []data.HeaderHandler{&block.MetaBlock{Nonce: 1}}, [][]byte{[]byte("hash")}, nil
			},
		}
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolsHolderStub
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	roundHandlerMock := &mock.RoundHandlerMock{}
	coreComponents.RoundField = roundHandlerMock

	sp, _ := blproc.NewShardProcessor(arguments)

	roundHandlerMock.RoundIndex = 20
	metaBlock := &block.MetaBlock{
		Round: 9,
		Nonce: 5,
	}
	sp.RequestMetaHeadersIfNeeded(0, metaBlock)

	expectedAddedNonces := []uint64{6, 7}
	assert.Equal(t, expectedAddedNonces, addedNonces)
}

func TestShardProcessor_CheckEpochCorrectnessShouldRemoveAndRequestStartOfEpochMetaBlockWhenEpochDoesNotMatch(t *testing.T) {
	t.Parallel()

	removeHeaderByHashWasCalled := false
	requestMetaHeaderWasCalled := false
	epochStartMetaHash := []byte("epoch start meta hash")

	currentHeader := &block.Header{
		Epoch: uint32(1),
	}
	nextHeader := &block.Header{
		Epoch:              uint32(2),
		EpochStartMetaHash: epochStartMetaHash,
	}

	blockChainMock := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return currentHeader
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	epochStartTriggerStub := &mock.EpochStartTriggerStub{
		MetaEpochCalled: func() uint32 {
			return currentHeader.Epoch
		},
	}
	poolsHolderStub := &dataRetrieverMock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{
				RemoveHeaderByHashCalled: func(headerHash []byte) {
					if bytes.Equal(headerHash, epochStartMetaHash) {
						removeHeaderByHashWasCalled = true
					}
				},
			}
		},
	}

	ch := make(chan struct{})

	requestHandlerStub := &testscommon.RequestHandlerStub{
		RequestMetaHeaderCalled: func(headerHash []byte) {
			if bytes.Equal(headerHash, epochStartMetaHash) {
				requestMetaHeaderWasCalled = true
				close(ch)
			}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.BlockChain = blockChainMock
	dataComponents.DataPool = poolsHolderStub
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = epochStartTriggerStub
	arguments.RequestHandler = requestHandlerStub
	sp, _ := blproc.NewShardProcessor(arguments)

	err := sp.CheckEpochCorrectness(nextHeader)

	select {
	case <-ch:
	case <-time.After(time.Minute):
		assert.Fail(t, "timeout while waiting the sending of the request for the meta header")
	}

	assert.True(t, removeHeaderByHashWasCalled)
	assert.True(t, requestMetaHeaderWasCalled)
	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_CreateNewHeaderErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	cc, dc, _, sc := createMockComponentHolders()

	boostrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
	}

	arguments := CreateMockArgumentsMultiShard(cc, dc, boostrapComponents, sc)

	sp, err := blproc.NewShardProcessor(arguments)
	assert.Nil(t, err)

	h, err := sp.CreateNewHeader(1, 1)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
	assert.Nil(t, h)
}

func TestShardProcessor_CreateNewHeaderValsOK(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root")
	round := uint64(7)
	epoch := uint64(5)

	coreComponents, dataComponents, _, statusComponents := createMockComponentHolders()

	boostrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.HeaderV2{}
			},
		},
	}

	boostrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
		CreateCalled: func(epoch uint32) data.HeaderHandler {
			return &block.HeaderV2{
				Header: &block.Header{},
			}
		},
	}

	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, boostrapComponents, statusComponents)
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}

	sp, err := blproc.NewShardProcessor(arguments)
	assert.Nil(t, err)

	h, err := sp.CreateNewHeader(round, epoch)
	assert.Nil(t, err)
	assert.IsType(t, &block.HeaderV2{}, h)
	assert.Equal(t, round, h.GetRound())

	headerV2 := h.(*block.HeaderV2)
	assert.Equal(t, rootHash, headerV2.ScheduledRootHash)

	zeroInt := big.NewInt(0)
	assert.Equal(t, zeroInt, headerV2.GetDeveloperFees())
	assert.Equal(t, zeroInt, headerV2.GetAccumulatedFees())
}
