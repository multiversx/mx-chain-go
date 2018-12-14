package block_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	blproc "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

func blockchainConfig() *blockchain.Config {
	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}
	persisterTxBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxBlockBodyStorage"}
	persisterStateBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "StateBlockBodyStorage"}
	persisterPeerBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "PeerBlockBodyStorage"}
	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockHeaderStorage"}
	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxStorage"}
	return &blockchain.Config{
		TxBlockBodyStorage:    storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxBlockBodyStorage, BloomConf: bloom},
		StateBlockBodyStorage: storage.UnitConfig{CacheConf: cacher, DBConf: persisterStateBlockBodyStorage, BloomConf: bloom},
		PeerBlockBodyStorage:  storage.UnitConfig{CacheConf: cacher, DBConf: persisterPeerBlockBodyStorage, BloomConf: bloom},
		BlockHeaderStorage:    storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockHeaderStorage, BloomConf: bloom},
		TxStorage:             storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxStorage, BloomConf: bloom},
		TxPoolStorage:         cacher,
		TxBadBlockBodyCache:   cacher,
	}
}

func TestNewBlockProcessor(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		1)

	assert.NotNil(t, be)
}

func TestBlockProc_GetTransactionFromPool(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		1)

	txHash := []byte("tx1_hash")

	tx := be.GetTransactionFromPool(1, txHash)
	assert.Nil(t, tp.MiniPoolTxStore(1))
	assert.Nil(t, tx)

	tp.NewMiniPool(1)

	tx = be.GetTransactionFromPool(1, txHash)
	assert.NotNil(t, tp.MiniPoolTxStore(1))
	assert.Nil(t, tx)

	tp.AddTransaction(txHash, &transaction.Transaction{Nonce: uint64(1)}, 1)

	tx = be.GetTransactionFromPool(1, txHash)
	assert.NotNil(t, tp.MiniPoolTxStore(1))
	assert.NotNil(t, tx)
	assert.Equal(t, uint64(1), tx.Nonce)
}

func TestBlockProc_RequestTransactionFromNetwork(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil,
		1)
	//1, []byte("tx1_hash1"), WaitTime

	shardId := uint32(1)
	txHash1 := []byte("tx1_hash1")

	blk := block.TxBlockBody{}
	mBlocks := make([]block.MiniBlock, 0)
	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash1)
	mBlk := block.MiniBlock{ShardID: shardId, TxHashes: txHashes}
	mBlocks = append(mBlocks, mBlk)
	blk.MiniBlocks = mBlocks
	tx1 := &transaction.Transaction{Nonce: 7}
	tp.AddTransaction(txHash1, tx1, 1)

	be.RequestTransactionFromNetwork(&blk)
	be.WaitForTxHashes()
	tx, _ := tp.MiniPool(shardId).TxStore.Get(txHash1)

	assert.Equal(t, tx1, tx)
}

func TestBlockProcessor_ProcessBlockWithNilTxBlockBodyShouldErr(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)
	tpm := mock.TxProcessorMock{}
	// set accounts dirty
	JournalLen := func() int { return 3 }

	blkc, _ := blockchain.NewBlockChain(blockchainConfig())

	hdr := block.Header{
		Nonce:         0,
		BlockBodyHash: []byte("bodyHash"),
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
	}

	// cleanup after tests
	defer func() {
		_ = blkc.Destroy()
	}()

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{JournalLenCalled: JournalLen},
		1,
	)

	// should return err
	err := be.ProcessBlock(blkc, &hdr, nil)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrInvalidTxBlockBody)
}

func TestBlockProc_ProcessBlockWithDirtyAccountShouldErr(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)
	tpm := mock.TxProcessorMock{}
	// set accounts dirty
	JournalLen := func() int { return 3 }

	blkc, _ := blockchain.NewBlockChain(blockchainConfig())

	hdr := block.Header{
		Nonce:         0,
		BlockBodyHash: []byte("bodyHash"),
		PubKeysBitmap: []byte("0100101"),
		PrevHash:      []byte(""),
		Signature:     []byte("signature"),
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

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{JournalLenCalled: JournalLen},
		1,
	)

	// should return err
	err := be.ProcessBlock(blkc, &hdr, &txBody)
	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrAccountStateDirty)
}

func TestBlockProcessor_ProcessBlockWithInvalidTransactionShouldErr(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)
	txHash := []byte("tx_hash1")
	tp.AddTransaction(txHash, &transaction.Transaction{Nonce: 1}, 0)

	// invalid transaction
	txProcess := func(transaction *transaction.Transaction) error {
		return process.ErrHigherNonceInTransaction
	}

	tpm := mock.TxProcessorMock{ProcessTransactionCalled: txProcess}
	blkc, _ := blockchain.NewBlockChain(blockchainConfig())
	hdr := block.Header{Nonce: 0, PrevHash: []byte(""), Signature: []byte("signature")}
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

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled:       journalLen,
			RevertToSnapshotCalled: revertToSnapshot,
		},
		1,
	)

	// should return err
	err := be.ProcessBlock(blkc, &hdr, &txBody)
	assert.NotNil(t, err)
}

