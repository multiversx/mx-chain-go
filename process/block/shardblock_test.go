package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	blproc "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

//------- NewShardProcessor

func initAccountsMock() *mock.AccountsStub {
	rootHashCalled := func() []byte {
		return []byte("rootHash")
	}
	return &mock.AccountsStub{
		RootHashCalled: rootHashCalled,
	}
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, sp)
}

func TestNewShardProcessor_NilTxProcessorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	sp, err := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilTxProcessor, err)
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
		&mock.TxProcessorMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		nil,
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		nil,
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		nil,
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
	assert.Equal(t, process.ErrNilTransactionPool, err)
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
	blk := make(block.Body, 0)
	err := sp.ProcessBlock(&blockchain.BlockChain{}, &block.Header{}, blk, nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestShardProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	tpm := mock.TxProcessorMock{}
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
		&tpm,
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
	// invalid transaction
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, nil
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return []byte("rootHash")
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHeaderBodyMismatch, err)
}

func TestShardProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
	txHash := []byte("tx_hash1")
	// invalid transaction
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, process.ErrHigherNonceInTransaction
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return []byte("rootHash")
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		hasher,
		marshalizer,
		&tpm,
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
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
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
	err := errors.New("process block transaction error")
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, err
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return []byte("rootHash")
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, nil
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return []byte("rootHashX")
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, nil
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return rootHash
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, nil
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return rootHash
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrCrossShardMBWithoutConfirmationFromMeta, err)
	assert.False(t, wasCalled)
}

func TestShardProcessor_ProcessBlockCrossShardWithMetaShouldPass(t *testing.T) {
	t.Parallel()

	tdp := mock.NewPoolsHolderFake()
	txHash := []byte("tx_hash1")
	tdp.Transactions().AddData(txHash, &transaction.Transaction{}, process.ShardCacherIdentifier(1, 0))
	txProcess := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, nil
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
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
	rootHashCalled := func() []byte {
		return rootHash
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	)

	// should return err
	err := sp.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrCrossShardMBWithoutConfirmationFromMeta, err)
	assert.False(t, wasCalled)
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
		&mock.TxProcessorMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		RootHashCalled: func() []byte {
			return rootHash
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
		&mock.TxProcessorMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		RootHashCalled: func() []byte {
			return rootHash
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
		&mock.TxProcessorMock{},
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
		RootHashCalled: func() []byte {
			return rootHash
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
		&mock.TxProcessorMock{},
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
		RootHashCalled: func() []byte {
			return rootHash
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
		&mock.TxProcessorMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
	tdp.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
		return nil
	}
	blkc := createTestBlockchain()
	err := sp.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestShardProcessor_CommitBlockNoTxInPoolShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))
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
		RootHashCalled: func() []byte {
			return rootHash
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

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		store,
		hasher,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewMultiShardsCoordinatorMock(3),
		fd,
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
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
		}
	}
	blkc := createTestBlockchain()

	err := sp.CommitBlock(blkc, hdr, body)
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
		RootHashCalled: func() []byte {
			return rootHash
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
		&mock.TxProcessorMock{ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
			return nil, nil
		}},
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
	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,

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
	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 0 }
	rootHashfunc := func() []byte { return []byte("roothash") }
	revToSnapshot := func(snapshot int) error { return nil }
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
	//process transaction. return nil for no error
	procTx := func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
		return nil, nil
	}
	tpm := mock.TxProcessorMock{
		ProcessTransactionCalled: procTx,
	}
	journalLen := func() int { return 0 }
	rootHashfunc := func() []byte { return []byte("roothash") }
	haveTime := func() bool {
		return true
	}
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		shardCoordinator,
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(shardCoordinator),
		true,
		&mock.RequestHandlerMock{},
	)
	assert.NotNil(t, sp)
	hdr.PrevHash = hasher.Compute("prev hash")
	blproc.DisplayLogInfo(hdr, txBlock, []byte("tx_hash1"), shardCoordinator.NumberOfShards(), shardCoordinator.SelfId(), tdp)
}

