package block_test

import (
	"math/big"
	"testing"

	"reflect"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var testCacherConfig = storage.CacheConfig{
	Size: 1000,
	Type: storage.LRUCache,
}

func createBlockchain() (*blockchain.BlockChain, error) {
	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}
	persisterTxBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxBlockBodyStorage"}
	persisterStateBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "StateBlockBodyStorage"}
	persisterPeerBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "PeerBlockBodyStorage"}
	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockHeaderStorage"}
	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxStorage"}

	var headerUnit, peerBlockUnit, stateBlockUnit, txBlockUnit, txUnit *storage.Unit

	badBlockCache, err := storage.NewCache(
		cacher.Type,
		cacher.Size)

	if err == nil {
		txUnit, err = storage.NewStorageUnitFromConf(
			cacher,
			persisterTxStorage,
			bloom)
	}

	if err == nil {
		txBlockUnit, err = storage.NewStorageUnitFromConf(
			cacher,
			persisterTxBlockBodyStorage,
			bloom)
	}

	if err == nil {
		stateBlockUnit, err = storage.NewStorageUnitFromConf(
			cacher,
			persisterStateBlockBodyStorage,
			bloom)
	}

	if err == nil {
		peerBlockUnit, err = storage.NewStorageUnitFromConf(
			cacher,
			persisterPeerBlockBodyStorage,
			bloom)
	}

	if err == nil {
		headerUnit, err = storage.NewStorageUnitFromConf(
			cacher,
			persisterBlockHeaderStorage,
			bloom)
	}

	if err == nil {
		blockChain, err := blockchain.NewBlockChain(
			badBlockCache,
			txUnit,
			txBlockUnit,
			stateBlockUnit,
			peerBlockUnit,
			headerUnit)

		return blockChain, err
	}

	// cleanup
	if err != nil {
		if headerUnit != nil {
			_ = headerUnit.DestroyUnit()
		}
		if peerBlockUnit != nil {
			_ = peerBlockUnit.DestroyUnit()
		}
		if stateBlockUnit != nil {
			_ = stateBlockUnit.DestroyUnit()
		}
		if txBlockUnit != nil {
			_ = txBlockUnit.DestroyUnit()
		}
		if txUnit != nil {
			_ = txUnit.DestroyUnit()
		}
	}
	return nil, err
}

func initDataPool() data.TransientDataHolder {
	tdp := &mock.TransientDataPoolStub{
		TransactionsCalled: func() data.ShardedDataCacherNotifier {
			return &mock.ShardedDataCacheNotifierStub{
				RegisterHandlerStub: func(i func(key []byte)) {},
				ShardDataStoreStub: func(shardID uint32) (c storage.Cacher) {
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
					}
				},
				RemoveSetOfDataFromPoolStub: func(keys [][]byte, destShardID uint32) {},
			}
		}}

	return tdp
}

func TestNewBlockProcessor(t *testing.T) {
	tdp := initDataPool()

	be, err := blproc.NewBlockProcessor(
		tdp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, be)
}

func TestBlockProc_GetTransactionFromPool(t *testing.T) {
	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock())

	txHash := []byte("tx1_hash")
	tx := be.GetTransactionFromPool(1, txHash)

	assert.NotNil(t, tx)
	assert.Equal(t, uint64(10), tx.Nonce)
}

func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock())

	shardId := uint32(1)
	txHash1 := []byte("tx1_hash1")

	blk := block.TxBlockBody{}
	mBlocks := make([]block.MiniBlock, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	mBlk := block.MiniBlock{ShardID: shardId, TxHashes: txHashes}
	mBlocks = append(mBlocks, mBlk)
	blk.MiniBlocks = mBlocks

	//TODO refactor the test

	if be.RequestTransactionFromNetwork(&blk) > 0 {
		be.WaitForTxHashes()
	}
}

func TestBlockProcessor_ProcessBlockWithNilTxBlockBodyShouldErr(t *testing.T) {
	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	blkc, _ := createBlockchain()

	hdr := block.Header{
		Nonce:         0,
		BlockBodyHash: []byte("bodyHash"),
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
	}

	// cleanup after tests
	defer func() {
		_ = blkc.Destroy()
	}()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot},
		mock.NewOneShardCoordinatorMock(),
	)

	// should return err
	err := be.ProcessAndCommit(blkc, &hdr, nil)

	assert.NotNil(t, err)
	assert.Equal(t, process.ErrNilTxBlockBody, err)
}