func TestBlockProc_CreateTxBlockBodyWithDirtyAccStateShouldErr(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)
	tpm := mock.TxProcessorMock{}
	JournalLen := func() int { return 3 }

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{JournalLenCalled: JournalLen},
		1,
	)

	bl, err := be.CreateTxBlockBody(0, 100, func() bool { return true })

	// nil block
	assert.Nil(t, bl)
	// error
	assert.NotNil(t, err)
}

func TestBlockProcessor_CreateTxBlockBodyWithNoTimeShouldEmptyBlock(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	tpm := mock.TxProcessorMock{}
	journalLen := func() int { return 0 }
	rootHashfunc := func() []byte { return []byte("roothash") }

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled: journalLen,
			RootHashCalled:   rootHashfunc,
		},
		1,
	)

	haveTime := func() bool {
		return false
	}

	bl, err := be.CreateTxBlockBody(0, 100, haveTime)

	// no error
	assert.Nil(t, err)
	// no miniblocks
	assert.Equal(t, len(bl.MiniBlocks), 0)
}

func TestBlockProcessor_CreateTxBlockBodydOK(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)
	tp.AddTransaction([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0)
	tp.AddTransaction([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 1)
	tp.AddTransaction([]byte("tx_hash3"), &transaction.Transaction{Nonce: 3}, 2)
	tp.AddTransaction([]byte("tx_hash4"), &transaction.Transaction{Nonce: 4}, 3)
	tp.AddTransaction([]byte("tx_hash5"), &transaction.Transaction{Nonce: 5}, 2)
	tp.AddTransaction([]byte("tx_hash6"), &transaction.Transaction{Nonce: 6}, 1)

	//process transaction. return nil for no error
	procTx := func(transaction *transaction.Transaction) error {
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

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		&mock.AccountsStub{
			JournalLenCalled: journalLen,
			RootHashCalled:   rootHashfunc,
		},
		1,
	)

	blk, err := be.CreateTxBlockBody(0, 100, haveTime)

	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithNilTxProcessorShouldPanic(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, nil,
		nil,
		nil,
		nil,
		1,
	)

	createGenesis := func() {
		be.CreateGenesisBlockBody(nil, 0)
	}

	assert.Panics(t, createGenesis)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithFailSetBalanceShouldPanic(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	txProcess := func(transaction *transaction.Transaction) error {
		return nil
	}

	setBalances := func(accBalance map[string]big.Int) (rootHash []byte, err error) {
		return nil, process.ErrAccountStateDirty
	}

	txProc := mock.TxProcessorMock{
		ProcessTransactionCalled: txProcess,
		SetBalancesToTrieCalled:  setBalances,
	}

	be := blproc.NewBlockProcessor(
		tp, nil,
		nil,
		&txProc,
		nil,
		1,
	)

	createGenesis := func() {
		be.CreateGenesisBlockBody(nil, 0)
	}

	assert.Panics(t, createGenesis)
}

func TestBlockProcessor_CreateGenesisBlockBodyOK(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	txProcess := func(transaction *transaction.Transaction) error {
		return nil
	}

	setBalances := func(accBalance map[string]big.Int) (rootHash []byte, err error) {
		return []byte("stateRootHash"), nil
	}

	txProc := mock.TxProcessorMock{
		ProcessTransactionCalled: txProcess,
		SetBalancesToTrieCalled:  setBalances,
	}

	be := blproc.NewBlockProcessor(
		tp, nil,
		nil,
		&txProc,
		nil,
		1,
	)

	stBlock := be.CreateGenesisBlockBody(nil, 0)
	assert.NotNil(t, stBlock)
	assert.Equal(t, stBlock.RootHash, []byte("stateRootHash"))
}

func TestBlockProcessor_RemoveBlockTxsFromPoolNilBlockShouldErr(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, nil,
		nil,
		nil,
		nil,
		1,
	)

	err := be.RemoveBlockTxsFromPool(nil)

	assert.NotNil(t, err)
	assert.Equal(t, err, process.ErrNilBlockBody)
}

func TestBlockProcessor_RemoveBlockTxsFromPoolOK(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, nil,
		nil,
		nil,
		nil,
		1,
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

func TestBlockProcessor_GetNbShards(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)
	be := blproc.NewBlockProcessor(
		tp, nil,
		nil,
		nil,
		nil,
		1,
	)

	be.SetNbShards(20)
	nb := be.GetNbShards()

	assert.Equal(t, uint32(20), nb)
}