//------- SortTxByNonce

func TestSortTxByNonce_NilCacherShouldErr(t *testing.T) {
	t.Parallel()
	transactions, txHashes, err := blproc.SortTxByNonce(nil)
	assert.Nil(t, transactions)
	assert.Nil(t, txHashes)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestSortTxByNonce_EmptyCacherShouldReturnEmpty(t *testing.T) {
	t.Parallel()
	cacher, _ := storageUnit.NewCache(storageUnit.LRUCache, 100, 1)
	transactions, txHashes, err := blproc.SortTxByNonce(cacher)
	assert.Equal(t, 0, len(transactions))
	assert.Equal(t, 0, len(txHashes))
	assert.Nil(t, err)
}

func TestSortTxByNonce_OneTxShouldWork(t *testing.T) {
	t.Parallel()
	cacher, _ := storageUnit.NewCache(storageUnit.LRUCache, 100, 1)
	hash, tx := createRandTx(r)
	cacher.HasOrAdd(hash, tx)
	transactions, txHashes, err := blproc.SortTxByNonce(cacher)
	assert.Equal(t, 1, len(transactions))
	assert.Equal(t, 1, len(txHashes))
	assert.Nil(t, err)
	assert.True(t, hashInSlice(hash, txHashes))
	assert.True(t, txInSlice(tx, transactions))
}

func createRandTx(rand *rand.Rand) ([]byte, *transaction.Transaction) {
	mutex.Lock()
	nonce := rand.Uint64()
	mutex.Unlock()
	tx := &transaction.Transaction{
		Nonce: nonce,
	}
	marshalizer := &mock.MarshalizerMock{}
	buffTx, _ := marshalizer.Marshal(tx)
	hash := mock.HasherMock{}.Compute(string(buffTx))
	return hash, tx
}

func hashInSlice(hash []byte, hashes [][]byte) bool {
	for _, h := range hashes {
		if bytes.Equal(h, hash) {
			return true
		}
	}
	return false
}

func txInSlice(tx *transaction.Transaction, transactions []*transaction.Transaction) bool {
	for _, t := range transactions {
		if reflect.DeepEqual(tx, t) {
			return true
		}
	}
	return false
}

func TestSortTxByNonce_MoreTransactionsShouldNotErr(t *testing.T) {
	t.Parallel()
	cache, _, _ := genCacherTransactionsHashes(100)
	_, _, err := blproc.SortTxByNonce(cache)
	assert.Nil(t, err)
}

func TestSortTxByNonce_MoreTransactionsShouldRetSameSize(t *testing.T) {
	t.Parallel()
	cache, genTransactions, _ := genCacherTransactionsHashes(100)
	transactions, txHashes, _ := blproc.SortTxByNonce(cache)
	assert.Equal(t, len(genTransactions), len(transactions))
	assert.Equal(t, len(genTransactions), len(txHashes))
}

func TestSortTxByNonce_MoreTransactionsShouldContainSameElements(t *testing.T) {
	t.Parallel()
	cache, genTransactions, genHashes := genCacherTransactionsHashes(100)
	transactions, txHashes, _ := blproc.SortTxByNonce(cache)
	for i := 0; i < len(genTransactions); i++ {
		assert.True(t, hashInSlice(genHashes[i], txHashes))
		assert.True(t, txInSlice(genTransactions[i], transactions))
	}
}

func TestSortTxByNonce_MoreTransactionsShouldContainSortedElements(t *testing.T) {
	t.Parallel()
	cache, _, _ := genCacherTransactionsHashes(100)
	transactions, _, _ := blproc.SortTxByNonce(cache)
	lastNonce := uint64(0)
	for i := 0; i < len(transactions); i++ {
		tx := transactions[i]
		assert.True(t, lastNonce <= tx.Nonce)
		fmt.Println(tx.Nonce)
		lastNonce = tx.Nonce
	}
}

func TestSortTxByNonce_TransactionsWithSameNonceShouldGetSorted(t *testing.T) {
	t.Parallel()
	transactions := []*transaction.Transaction{
		{Nonce: 1, Signature: []byte("sig1")},
		{Nonce: 2, Signature: []byte("sig2")},
		{Nonce: 1, Signature: []byte("sig3")},
		{Nonce: 2, Signature: []byte("sig4")},
		{Nonce: 3, Signature: []byte("sig5")},
	}
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, uint32(len(transactions)), 1)
	for _, tx := range transactions {
		marshalizer := &mock.MarshalizerMock{}
		buffTx, _ := marshalizer.Marshal(tx)
		hash := mock.HasherMock{}.Compute(string(buffTx))

		cache.Put(hash, tx)
	}
	sortedTxs, _, _ := blproc.SortTxByNonce(cache)
	lastNonce := uint64(0)
	for i := 0; i < len(sortedTxs); i++ {
		tx := sortedTxs[i]
		assert.True(t, lastNonce <= tx.Nonce)
		fmt.Printf("tx.Nonce: %d, tx.Sig: %s\n", tx.Nonce, tx.Signature)
		lastNonce = tx.Nonce
	}
	assert.Equal(t, len(sortedTxs), len(transactions))
	//test if one transaction from transactions might not be in sortedTx
	for _, tx := range transactions {
		found := false
		for _, stx := range sortedTxs {
			if reflect.DeepEqual(tx, stx) {
				found = true
				break
			}
		}
		if !found {
			assert.Fail(t, "Not found tx in sorted slice for sig: "+string(tx.Signature))
		}
	}
}

