package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"math/big"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

var r *rand.Rand
var mutex sync.Mutex

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func haveTime() time.Duration {
	return time.Duration(2000 * time.Millisecond)
}

func createTestBlockchain() *mock.BlockChainMock {
	return &mock.BlockChainMock{}
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

func initDataPool() *mock.PoolsHolderStub {
	tdp := &mock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{
				RegisterHandlerCalled: func(i func(key []byte)) {},
				ShardDataStoreCalled: func(id string) (c storage.Cacher) {
					return &mock.CacherStub{
						PeekCalled: func(key []byte) (value interface{}, ok bool) {
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
				RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
			}
		},
		HeadersNoncesCalled: func() dataRetriever.Uint64Cacher {
			return &mock.Uint64CacherStub{
				PutCalled: func(u uint64, i []byte) bool {
					return true
				},
			}
		},
		MetaBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte)) {},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			cs := &mock.CacherStub{}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {
			}
			cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.RegisterHandlerCalled = func(i func(key []byte)) {}
			cs.RemoveCalled = func(key []byte) {}
			return cs
		},
	}
	return tdp
}

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	return store
}

//------- NewBlockProcessor

func TestNewBlockProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()
	be, err := blproc.NewBlockProcessor(
		nil,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, mbHash []byte) {},
	)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilStoreShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		nil,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, mbHash []byte) {},
	)
	assert.Equal(t, process.ErrNilTxStorage, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		nil,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilHasher, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilMarshalizerShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		nil,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilTxProcessorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		nil,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilTxProcessor, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilAccountsAdapterShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilAccountsAdapter, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		nil,
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilForkDetectorShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		nil,
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilForkDetector, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilRequestTransactionHandlerShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		nil,
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilTransactionHandler, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_NilTransactionPoolShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return nil
	}
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Equal(t, process.ErrNilTransactionPool, err)
	assert.Nil(t, be)
}

func TestNewBlockProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.Nil(t, err)
	assert.NotNil(t, be)
}

//------- ProcessBlock

func TestBlockProcessor_ProcessBlockWithNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	blk := make(block.Body, 0)
	err := be.ProcessBlock(nil, &block.Header{}, blk, haveTime)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestBlockProcessor_ProcessBlockWithNilHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	body := make(block.Body, 0)
	err := be.ProcessBlock(createTestBlockchain(), nil, body, haveTime)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestBlockProcessor_ProcessBlockWithNilBlockBodyShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	err := be.ProcessBlock(createTestBlockchain(), &block.Header{}, nil, haveTime)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
}

func TestBlockProcessor_ProcessBlockWithNilHaveTimeFuncShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	blk := make(block.Body, 0)
	err := be.ProcessBlock(createTestBlockchain(), &block.Header{}, blk, nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestBlockProcessor_ProcessWithDirtyAccountShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	tpm := mock.TxProcessorMock{}
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }
	blkc := &blockchain.BlockChain{}
	hdr := block.Header{
		Nonce:         0,
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		RootHash:      []byte("roothash"),
	}
	body := make(block.Body, 0)
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
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
		func(destShardID uint32, txHash []byte) {},
	)
	// should return err
	err := be.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestBlockProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	txHash := []byte("tx_hash1")
	// invalid transaction
	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return process.ErrHigherNonceInTransaction
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	blkc := &blockchain.BlockChain{}
	hdr := block.Header{
		Nonce:         0,
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
	be, _ := blproc.NewBlockProcessor(
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	go func() {
		be.ChRcvAllTxs <- true
	}()
	// should return err
	err := be.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestBlockProcessor_ProcessWithHeaderNotFirstShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
	err := be.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestBlockProcessor_ProcessWithHeaderNotCorrectNonceShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
	err := be.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)
}

func TestBlockProcessor_ProcessWithHeaderNotCorrectPrevHashShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
	err := be.ProcessBlock(blkc, hdr, body, haveTime)
	assert.Equal(t, process.ErrInvalidBlockHash, err)
}

