package block_test

import (
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
		nil)

	assert.NotNil(t, be)
}

func TestBlockProc_GetTransactionFromPool(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil)

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
		nil)
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

func TestBlockProc_VerifyBlockSignature(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.TxProcessorMock{},
		nil)

	b := be.VerifyBlockSignature(nil)
	assert.Equal(t, false, b)

	hdr := block.Header{Signature: []byte("blk_sig0")}

	b = be.VerifyBlockSignature(&hdr)
	assert.Equal(t, true, b)
}

func TestBlockProc_ProcessBlock(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	tpm := mock.TxProcessorMock{}
	tpm.ProcessTransactionCalled = func(transaction *transaction.Transaction) error {
		return nil
	}
	accountsAdapter := &mock.AccountsStub{}
	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&tpm,
		accountsAdapter,
	)

	err := be.ProcessBlock(nil, nil, nil)
	assert.Equal(t, process.ErrNilBlockChain, err)

	blkc, err := blockchain.NewBlockChain(blockchainConfig())
	assert.Nil(t, err)

	// cleanup after tests
	defer func() {
		_ = blkc.Destroy()
	}()

	err = be.ProcessBlock(blkc, nil, nil)
	assert.Equal(t, process.ErrNilBlockHeader, err)

	hdr := block.Header{Nonce: 0, PrevHash: []byte("")}

	err = be.ProcessBlock(blkc, &hdr, nil)
	assert.Equal(t, process.ErrNilBlockBody, err)

	hdr.Nonce = 2
	hdr.Round = 2
	blk := block.TxBlockBody{}

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	blkc.CurrentBlockHeader = &block.Header{Nonce: 1, BlockBodyHash: []byte("blk_hash1")}
	hdr.Nonce = 1

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 3

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("blk_hash2")

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, process.ErrInvalidBlockHash, err)

	hdr.PrevHash = []byte("blk_hash1")

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, process.ErrInvalidBlockSignature, err)

	hdr.Signature = []byte("blk_sig1")
	blk.MiniBlocks = append(blk.MiniBlocks, block.MiniBlock{ShardID: 0})
	blk.MiniBlocks[0].TxHashes = append(blk.MiniBlocks[0].TxHashes, []byte("tx_hash1"))

	tp.AddTransaction([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0)

	accountsAdapter.RootHashCalled = func() []byte {
		return []byte("")
	}

	accountsAdapter.CommitCalled = func() ([]byte, error) {
		return []byte(""), nil
	}

	accountsAdapter.JournalLenCalled = func() int {
		return 0
	}

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Nil(t, err)
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
	)

	bl, err := be.CreateTxBlockBody(1, 0, 100, func() bool { return true })

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
	)

	haveTime := func() bool {
		return false
	}

	bl, err := be.CreateTxBlockBody(1, 0, 100, haveTime)

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
	)

	blk, err := be.CreateTxBlockBody(1, 0, 100, haveTime)

	assert.NotNil(t, blk)
	assert.Nil(t, err)
}

func TestBlockProcessor_CreateGenesisBlockBodyWithFailSetBalanceShouldPanic(t *testing.T) {

}

func TestBlockProcessor_CreateGenesisBlockBodyOK(t *testing.T) {

}

func TestBlockProcessor_RemoveBlockTxsFromPoolNilBlockOK(t *testing.T) {

}

func TestBlockProcessor_RemoveBlockTxsFromPoolOK(t *testing.T) {

}