func genCacherTransactionsHashes(noOfTx int) (storage.Cacher, []*transaction.Transaction, [][]byte) {
	cacher, _ := storageUnit.NewCache(storageUnit.LRUCache, uint32(noOfTx), 1)
	genHashes := make([][]byte, 0)
	genTransactions := make([]*transaction.Transaction, 0)
	for i := 0; i < noOfTx; i++ {
		hash, tx := createRandTx(r)
		cacher.HasOrAdd(hash, tx)

		genHashes = append(genHashes, hash)
		genTransactions = append(genTransactions, tx)
	}
	return cacher, genTransactions, genHashes
}

func BenchmarkSortTxByNonce1(b *testing.B) {
	cache, _, _ := genCacherTransactionsHashes(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = blproc.SortTxByNonce(cache)
	}
}

func TestBlockProcessor_CreateBlockHeaderShouldNotReturnNil(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		initDataPool([]byte("tx_hash1")),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		&mock.AccountsStub{
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	msh, mstx, err := sp.MarshalizedDataToBroadcast(&block.Header{}, body)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
	assert.Equal(t, 0, len(msh))
	assert.Equal(t, 0, len(mstx))
}

//------- receivedMiniBlock

func TestShardProcessor_ReceivedMiniBlockShouldRequestMissingTransactions(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	//we will have a miniblock that will have 3 tx hashes
	//1 tx hash will be in cache
	//2 will be requested on network

	txHash1 := []byte("tx hash 1 found in cache")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(1)
	receiverShardId := uint32(2)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	//put this miniblock inside datapool
	miniBlockHash := []byte("miniblock hash")
	dataPool.MiniBlocks().Put(miniBlockHash, miniBlock)

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{}, cacheId)

	txHash1Requested := int32(0)
	txHash2Requested := int32(0)
	txHash3Requested := int32(0)

	requestHandler := &mock.RequestHandlerMock{}
	requestHandler.RequestTransactionHandlerCalled = func(destShardID uint32, txHashes [][]byte) {
		if containsHash(txHashes, txHash1) {
			atomic.AddInt32(&txHash1Requested, 1)
		}
		if containsHash(txHashes, txHash2) {
			atomic.AddInt32(&txHash2Requested, 1)
		}
		if containsHash(txHashes, txHash3) {
			atomic.AddInt32(&txHash3Requested, 1)
		}
	}

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		requestHandler,
	)

	bp.ReceivedMiniBlock(miniBlockHash)

	//we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&txHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&txHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&txHash2Requested))
}