func TestBlockProc_ProcessBlockWithDirtyAccountShouldErr(t *testing.T) {
	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	// set accounts dirty
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	blkc, _ := createBlockchain()

	hdr := block.Header{
		Nonce:         0,
		BlockBodyHash: []byte("bodyHash"),
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		Commitment:    []byte("commitment"),
	}

	miniblocks := make([]block.MiniBlock, 0)

	txBody := block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{
			RootHash: []byte("root hash"),
			ShardID:  0,
		},
		MiniBlocks: miniblocks,
	}

	// cleanup after tests
	defer func() {
		_ = blkc.Destroy()
	}()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
	)

	// should return err
	err := be.ProcessAndCommit(blkc, &hdr, &txBody)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestBlockProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
	tdp := initDataPool()

	txHash := []byte("tx_hash1")

	// invalid transaction
	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return process.ErrHigherNonceInTransaction
	}

	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	blkc, _ := createBlockchain()
	hdr := block.Header{
		Nonce:         0,
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("00110"),
		BlockBodyHash: []byte("bodyHash"),
		ShardId:       0,
		Commitment:    []byte("commitment"),
	}
	miniblocks := make([]block.MiniBlock, 0)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}
	miniblocks = append(miniblocks, miniblock)

	txBody := block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{
			RootHash: []byte("root hash"),
			ShardID:  0,
		},
		MiniBlocks: miniblocks,
	}

	// cleanup after tests
	defer func() {
		_ = blkc.Destroy()
	}()

	// set accounts dirty
	journalLen := func() int { return 0 }
	// set revertToSnapshot
	revertToSnapshot := func(snapshot int) error { return nil }

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
	)

	// should return err
	err := be.ProcessAndCommit(blkc, &hdr, &txBody)
	assert.Equal(t, process.ErrHigherNonceInTransaction, err)
}

func TestBlockProc_CreateTxBlockBodyWithDirtyAccStateShouldErr(t *testing.T) {
	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 3 }
	revToSnapshot := func(snapshot int) error { return nil }

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,

		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
	)

	bl, err := be.CreateTxBlockBody(0, 100, 0, func() bool { return true })

	// nil block
	assert.Nil(t, bl)
	// error
	assert.Equal(t, process.ErrAccountStateDirty, err)
}

func TestBlockProcessor_CreateTxBlockBodyWithNoTimeShouldEmptyBlock(t *testing.T) {
	tdp := initDataPool()

	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 0 }
	rootHashfunc := func() []byte { return []byte("roothash") }
	revToSnapshot := func(snapshot int) error { return nil }

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RootHashCalled:         rootHashfunc,
			RevertToSnapshotCalled: revToSnapshot,
		},
		mock.NewOneShardCoordinatorMock(),
	)

	haveTime := func() bool {
		return false
	}

	bl, err := be.CreateTxBlockBody(0, 100, 0, haveTime)

	// no error
	assert.Nil(t, err)
	// no miniblocks
	assert.Equal(t, len(bl.MiniBlocks), 0)
}

func TestBlockProcessor_CreateTxBlockBodyOK(t *testing.T) {
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
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled: journalLen,
			RootHashCalled:   rootHashfunc,
		},
		mock.NewOneShardCoordinatorMock(),
	)

	blk, err := be.CreateTxBlockBody(0, 100, 0, haveTime)

	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithNilTxProcessorShouldPanic(t *testing.T) {
	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, nil,
		nil,
		nil,
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	createGenesis := func() {
		be.CreateGenesisBlockBody(nil, 0)
	}

	assert.Panics(t, createGenesis)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithFailSetBalanceShouldPanic(t *testing.T) {
	tdp := initDataPool()

	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return nil
	}

	setBalances := func(accBalance map[string]big.Int) (rootHash []byte, err error) {
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
	)

	createGenesis := func() {
		be.CreateGenesisBlockBody(nil, 0)
	}

	assert.Panics(t, createGenesis)
}