func TestBlockProcessor_ProcessBlockWithErrOnProcessBlockTransactionsCallShouldRevertState(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	txHash := []byte("tx_hash1")
	err := errors.New("process block transaction error")
	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return err
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
		},
	}
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
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() []byte {
		return []byte("rootHash")
	}
	be, _ := blproc.NewBlockProcessor(
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	go func() {
		be.ChRcvAllTxs <- true
	}()
	// should return err
	err2 := be.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, err, err2)
	assert.True(t, wasCalled)
}

func TestBlockProcessor_ProcessBlockWithErrOnVerifyStateRootCallShouldRevertState(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	txHash := []byte("tx_hash1")
	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return nil
	}
	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	blkc := &blockchain.BlockChain{
		CurrentBlockHeader: &block.Header{
			Nonce: 0,
		},
	}
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
	wasCalled := false
	revertToSnapshot := func(snapshot int) error {
		wasCalled = true
		return nil
	}
	rootHashCalled := func() []byte {
		return []byte("rootHashX")
	}
	be, _ := blproc.NewBlockProcessor(
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	go func() {
		be.ChRcvAllTxs <- true
	}()
	// should return err
	err := be.ProcessBlock(blkc, &hdr, body, haveTime)
	assert.Equal(t, process.ErrRootStateMissmatch, err)
	assert.True(t, wasCalled)
}

//------- CommitBlock

func TestBlockProcessor_CommitBlockNilBlockchainShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	accounts := &mock.AccountsStub{}
	accounts.RevertToSnapshotCalled = func(snapshot int) error {
		return nil
	}
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	blk := make(block.Body, 0)
	err := be.CommitBlock(nil, &block.Header{}, blk)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestBlockProcessor_CommitBlockMarshalizerFailForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
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
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	blkc := createTestBlockchain()
	err := be.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errMarshalizer, err)
}

func TestBlockProcessor_CommitBlockStorageFailsForHeaderShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	errPersister := errors.New("failure")
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
	hdrUnit := &mock.StorerStub{
		PutCalled: func(key, data []byte) error {
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.BlockHeaderUnit, hdrUnit)

	be, _ := blproc.NewBlockProcessor(
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
	)
	err := be.CommitBlock(blkc, hdr, body)
	assert.Equal(t, errPersister, err)
}

func TestBlockProcessor_CommitBlockStorageFailsForBodyShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
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
			return errPersister
		},
	}
	store := initStore()
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)

	be, _ := blproc.NewBlockProcessor(
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{
			AddHeaderCalled: func(header *block.Header, hash []byte, isProcessed bool) error {
				return nil
			},
		},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)

	blkc, _ := blockchain.NewBlockChain(
		generateTestCache(),
	)
	err := be.CommitBlock(blkc, hdr, body)

	assert.Equal(t, errPersister, err)
}

func TestBlockProcessor_CommitBlockNilNoncesDataPoolShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
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

	be, _ := blproc.NewBlockProcessor(
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	tdp.HeadersNoncesCalled = func() dataRetriever.Uint64Cacher {
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
		AddHeaderCalled: func(header *block.Header, hash []byte, isProcessed bool) error {
			return nil
		},
	}
	hasher := &mock.HasherStub{}
	hasher.ComputeCalled = func(s string) []byte {
		return hdrHash
	}
	store := initStore()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		store,
		hasher,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		fd,
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
	}
	forkDetectorAddCalled := false
	fd := &mock.ForkDetectorMock{
		AddHeaderCalled: func(header *block.Header, hash []byte, isProcessed bool) error {
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

	be, _ := blproc.NewBlockProcessor(
		tdp,
		store,
		hasher,
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		fd,
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	txCache := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
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
	tdp.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return &mock.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return txCache
			},

			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {
				if bytes.Equal(keys[0], []byte(txHash)) && len(keys) == 1 {
					removeTxWasCalled = true
				}
			},
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if reflect.DeepEqual(key, []byte(txHash)) {
					return &transaction.Transaction{Nonce: 10}, true
				}
				return nil, false
			},
		}

	}
	blkc := createTestBlockchain()
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return hdr
	}
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return hdrHash
	}
	err := be.CommitBlock(blkc, hdr, body)
	assert.Nil(t, err)
	assert.True(t, removeTxWasCalled)
	assert.True(t, forkDetectorAddCalled)
	assert.True(t, blkc.GetCurrentBlockHeader() == hdr)
	assert.Equal(t, hdrHash, blkc.GetCurrentBlockHeaderHash())
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
	store := initStore()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		store,
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		accounts,
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	assert.True(t, be.VerifyStateRoot(rootHash))
}

