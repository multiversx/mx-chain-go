package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

//------- NewShardProcessor

func initAccountsMock() *mock.AccountsStub {
	rootHashCalled := func() ([]byte, error) {
		return []byte("rootHash"), nil
	}
	return &mock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
}

func initBasicTestData() (*mock.PoolsHolderFake, *blockchain.BlockChain, []byte, block.Body, [][]byte, *mock.HasherMock, *mock.MarshalizerMock, error, []byte) {
	tdp := mock.NewPoolsHolderFake()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 1,
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

func initBlockHeader(prevHash []byte, rootHash []byte, mbHdrs []block.MiniBlockHeader) block.Header {
	hdr := block.Header{
		Nonce:            2,
		PrevHash:         prevHash,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	return hdr
}

//------- NewBlockProcessor

func TestNewBlockProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		nil,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		nil,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilStorage, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		nil,
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilMarshalizerShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		nil,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		nil,
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilBlocksTrackerShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		nil,
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilBlocksTracker, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilRequestTransactionHandlerShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		nil,
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilTransactionPool, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilTxCoordinator(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		nil,
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilUint64Converter(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		nil,
	)
	assert.Equal(t, process.ErrNilUint64Converter, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.Nil(t, err)
	assert.NotNil(t, sp)
}

//------- ProcessBlock

func TestShardProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	blk := make(block.Body, 0)
	err := sp.ProcessBlock(nil, &block.Header{}, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestShardProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	body := make(block.Body, 0)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, nil, body, haveTime)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestShardProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, nil, haveTime)
	assert.Equal(t, process.ErrNilBlockBody, err)
}

func TestShardProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	blk := make(block.Body, 0)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, blk, nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestShardProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	blkc := &blockchain.BlockChain{}
	hdr := block.Header{
		Nonce:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
	}
	body := make(block.Body, 0)
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestShardProcessor_ProcessBlockHeaderBodyMismatchShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{}
	hdr := block.Header{
		Nonce:         1,
		PrevHash:      []byte(""),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

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
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) error {
				return process.ErrHigherNonceInTransaction
			},
		},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		accounts,
		tdp,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)

	// should return err
	err = sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestShardProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	hdr := &block.Header{
		Nonce:         0,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
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
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	hdr := &block.Header{
		Nonce:         0,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
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
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		RootHash:      []byte("root hash"),
	}
	body := make(block.Body, 0)
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
		},
	}
	err := sp.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrInvalidBlockHash, err)
}

func TestShardProcessor_ProcessBlockWithErrOnProcessBlockTransactionsCallShouldRevertState(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
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
		Nonce:            1,
		PrevHash:         []byte(""),
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
	txProcess := func(transaction *transaction.Transaction, round uint32) error {
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
	)
	container, _ := factory.Create()

	tc, _ := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		accounts,
		tdp,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		hasher,
		&mock.MarshalizerMock{},
		accounts,
		shardCoordinator,
		&mock.ForkDetectorMock{
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)

	// should return err
	err2 := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, err, err2)
	assert.True(t, wasCalled)
}

func TestShardProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
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
		Nonce:            1,
		PrevHash:         []byte(""),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrRootStateMissmatch, err)
	assert.True(t, wasCalled)
}