func containsHash(txHashes [][]byte, hash []byte) bool {
	for _, txHash := range txHashes {
		if bytes.Equal(hash, txHash) {
			return true
		}
	}

	return false
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

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{RequestMiniBlockHandlerCalled: func(destShardID uint32, miniblockHash []byte) {
			if bytes.Equal(miniBlockHash1, miniblockHash) {
				atomic.AddInt32(&miniBlockHash1Requested, 1)
			}
			if bytes.Equal(miniBlockHash2, miniblockHash) {
				atomic.AddInt32(&miniBlockHash2Requested, 1)
			}
			if bytes.Equal(miniBlockHash3, miniblockHash) {
				atomic.AddInt32(&miniBlockHash3Requested, 1)
			}
		}},
	)
	bp.ReceivedMetaBlock(metaBlockHash)

	//we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&miniBlockHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
}

//------- processMiniBlockComplete

func TestShardProcessor_ProcessMiniBlockCompleteWithOkTxsShouldExecuteThemAndNotRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	//we will have a miniblock that will have 3 tx hashes
	//all txs will be in datapool and none of them will return err when processed
	//so, tx processor will return nil on processing tx

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(1)
	receiverShardId := uint32(2)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
				//execution, in this context, means moving the tx nonce to itx corresponding execution result variable
				if bytes.Equal(transaction.Data, txHash1) {
					tx1ExecutionResult = transaction.Nonce
				}
				if bytes.Equal(transaction.Data, txHash2) {
					tx2ExecutionResult = transaction.Nonce
				}
				if bytes.Equal(transaction.Data, txHash3) {
					tx3ExecutionResult = transaction.Nonce
				}

				return nil, nil
			},
		},
		&mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				assert.Fail(t, "revert should have not been called")
				return nil
			},
			JournalLenCalled: func() int {
				return 0
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	err := bp.ProcessMiniBlockComplete(&miniBlock, 0, func() bool {
		return true
	})

	assert.Nil(t, err)
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
}

