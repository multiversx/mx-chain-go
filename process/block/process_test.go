package block_test

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func haveTime() time.Duration {
	return time.Duration(2000 * time.Millisecond)
}

func createTestBlockchain() *blockchain.BlockChain {
	blockChain, _ := blockchain.NewBlockChain(
		generateTestCache(),
		generateTestUnit(),
		generateTestUnit(),
		generateTestUnit(),
		generateTestUnit(),
	)

	return blockChain
}

func generateTestCache() storage.Cacher {
	cache, _ := storage.NewCache(storage.LRUCache, 1000)
	return cache
}

func generateTestUnit() storage.Storer {
	memDB, _ := memorydb.New()

	storer, _ := storage.NewStorageUnit(
		generateTestCache(),
		memDB,
	)

	return storer
}

func initDataPool() *mock.TransientDataPoolStub {
	tdp := &mock.TransientDataPoolStub{
		TransactionsCalled: func() data.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(shardID uint32) (c storage.Cacher) {
					return &mock.CacherStub{
						GetCalled: func(key []byte) (value interface{}, ok bool) {
							if reflect.DeepEqual(key, []byte("tx1_hash")) {
								return &transaction.Transaction{Nonce: 10}, true
							}
							return nil, false
						},
						KeysCalled: func() [][]byte {
							return [][]byte{[]byte("key1"), []byte("key2")}
						},
						LenCalled: func() int {
							return 0
						},
					}
				},
				RemoveSetOfDataFromPoolCalled: func(keys [][]byte, destShardID uint32) {},
			}
		},
		HeadersNoncesCalled: func() data.Uint64Cacher {
			return &mock.Uint64CacherStub{
				PutCalled: func(u uint64, i []byte) bool {
					return true
				},
			}
		},
	}

	return tdp
}

//------- NewBlockProcessor

func TestNewBlockProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	be, err := blproc.NewBlockProcessor(
		nil,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		nil,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilMarshalizerShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		nil,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilTxProcessorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		nil,
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilRequestTransactionHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		nil,
	)

	assert.Equal(t, process.ErrNilTransactionHandler, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()
	tdp.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return nil
	}

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Equal(t, process.ErrNilTransactionPool, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.Nil(t, err)
	assert.NotNil(t, be)
}

//------- ProcessAndCommit

func TestBlockProcessor_ProcessAndCommitNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.ProcessAndCommit(nil, &block.Header{}, make(block.Body, 0), haveTime)

	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestBlockProcessor_ProcessAndCommitNilHeaderShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.ProcessAndCommit(createTestBlockchain(), nil, make(block.Body, 0), haveTime)

	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBlockProcessor_ProcessAndCommitNilTxBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.ProcessAndCommit(createTestBlockchain(), &block.Header{}, nil, haveTime)

	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestBlockProcessor_ProcessAndCommitBlockWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	blkc := createTestBlockchain()

	hdr := block.Header{
		Nonce:         0,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("roothash"),
	}

	body := make(block.Body, 0)

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	// should return err
	err := be.ProcessAndCommit(blkc, &hdr, body, haveTime)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestBlockProcessor_ProcessAndCommitBlockWithInvalidTransactionShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	txHash := []byte("tx_hash1")

	// invalid transaction
	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return process.ErrHigherNonceInTransaction
	}

	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	blkc := createTestBlockchain()
	hdr := block.Header{
		Nonce:         0,
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("00110"),
		ShardId:       0,
		Commitment:    []byte("commitment"),
		RootHash:      []byte("rootHash"),
	}
	body := make(block.Body, 0)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}
	body = append(body, &miniblock)

	// set accounts dirty
	journalLen := func() int { return 0 }
	// set revertToSnapshot
	revertToSnapshot := func(snapshot int) error { return nil }

	rootHashCalled := func() []byte {
		return []byte("rootHash")
	}

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
			RootHashCalled:         rootHashCalled,
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	go func() {
		be.ChRcvAllTxs <- true
	}()

	// should return err
	err := be.ProcessAndCommit(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestBlockProcessor_ProcessAndCommitHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	hdr := &block.Header{
		Nonce:         0,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	body := make(block.Body, 0)

	blkc := createTestBlockchain()

	err := be.ProcessAndCommit(blkc, hdr, body, haveTime)

	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestBlockProcessor_ProcessAndCommitHeaderNotCorrectNonceShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	hdr := &block.Header{
		Nonce:         0,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	body := make(block.Body, 0)

	blkc := createTestBlockchain()
	blkc.CurrentBlockHeader = &block.Header{
		Nonce: 0,
	}

	err := be.ProcessAndCommit(blkc, hdr, body, haveTime)

	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestBlockProcessor_ProcessAndCommitHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	body := make(block.Body, 0)

	blkc := createTestBlockchain()
	blkc.CurrentBlockHeader = &block.Header{
		Nonce: 0,
	}
	blkc.CurrentBlockHeaderHash = []byte("bbb")

	err := be.ProcessAndCommit(blkc, hdr, body, haveTime)

	assert.Equal(t, process.ErrInvalidBlockHash, err)
}

//------- ProcessBlock

func TestBlockProcessor_ProcessBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.ProcessBlock(nil, &block.Header{}, make(block.Body, 0), func() time.Duration {
		return time.Second
	})

	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestBlockProcessor_ProcessBlockNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.ProcessBlock(createTestBlockchain(), &block.Header{}, make(block.Body, 0), nil)

	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

//------- CommitBlock

func TestBlockProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.CommitBlock(nil, &block.Header{}, make(block.Body, 0))

	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestBlockProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	errMarshalizer := errors.New("failure")

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
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

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	blkc := createTestBlockchain()

	err := be.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrMarshalWithoutSuccess, err)
}

func TestBlockProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	errPersister := errors.New("failure")

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	body := make(block.Body, 0)

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	hdrUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPersister
		},
	}

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
		generateTestUnit(),
		generateTestUnit(),
		generateTestUnit(),
		hdrUnit,
	)

	err := be.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrPersistWithoutSuccess, err)
}

func TestBlockProcessor_CommitBlockStorageFailsForBodyShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	errPersister := errors.New("failure")

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	mb := block.MiniBlock{}
	body := make(block.Body, 0)
	body = append(body, &mb)

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{
			CommitCalled: func() (i []byte, e error) {
				return nil, nil
			},
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header *block.Header, hash []byte, isReceived bool) error {
				return nil
			},
		},
		func(destShardID uint32, txHash []byte) {
		},
	)

	txBlockUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPersister
		},
	}

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
		generateTestUnit(),
		txBlockUnit,
		generateTestUnit(),
		generateTestUnit(),
	)

	err := be.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrPersistWithoutSuccess, err)
}

func TestBlockProcessor_CommitBlockNilNoncesDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	body := make(block.Body, 0)

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	tdp.HeadersNoncesCalled = func() data.Uint64Cacher {
		return nil
	}

	blkc := createTestBlockchain()
	err := be.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestBlockProcessor_CommitBlockNoTxInPoolShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	txHash := []byte("txHash")
	rootHash := []byte("root hash")
	hdrHash := []byte("header hash")

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
	}
	body := block.Body{&mb}

	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
	}

	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header *block.Header, hash []byte, isReceived bool) error {
			return nil
		},
	}

	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}

	be, _ := blproc.NewBlockProcessor(
		tdp,
		hasher,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		fd,
		func(destShardID uint32, txHash []byte) {
		},
	)

	txCache := &mock.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		LenCalled: func() int {
			return 0
		},
	}

	tdp.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			ShardDataStoreCalled: func(shardID uint32) (c storage.Cacher) {
				return txCache
			},

			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, destShardID uint32) {
			},
		}

	}

	blkc := createTestBlockchain()
	err := be.CommitBlock(blkc, hdr, body)

	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestBlockProcessor_CommitBlockOkValsShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	txHash := []byte("txHash")
	tx := &transaction.Transaction{}
	rootHash := []byte("root hash")
	hdrHash := []byte("header hash")

	hdr := &block.Header{
		Nonce:         1,
		Round:         1,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte("zzz"),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
		RootHash:      []byte("root hash"),
	}

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte(txHash)},
	}
	body := block.Body{&mb}

	accounts := &mock.AccountsStub{
		CommitCalled: func() (i []byte, e error) {
			return rootHash, nil
		},
	}

	forkDetectorAddCalled := false

	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header *block.Header, hash []byte, isReceived bool) error {
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

	be, _ := blproc.NewBlockProcessor(
		tdp,
		hasher,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		fd,
		func(destShardID uint32, txHash []byte) {
		},
	)

	txCache := &mock.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(txHash, key) {
				return tx, true
			}
			return nil, false
		},
		LenCalled: func() int {
			return 0
		},
	}

	removeTxWasCalled := false

	tdp.TransactionsCalled = func() data.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			ShardDataStoreCalled: func(shardID uint32) (c storage.Cacher) {
				return txCache
			},

			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, destShardID uint32) {
				if bytes.Equal(keys[0], []byte(txHash)) && len(keys) == 1 {
					removeTxWasCalled = true
				}
			},
		}

	}

	blkc := createTestBlockchain()
	err := be.CommitBlock(blkc, hdr, body)

	assert.Nil(t, err)
	assert.True(t, removeTxWasCalled)
	assert.True(t, forkDetectorAddCalled)
	assert.True(t, blkc.CurrentBlockHeader == hdr)
	assert.Equal(t, hdrHash, blkc.CurrentBlockHeaderHash)

	//this should sleep as there is an async call to display current header and block in CommitBlock
	time.Sleep(time.Second)
}