func TestShardProcessor_ProcessBlockOnlyIntraShardShouldPass(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
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
		Nonce:            1,
		PrevHash:         []byte(""),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockCrossShardWithoutMetaShouldFail(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
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
		Nonce:            1,
		PrevHash:         []byte(""),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

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

	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardId: mbHdr.ReceiverShardID,
		SenderShardId:   mbHdr.SenderShardID,
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
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)

	meta = block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrCrossShardMBWithoutConfirmationFromMeta, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockHaveTimeLessThanZeroShouldErr(t *testing.T) {
	t.Parallel()
	txHash := []byte("tx_hash1")
	tdp := initDataPool(txHash)

	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 1,
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
		Nonce:            2,
		PrevHash:         preHash,
		Signature:        []byte("signature"),
		PubKeysBitmap:    []byte("00110"),
		ShardId:          0,
		RootHash:         rootHash,
		MiniBlockHeaders: mbHdrs,
	}
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(tdp, genesisBlocks)
	haveTimeLessThanZero := func() time.Duration {
		return time.Duration(-1 * time.Millisecond)
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

	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardId: mbHdr.ReceiverShardID,
		SenderShardId:   mbHdr.SenderShardID,
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
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	meta = block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestShardProcessor_ProcessBlockWithWrongMiniBlockHeaderShouldErr(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx_hash1")
	tdp := initDataPool(txHash)
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 1,
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
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			RootHashCalled: rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

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
		Hash:            mbHash}
	mbHdrs := make([]block.MiniBlockHeader, 0)
	mbHdrs = append(mbHdrs, mbHdr)

	lastHdr := blkc.GetCurrentBlockHeader()
	prevHash, _ := core.CalculateHash(marshalizer, hasher, lastHdr)
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardId: mbHdr.ReceiverShardID,
		SenderShardId:   mbHdr.SenderShardID,
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
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	meta = &block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
		Round:     2,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)

	// set accounts not dirty
	journalLen := func() int { return 0 }
	revertToSnapshot := func(snapshot int) error {
		return nil
	}
	rootHashCalled := func() ([]byte, error) {
		return rootHash, nil
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{
			RequestHeaderHandlerByNonceCalled: func(destShardID uint32, nonce uint64) {
				atomic.AddInt32(&hdrNoncesRequestCalled, 1)
			},
		},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)

	sp.CheckAndRequestIfMetaHeadersMissing(2)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&hdrNoncesRequestCalled))
	assert.Equal(t, err, process.ErrTimeIsOut)
}

//-------- isMetaHeaderFinal
func TestShardProcessor_IsMetaHeaderFinalShouldPass(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderFake()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 1,
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
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardId: mbHdr.ReceiverShardID,
		SenderShardId:   mbHdr.SenderShardID,
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
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)
	tdp.MetaBlocks().Put(metaHash, meta)

	meta = &block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
		Round:     2,
		PrevHash:  metaHash,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))
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
		Nonce:     1,
		ShardInfo: shardHdrs,
		Round:     1,
	}
	ordered, _ := sp.GetOrderedMetaBlocks(3)
	res = sp.IsMetaHeaderFinal(meta, ordered, 0)
	assert.True(t, res)
}

//-------- requestFinalMissingHeaders
func TestShardProcessor_RequestFinalMissingHeaders(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderFake()
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	res := sp.RequestFinalMissingHeaders()
	assert.Equal(t, res > 0, true)
}

//--------- verifyIncludedMetaBlocksFinality
func TestShardProcessor_CheckMetaHeadersValidityAndFinalityShouldPass(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderFake()
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
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	shardMiniBlock := block.ShardMiniBlockHeader{
		ReceiverShardId: mbHdr.ReceiverShardID,
		SenderShardId:   mbHdr.SenderShardID,
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
	meta := &block.MetaBlock{
		Nonce:        1,
		ShardInfo:    shardHdrs,
		Round:        1,
		PrevHash:     prevHash,
		PrevRandSeed: prevMeta.GetRandSeed(),
	}
	metaBytes, _ := marshalizer.Marshal(meta)
	metaHash := hasher.Compute(string(metaBytes))
	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, metaHash)

	tdp.MetaBlocks().Put(metaHash, meta)

	prevHash, _ = core.CalculateHash(marshalizer, hasher, meta)
	meta = &block.MetaBlock{
		Nonce:     2,
		ShardInfo: make([]block.ShardData, 0),
		Round:     2,
		PrevHash:  prevHash,
	}
	metaBytes, _ = marshalizer.Marshal(meta)
	metaHash = hasher.Compute(string(metaBytes))

	tdp.MetaBlocks().Put(metaHash, meta)
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		genesisBlocks,
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	hdr.Round = 4

	err := sp.CheckMetaHeadersValidityAndFinality(&hdr)
	assert.Nil(t, err)
}

func TestShardProcessor_CheckMetaHeadersValidityAndFinalityShouldErr(t *testing.T) {
	t.Parallel()

	mbHdrs := make([]block.MiniBlockHeader, 0)
	rootHash := []byte("rootHash")
	txHash := []byte("txhash1")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	tdp := mock.NewPoolsHolderFake()
	genesisBlocks := createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3))
	sp, _ := blproc.NewShardProcessorEmptyWith3shards(tdp, genesisBlocks)

	lastHdr := genesisBlocks[0]
	prevHash, _ := core.CalculateHash(&mock.MarshalizerMock{}, &mock.HasherMock{}, lastHdr)
	hdr := initBlockHeader(prevHash, rootHash, mbHdrs)

	hdr.MetaBlockHashes = append(hdr.MetaBlockHashes, []byte("meta"))
	hdr.Round = 0
	err := sp.CheckMetaHeadersValidityAndFinality(&hdr)
	assert.Equal(t, err, process.ErrNilMetaBlockHeader)
}

