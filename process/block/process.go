package block

import (
	"bytes"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(2000 * time.Millisecond)

var log = logger.NewDefaultLogger()

// blockProcessor implements BlockProcessor interface and actually it tries to execute block
type blockProcessor struct {
	txPool               data.ShardedDataCacherNotifier
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	txProcessor          process.TransactionProcessor
	ChRcvAllTxs          chan bool
	OnRequestTransaction func(destShardID uint32, txHash []byte)
	requestedTxHashes    map[string]bool
	mut                  sync.RWMutex
	accounts             state.AccountsAdapter
	shardCoordinator     sharding.ShardCoordinator
}

// NewBlockProcessor creates a new blockProcessor object
func NewBlockProcessor(
	txPool data.ShardedDataCacherNotifier,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionProcessor,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.ShardCoordinator,
) *blockProcessor {
	//TODO: check nil values

	bp := blockProcessor{
		txPool:           txPool,
		hasher:           hasher,
		marshalizer:      marshalizer,
		txProcessor:      txProcessor,
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
	}

	bp.ChRcvAllTxs = make(chan bool)

	bp.txPool.RegisterHandler(bp.receivedTransaction)

	return &bp
}

// ProcessAndCommit takes each transaction from the transactions block body received as parameter
// and processes it, updating at the same time the state trie and the associated root hash
// if transaction is not valid or not found it will return error.
// If all ok it will commit the block and state.
func (bp *blockProcessor) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	err := bp.validateHeader(blockChain, header)
	if err != nil {
		return err
	}

	err = bp.ProcessBlock(blockChain, header, body)

	defer func() {
		if err != nil {
			bp.RevertAccountState()
		}
	}()

	if err != nil {
		return err
	}

	if !bp.VerifyStateRoot(bp.accounts.RootHash()) {
		err = process.ErrRootStateMissmatch
		return err
	}

	err = bp.CommitBlock(blockChain, header, body)
	if err != nil {
		return err
	}

	return nil
}

// RevertAccountState reverets the account state for cleanup failed process
func (bp *blockProcessor) RevertAccountState() {
	err := bp.accounts.RevertToSnapshot(0)

	if err != nil {
		log.Error(err.Error())
	}
}

// ProcessBlock processes a block. It returns nil if all ok or the speciffic error
func (bp *blockProcessor) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	err := bp.validateBlockBody(body)
	if err != nil {
		return err
	}

	bp.requestBlockTransactions(body)
	bp.waitForTxHashes()

	if bp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		if err != nil {
			bp.RevertAccountState()
		}
	}()

	err = bp.processBlockTransactions(body, int32(header.Round))

	if err != nil {
		return err
	}

	return nil
}

// RemoveBlockTxsFromPool removes the TxBlock transactions from associated tx pools
func (bp *blockProcessor) RemoveBlockTxsFromPool(body *block.TxBlockBody) error {
	if body == nil {
		return process.ErrNilTxBlockBody
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		bp.txPool.RemoveSetOfDataFromPool(body.MiniBlocks[i].TxHashes,
			body.MiniBlocks[i].ShardID)
	}

	return nil
}

// VerifyStateRoot verifies the state root hash given as parameter agains the
// Merkle trie root hash stored for accounts and returns if equal or not
func (bp *blockProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bytes.Equal(bp.accounts.RootHash(), rootHash)
}