func TestBlockProc_GetTransactionFromPool(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	txHash := []byte("tx1_hash")
	tx := be.GetTransactionFromPool(1, 1, txHash)
	assert.NotNil(t, tx)
	assert.Equal(t, uint64(10), tx.Nonce)
}

func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	shardId := uint32(1)
	txHash1 := []byte("tx1_hash1")
	body := make(block.Body, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	mBlk := block.MiniBlock{ReceiverShardID: shardId, TxHashes: txHashes}
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
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
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
		func(destShardID uint32, txHash []byte) {},
	)
	bl, err := be.CreateBlockBody(0, func() bool { return true })
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	haveTime := func() bool {
		return false
	}
	bl, err := be.CreateBlockBody(0, haveTime)
	// no error
	assert.Nil(t, err)
	// no miniblocks
	assert.Equal(t, len(bl.(block.Body)), 0)
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
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherStub{},
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
		func(destShardID uint32, txHash []byte) {},
	)
	blk, err := be.CreateBlockBody(0, haveTime)
	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithFailSetBalanceShouldErr(t *testing.T) {
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
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&txProc,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	_, err := be.CreateGenesisBlock(nil)
	assert.Equal(t, process.ErrAccountStateDirty, err)
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
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&txProc,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	err := be.RemoveTxBlockFromPools(nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestBlockProcessor_RemoveBlockTxsFromPoolOK(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	body := make(block.Body, 0)
	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        txHashes,
	}
	body = append(body, &miniblock)
	err := be.RemoveTxBlockFromPools(body)
	assert.Nil(t, err)
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	marshalizer := &mock.MarshalizerStub{}
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
		tdp,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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

func TestBlockProcessor_DisplayLogInfo(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	hasher := mock.HasherMock{}
	hdr, txBlock := createTestHdrTxBlockBody()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
		initStore(),
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	blkc := createTestBlockchain()
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = 0
	hdr.PrevHash = []byte("X")
	r := bp.CheckBlockValidity(blkc, hdr, nil)
	assert.False(t, r)

	hdr.PrevHash = []byte("")
	r = bp.CheckBlockValidity(blkc, hdr, nil)
	assert.True(t, r)

	hdr.Nonce = 2
	r = bp.CheckBlockValidity(blkc, hdr, nil)
	assert.False(t, r)

	hdr.Nonce = 1
	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Nonce: 1}
	}
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = 0
	r = bp.CheckBlockValidity(blkc, hdr, nil)
	assert.False(t, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")
	r = bp.CheckBlockValidity(blkc, hdr, nil)
	assert.False(t, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")
	r = bp.CheckBlockValidity(blkc, hdr, nil)
	assert.False(t, r)

	hdr.Nonce = 2
	marshalizerMock := mock.MarshalizerMock{}
	hasherMock := mock.HasherMock{}
	prevHeader, _ := marshalizerMock.Marshal(blkc.GetCurrentBlockHeader())
	hdr.PrevHash = hasherMock.Compute(string(prevHeader))
	r = bp.CheckBlockValidity(blkc, hdr, nil)
	assert.True(t, r)
}

func TestBlockProcessor_CreateBlockHeaderShouldNotReturnNil(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	mbHeaders, err := bp.CreateBlockHeader(nil)
	assert.Nil(t, err)
	assert.NotNil(t, mbHeaders)
	assert.Equal(t, 0, len(mbHeaders.(*block.Header).MiniBlockHeaders))
}

func TestBlockProcessor_CreateBlockHeaderShouldErrWhenMarshalizerErrors(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{Fail: true},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
	mbHeaders, err := bp.CreateBlockHeader(body)
	assert.NotNil(t, err)
	assert.Nil(t, mbHeaders)
}

func TestBlockProcessor_CreateBlockHeaderReturnsOK(t *testing.T) {
	t.Parallel()
	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
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
	mbHeaders, err := bp.CreateBlockHeader(body)
	assert.Nil(t, err)
	assert.Equal(t, len(body), len(mbHeaders.(*block.Header).MiniBlockHeaders))
}

func TestBlockProcessor_CommitBlockShouldRevertAccountStateWhenErr(t *testing.T) {
	t.Parallel()
	// set accounts dirty
	journalEntries := 3
	revToSnapshot := func(snapshot int) error {
		journalEntries = 0
		return nil
	}
	bp, _ := blproc.NewBlockProcessor(
		initDataPool(),
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	err := bp.CommitBlock(nil, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, 0, journalEntries)
}

func TestBlockProcessor_MarshalizedDataForCrossShardShouldWork(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
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
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, txHash []byte) {},
	)
	msh, mstx, err := be.MarshalizedDataForCrossShard(body)
	assert.Nil(t, err)
	assert.NotNil(t, msh)
	assert.NotNil(t, mstx)
	_, found := msh[0]
	assert.False(t, found)

	expectedBody := make(block.Body, 0)
	err = marshalizer.Unmarshal(&expectedBody, msh[1])
	assert.Nil(t, err)
	assert.Equal(t, 2, len(expectedBody))
	assert.Equal(t, &mb1, expectedBody[0])
	assert.Equal(t, &mb1, expectedBody[1])
}

type wrongBody struct {
}

func (wr wrongBody) IntegrityAndValidity() error {
	return nil
}

func TestBlockProcessor_MarshalizedDataWrongType(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	wr := wrongBody{}
	msh, mstx, err := be.MarshalizedDataForCrossShard(wr)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
	assert.Nil(t, msh)
	assert.Nil(t, mstx)
}

func TestBlockProcessor_MarshalizedDataNilInput(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	marshalizer := &mock.MarshalizerMock{
		Fail: false,
	}
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	msh, mstx, err := be.MarshalizedDataForCrossShard(nil)
	assert.Equal(t, process.ErrNilMiniBlocks, err)
	assert.Nil(t, msh)
	assert.Nil(t, mstx)
}

func TestBlockProcessor_MarshalizedDataMarshalWithoutSuccess(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
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
			return nil, process.ErrMarshalWithoutSuccess
		},
	}
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	msh, mstx, err := be.MarshalizedDataForCrossShard(body)
	assert.Equal(t, process.ErrMarshalWithoutSuccess, err)
	assert.Nil(t, msh)
	assert.Nil(t, mstx)
}