//------- CommitBlock

func TestShardProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
		PrevHash:      []byte("zzz"),
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		marshalizer,
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	body := make(block.Body, 0)
	hdrUnit := &mock.StorerStub{
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

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
		},
		&mock.BlocksTrackerMock{
			AddBlockCalled: func(headerHandler data.HeaderHandler) {
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
	)

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
		PrevHash:      []byte("zzz"),
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

	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
				return nil
			},
		},
		&mock.BlocksTrackerMock{
			AddBlockCalled: func(headerHandler data.HeaderHandler) {
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	assert.Nil(t, err)

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
	)

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
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	body := make(block.Body, 0)
	store := initStore()

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	tdp.HeadersNoncesCalled = func() dataRetriever.Uint64SyncMapCacher {
		return nil
	}
	blkc := createTestBlockchain()
	err := sp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
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
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		RootHash:      rootHash,
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
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
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
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
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		tdp,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		hasher,
		&mock.MarshalizerMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		fd,
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)

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

	prevHdr := &block.Header{
		Nonce:         0,
		Round:         0,
		PubKeysBitmap: rootHash,
		PrevHash:      hdrHash,
		Signature:     rootHash,
		RootHash:      rootHash,
	}

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: rootHash,
		PrevHash:      hdrHash,
		Signature:     rootHash,
		RootHash:      rootHash,
	}
	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
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
		AddHeaderCalled: func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
			if header == hdr {
				forkDetectorAddCalled = true
				return nil
			}

			return errors.New("should have not got here")
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		hasher,
		&mock.MarshalizerMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		fd,
		&mock.BlocksTrackerMock{
			AddBlockCalled: func(headerHandler data.HeaderHandler) {
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

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
	//this should sleep as there is an async call to display current header and block in CommitBlock
	time.Sleep(time.Second)
}

func TestShardProcessor_CreateTxBlockBodyWithDirtyAccStateShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	bl, err := sp.CreateBlockBody(0, func() bool { return true })
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RootHashCalled:         rootHashfunc,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	haveTime := func() bool {
		return false
	}
	bl, err := sp.CreateBlockBody(0, haveTime)
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			JournalLenCalled: journalLen,
			RootHashCalled:   rootHashfunc,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	blk, err := sp.CreateBlockBody(0, haveTime)
	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

//------- ComputeNewNoncePrevHash

func TestNode_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}
	be, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
			return []byte("hdr hash")
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		shardCoordinator,
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(shardCoordinator),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	assert.NotNil(t, sp)
	hdr.PrevHash = hasher.Compute("prev hash")
	sp.DisplayLogInfo(hdr, txBlock, []byte("tx_hash1"), shardCoordinator.NumberOfShards(), shardCoordinator.SelfId(), tdp)
}

func TestBlockProcessor_CreateBlockHeaderShouldNotReturnNil(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	mbHeaders, err := bp.CreateBlockHeader(nil, 0, func() bool {
		return true
	})
	assert.Nil(t, err)
	assert.NotNil(t, mbHeaders)
	assert.Equal(t, 0, len(mbHeaders.(*block.Header).MiniBlockHeaders))
}

func TestShardProcessor_CreateBlockHeaderShouldErrWhenMarshalizerErrors(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{Fail: true},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
	mbHeaders, err := bp.CreateBlockHeader(body, 0, func() bool {
		return true
	})
	assert.NotNil(t, err)
	assert.Nil(t, mbHeaders)
}

func TestShardProcessor_CreateBlockHeaderReturnsOK(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
	mbHeaders, err := bp.CreateBlockHeader(body, 0, func() bool {
		return true
	})
	assert.Nil(t, err)
	assert.Equal(t, len(body), len(mbHeaders.(*block.Header).MiniBlockHeaders))
}