func TestBlockProcessor_CreateGenesisBlockBodyOK(t *testing.T) {
	tdp := initDataPool()

	txProcess := func(transaction *transaction.Transaction, round int32) error {
		return nil
	}

	setBalances := func(accBalance map[string]big.Int) (rootHash []byte, err error) {
		return []byte("stateRootHash"), nil
	}

	txProc := mock.TxProcessorMock{
		ProcessTransactionCalled: txProcess,
		SetBalancesToTrieCalled:  setBalances,
	}

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&txProc,
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	stBlock := be.CreateGenesisBlockBody(nil, 0)
	assert.NotNil(t, stBlock)
	assert.Equal(t, stBlock.RootHash, []byte("stateRootHash"))
}

func TestBlockProcessor_RemoveBlockTxsFromPoolNilBlockShouldErr(t *testing.T) {
	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := be.RemoveBlockTxsFromPool(nil)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilTxBlockBody)
}

func TestBlockProcessor_RemoveBlockTxsFromPoolOK(t *testing.T) {
	tdp := initDataPool()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
	)
	miniblocks := make([]block.MiniBlock, 0)

	txHash := []byte("txHash")
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)

	miniblock := block.MiniBlock{
		ShardID:  0,
		TxHashes: txHashes,
	}

	miniblocks = append(miniblocks, miniblock)

	txBody := block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{},
		MiniBlocks:     miniblocks,
	}

	err := be.RemoveBlockTxsFromPool(&txBody)

	assert.Nil(t, err)
}

func createStubBlockchain() *blockchain.BlockChain {
	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.StorerStub{})

	return blkc
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	tdp := initDataPool()

	marshalizer := &mock.MarshalizerMock2{}

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock2{},
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	hdr, txBlock := createTestHdrTxBlockBody()

	expectedError := errors.New("marshalizer fail")

	marshalizer.MarshalHandler = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return nil, expectedError
		}

		if txBlock == obj {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}

	_, err := be.ComputeHeaderHash(hdr)

	assert.Equal(t, expectedError, err)
}

func TestNode_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	tdp := initDataPool()

	sposWrk := &spos.SPOSConsensusWorker{}
	sposWrk.BlockChain = createStubBlockchain()

	marshalizer := &mock.MarshalizerMock2{}
	hasher := &mock.HasherMock2{}

	be, _ := blproc.NewBlockProcessor(
		tdp, hasher,
		marshalizer,
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	hdr, txBlock := createTestHdrTxBlockBody()

	marshalizer.MarshalHandler = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return []byte("hdrHeaderMarshalized"), nil
		}
		if txBlock == obj {
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

func createTestHdrTxBlockBody() (*block.Header, *block.TxBlockBody) {
	hasher := mock.HasherFake{}

	hdr := &block.Header{
		Nonce:         1,
		ShardId:       2,
		Epoch:         3,
		Round:         4,
		TimeStamp:     uint64(11223344),
		PrevHash:      hasher.Compute("prev hash"),
		BlockBodyHash: hasher.Compute("tx block body hash"),
		PubKeysBitmap: []byte{255, 0, 128},
		Commitment:    hasher.Compute("commitment"),
		Signature:     hasher.Compute("signature"),
	}

	txBlock := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{
			RootHash: hasher.Compute("root hash"),
		},
		MiniBlocks: []block.MiniBlock{
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
		},
	}

	return hdr, txBlock
}

//------- ComputeNewNoncePrevHash

func TestBlockProcessor_DisplayLogInfo(t *testing.T) {
	tdp := initDataPool()

	hasher := mock.HasherFake{}
	hdr, txBlock := createTestHdrTxBlockBody()

	be, _ := blproc.NewBlockProcessor(
		tdp, &mock.HasherMock2{},
		&mock.MarshalizerMock2{},
		&mock.TxProcessorMock{},
		&mock.AccountsStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	hdr.PrevHash = hasher.Compute("prev hash")
	hdr.BlockBodyHash = hasher.Compute("block hash")

	be.DisplayLogInfo(hdr, txBlock, hasher.Compute("header hash"))
}