//------- GetAllTxsFromMiniBlock

func TestBlockProcessor_GetAllTxsFromMiniBlockShouldWork(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()
	senderShardId := uint32(0)
	destinationShardId := uint32(1)

	transactions := []*transaction.Transaction{
		{Nonce: 1},
		{Nonce: 2},
		{Nonce: 3},
	}
	transactionsHashes := make([][]byte, len(transactions))

	//add defined transactions to sender-destination cacher
	for idx, tx := range transactions {
		transactionsHashes[idx] = computeHash(tx, marshalizer, hasher)

		dataPool.Transactions().AddData(
			transactionsHashes[idx],
			tx,
			process.ShardCacherIdentifier(senderShardId, destinationShardId),
		)
	}

	//add some random data
	txRandom := &transaction.Transaction{Nonce: 4}
	dataPool.Transactions().AddData(
		computeHash(txRandom, marshalizer, hasher),
		txRandom,
		process.ShardCacherIdentifier(3, 4),
	)

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, txHash []byte) {},
	)

	mb := &block.MiniBlock{
		SenderShardID:   senderShardId,
		ReceiverShardID: destinationShardId,
		TxHashes:        transactionsHashes,
	}

	txsRetrieved, txHashesRetrieved, err := bp.GetAllTxsFromMiniBlock(mb, func() bool {
		return true
	})

	assert.Nil(t, err)
	assert.Equal(t, len(transactions), len(txsRetrieved))
	assert.Equal(t, len(transactions), len(txHashesRetrieved))
	for idx, tx := range transactions {
		//txReceived should be all txs in the same order
		assert.Equal(t, txsRetrieved[idx], tx)
		//verify corresponding transaction hashes
		assert.Equal(t, txHashesRetrieved[idx], computeHash(tx, marshalizer, hasher))
	}

}