func TestVerifyStateRoot_ShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	rootHash := []byte("root hash to be tested")

	accounts := &mock.AccountsStub{
		RootHashCalled: func() []byte {
			return rootHash
		},
	}

	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	assert.True(t, be.VerifyStateRoot(rootHash))
}

func TestBlockProc_GetTransactionFromPool(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	txHash := []byte("tx1_hash")
	tx := be.GetTransactionFromPool(1, txHash)

	assert.NotNil(t, tx)
	assert.Equal(t, uint64(10), tx.Nonce)
}

func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	shardId := uint32(1)
	txHash1 := []byte("tx1_hash1")


	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	mBlk := block.MiniBlock{ShardID: shardId, TxHashes: txHashes}
	body = append(body, &mBlk)

	//TODO refactor the test

	if be.RequestTransactionFromNetwork(body) > 0 {
		be.WaitForTxHashes(haveTime())
	}
}

func TestBlockProc_CreateTxBlockBodyWithDirtyAccStateShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,

		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	bl, err := be.CreateTxBlockBody(0, 100, 0, func() bool { return true })

	// nil block
	assert.Nil(t, bl)
	// error
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestBlockProcessor_CreateTxBlockBodyWithNoTimeShouldEmptyBlock(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 0 }
	rootHashfunc := func() []byte { return []byte("roothash") }
	revToSnapshot := func(snapshot int) error { return nil }

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RootHashCalled:         rootHashfunc,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	haveTime := func() bool {
		return false
	}

	bl, err := be.CreateTxBlockBody(0, 100, 0, haveTime)

	// no error
	assert.Nil(t, err)
	// no miniblocks
	assert.Equal(t, len(bl), 0)
}

func TestBlockProcessor_CreateTxBlockBodyOK(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	//process transaction. return nil for no error
	procTx := func(transaction *transaction.Transaction, round int32) error {
		return nil
	}

	tpm := mock.TxProcessorMock{
		ProcessTransactionCalled: procTx,
	}

	journalLen := func() int { return 0 }
	rootHashfunc := func() []byte { return []byte("roothash") }

	haveTime := func() bool {
		return true
	}

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled: journalLen,
			RootHashCalled:   rootHashfunc,
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	blk, err := be.CreateTxBlockBody(0, 100, 0, haveTime)

	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithNilTxProcessorShouldPanic(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, nil,
		nil,
		nil,
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	createGenesis := func() {
		be.CreateGenesisBlock(nil)
	}

	assert.Panics(t, createGenesis)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithFailSetBalanceShouldPanic(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return nil
	}

	setBalances := func(accBalance map[string]*big.Int) (rootHash []byte, err error) {
		return nil, process.ErrAccountStateDirty
	}

	txProc := mock.TxProcessorMock{
		ProcessTransactionCalled: txProcess,
		SetBalancesToTrieCalled:  setBalances,
	}

	be, _ := blproc.NewBlockProcessor(
		tdp, nil,
		nil,
		&txProc,
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	createGenesis := func() {
		be.CreateGenesisBlock(nil)
	}

	assert.Panics(t, createGenesis)
}

func TestBlockProcessor_CreateGenesisBlockBodyOK(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return nil
	}

	setBalances := func(accBalance map[string]*big.Int) (rootHash []byte, err error) {
		return []byte("stateRootHash"), nil
	}

	txProc := mock.TxProcessorMock{
		ProcessTransactionCalled: txProcess,
		SetBalancesToTrieCalled:  setBalances,
	}

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&txProc,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	rootHash, err := be.CreateGenesisBlock(nil)
	assert.Nil(t, err)
	assert.NotNil(t, rootHash)
	assert.Equal(t, rootHash, []byte("stateRootHash"))
}

func TestBlockProcessor_RemoveBlockTxsFromPoolNilBlockShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	err := be.RemoveBlockTxsFromPool(nil)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestBlockProcessor_RemoveBlockTxsFromPoolOK(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)
	body := make(block.Body, 0)

	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}

	body = append(body, &miniblock)

	err := be.RemoveBlockTxsFromPool(body)

	assert.Nil(t, err)
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	marshalizer := &mock.MarshalizerStub{}

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	hdr, txBlock := createTestHdrTxBlockBody()

	expectedError := errors.New("marshalizer fail")

	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return nil, expectedError
		}

		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}

	_, err := be.ComputeHeaderHash(hdr)

	assert.Equal(t, expectedError, err)
}