func TestShardProcessor_CommitBlockShouldRevertAccountStateWhenErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalEntries := 3
	revToSnapshot := func(snapshot int) error {
		journalEntries = 0
		return nil
	}
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
		TxHashes:        [][]byte{[]byte(txHash0)},
	}
	txHash1 := []byte("txHash1")
	mb1 := block.MiniBlock{
		ReceiverShardID: 1,
		SenderShardID:   0,
		TxHashes:        [][]byte{[]byte(txHash1)},
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
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		tdp,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	wr := wrongBody{}
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
		TxHashes:        [][]byte{[]byte(txHash0)},
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
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		tdp,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)

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
	dataPool := mock.NewPoolsHolderFake()

	//we will have a metablock that will return 3 miniblock hashes
	//1 miniblock hash will be in cache
	//2 will be requested on network

	miniBlockHash1 := []byte("miniblock hash 1 found in cache")
	miniBlockHash2 := []byte("miniblock hash 2")
	miniBlockHash3 := []byte("miniblock hash 3")

	metaBlock := mock.HeaderHandlerStub{
		GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
			return map[string]uint32{
				string(miniBlockHash1): 0,
				string(miniBlockHash2): 0,
				string(miniBlockHash3): 0,
			}
		},
	}

	//put this metaBlock inside datapool
	metaBlockHash := []byte("metablock hash")
	dataPool.MetaBlocks().Put(metaBlockHash, &metaBlock)
	//put the existing miniblock inside datapool
	dataPool.MiniBlocks().Put(miniBlockHash1, &block.MiniBlock{})

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
		dataPool,
		requestHandler,
		&mock.PreProcessorContainerMock{},
		&mock.InterimProcessorContainerMock{},
	)

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		requestHandler,
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)
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
	dataPool := mock.NewPoolsHolderFake()

	//we will have a metablock that will return 3 miniblock hashes
	//1 miniblock hash will be in cache
	//2 will be requested on network

	miniBlockHash1 := []byte("miniblock hash 1 found in cache")

	metaBlock := mock.HeaderHandlerStub{
		GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
			return map[string]uint32{
				string(miniBlockHash1): 0,
			}
		},
	}

	//put this metaBlock inside datapool
	metaBlockHash := []byte("metablock hash")
	dataPool.MetaBlocks().Put(metaBlockHash, &metaBlock)
	//put the existing miniblock inside datapool
	dataPool.MiniBlocks().Put(miniBlockHash1, &block.MiniBlock{})

	noOfMissingMiniBlocks := int32(0)

	requestHandler := &mock.RequestHandlerMock{RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
		atomic.AddInt32(&noOfMissingMiniBlocks, 1)
	}}

	tc, _ := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		dataPool,
		requestHandler,
		&mock.PreProcessorContainerMock{},
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		requestHandler,
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)
	sp.ReceivedMetaBlock(metaBlockHash)
	assert.Equal(t, int32(0), atomic.LoadInt32(&noOfMissingMiniBlocks))
}