func computeHash(data interface{}, marshalizer marshal.Marshalizer, hasher hashing.Hasher) []byte {
	buff, _ := marshalizer.Marshal(data)
	return hasher.Compute(string(buff))
}

//------- receivedTransaction

func TestBlockProcessor_ReceivedTransactionShouldEraseRequested(t *testing.T) {
	t.Parallel()

	hasher := mock.HasherMock{}
	marshalizer := &mock.MarshalizerMock{}
	dataPool := mock.NewPoolsHolderFake()

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, miniblockHash []byte) {},
	)

	//add 3 tx hashes on requested list
	txHash1 := []byte("tx hash 1")
	txHash2 := []byte("tx hash 2")
	txHash3 := []byte("tx hash 3")

	bp.AddTxHashToRequestedList(txHash1)
	bp.AddTxHashToRequestedList(txHash2)
	bp.AddTxHashToRequestedList(txHash3)

	//received txHash2
	bp.ReceivedTransaction(txHash2)

	assert.True(t, bp.IsTxHashRequested(txHash1))
	assert.False(t, bp.IsTxHashRequested(txHash2))
	assert.True(t, bp.IsTxHashRequested(txHash3))
}

//------- receivedMiniBlock

func TestBlockProcessor_ReceivedMiniBlockShouldRequestMissingTransactions(t *testing.T) {
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

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
			if bytes.Equal(txHash1, txHash) {
				atomic.AddInt32(&txHash1Requested, 1)
			}
			if bytes.Equal(txHash2, txHash) {
				atomic.AddInt32(&txHash2Requested, 1)
			}
			if bytes.Equal(txHash3, txHash) {
				atomic.AddInt32(&txHash3Requested, 1)
			}
		},
		func(destShardID uint32, miniblockHash []byte) {},
	)

	bp.ReceivedMiniBlock(miniBlockHash)

	//we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&txHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&txHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&txHash2Requested))
}

//------- receivedMetaBlock

func TestBlockProcessor_ReceivedMetaBlockShouldRequestMissingMiniBlocks(t *testing.T) {
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

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, miniblockHash []byte) {
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
	)

	bp.ReceivedMetaBlock(metaBlockHash)

	//we have to wait to be sure txHash1Requested is not incremented by a late call
	time.Sleep(time.Second)

	assert.Equal(t, int32(0), atomic.LoadInt32(&miniBlockHash1Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
	assert.Equal(t, int32(1), atomic.LoadInt32(&miniBlockHash2Requested))
}

//------- processMiniBlockComplete

func TestBlockProcessor_ProcessMiniBlockCompleteWithOkTxsShouldExecuteThemAndNotRevertAccntState(t *testing.T) {
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

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round int32) error {
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

				return nil
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, miniblockHash []byte) {},
	)

	err := bp.ProcessMiniBlockComplete(&miniBlock, 0, func() bool {
		return true
	})

	assert.Nil(t, err)
	assert.Equal(t, tx1Nonce, tx1ExecutionResult)
	assert.Equal(t, tx2Nonce, tx2ExecutionResult)
	assert.Equal(t, tx3Nonce, tx3ExecutionResult)
}

func TestBlockProcessor_ProcessMiniBlockCompleteWithErrorWhileProcessShouldCallRevertAccntState(t *testing.T) {
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

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round int32) error {
				if bytes.Equal(transaction.Data, txHash2) {
					return errTxProcessor
				}

				return nil
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, miniblockHash []byte) {},
	)

	err := bp.ProcessMiniBlockComplete(&miniBlock, 0, func() bool {
		return true
	})

	assert.Equal(t, errTxProcessor, err)
	assert.True(t, revertAccntStateCalled)
}

//------- createMiniBlocks

func TestBlockProcessor_CreateMiniBlocksShouldWorkWithIntraShardTxs(t *testing.T) {
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

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		initStore(),
		hasher,
		marshalizer,
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction, round int32) error {
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

				return nil
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
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, miniblockHash []byte) {},
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

func isInTxHashes(searched []byte, list [][]byte) bool {
	for _, txHash := range list {
		if bytes.Equal(txHash, searched) {
			return true
		}
	}
	return false
}

