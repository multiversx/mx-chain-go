package exBlock_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/exBlock"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/stretchr/testify/assert"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(100 * time.Millisecond)

func TestNewBlockExec(t *testing.T) {
	tp := transactionPool.NewTransactionPool()

	be := exBlock.NewExecBlock(
		tp,
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{})

	assert.NotNil(t, be)
}

func TestBlockExec_GetTransactionFromPool(t *testing.T) {
	tp := transactionPool.NewTransactionPool()

	be := exBlock.NewExecBlock(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{})

	txHash := []byte("tx1_hash")

	tx := be.GetTransactionFromPool(1, txHash)
	assert.Nil(t, tp.MiniPoolTxStore(1))
	assert.Nil(t, tx)

	tp.NewMiniPool(1)

	tx = be.GetTransactionFromPool(1, txHash)
	assert.NotNil(t, tp.MiniPoolTxStore(1))
	assert.Nil(t, tx)

	tp.OnAddTransaction = nil
	tp.AddTransaction(txHash, &transaction.Transaction{Nonce: uint64(1)}, 1)

	tx = be.GetTransactionFromPool(1, txHash)
	assert.NotNil(t, tp.MiniPoolTxStore(1))
	assert.NotNil(t, tx)
	assert.Equal(t, uint64(1), tx.Nonce)
}

func TestBlockExec_RequestTransactionFromNetwork(t *testing.T) {
	tp := transactionPool.NewTransactionPool()

	be := exBlock.NewExecBlock(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{})

	tx := be.RequestTransactionFromNetwork(1, []byte("tx1_hash1"), WaitTime)
	assert.Nil(t, tx)

	go be.ReceivedTransaction([]byte("tx1_hash2"))

	tx = be.RequestTransactionFromNetwork(1, []byte("tx1_hash1"), WaitTime)
	assert.Nil(t, tx)

	go be.ReceivedTransaction([]byte("tx1_hash1"))

	tx = be.RequestTransactionFromNetwork(1, []byte("tx1_hash1"), WaitTime)
	assert.Nil(t, tx)

	tp.OnAddTransaction = nil
	tp.AddTransaction([]byte("tx1_hash1"), &transaction.Transaction{Nonce: uint64(1)}, 1)

	go be.ReceivedTransaction([]byte("tx1_hash1"))

	tx = be.RequestTransactionFromNetwork(1, []byte("tx1_hash1"), WaitTime)
	assert.NotNil(t, tx)
	assert.Equal(t, uint64(1), tx.Nonce)
}

func TestBlockExec_VerifyBlockSignature(t *testing.T) {
	tp := transactionPool.NewTransactionPool()

	be := exBlock.NewExecBlock(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.ExecTransactionMock{})

	b := be.VerifyBlockSignature(nil)
	assert.Equal(t, false, b)

	hdr := block.Header{Signature: []byte("blk_sig0")}

	b = be.VerifyBlockSignature(&hdr)
	assert.Equal(t, true, b)
}

func TestBlockExec_ProcessBlock(t *testing.T) {
	tp := transactionPool.NewTransactionPool()

	etm := mock.ExecTransactionMock{}
	etm.ProcessTransactionCalled = func(transaction *transaction.Transaction) error {
		return nil
	}

	be := exBlock.NewExecBlock(
		tp, &mock.HasherMock{},
		&mock.MarshalizerMock{},
		&etm)

	err := be.ProcessBlock(nil, nil, nil)
	assert.Equal(t, execution.ErrNilBlockChain, err)

	blkc, err := blockchain.NewData()
	assert.Nil(t, err)

	err = be.ProcessBlock(blkc, nil, nil)
	assert.Equal(t, execution.ErrNilBlockHeader, err)

	hdr := block.Header{Nonce: 0, PrevHash: []byte("")}

	err = be.ProcessBlock(blkc, &hdr, nil)
	assert.Equal(t, execution.ErrNilBlockBody, err)

	hdr.Nonce = 2
	blk := block.Block{}

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, execution.ErrHigherNonceInBlock, err)

	blkc.CurrentBlock = &block.Header{Nonce: 1, BlockHash: []byte("blk_hash1")}
	hdr.Nonce = 1

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, execution.ErrLowerNonceInBlock, err)

	hdr.Nonce = 3

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, execution.ErrHigherNonceInBlock, err)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("blk_hash2")

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, execution.ErrInvalidBlockHash, err)

	hdr.PrevHash = []byte("blk_hash1")

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Equal(t, execution.ErrInvalidBlockSignature, err)

	hdr.Signature = []byte("blk_sig1")
	blk.MiniBlocks = append(blk.MiniBlocks, block.MiniBlock{DestShardID: 0})
	blk.MiniBlocks[0].TxHashes = append(blk.MiniBlocks[0].TxHashes, []byte("tx_hash1"))
	tp.OnAddTransaction = nil
	tp.AddTransaction([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0)

	err = be.ProcessBlock(blkc, &hdr, &blk)
	assert.Nil(t, err)
}
