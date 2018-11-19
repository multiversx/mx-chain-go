package execution

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
)

const (
	// BlockSuccess signals that the block has been executed correctly
	BlockSuccess ExecCode = iota
	// BlockInvalidParameters signals that the parameters are invalid
	BlockInvalidParameters
	// BlockLowerNonce signals that the block's nonce is lower, block can be discarded
	BlockLowerNonce
	// BlockHigherNonce signals that the block's nonce is higher, block can not be executed right now,
	// maybe in the future
	BlockHigherNonce
	// BlockInvalidHash signals that the received block's previous hash does not match with the last block hash
	BlockInvalidHash
	// BlockTransactionMissing signals that a transaction from the block is not found
	BlockTransactionMissing
	// BlockAccountsError signals a general failure in accounts DB. Process should return immediately and restore
	// the DB trie as there might be inconsistencies
	BlockAccountsError
)

// BlockExec implements BlockExecutor interface and can modify account states according to a block
type BlockExec struct {
	tp   *transactionPool.TransactionPool
	accs state.AccountsHandler
	blkc *blockchain.BlockChain

	hdr *block.Header
	blk *block.Block

	chRcvTx  chan []byte
	chRcvBlk chan uint64

	mut sync.RWMutex
}

// NewBlockExec creates a new BlockExec engine
func NewBlockExec(
	tp *transactionPool.TransactionPool,
	accs state.AccountsHandler,
	blck *blockchain.BlockChain) *BlockExec {

	be := BlockExec{
		tp:   tp,
		accs: accs,
		blkc: blck}

	be.chRcvTx = make(chan []byte)
	be.chRcvBlk = make(chan uint64)

	be.tp.OnAddTransaction = be.receivedTransaction

	go be.work()

	return &be
}

func (be *BlockExec) work() {
	for {
		nounce := uint64(0)
		if be.blkc != nil && be.blkc.CurrentBlock != nil {
			nounce = be.blkc.CurrentBlock.Nonce + 1
		}

		if !be.getBlockFromPool(nounce) {
			if !be.requestBlockFromNetwork(nounce) {
				continue
			}
		}

		es := be.processBlock()

		if es.err == nil {
			fmt.Println("Block processed successfully")
		}
	}
}

// ProcessBlock modifies the account states in respect with the transaction data
func (be *BlockExec) processBlock() *ExecSummary {
	if be.accs == nil {
		return NewExecSummary(BlockInvalidParameters, errors.New("nil accounts"))
	}

	if be.hdr == nil {
		return NewExecSummary(BlockInvalidParameters, errors.New("nil header"))
	}

	if be.blk == nil {
		return NewExecSummary(BlockInvalidParameters, errors.New("nil block"))
	}

	if be.blkc == nil {
		if be.hdr.Nonce > 0 {
			return NewExecSummary(BlockHigherNonce, errors.New("higher Nonce in block: wait for the right one"))
		}
	} else {
		if be.hdr.Nonce < be.blkc.CurrentBlock.Nonce+1 {
			return NewExecSummary(BlockLowerNonce, errors.New("lower Nonce in block: drop block"))
		}

		if be.hdr.Nonce > be.blkc.CurrentBlock.Nonce+1 {
			return NewExecSummary(BlockHigherNonce, errors.New("higher Nonce in block: wait for the right one"))
		}

		if !bytes.Equal(be.hdr.PrevHash, be.blkc.CurrentBlock.BlockHash) {
			return NewExecSummary(BlockInvalidHash, errors.New("invalid hash: drop block"))
		}
	}

	txExec := NewTxExec()

	for i := 0; i < len(be.blk.MiniBlocks); i++ {
		for j := 0; j < len(be.blk.MiniBlocks[i].TxHashes); j++ {
			tx := be.getTransactionFromPool(be.blk.MiniBlocks[i].DestShardID, be.blk.MiniBlocks[i].TxHashes[j])

			if tx == nil {
				tx = be.requestTransactionFromNetwork(be.blk.MiniBlocks[i].DestShardID, be.blk.MiniBlocks[i].TxHashes[j])

				if tx == nil {
					return NewExecSummary(BlockTransactionMissing, errors.New("transaction missing: drop block"))
				}
			}

			es := txExec.ProcessTransaction(be.accs, tx)

			if es.err != nil {
				return NewExecSummary(es.ec, es.err)
			}
		}
	}

	return NewExecSummary(BlockSuccess, nil)
}

func (be *BlockExec) getTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	txStore := be.tp.MiniPoolTxStore(destShardID)
	val, ok := txStore.Get(txHash)

	if !ok {
		return nil
	}

	return val.(*transaction.Transaction)
}

func (be *BlockExec) requestTransactionFromNetwork(destShardID uint32, txHash []byte) *transaction.Transaction {
	// to do request transaction from network than wait for it

	for {
		select {
		case rcvTxhash := <-be.chRcvTx:
			if bytes.Equal(txHash, rcvTxhash) {
				tx := be.getTransactionFromPool(destShardID, txHash)
				if tx != nil {
					return tx
				}
			}
		}
	}

	return nil
}

func (be *BlockExec) receivedTransaction(txHash []byte) {
	be.chRcvTx <- txHash
}

func (be *BlockExec) getBlockFromPool(nounce uint64) bool {

	be.hdr = nil
	be.blk = nil

	return false
}

func (be *BlockExec) requestBlockFromNetwork(nounce uint64) bool {
	// to do request block from network than wait for it

	for {
		select {
		case rcvBlkNounce := <-be.chRcvBlk:
			if nounce == rcvBlkNounce {
				if be.getBlockFromPool(nounce) {
					return true
				}
			}
		}
	}

	return false
}

func (be *BlockExec) receivedBlock(nounce uint64) {
	be.chRcvBlk <- nounce
}