func TestNode_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}

	be, _ := blproc.NewBlockProcessor(
		tdp, hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
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
		Commitment:    hasher.Compute("commitment"),
		Signature:     hasher.Compute("signature"),
		RootHash:      hasher.Compute("root hash"),
	}

	txBlock := block.Body{
		{
			ShardID: 0,
			TxHashes: [][]byte{
				hasher.Compute("txHash_0_1"),
				hasher.Compute("txHash_0_2"),
			},
		},
		{
			ShardID: 1,
			TxHashes: [][]byte{
				hasher.Compute("txHash_1_1"),
				hasher.Compute("txHash_1_2"),
			},
		},
		{
			ShardID: 2,
			TxHashes: [][]byte{
				hasher.Compute("txHash_2_1"),
			},
		},
		{
			ShardID:  3,
			TxHashes: make([][]byte, 0),
		},
	}

	return hdr, txBlock
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_DisplayLogInfo(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	hasher := mock.HasherMock{}
	hdr, txBlock := createTestHdrTxBlockBody()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	hdr.PrevHash = hasher.Compute("prev hash")

	be.DisplayLogInfo(hdr, txBlock, hasher.Compute("header hash"))
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

	cacher, _ := storage.NewCache(storage.LRUCache, 100)
	transactions, txHashes, err := blproc.SortTxByNonce(cacher)

	assert.Equal(t, 0, len(transactions))
	assert.Equal(t, 0, len(txHashes))
	assert.Nil(t, err)
}

func TestSortTxByNonce_OneTxShouldWork(t *testing.T) {
	t.Parallel()

	cacher, _ := storage.NewCache(storage.LRUCache, 100)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

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
	tx := &transaction.Transaction{
		Nonce: rand.Uint64(),
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

	cache, _ := storage.NewCache(storage.LRUCache, uint32(len(transactions)))

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
	cacher, _ := storage.NewCache(storage.LRUCache, uint32(noOfTx))

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

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

func TestBlockProcessor_CheckBlockValidity(t *testing.T) {
	t.Parallel()

	tdp := initDataPool()

	bp, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	blkc := createTestBlockchain()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = 0

	hdr.PrevHash = []byte("X")

	r := bp.CheckBlockValidity(blkc, hdr)
	assert.False(t, r)

	hdr.PrevHash = []byte("")

	r = bp.CheckBlockValidity(blkc, hdr)
	assert.True(t, r)

	hdr.Nonce = 2

	r = bp.CheckBlockValidity(blkc, hdr)
	assert.False(t, r)

	hdr.Nonce = 1
	blkc.CurrentBlockHeader = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = 0

	r = bp.CheckBlockValidity(blkc, hdr)
	assert.False(t, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = bp.CheckBlockValidity(blkc, hdr)
	assert.False(t, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = bp.CheckBlockValidity(blkc, hdr)
	assert.False(t, r)

	hdr.Nonce = 2

	marshalizerMock := mock.MarshalizerMock{}
	hasherMock := mock.HasherMock{}

	prevHeader, _ := marshalizerMock.Marshal(blkc.CurrentBlockHeader)
	hdr.PrevHash = hasherMock.Compute(string(prevHeader))

	r = bp.CheckBlockValidity(blkc, hdr)
	assert.True(t, r)
}

func TestBlockProcessor_CreateMiniBlockHeadersShouldNotReturnNil(t *testing.T) {
	t.Parallel()

	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	mbHeaders, err := bp.CreateMiniBlockHeaders(nil)
	assert.Nil(t, err)
	assert.NotNil(t, mbHeaders)
	assert.Equal(t, 0, len(mbHeaders))
}

func TestBlockProcessor_CreateMiniBlockHeadersShouldErrWhenMarshalizerErrors(t *testing.T) {
	t.Parallel()

	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{Fail:true},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	body := block.Body{
		{ShardID: 1, TxHashes: make([][]byte, 0)},
		{ShardID: 2, TxHashes: make([][]byte, 0)},
		{ShardID: 3, TxHashes: make([][]byte, 0)},
	}
	mbHeaders, err := bp.CreateMiniBlockHeaders(body)
	assert.NotNil(t, err)
	assert.Nil(t, mbHeaders)
}

func TestBlockProcessor_CreateMiniBlockHeadersReturnsOK(t *testing.T) {
	t.Parallel()

	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
	)

	body := block.Body{
		{ShardID: 1, TxHashes: make([][]byte, 0)},
		{ShardID: 2, TxHashes: make([][]byte, 0)},
		{ShardID: 3, TxHashes: make([][]byte, 0)},
	}
	mbHeaders, err := bp.CreateMiniBlockHeaders(body)
	assert.Nil(t, err)
	assert.Equal(t, len(body), len(mbHeaders))
}