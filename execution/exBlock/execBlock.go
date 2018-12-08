package exBlock

import (
	"bytes"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(2000 * time.Millisecond)

// execBlock implements BlockExecutor interface and actually it tries to execute block
type execBlock struct {
	tp                   *transactionPool.TransactionPool
	hasher               hashing.Hasher
	marsher              marshal.Marshalizer
	txExecutor           execution.TransactionExecutor
	ChRcvTx              chan []byte
	OnRequestTransaction func(destShardID uint32, txHash []byte)
}

// NewExecBlock creates a new execBlock object
func NewExecBlock(
	tp *transactionPool.TransactionPool,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txExecutor execution.TransactionExecutor,
) *execBlock {

	eb := execBlock{
		tp:         tp,
		hasher:     hasher,
		marsher:    marshalizer,
		txExecutor: txExecutor,
	}

	eb.ChRcvTx = make(chan []byte)

	eb.tp.OnAddTransaction = eb.receivedTransaction

	return &eb
}

// ProcessBlock process the block and the transactions inside
func (eb *execBlock) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.Block) error {

	if blockChain == nil {
		return execution.ErrNilBlockChain
	}

	if header == nil {
		return execution.ErrNilBlockHeader
	}

	if body == nil {
		return execution.ErrNilBlockBody
	}

	if blockChain.CurrentBlock == nil {
		if header.Nonce > 1 {
			return execution.ErrHigherNonceInBlock
		}
	} else {
		if header.Nonce < blockChain.CurrentBlock.Nonce+1 {
			return execution.ErrLowerNonceInBlock
		}

		if header.Nonce > blockChain.CurrentBlock.Nonce+1 {
			return execution.ErrHigherNonceInBlock
		}

		if !bytes.Equal(header.PrevHash, blockChain.CurrentBlock.BlockHash) {
			return execution.ErrInvalidBlockHash
		}
	}

	if !eb.verifyBlockSignature(header) {
		return execution.ErrInvalidBlockSignature
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		for j := 0; j < len(body.MiniBlocks[i].TxHashes); j++ {
			tx := eb.getTransactionFromPool(body.MiniBlocks[i].DestShardID, body.MiniBlocks[i].TxHashes[j])

			if tx == nil {
				tx = eb.requestTransactionFromNetwork(
					body.MiniBlocks[i].DestShardID,
					body.MiniBlocks[i].TxHashes[j],
					WaitTime)

				if tx == nil {
					return execution.ErrMissingTransaction
				}
			}

			err := eb.txExecutor.ProcessTransaction(tx)

			if err != nil {
				return err
			}
		}
	}

	// TODO: Check app state root hash

	// commit body in the blockchain
	return eb.commitBlock(blockChain, header, body)
}

// commitBlock commits the block in the blockchain if everything was checked successfully
func (eb *execBlock) commitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.Block) error {
	if header.Nonce == 0 {
		blockChain.GenesisBlock = header
	}

	blockChain.CurrentBlock = header
	blockChain.LocalHeight = big.NewInt(int64(header.Nonce)) // should be refactored to hold the same data type
	// (big.Int or uint64?)
	buff, err := eb.marsher.Marshal(header)

	if err != nil {
		return execution.ErrMarshalWithoutSuccess
	}

	err = blockChain.Put(blockchain.BlockHeaderUnit, header.BlockHash, buff)

	if err != nil {
		return execution.ErrPersistWithoutSuccess
	}

	buff, err = eb.marsher.Marshal(block)

	if err != nil {
		return execution.ErrMarshalWithoutSuccess
	}

	err = blockChain.Put(blockchain.BlockUnit, header.BlockHash, buff)

	if err != nil {
		return execution.ErrPersistWithoutSuccess
	}

	for i := 0; i < len(block.MiniBlocks); i++ {
		for j := 0; j < len(block.MiniBlocks[i].TxHashes); j++ {
			tx := eb.getTransactionFromPool(block.MiniBlocks[i].DestShardID, block.MiniBlocks[i].TxHashes[j])
			if tx == nil {
				return execution.ErrMissingTransaction
			}

			buff, err = eb.marsher.Marshal(tx)

			if err != nil {
				return execution.ErrMarshalWithoutSuccess
			}

			err = blockChain.Put(blockchain.TransactionUnit, block.MiniBlocks[i].TxHashes[j], buff)

			if err != nil {
				return execution.ErrPersistWithoutSuccess
			}
		}
	}

	return nil
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (eb *execBlock) getTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	txStore := eb.tp.MiniPoolTxStore(destShardID)

	if txStore == nil {
		return nil
	}

	val, ok := txStore.Get(txHash)

	if !ok {
		return nil
	}

	return val.(*transaction.Transaction)
}

// requestTransactionFromNetwork method requests a transaction from network when it is not found in the pool
func (eb *execBlock) requestTransactionFromNetwork(destShardID uint32, txHash []byte, waitTime time.Duration,
) *transaction.Transaction {
	// Request transaction from network than wait for it, after topic request will be implemented
	if eb.OnRequestTransaction != nil {
		eb.OnRequestTransaction(destShardID, txHash)
	}

	for {
		select {
		case rcvTxhash := <-eb.ChRcvTx:
			if bytes.Equal(txHash, rcvTxhash) {
				tx := eb.getTransactionFromPool(destShardID, txHash)
				if tx != nil {
					return tx
				}
			}
		case <-time.After(waitTime):
			return nil
		}
	}

	return nil
}

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (eb *execBlock) receivedTransaction(txHash []byte) {
	eb.ChRcvTx <- txHash
}

// verifyBlockSignature verifies if the block has all the valid signatures needed
func (eb *execBlock) verifyBlockSignature(header *block.Header) bool {
	if header == nil || header.Signature == nil {
		return false
	}

	// TODO: Check block signature after multisig will be implemented
	return true
}