func TestShardProcessor_ProcessMiniBlockCompleteWithErrorWhileProcessShouldCallRevertAccntState(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	//we will have a miniblock that will have 3 tx hashes
	//all txs will be in datapool and none of them will return err when processed
	//so, tx processor will return nil on processing tx

	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2 - this will cause the tx processor to err")
	txHash3 := []byte("tx hash 3")

	senderShardId := uint32(1)
	receiverShardId := uint32(2)

	miniBlock := block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: receiverShardId,
		TxHashes:        [][]byte{txHash1, txHash2, txHash3},
	}

	tx1Nonce := uint64(45)
	tx2Nonce := uint64(46)
	tx3Nonce := uint64(47)

	errTxProcessor := errors.New("tx processor failing")

	//put the existing tx inside datapool
	cacheId := process.ShardCacherIdentifier(senderShardId, receiverShardId)
	dataPool.Transactions().AddData(txHash1, &transaction.Transaction{
		Nonce: tx1Nonce,
		Data:  txHash1,
	}, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, cacheId)

	currentJournalLen := 445
	revertAccntStateCalled := false

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
				if bytes.Equal(transaction.Data, txHash2) {
					return nil, errTxProcessor
				}

				return nil, nil
			},
		},
		&mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				if snapshot == currentJournalLen {
					revertAccntStateCalled = true
				}

				return nil
			},
			JournalLenCalled: func() int {
				return currentJournalLen
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	err := bp.ProcessMiniBlockComplete(&miniBlock, 0, func() bool { return true })

	assert.Equal(t, errTxProcessor, err)
	assert.True(t, revertAccntStateCalled)
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
		Data:  txHash1,
	}, cacheId)
	dataPool.Transactions().AddData(txHash2, &transaction.Transaction{
		Nonce: tx2Nonce,
		Data:  txHash2,
	}, cacheId)
	dataPool.Transactions().AddData(txHash3, &transaction.Transaction{
		Nonce: tx3Nonce,
		Data:  txHash3,
	}, cacheId)

	tx1ExecutionResult := uint64(0)
	tx2ExecutionResult := uint64(0)
	tx3ExecutionResult := uint64(0)

	bp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round uint32) ([]*smartContractResult.SmartContractResult, error) {
				//execution, in this context, means moving the tx nonce to itx corresponding execution result variable
				if bytes.Equal(transaction.Data, txHash1) {
					tx1ExecutionResult = transaction.Nonce
				}
				if bytes.Equal(transaction.Data, txHash2) {
					tx2ExecutionResult = transaction.Nonce
				}
				if bytes.Equal(transaction.Data, txHash3) {
					tx3ExecutionResult = transaction.Nonce
				}

				return nil, nil
			},
		},
		&mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				assert.Fail(t, "revert should have not been called")
				return nil
			},
			JournalLenCalled: func() int {
				return 0
			},
		},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	blockBody, err := bp.CreateMiniBlocks(1, 15000, 0, func() bool {
		return true
	})

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
		&mock.TxProcessorMock{},
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
	)

	//create block body with first 3 miniblocks from miniblocks var
	blockBody := block.Body{miniblocks[0], miniblocks[1], miniblocks[2]}

	_, err := bp.GetProcessedMetaBlocksFromPool(blockBody)

	assert.Nil(t, err)
	//check WasMiniBlockProcessed for remaining metablocks
	metaBlock2Recov, _ := dataPool.MetaBlocks().Get(mb2Hash)
	assert.True(t, (metaBlock2Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[2]))
	assert.False(t, (metaBlock2Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[3]))

	metaBlock3Recov, _ := dataPool.MetaBlocks().Get(mb3Hash)
	assert.False(t, (metaBlock3Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[4]))
	assert.False(t, (metaBlock3Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[5]))
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilBlockChain(t *testing.T) {
	t.Parallel()
	tdp := initDataPool([]byte("tx_hash1"))

	be, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)
	err := be.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilTxBlockBody, err)
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	err := sp.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestShardProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()
	dataPool := mock.NewPoolsHolderFake()
	marshalizerMock := &mock.MarshalizerMock{}
	hasherMock := &mock.HasherStub{}

	body := make(block.Body, 0)
	tx := transaction.Transaction{Nonce: 1}
	buffTx, _ := marshalizerMock.Marshal(tx)
	txHash := []byte("tx hash 1")

	store := &mock.ChainStorerMock{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			m := make(map[string][]byte, 0)
			m[string(txHash)] = buffTx
			return m, nil
		},
	}

	sp, _ := blproc.NewShardProcessor(
		&mock.ServiceContainerMock{},
		dataPool,
		store,
		hasherMock,
		marshalizerMock,
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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

	err := sp.RestoreBlockIntoPools(nil, body)

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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		initAccountsMock(),
		mock.NewMultiShardsCoordinatorMock(shardNr),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(shardNr)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
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

	// test header not in pool and defer called
	processedMetaHdrs, err := sp.GetProcessedMetaBlocksFromPool(shardBlock)
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

	processedMetaHdrs, err = sp.GetProcessedMetaBlocksFromPool(shardBlock)
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

	processedMetaHdrs, err = sp.GetProcessedMetaBlocksFromPool(shardBlock)
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
		&mock.TxProcessorMock{},
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

	processedMetaHdrs, err := sp.GetProcessedMetaBlocksFromPool(shardBlock)
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
		&mock.TxProcessorMock{},
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
	dataPool.MetaBlocks().Put([]byte("shouldNotRemove"), &block.MetaBlock{Nonce: 47})

	processedMetaHdrs, err := sp.GetProcessedMetaBlocksFromPool(shardBlock)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(processedMetaHdrs))

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
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
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
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		createGenesisBlocks(mock.NewMultiShardsCoordinatorMock(3)),
		true,
		&mock.RequestHandlerMock{},
	)

	err := sp.CheckHeaderBodyCorrelation(hdr, body)
	assert.Nil(t, err)
}
