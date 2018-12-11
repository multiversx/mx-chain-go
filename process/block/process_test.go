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

func TestNewBlockExec(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{},
		nil)

	assert.NotNil(t, be)
}

func TestBlockExec_GetTransactionFromPool(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{},
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

func TestBlockExec_RequestTransactionFromNetwork(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{},
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

func TestBlockExec_VerifyBlockSignature(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{},
		nil)

	b := be.VerifyBlockSignature(nil)
	assert.Equal(t, false, b)

	hdr := block.Header{Signature: []byte("blk_sig0")}

	b = be.VerifyBlockSignature(&hdr)
	assert.Equal(t, true, b)
}

func TestBlockExec_ProcessBlock(t *testing.T) {
	tp := transactionPool.NewTransactionPool(nil)

	etm := mock.ExecTransactionMock{}
	etm.ProcessTransactionCalled = func(transaction *transaction.Transaction) error {
		return nil
	}
	accountsAdapter := &mock.AccountsStub{}
	be := blproc.NewBlockProcessor(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&etm,
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