//--------- createAndProcessCrossMiniBlocksDstMe
func TestShardProcessor_CreateAndProcessCrossMiniBlocksDstMe(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderFake()
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
		ReceiverShardId: mbHdr.ReceiverShardID,
		SenderShardId:   mbHdr.SenderShardID,
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

	haveTimeReturnsBool := func() bool {
		return true
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	miniBlockSlice, usedMetaHdrsHashes, noOfTxs, err := sp.CreateAndProcessCrossMiniBlocksDstMe(3, 2, 2, haveTimeReturnsBool)
	assert.Equal(t, err == nil, true)
	assert.Equal(t, len(miniBlockSlice) == 0, true)
	assert.Equal(t, len(usedMetaHdrsHashes) == 0, true)
	assert.Equal(t, noOfTxs, uint32(0))
}

//------- createMiniBlocks

func TestShardProcessor_CreateMiniBlocksShouldWorkWithIntraShardTxs(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

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
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  string(txHash1),
	}, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  string(txHash2),
	}, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  string(txHash3),
	}, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	txProcessorMock := &mock.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) error {
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

	factory, _ := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		initStore(),
		marshalizer,
		hasher,
		dataPool,
		&mock.AddressConverterMock{},
		accntAdapter,
		&mock.RequestHandlerMock{},
		txProcessorMock,
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		accntAdapter,
		dataPool,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		accntAdapter,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)

	blockBody, err := bp.CreateMiniBlocks(1, 15000, 0, func() bool { return true })

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
	dataPool := mock.NewPoolsHolderFake()

	miniblocks := make([]*block.MiniBlock, 6)
	miniblockHashes := make([][]byte, 6)

	destShards := []uint32{1, 3, 4}
	for i := 0; i < 6; i++ {
		mb, hash := createDummyMiniBlock(fmt.Sprintf("tx hash %d", i), marshalizer, hasher, destShardId, destShards[i/2])
		miniblocks[i] = mb
		miniblockHashes[i] = hash
	}

	//put 3 metablocks in pool
	mb1Hash := []byte("meta block 1")
	dataPool.MetaBlocks().Put(
		mb1Hash,
		createDummyMetaBlock(destShardId, destShards[0], miniblockHashes[0], miniblockHashes[1]),
	)
	mb2Hash := []byte("meta block 2")
	dataPool.MetaBlocks().Put(
		mb2Hash,
		createDummyMetaBlock(destShardId, destShards[1], miniblockHashes[2], miniblockHashes[3]),
	)
	mb3Hash := []byte("meta block 3")
	dataPool.MetaBlocks().Put(
		mb3Hash,
		createDummyMetaBlock(destShardId, destShards[2], miniblockHashes[4], miniblockHashes[5]),
	)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = destShardId
	shardCoordinator.SetNoShards(destShardId + 1)

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		initAccountsMock(),
		shardCoordinator,
		&mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		&mock.BlocksTrackerMock{
			RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
				return nil
			},
		},
		createGenesisBlocks(shardCoordinator),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	//create block body with first 3 miniblocks from miniblocks var
	blockBody := block.Body{miniblocks[0], miniblocks[1], miniblocks[2]}

	hashes := make([][]byte, 0)
	hashes = append(hashes, mb1Hash)
	hashes = append(hashes, mb2Hash)
	hashes = append(hashes, mb3Hash)
	blockHeader := &block.Header{MetaBlockHashes: hashes}

	_, err := bp.GetProcessedMetaBlocksFromPool(blockBody, blockHeader)

	assert.Nil(t, err)
	//check WasMiniBlockProcessed for remaining metablocks
	metaBlock2Recov, _ := dataPool.MetaBlocks().Get(mb2Hash)
	assert.True(t, (metaBlock2Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[2]))
	assert.False(t, (metaBlock2Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[3]))

	metaBlock3Recov, _ := dataPool.MetaBlocks().Get(mb3Hash)
	assert.False(t, (metaBlock3Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[4]))
	assert.False(t, (metaBlock3Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[5]))
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilBlockHeader(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	be, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	err := be.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilTxBlockBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	err := sp.RestoreBlockIntoPools(&block.Header{}, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestShardProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx hash 1")

	dataPool := mock.NewPoolsHolderFake()
	marshalizerMock := &mock.MarshalizerMock{}
	hasherMock := &mock.HasherStub{}

	body := make(block.Body, 0)
	tx := transaction.Transaction{Nonce: 1}
	buffTx, _ := marshalizerMock.Marshal(tx)

	store := &mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			m := make(map[string][]byte, 0)
			m[string(txHash)] = buffTx
			return m, nil
		},
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{
				RemoveCalled: func(key []byte) error {
					return nil
				},
			}
		},
	}

	factory, _ := shard.NewPreProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		store,
		marshalizerMock,
		hasherMock,
		dataPool,
		&mock.AddressConverterMock{},
		initAccountsMock(),
		&mock.RequestHandlerMock{},
		&mock.TxProcessorMock{},
		&mock.SCProcessorMock{},
		&mock.SmartContractResultsProcessorMock{},
	)
	container, _ := factory.Create()

	tc, err := coordinator.NewTransactionCoordinator(
		mock.NewMultiShardsCoordinatorMock(3),
		initAccountsMock(),
		dataPool,
		&mock.RequestHandlerMock{},
		container,
		&mock.InterimProcessorContainerMock{},
	)

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		store,
		hasherMock,
		marshalizerMock,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		tc,
		&mock.Uint64ByteSliceConverterMock{},
	)

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
	metablockHeader.SetMiniBlockProcessed(metablockHash, true)
	dataPool.MetaBlocks().Put(
		metablockHash,
		metablockHeader,
	)

	err = sp.RestoreBlockIntoPools(&block.Header{}, body)

	miniblockFromPool, _ := dataPool.MiniBlocks().Get(miniblockHash)
	txFromPool, _ := dataPool.Transactions().SearchFirstData(txHash)
	metablockFromPool, _ := dataPool.MetaBlocks().Get(metablockHash)
	metablock := metablockFromPool.(*block.MetaBlock)
	assert.Nil(t, err)
	assert.Equal(t, &miniblock, miniblockFromPool)
	assert.Equal(t, &tx, txFromPool)
	assert.Equal(t, false, metablock.GetMiniBlockProcessed(miniblockHash))
}