//------- removeMetaBlockFromPool

func TestBlockProcessor_RemoveMetaBlockFromPoolShouldWork(t *testing.T) {
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

	wasCalledPut := false

	blkc := &mock.BlockChainMock{}
	store := &mock.ChainStorerMock{
		PutCalled: func(unitType dataRetriever.UnitType, key []byte, value []byte) error {
			if bytes.Equal(key, mb1Hash) {
				wasCalledPut = true
			}

			return nil
		},
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = destShardId
	shardCoordinator.SetNoShards(destShardId + 1)

	bp, _ := blproc.NewBlockProcessor(
		dataPool,
		store,
		hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		shardCoordinator,
		&mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		func(destShardID uint32, txHash []byte) {},
		func(destShardID uint32, miniblockHash []byte) {},
	)

	//create block body with first 3 miniblocks from miniblocks var
	blockBody := block.Body{miniblocks[0], miniblocks[1], miniblocks[2]}

	err := bp.RemoveMetaBlockFromPool(blockBody, blkc)

	assert.Nil(t, err)
	assert.True(t, wasCalledPut)
	//check WasMiniBlockProcessed for remaining metablocks
	metaBlock2Recov, _ := dataPool.MetaBlocks().Get(mb2Hash)
	assert.True(t, (metaBlock2Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[2]))
	assert.False(t, (metaBlock2Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[3]))

	metaBlock3Recov, _ := dataPool.MetaBlocks().Get(mb3Hash)
	assert.False(t, (metaBlock3Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[4]))
	assert.False(t, (metaBlock3Recov.(data.HeaderHandler)).GetMiniBlockProcessed(miniblockHashes[5]))
}

func createDummyMiniBlock(
	txHash string,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	destShardId uint32,
	senderShardId uint32) (*block.MiniBlock, []byte) {

	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte(txHash)},
		ReceiverShardID: destShardId,
		SenderShardID:   senderShardId,
	}

	buff, _ := marshalizer.Marshal(miniblock)
	hash := hasher.Compute(string(buff))

	return miniblock, hash
}

func createDummyMetaBlock(destShardId uint32, senderShardId uint32, miniBlockHashes ...[]byte) data.HeaderHandler {
	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardMiniBlockHeaders: make([]block.ShardMiniBlockHeader, len(miniBlockHashes)),
			},
		},
	}

	for idx, mbHash := range miniBlockHashes {
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].ReceiverShardId = destShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].SenderShardId = senderShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].Hash = mbHash
	}

	return metaBlock
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilBlockChain(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)
	err := be.RestoreBlockIntoPools(nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilBlockChain)
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldErrNilTxBlockBody(t *testing.T) {
	t.Parallel()
	tdp := initDataPool()
	be, _ := blproc.NewBlockProcessor(
		tdp,
		initStore(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {},
	)

	blkc := &mock.BlockChainMock{}

	err := be.RestoreBlockIntoPools(blkc, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestBlockProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
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

	be, _ := blproc.NewBlockProcessor(
		dataPool,
		store,
		hasherMock,
		marshalizerMock,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {
		},
		func(destShardID uint32, txHash []byte) {
		},
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

	blkc := &mock.BlockChainMock{}

	metablockHash := []byte("meta block hash 1")
	metablockHeader := createDummyMetaBlock(0, 1, miniblockHash)
	metablockHeader.SetMiniBlockProcessed(metablockHash, true)
	dataPool.MetaBlocks().Put(
		metablockHash,
		metablockHeader,
	)

	err := be.RestoreBlockIntoPools(blkc, body)

	miniblockFromPool, _ := dataPool.MiniBlocks().Get(miniblockHash)
	txFromPool, _ := dataPool.Transactions().SearchFirstData(txHash)
	metablockFromPool, _ := dataPool.MetaBlocks().Get(metablockHash)
	metablock := metablockFromPool.(*block.MetaBlock)
	assert.Nil(t, err)
	assert.Equal(t, &miniblock, miniblockFromPool)
	assert.Equal(t, &tx, txFromPool)
	assert.Equal(t, false, metablock.GetMiniBlockProcessed(miniblockHash))
}
