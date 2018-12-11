package block

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(2000 * time.Millisecond)

// blockProcessor implements BlockExecutor interface and actually it tries to execute block
type blockProcessor struct {
	tp                   *transactionPool.TransactionPool
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	txProcessor          process.TransactionExecutor
	ChRcvAllTxs          chan bool
	OnRequestTransaction func(destShardID uint32, txHash []byte)
	requestedTxHashes    map[string]bool
	mut                  sync.RWMutex
	accounts             state.AccountsAdapter
}

// NewBlockProcessor creates a new blockProcessor object
func NewBlockProcessor(
	tp *transactionPool.TransactionPool,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionExecutor,
	accounts state.AccountsAdapter,
) *blockProcessor {
	//TODO: check nil values

	eb := blockProcessor{
		tp:          tp,
		hasher:      hasher,
		marshalizer: marshalizer,
		txProcessor: txProcessor,
		accounts:    accounts,
	}

	eb.ChRcvAllTxs = make(chan bool)

	eb.tp.RegisterTransactionHandler(eb.receivedTransaction)

	return &eb
}

// ProcessBlock process the block and the transactions inside
func (eb *blockProcessor) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	err := eb.validateBlock(blockChain, header, body)
	if err != nil {
		return err
	}

	eb.requestBlockTransactions(body)
	eb.waitForTxHashes()

	if eb.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		if err != nil {
			err2 := eb.accounts.RevertToSnapshot(0)
			if err2 != nil {
				fmt.Println(err2.Error())
			}
		}
	}()

	err = eb.ProcessBlockTransactions(body)
	if err != nil {
		return err
	}

	// TODO: Check app state root hash
	if !eb.VerifyStateRoot(eb.accounts.RootHash()) {
		return process.ErrRootStateMissmatch
	}

	err = eb.commitBlock(blockChain, header, body)
	if err != nil {
		return err
	}

	return nil
}

func (eb *blockProcessor) validateBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	if blockChain == nil {
		return process.ErrNilBlockChain
	}

	if header == nil {
		return process.ErrNilBlockHeader
	}

	if body == nil {
		return process.ErrNilBlockBody
	}

	if blockChain.CurrentBlockHeader == nil {
		if !eb.IsFirstBlockInEpoch(header) {
			return process.ErrWrongNonceInBlock
		}
	} else {
		if eb.IsCorrectNonce(blockChain.CurrentBlockHeader.Nonce, header.Nonce) {
			return process.ErrWrongNonceInBlock
		}

		if !bytes.Equal(header.PrevHash, blockChain.CurrentBlockHeader.BlockBodyHash) {
			return process.ErrInvalidBlockHash
		}
	}

	if !eb.verifyBlockSignature(header) {
		return process.ErrInvalidBlockSignature
	}

	return nil
}

func (eb *blockProcessor) IsCorrectNonce(currentBlockNonce, receivedBlockNonce uint64) bool {
	return currentBlockNonce+1 != receivedBlockNonce
}

func (eb *blockProcessor) IsFirstBlockInEpoch(header *block.Header) bool {
	return header.Round == 0
}

//TODO: do not marshal and unmarshal: wrapper struct with marshaled data
// commitBlock commits the block in the blockchain if everything was checked successfully
func (eb *blockProcessor) commitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {

	blockChain.CurrentBlockHeader = header
	blockChain.LocalHeight = int64(header.Nonce)
	buff, err := eb.marshalizer.Marshal(header)
	if err != nil {
		return process.ErrMarshalWithoutSuccess
	}

	headerHash := eb.hasher.Compute(string(buff))
	err = blockChain.Put(blockchain.BlockHeaderUnit, headerHash, buff)

	if err != nil {
		return process.ErrPersistWithoutSuccess
	}

	buff, err = eb.marshalizer.Marshal(block)

	if err != nil {
		return process.ErrMarshalWithoutSuccess
	}

	err = blockChain.Put(blockchain.TxBlockBodyUnit, header.BlockBodyHash, buff)

	if err != nil {
		return process.ErrPersistWithoutSuccess
	}

	for i := 0; i < len(block.MiniBlocks); i++ {
		miniBlock := block.MiniBlocks[i]
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := eb.getTransactionFromPool(miniBlock.ShardID, txHash)
			if tx == nil {
				return process.ErrMissingTransaction
			}

			buff, err = eb.marshalizer.Marshal(tx)

			if err != nil {
				return process.ErrMarshalWithoutSuccess
			}

			err = blockChain.Put(blockchain.TransactionUnit, txHash, buff)

			if err != nil {
				return process.ErrPersistWithoutSuccess
			}
		}
	}

	_, err = eb.accounts.Commit()

	return err
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (eb *blockProcessor) getTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
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

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (eb *blockProcessor) receivedTransaction(txHash []byte) {
	eb.mut.Lock()
	if eb.requestedTxHashes[string(txHash)] {
		delete(eb.requestedTxHashes, string(txHash))
	}
	eb.mut.Unlock()

	if len(eb.requestedTxHashes) == 0 {
		eb.ChRcvAllTxs <- true
	}
}

// verifyBlockSignature verifies if the block has all the valid signatures needed
func (eb *blockProcessor) verifyBlockSignature(header *block.Header) bool {
	if header == nil || header.Signature == nil {
		return false
	}

	// TODO: Check block signature after multisig will be implemented
	return true
}

func (eb *blockProcessor) requestBlockTransactions(body *block.TxBlockBody) {
	missingTxsForShards := eb.computeMissingTxsForShards(body)
	eb.requestedTxHashes = make(map[string]bool)
	if eb.OnRequestTransaction != nil {
		for shardId, txHashes := range missingTxsForShards {
			for _, txHash := range txHashes {
				eb.requestedTxHashes[string(txHash)] = true
				eb.OnRequestTransaction(shardId, txHash)
			}
		}
	}
}

func (eb *blockProcessor) computeMissingTxsForShards(body *block.TxBlockBody) map[uint32][][]byte {
	missingTxsForShard := make(map[uint32][][]byte)
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		shardId := miniBlock.ShardID
		currentShardMissingTransactions := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := eb.getTransactionFromPool(shardId, txHash)

			if tx == nil {
				currentShardMissingTransactions = append(currentShardMissingTransactions, txHash)
			}
		}
		missingTxsForShard[shardId] = currentShardMissingTransactions
	}

	return missingTxsForShard
}

func (eb *blockProcessor) ProcessBlockTransactions(body *block.TxBlockBody) error {
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		shardId := miniBlock.ShardID

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := eb.getTransactionFromPool(shardId, txHash)
			err := eb.txProcessor.ProcessTransaction(tx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (eb *blockProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bytes.Equal(eb.accounts.RootHash(), rootHash)
}

func (eb *blockProcessor) waitForTxHashes() {
	select {
	case <-eb.ChRcvAllTxs:
		return
	case <-time.After(WaitTime):
		return
	}
}