func TestShardProcessor_DecodeBlockBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	marshalizerMock := &mock.MarshalizerMock{}
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		marshalizerMock,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		marshalizerMock,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
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
	dataPool := initDataPool([]byte("tx_hash1"))

	shardNr := uint32(5)
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(shardNr),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := sp.LastNotarizedHdrs()
	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    44,
		RandSeed: prevRandSeed}
	lastNodesHdrs[sharding.MetachainShardId] = lastHdr

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(lastNodesHdrs[sharding.MetachainShardId].(*block.MetaBlock))
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
	assert.Equal(t, err, process.ErrRootStateMissmatch)

	currHdr.Nonce = 0
	prevHdr.Nonce = 0
	prevHdr.RootHash = nil
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)

	currHdr.Nonce = 46
	prevHdr.Nonce = 45
	prevHdr.Round = currHdr.Round + 1
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrLowShardHeaderRound)

	prevHdr.Round = currHdr.Round - 1
	currHdr.Nonce = prevHdr.Nonce + 2
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrWrongNonceInBlock)

	currHdr.Nonce = prevHdr.Nonce + 1
	prevHdr.RandSeed = []byte("randomwrong")
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrRandSeedMismatch)

	prevHdr.RandSeed = currRandSeed
	currHdr.PrevHash = []byte("wronghash")
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Equal(t, err, process.ErrInvalidBlockHash)

	currHdr.PrevHash = prevHash
	prevHdr.RootHash = []byte("prevRootHash")
	err = sp.IsHdrConstructionValid(currHdr, prevHdr)
	assert.Nil(t, err)
}