// CreateTxBlockBody creates a transactions block body by filling it with transactions out of the transactions pools
// as long as the transactions limit for the block has not been reached and there is still time to add transactions
func (bp *blockProcessor) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
	mblks, err := bp.createMiniBlocks(bp.shardCoordinator.NoShards(), maxTxInBlock, round, haveTime)

	if err != nil {
		return nil, err
	}

	rootHash := bp.accounts.RootHash()

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
func (bp *blockProcessor) CreateGenesisBlockBody(balances map[string]big.Int, shardId uint32) *block.StateBlockBody {
	if bp.txProcessor == nil {
		panic("transaction Processor is nil")
	}

	rootHash, err := bp.txProcessor.SetBalancesToTrie(balances)

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

// GetRootHash returns the accounts merkle tree root hash
func (bp *blockProcessor) GetRootHash() []byte {
	return bp.accounts.RootHash()
}

func (bp *blockProcessor) validateHeader(blockChain *blockchain.BlockChain, header *block.Header) error {
	headerWrapper := HeaderWrapper{
		Header: header,
	}

	err := headerWrapper.IntegrityAndValidity(bp.shardCoordinator)
	if err != nil {
		return err
	}

	if blockChain.CurrentBlockHeader == nil {
		if !bp.isFirstBlockInEpoch(header) {
			return process.ErrWrongNonceInBlock
		}
	} else {
		if bp.isCorrectNonce(blockChain.CurrentBlockHeader.Nonce, header.Nonce) {
			return process.ErrWrongNonceInBlock
		}

		if !bytes.Equal(header.PrevHash, blockChain.CurrentBlockHeader.BlockBodyHash) {
			return process.ErrInvalidBlockHash
		}
	}

	if headerWrapper.VerifySig() != nil {
		return process.ErrInvalidBlockSignature
	}

	return nil
}

func (bp *blockProcessor) validateBlockBody(body *block.TxBlockBody) error {
	txbWrapper := TxBlockBodyWrapper{
		TxBlockBody: body,
	}

	err := txbWrapper.IntegrityAndValidity(bp.shardCoordinator)
	if err != nil {
		return err
	}

	return nil
}

func (bp *blockProcessor) isCorrectNonce(currentBlockNonce, receivedBlockNonce uint64) bool {
	return currentBlockNonce+1 != receivedBlockNonce
}

func (bp *blockProcessor) isFirstBlockInEpoch(header *block.Header) bool {
	return header.Round == 0
}

func (bp *blockProcessor) processBlockTransactions(body *block.TxBlockBody, round int32) error {
	txbWrapper := TxBlockBodyWrapper{
		TxBlockBody: body,
	}

	err := txbWrapper.IntegrityAndValidity(bp.shardCoordinator)
	if err != nil {
		return err
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		shardId := miniBlock.ShardID

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := bp.getTransactionFromPool(shardId, txHash)
			err := bp.txProcessor.ProcessTransaction(tx, round)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (bp *blockProcessor) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {

	buff, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return process.ErrMarshalWithoutSuccess
	}

	headerHash := bp.hasher.Compute(string(buff))
	err = blockChain.Put(blockchain.BlockHeaderUnit, headerHash, buff)

	if err != nil {
		return process.ErrPersistWithoutSuccess
	}

	buff, err = bp.marshalizer.Marshal(block)

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
			tx := bp.getTransactionFromPool(miniBlock.ShardID, txHash)
			if tx == nil {
				return process.ErrMissingTransaction
			}

			buff, err = bp.marshalizer.Marshal(tx)

			if err != nil {
				return process.ErrMarshalWithoutSuccess
			}

			err = blockChain.Put(blockchain.TransactionUnit, txHash, buff)

			if err != nil {
				return process.ErrPersistWithoutSuccess
			}
		}
	}

	_, err = bp.accounts.Commit()

	if err == nil {
		blockChain.CurrentBlockHeader = header
		blockChain.LocalHeight = int64(header.Nonce)
	}

	return err
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (bp *blockProcessor) getTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	txStore := bp.txPool.ShardDataStore(destShardID)

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
func (bp *blockProcessor) receivedTransaction(txHash []byte) {
	bp.mut.Lock()
	if bp.requestedTxHashes[string(txHash)] {
		delete(bp.requestedTxHashes, string(txHash))
	}

	if len(bp.requestedTxHashes) == 0 {
		bp.ChRcvAllTxs <- true
	}
	bp.mut.Unlock()
}

func (bp *blockProcessor) requestBlockTransactions(body *block.TxBlockBody) {
	bp.mut.Lock()
	missingTxsForShards := bp.computeMissingTxsForShards(body)
	bp.requestedTxHashes = make(map[string]bool)
	if bp.OnRequestTransaction != nil {
		for shardId, txHashes := range missingTxsForShards {
			for _, txHash := range txHashes {
				bp.requestedTxHashes[string(txHash)] = true
				bp.OnRequestTransaction(shardId, txHash)
			}
		}
	}
	bp.mut.Unlock()
}

func (bp *blockProcessor) computeMissingTxsForShards(body *block.TxBlockBody) map[uint32][][]byte {
	missingTxsForShard := make(map[uint32][][]byte)
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		shardId := miniBlock.ShardID
		currentShardMissingTransactions := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := bp.getTransactionFromPool(shardId, txHash)

			if tx == nil {
				currentShardMissingTransactions = append(currentShardMissingTransactions, txHash)
			}
		}
		missingTxsForShard[shardId] = currentShardMissingTransactions
	}

	return missingTxsForShard
}

func (bp *blockProcessor) createMiniBlocks(noShards uint32, maxTxInBlock int, round int32, haveTime func() bool) ([]block.MiniBlock, error) {
	miniBlocks := make([]block.MiniBlock, 0)

	if bp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		return miniBlocks, nil
	}

	for i, txs := 0, 0; i < int(noShards); i++ {
		txStore := bp.txPool.ShardDataStore(uint32(i))

		if txStore == nil {
			continue
		}

		miniBlock := block.MiniBlock{}
		miniBlock.ShardID = uint32(i)
		miniBlock.TxHashes = make([][]byte, 0)

		for _, txHash := range txStore.Keys() {
			snapshot := bp.accounts.JournalLen()

			tx := bp.getTransactionFromPool(miniBlock.ShardID, txHash)

			if tx == nil {
				log.Error("did not find transaction in pool")
				continue
			}

			// execute transaction to change the trie root hash
			err := bp.txProcessor.ProcessTransaction(tx, round)

			if err != nil {
				err = bp.accounts.RevertToSnapshot(snapshot)
				log.LogIfError(err)
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

func (bp *blockProcessor) waitForTxHashes() {
	select {
	case <-bp.ChRcvAllTxs:
		return
	case <-time.After(WaitTime):
		return
	}
}
