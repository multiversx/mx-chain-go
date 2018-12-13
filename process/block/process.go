package block

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(2000 * time.Millisecond)

type StateBlockBodyWrapper struct {
	block.StateBlockBody
}

type PeerBlockBodyWrapper struct {
	block.PeerBlockBody
}

type TxBlockBodyWrapper struct {
	block.TxBlockBody
}

type HeaderWrapper struct {
	block.Header
}

var log = logger.NewDefaultLogger()

// blockProcessor implements BlockProcessor interface and actually it tries to execute block
type blockProcessor struct {
	tp                   *transactionPool.TransactionPool
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	txProcessor          process.TransactionProcessor
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
	txProcessor process.TransactionProcessor,
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

// ProcessBlock takes each transaction from the transactions block body received as parameter
// and processes it, updating at the same time the state trie and the associated root hash
// if transaction is not valid or not found it will return error
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

	err = eb.processBlockTransactions(body)
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

// RemoveBlockTxsFromPool removes the TxBlock transactions from associated tx pools
func (eb *blockProcessor) RemoveBlockTxsFromPool(body *block.TxBlockBody) {
	if body != nil {
		for i := 0; i < len(body.MiniBlocks); i++ {
			eb.tp.RemoveTransactionsFromPool(body.MiniBlocks[i].TxHashes,
				body.MiniBlocks[i].ShardID)
		}
	}
}

// VerifyStateRoot verifies the state root hash given as parameter agains the
// Merkle trie root hash stored for accounts and returns if equal or not
func (eb *blockProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bytes.Equal(eb.accounts.RootHash(), rootHash)
}

// CreateTxBlockBody creates a transactions block body by filling it with transactions out of the transactions pools
// as long as the transactions limit for the block has not been reached and there is still time to add transactions
func (eb *blockProcessor) CreateTxBlockBody(nbShards int, shardId uint32, maxTxInBlock int, haveTime func() bool) (*block.TxBlockBody, error) {
	mblks, err := eb.createMiniBlocks(nbShards, maxTxInBlock, haveTime)

	if err != nil {
		return nil, err
	}

	rootHash := eb.accounts.RootHash()

	blk := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{
			RootHash: rootHash,
			ShardID:  shardId,
		},
		MiniBlocks: mblks,
	}

	return blk, nil
}

// CreateGenesisBlockBody creates the genesis block body from map of account balances
func (eb *blockProcessor) CreateGenesisBlockBody(balances map[string]big.Int, shardId uint32) *block.StateBlockBody {

	rootHash, err := eb.txProcessor.SetBalancesToTrie(balances)

	if err != nil {
		// cannot create Genesis block
		panic(err)
	}

	stateBlockBody := &block.StateBlockBody{
		RootHash: rootHash,
		ShardID:  shardId,
	}

	return stateBlockBody
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
		if !eb.isFirstBlockInEpoch(header) {
			return process.ErrWrongNonceInBlock
		}
	} else {
		if eb.isCorrectNonce(blockChain.CurrentBlockHeader.Nonce, header.Nonce) {
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

func (eb *blockProcessor) isCorrectNonce(currentBlockNonce, receivedBlockNonce uint64) bool {
	return currentBlockNonce+1 != receivedBlockNonce
}

func (eb *blockProcessor) isFirstBlockInEpoch(header *block.Header) bool {
	return header.Round == 0
}

func (eb *blockProcessor) processBlockTransactions(body *block.TxBlockBody) error {
	if body == nil {
		return process.ErrNilBlockBody
	}

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

func (eb *blockProcessor) createMiniBlocks(nbShards int, maxTxInBlock int, haveTime func() bool) ([]block.MiniBlock, error) {
	miniBlocks := make([]block.MiniBlock, 0)

	if eb.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		return miniBlocks, nil
	}

	for i, txs := 0, 0; i < nbShards; i++ {
		txStore := eb.tp.MiniPoolTxStore(uint32(i))

		if txStore == nil {
			continue
		}

		miniBlock := block.MiniBlock{}
		miniBlock.ShardID = uint32(i)
		miniBlock.TxHashes = make([][]byte, 0)

		for _, txHash := range txStore.Keys() {
			tx := eb.getTransactionFromPool(miniBlock.ShardID, txHash)

			if tx == nil {
				log.Error("did not find transaction in pool")
				continue
			}

			// execute transaction to change the trie root hash
			err := eb.txProcessor.ProcessTransaction(tx)

			if err != nil {
				continue
			}

			miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
			txs++

			if txs >= maxTxInBlock { // max transactions count in one block was reached
				miniBlocks = append(miniBlocks, miniBlock)
				return miniBlocks, nil
			}
		}

		if !haveTime() {
			miniBlocks = append(miniBlocks, miniBlock)
			return miniBlocks, nil
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks, nil
}

func (eb *blockProcessor) waitForTxHashes() {
	select {
	case <-eb.ChRcvAllTxs:
		return
	case <-time.After(WaitTime):
		return
	}
}