func TestShardProcessor_RemoveAndSaveLastNotarizedMetaHdrNoDstMB(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		store,
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(shardNr),
		forkDetector,
		&mock.BlocksTrackerMock{
			RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
				return nil
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := sp.LastNotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	lastNodesHdrs[sharding.MetachainShardId] = lastHdr

	//put the existing headers inside datapool

	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(lastNodesHdrs[sharding.MetachainShardId].(*block.MetaBlock))
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
	shardBlock := block.Body{}

	blockHeader := &block.Header{}

	// test header not in pool and defer called
	processedMetaHdrs, err := sp.GetProcessedMetaBlocksFromPool(shardBlock, blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetablocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 0, putCalledNr)

	lastNodesHdrs = sp.LastNotarizedHdrs()
	assert.Equal(t, firstNonce, lastNodesHdrs[sharding.MetachainShardId].GetNonce())
	assert.Equal(t, 0, len(processedMetaHdrs))

	// wrong header type in pool and defer called
	dataPool.MetaBlocks().Put(currHash, shardHdr)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	blockHeader = &block.Header{MetaBlockHashes: hashes}

	processedMetaHdrs, err = sp.GetProcessedMetaBlocksFromPool(shardBlock, blockHeader)
	assert.Equal(t, nil, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetablocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 0, putCalledNr)

	lastNodesHdrs = sp.LastNotarizedHdrs()
	assert.Equal(t, firstNonce, lastNodesHdrs[sharding.MetachainShardId].GetNonce())

	// put headers in pool
	dataPool.MetaBlocks().Put(currHash, currHdr)
	dataPool.MetaBlocks().Put(prevHash, prevHdr)

	hashes = make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader = &block.Header{MetaBlockHashes: hashes}

	processedMetaHdrs, err = sp.GetProcessedMetaBlocksFromPool(shardBlock, blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetablocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 4, putCalledNr)

	lastNodesHdrs = sp.LastNotarizedHdrs()
	assert.Equal(t, currHdr, lastNodesHdrs[sharding.MetachainShardId])
}

func createShardData(hasher hashing.Hasher, marshalizer marshal.Marshalizer, miniBlocks []block.MiniBlock) []block.ShardData {
	shardData := make([]block.ShardData, len(miniBlocks))
	for i := 0; i < len(miniBlocks); i++ {
		marshaled, _ := marshalizer.Marshal(miniBlocks[i])
		hashed := hasher.Compute(string(marshaled))

		shardMBHeader := block.ShardMiniBlockHeader{
			ReceiverShardId: miniBlocks[i].ReceiverShardID,
			SenderShardId:   miniBlocks[i].SenderShardID,
			TxCount:         uint32(len(miniBlocks[i].TxHashes)),
			Hash:            hashed,
		}
		shardMBHeaders := make([]block.ShardMiniBlockHeader, 0)
		shardMBHeaders = append(shardMBHeaders, shardMBHeader)

		shardData[0].ShardId = miniBlocks[i].SenderShardID
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
	dataPool := mock.NewPoolsHolderFake()
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		store,
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(shardNr),
		forkDetector,
		&mock.BlocksTrackerMock{
			RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
				return nil
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := sp.LastNotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	lastNodesHdrs[sharding.MetachainShardId] = lastHdr

	shardBlock := make(block.Body, 0)
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
	shardBlock = append(shardBlock, &miniblock1, &miniblock2, &miniblock3)

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)
	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(lastNodesHdrs[sharding.MetachainShardId].(*block.MetaBlock))
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
	dataPool.MetaBlocks().Put(currHash, currHdr)
	dataPool.MetaBlocks().Put(prevHash, prevHdr)

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader := &block.Header{MetaBlockHashes: hashes}

	processedMetaHdrs, err := sp.GetProcessedMetaBlocksFromPool(shardBlock, blockHeader)
	assert.Nil(t, err)

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetablocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 2, putCalledNr)

	lastNodesHdrs = sp.LastNotarizedHdrs()
	assert.Equal(t, prevHdr, lastNodesHdrs[sharding.MetachainShardId])
}

func TestShardProcessor_RemoveAndSaveLastNotarizedMetaHdrAllMBFinished(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		store,
		hasher,
		marshalizer,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(shardNr),
		forkDetector,
		&mock.BlocksTrackerMock{
			RemoveNotarisedBlocksCalled: func(headerHandler data.HeaderHandler) error {
				return nil
			},
			UnnotarisedBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, 0)
			},
		},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	prevRandSeed := []byte("prevrand")
	currRandSeed := []byte("currrand")
	lastNodesHdrs := sp.LastNotarizedHdrs()
	firstNonce := uint64(44)

	lastHdr := &block.MetaBlock{Round: 9,
		Nonce:    firstNonce,
		RandSeed: prevRandSeed}
	lastNodesHdrs[sharding.MetachainShardId] = lastHdr

	shardBlock := make(block.Body, 0)
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
	shardBlock = append(shardBlock, &miniblock1, &miniblock2, &miniblock3, &miniblock4)

	miniBlocks := make([]block.MiniBlock, 0)
	miniBlocks = append(miniBlocks, miniblock1, miniblock2)
	//header shard 0
	prevHash, _ := sp.ComputeHeaderHash(lastNodesHdrs[sharding.MetachainShardId].(*block.MetaBlock))
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
	dataPool.MetaBlocks().Put(currHash, currHdr)
	dataPool.MetaBlocks().Put(prevHash, prevHdr)
	dataPool.MetaBlocks().Put([]byte("shouldNotRemove"), &block.MetaBlock{
		Round:        12,
		PrevRandSeed: []byte("nextrand"),
		PrevHash:     currHash,
		Nonce:        47})

	hashes := make([][]byte, 0)
	hashes = append(hashes, currHash)
	hashes = append(hashes, prevHash)
	blockHeader := &block.Header{MetaBlockHashes: hashes}

	processedMetaHdrs, err := sp.GetProcessedMetaBlocksFromPool(shardBlock, blockHeader)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(processedMetaHdrs))

	err = sp.SaveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	assert.Nil(t, err)

	err = sp.RemoveProcessedMetablocksFromPool(processedMetaHdrs)
	assert.Nil(t, err)
	assert.Equal(t, 4, putCalledNr)

	lastNodesHdrs = sp.LastNotarizedHdrs()
	assert.Equal(t, currHdr, lastNodesHdrs[sharding.MetachainShardId])
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	hdr.MiniBlockHeaders[0].ReceiverShardID = body[0].ReceiverShardID + 1
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationSenderMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	hdr.MiniBlockHeaders[0].SenderShardID = body[0].SenderShardID + 1
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationTxCountMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	hdr.MiniBlockHeaders[0].TxCount = uint32(len(body[0].TxHashes) + 1)
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationHashMissmatch(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	hdr.MiniBlockHeaders[0].Hash = []byte("wrongHash")
	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_CheckHeaderBodyCorrelationShouldPass(t *testing.T) {
	t.Parallel()

	hdr, body := createOneHeaderOneBody()
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)

	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Nil(t, err)
}
