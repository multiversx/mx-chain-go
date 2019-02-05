package block

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/pkg/errors"
)

var log = logger.NewDefaultLogger()

var txsCurrentBlockProcessed = 0
var txsTotalProcessed = 0

// blockProcessor implements BlockProcessor interface and actually it tries to execute block
type blockProcessor struct {
	dataPool             data.TransientDataHolder
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	txProcessor          process.TransactionProcessor
	ChRcvAllTxs          chan bool
	OnRequestTransaction func(destShardID uint32, txHash []byte)
	requestedTxHashes    map[string]bool
	mut                  sync.RWMutex
	accounts             state.AccountsAdapter
	shardCoordinator     sharding.ShardCoordinator
	forkDetector         process.ForkDetector
}

// NewBlockProcessor creates a new blockProcessor object
func NewBlockProcessor(
	dataPool data.TransientDataHolder,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionProcessor,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.ShardCoordinator,
	forkDetector process.ForkDetector,
	requestTransactionHandler func(destShardID uint32, txHash []byte),
) (*blockProcessor, error) {

	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	if txProcessor == nil {
		return nil, process.ErrNilTxProcessor
	}

	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	if forkDetector == nil {
		return nil, process.ErrNilForkDetector
	}

	if requestTransactionHandler == nil {
		return nil, process.ErrNilTransactionHandler
	}

	bp := blockProcessor{
		dataPool:         dataPool,
		hasher:           hasher,
		marshalizer:      marshalizer,
		txProcessor:      txProcessor,
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		forkDetector:     forkDetector,
	}

	bp.ChRcvAllTxs = make(chan bool)
	bp.OnRequestTransaction = requestTransactionHandler

	transactionPool := bp.dataPool.Transactions()

	if transactionPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	transactionPool.RegisterHandler(bp.receivedTransaction)

	return &bp, nil
}

// ProcessAndCommit takes each transaction from the transactions block body received as parameter
// and processes it, updating at the same time the state trie and the associated root hash
// if transaction is not valid or not found it will return error.
// If all ok it will commit the block and state.
func (bp *blockProcessor) ProcessAndCommit(
	blockChain *blockchain.BlockChain,
	header *block.Header,
	body *block.TxBlockBody,
	haveTime func() time.Duration,
) error {
	err := checkForNils(blockChain, header, body)
	if err != nil {
		return err
	}

	err = bp.validateHeader(blockChain, header)
	if err != nil {
		return err
	}

	err = bp.processBlock(blockChain, header, body, haveTime)

	defer func() {
		if err != nil {
			bp.RevertAccountState()
		}
	}()

	if err != nil {
		return err
	}

	if !bp.VerifyStateRoot(body.RootHash) {
		err = process.ErrRootStateMissmatch
		return err
	}

	err = bp.CommitBlock(blockChain, header, body)
	if err != nil {
		return err
	}

	return nil
}

func checkForNils(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	if blockChain == nil {
		return process.ErrNilBlockChain
	}

	if header == nil {
		return process.ErrNilBlockHeader
	}

	if body == nil {
		return process.ErrNilTxBlockBody
	}

	return nil
}

// RevertAccountState reverts the account state for cleanup failed process
func (bp *blockProcessor) RevertAccountState() {
	err := bp.accounts.RevertToSnapshot(0)

	if err != nil {
		log.Error(err.Error())
	}
}

// ProcessBlock processes a block. It returns nil if all ok or the speciffic error
func (bp *blockProcessor) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
	err := checkForNils(blockChain, header, body)
	if err != nil {
		return err
	}

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	return bp.processBlock(blockChain, header, body, haveTime)
}

func (bp *blockProcessor) processBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
	err := bp.validateBlockBody(body)
	if err != nil {
		return err
	}

	requestedTxs := bp.requestBlockTransactions(body)

	if requestedTxs > 0 {

		log.Info(fmt.Sprintf("requested %d missing txs\n", requestedTxs))

		err := bp.waitForTxHashes(haveTime())

		log.Info(fmt.Sprintf("received %d missing txs\n", requestedTxs-len(bp.requestedTxHashes)))

		if err != nil {
			return err
		}
	}

	if bp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		if err != nil {
			bp.RevertAccountState()
		}
	}()

	err = bp.processBlockTransactions(body, int32(header.Round), haveTime)

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

	transactionPool := bp.dataPool.Transactions()

	if transactionPool == nil {
		return process.ErrNilTransactionPool
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		transactionPool.RemoveSetOfDataFromPool(body.MiniBlocks[i].TxHashes,
			body.MiniBlocks[i].ShardID)
	}

	return nil
}

// VerifyStateRoot verifies the state root hash given as parameter against the
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

// CreateEmptyBlockBody creates a new block body without any tx hash
func (bp *blockProcessor) CreateEmptyBlockBody(shardId uint32, round int32) *block.TxBlockBody {
	miniBlocks := make([]block.MiniBlock, 0)

	rootHash := bp.accounts.RootHash()

	blk := &block.TxBlockBody{
		StateBlockBody: block.StateBlockBody{
			RootHash: rootHash,
			ShardID:  shardId,
		},
		MiniBlocks: miniBlocks,
	}

	return blk
}

// CreateGenesisBlockBody creates the genesis block body from map of account balances
func (bp *blockProcessor) CreateGenesisBlockBody(balances map[string]*big.Int, shardId uint32) (*block.StateBlockBody, error) {
	rootHash, err := bp.txProcessor.SetBalancesToTrie(balances)

	if err != nil {
		return nil, errors.New("can not create genesis block body " + err.Error())
	}

	stateBlockBody := &block.StateBlockBody{
		RootHash: rootHash,
		ShardID:  shardId,
	}

	return stateBlockBody, nil
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

		if !bytes.Equal(header.PrevHash, blockChain.CurrentBlockHeaderHash) {

			log.Info(fmt.Sprintf(
				"header.Nonce = %d has header.PrevHash = %s and blockChain.CurrentBlockHeader.Nonce = %d has blockChain.CurrentBlockHeaderHash = %s\n",
				header.Nonce,
				toB64(header.PrevHash),
				blockChain.CurrentBlockHeader.Nonce,
				toB64(blockChain.CurrentBlockHeaderHash)))

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

func (bp *blockProcessor) processBlockTransactions(body *block.TxBlockBody, round int32, haveTime func() time.Duration) error {
	txbWrapper := TxBlockBodyWrapper{
		TxBlockBody: body,
	}

	err := txbWrapper.IntegrityAndValidity(bp.shardCoordinator)
	if err != nil {
		return err
	}

	txPool := bp.dataPool.Transactions()

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		shardId := miniBlock.ShardID

		//TODO: Remove this display
		bp.displayTxsInfo(&miniBlock, shardId)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			tx := bp.getTransactionFromPool(shardId, txHash)

			err := bp.processAndRemoveBadTransaction(
				txHash,
				tx,
				txPool,
				round,
				miniBlock.ShardID,
			)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (bp *blockProcessor) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody) error {
	err := checkForNils(blockChain, header, body)
	if err != nil {
		return err
	}

	buff, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return process.ErrMarshalWithoutSuccess
	}

	headerHash := bp.hasher.Compute(string(buff))
	err = blockChain.Put(blockchain.BlockHeaderUnit, headerHash, buff)
	if err != nil {
		return process.ErrPersistWithoutSuccess
	}

	buff, err = bp.marshalizer.Marshal(body)
	if err != nil {
		return process.ErrMarshalWithoutSuccess
	}

	err = blockChain.Put(blockchain.TxBlockBodyUnit, header.BlockBodyHash, buff)
	if err != nil {
		return process.ErrPersistWithoutSuccess
	}

	headerNoncePool := bp.dataPool.HeadersNonces()
	if headerNoncePool == nil {
		return process.ErrNilDataPoolHolder
	}

	_ = headerNoncePool.Put(header.Nonce, headerHash)

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
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

	err = bp.RemoveBlockTxsFromPool(body)
	if err != nil {
		log.Error(err.Error())
	}

	_, err = bp.accounts.Commit()
	if err != nil {
		return err
	}

	blockChain.CurrentTxBlockBody = body
	blockChain.CurrentBlockHeader = header
	blockChain.CurrentBlockHeaderHash = headerHash
	err = bp.forkDetector.AddHeader(header, headerHash, false)

	go bp.displayBlockchain(blockChain)

	return err
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (bp *blockProcessor) getTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	txPool := bp.dataPool.Transactions()
	if txPool == nil {
		log.Error(process.ErrNilTransactionPool.Error())
		return nil
	}

	txStore := txPool.ShardDataStore(destShardID)
	if txStore == nil {
		log.Error(process.ErrNilTxStorage.Error())
		return nil
	}

	val, ok := txStore.Get(txHash)
	if !ok {
		return nil
	}

	v := val.(*transaction.Transaction)

	return v
}

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (bp *blockProcessor) receivedTransaction(txHash []byte) {
	bp.mut.Lock()
	if len(bp.requestedTxHashes) > 0 {
		if bp.requestedTxHashes[string(txHash)] {
			delete(bp.requestedTxHashes, string(txHash))
		}
		lenReqTxHashes := len(bp.requestedTxHashes)
		bp.mut.Unlock()

		if lenReqTxHashes == 0 {
			bp.ChRcvAllTxs <- true
		}
		return
	}
	bp.mut.Unlock()
}

func (bp *blockProcessor) requestBlockTransactions(body *block.TxBlockBody) int {
	bp.mut.Lock()
	requestedTxs := 0
	missingTxsForShards := bp.computeMissingTxsForShards(body)
	bp.requestedTxHashes = make(map[string]bool)
	if bp.OnRequestTransaction != nil {
		for shardId, txHashes := range missingTxsForShards {
			for _, txHash := range txHashes {
				requestedTxs++
				bp.requestedTxHashes[string(txHash)] = true
				bp.OnRequestTransaction(shardId, txHash)
			}
		}
	}
	bp.mut.Unlock()
	return requestedTxs
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

		if len(currentShardMissingTransactions) > 0 {
			missingTxsForShard[shardId] = currentShardMissingTransactions
		}
	}

	return missingTxsForShard
}

func (bp *blockProcessor) processAndRemoveBadTransaction(
	transactionHash []byte,
	transaction *transaction.Transaction,
	txPool data.ShardedDataCacherNotifier,
	round int32,
	shardId uint32,
) error {
	if txPool == nil {
		return process.ErrNilTransactionPool
	}

	err := bp.txProcessor.ProcessTransaction(transaction, round)

	if err == process.ErrLowerNonceInTransaction {
		txPool.RemoveData(transactionHash, shardId)
	}

	return err
}

func (bp *blockProcessor) createMiniBlocks(noShards uint32, maxTxInBlock int, round int32, haveTime func() bool) ([]block.MiniBlock, error) {
	miniBlocks := make([]block.MiniBlock, 0)

	if bp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createMiniBlocks method\n"))
		return miniBlocks, nil
	}

	txPool := bp.dataPool.Transactions()

	if txPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	for i, txs := 0, 0; i < int(noShards); i++ {
		txStore := txPool.ShardDataStore(uint32(i))

		timeBefore := time.Now()
		orderedTxes, orderedTxHashes, err := getTxs(txStore)
		timeAfter := time.Now()

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up after ordered %d txs in %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))
			return miniBlocks, nil
		}

		log.Info(fmt.Sprintf("time elapsed to ordered %d txs: %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))

		if err != nil {
			log.Debug(fmt.Sprintf("when trying to order txs: %s", err.Error()))
			continue
		}

		miniBlock := block.MiniBlock{}
		miniBlock.ShardID = uint32(i)
		miniBlock.TxHashes = make([][]byte, 0)

		log.Info(fmt.Sprintf("creating mini blocks has been started: have %d txs in pool for shard id %d\n", len(orderedTxes), miniBlock.ShardID))

		for index, tx := range orderedTxes {
			if !haveTime() {
				break
			}

			snapshot := bp.accounts.JournalLen()

			if tx == nil {
				log.Error("did not find transaction in pool")
				continue
			}

			// execute transaction to change the trie root hash
			err := bp.processAndRemoveBadTransaction(
				orderedTxHashes[index],
				orderedTxes[index],
				txPool,
				round,
				miniBlock.ShardID,
			)

			if err != nil {
				err = bp.accounts.RevertToSnapshot(snapshot)
				log.LogIfError(err)
				continue
			}

			miniBlock.TxHashes = append(miniBlock.TxHashes, orderedTxHashes[index])
			txs++

			if txs >= maxTxInBlock { // max transactions count in one block was reached
				log.Info(fmt.Sprintf("max txs accepted in one block is reached: added %d txs from %d txs\n", len(miniBlock.TxHashes), len(orderedTxes)))

				if len(miniBlock.TxHashes) > 0 {
					miniBlocks = append(miniBlocks, miniBlock)
				}

				log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))

				return miniBlocks, nil
			}
		}

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up: added %d txs from %d txs\n", len(miniBlock.TxHashes), len(orderedTxes)))

			if len(miniBlock.TxHashes) > 0 {
				miniBlocks = append(miniBlocks, miniBlock)
			}

			log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))

			return miniBlocks, nil
		}

		if len(miniBlock.TxHashes) > 0 {
			miniBlocks = append(miniBlocks, miniBlock)
		}
	}

	log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))

	return miniBlocks, nil
}

func (bp *blockProcessor) waitForTxHashes(waitTime time.Duration) error {
	select {
	case <-bp.ChRcvAllTxs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

func (bp *blockProcessor) displayBlockchain(blkc *blockchain.BlockChain) {
	if blkc == nil {
		return
	}

	blockHeader := blkc.CurrentBlockHeader
	txBlockBody := blkc.CurrentTxBlockBody

	if blockHeader == nil || txBlockBody == nil {
		return
	}

	headerHash, err := bp.computeHeaderHash(blockHeader)

	if err != nil {
		log.Error(err.Error())
		return
	}

	bp.displayLogInfo(blockHeader, txBlockBody, headerHash)
}

func (bp *blockProcessor) computeHeaderHash(hdr *block.Header) ([]byte, error) {
	headerMarsh, err := bp.marshalizer.Marshal(hdr)
	if err != nil {
		return nil, err
	}

	headerHash := bp.hasher.Compute(string(headerMarsh))

	return headerHash, nil
}

func (bp *blockProcessor) displayLogInfo(
	header *block.Header,
	txBlock *block.TxBlockBody,
	headerHash []byte,
) {
	dispHeader, dispLines := createDisplayableHeaderAndBlockBody(header, txBlock)

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Error(err.Error())
	}

	tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\nTotal txs "+
		"processed until now: %d. Total txs processed for this block: %d. Total txs remained in pool: %d\n",
		toB64(headerHash),
		txsTotalProcessed,
		txsCurrentBlockProcessed,
		bp.getTxsFromPool(txBlock.ShardID))

	log.Info(tblString)
}

func createDisplayableHeaderAndBlockBody(
	header *block.Header,
	txBlockBody *block.TxBlockBody,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	lines := displayHeader(header)

	if header.BlockBodyType == block.TxBlock {
		lines = displayTxBlockBody(lines, txBlockBody, header.BlockBodyHash)

		return tableHeader, lines
	}

	//TODO: implement the other block bodies

	lines = append(lines, display.NewLineData(false, []string{"Unknown", "", ""}))
	return tableHeader, lines
}

func displayHeader(header *block.Header) []*display.LineData {
	lines := make([]*display.LineData, 0)

	//TODO really remove this mock prints
	aggrCommits := sha256.Sha256{}.Compute(string(sha256.Sha256{}.Compute(string(header.Commitment) + strconv.Itoa(int(header.Round)))))
	aggrSigs := sha256.Sha256{}.Compute(string(sha256.Sha256{}.Compute(string(aggrCommits))))

	lines = append(lines, display.NewLineData(false, []string{
		"Header",
		"Nonce",
		fmt.Sprintf("%d", header.Nonce)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Shard",
		fmt.Sprintf("%d", header.ShardId)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Epoch",
		fmt.Sprintf("%d", header.Epoch)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Round",
		fmt.Sprintf("%d", header.Round)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Timestamp",
		fmt.Sprintf("%d", header.TimeStamp)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev hash",
		toB64(header.PrevHash)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Body type",
		header.BlockBodyType.String()}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Body hash",
		toB64(header.BlockBodyHash)}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Pub keys bitmap",
		toHex(header.PubKeysBitmap)}))

	//TODO uncomment as this
	//lines = append(lines, display.NewLineData(false, []string{
	//	"",
	//	"Commitment",
	//	toB64(header.Commitment)}))
	//lines = append(lines, display.NewLineData(true, []string{
	//	"",
	//	"Signature",
	//	toB64(header.Signature)}))

	//TODO remove this
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Commitment",
		toB64(aggrCommits)}))
	lines = append(lines, display.NewLineData(true, []string{
		"",
		"Signature",
		toB64(aggrSigs)}))

	return lines
}

func displayTxBlockBody(lines []*display.LineData, txBlockBody *block.TxBlockBody, blockBodyHash []byte) []*display.LineData {
	lines = append(lines, display.NewLineData(false, []string{"TxBody", "Block blockBodyHash", toB64(blockBodyHash)}))
	lines = append(lines, display.NewLineData(true, []string{"", "Root blockBodyHash", toB64(txBlockBody.RootHash)}))

	txsCurrentBlockProcessed = 0

	for i := 0; i < len(txBlockBody.MiniBlocks); i++ {
		miniBlock := txBlockBody.MiniBlocks[i]

		part := fmt.Sprintf("TxBody_%d", miniBlock.ShardID)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<NIL> or <EMPTY>"}))
		}

		txsCurrentBlockProcessed += len(miniBlock.TxHashes)
		txsTotalProcessed += len(miniBlock.TxHashes)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if j == 0 || j >= len(miniBlock.TxHashes)-1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("Tx blockBodyHash %d", j+1),
					toB64(miniBlock.TxHashes[j])}))

				part = ""
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("..."),
					fmt.Sprintf("...")}))

				part = ""
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}

func toHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return "0x" + hex.EncodeToString(buff)
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

func sortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	mTxHashes := make(map[uint64][][]byte)
	mTransactions := make(map[uint64][]*transaction.Transaction)

	nonces := make([]uint64, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Get(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*transaction.Transaction)
		if !ok {
			continue
		}

		if mTxHashes[tx.Nonce] == nil {
			nonces = append(nonces, tx.Nonce)
			mTxHashes[tx.Nonce] = make([][]byte, 0)
			mTransactions[tx.Nonce] = make([]*transaction.Transaction, 0)
		}

		mTxHashes[tx.Nonce] = append(mTxHashes[tx.Nonce], key)
		mTransactions[tx.Nonce] = append(mTransactions[tx.Nonce], tx)
	}

	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i] < nonces[j]
	})

	for _, nonce := range nonces {
		keys := mTxHashes[nonce]

		for idx, key := range keys {
			txHashes = append(txHashes, key)
			transactions = append(transactions, mTransactions[nonce][idx])
		}
	}

	return transactions, txHashes, nil
}

func (bp *blockProcessor) displayTxsInfo(miniBlock *block.MiniBlock, shardId uint32) {
	if miniBlock == nil || miniBlock.TxHashes == nil {
		return
	}

	txsInPool := bp.getTxsFromPool(shardId)

	log.Info(fmt.Sprintf("PROCESS BLOCK TRANSACTION STARTED: Have %d txs in pool and need to process %d txs from the received block for shard id %d\n", txsInPool, len(miniBlock.TxHashes), shardId))
}

func (bp *blockProcessor) getTxsFromPool(shardId uint32) int {
	txPool := bp.dataPool.Transactions()

	if txPool == nil {
		return 0
	}

	txStore := txPool.ShardDataStore(shardId)

	if txStore == nil {
		return 0
	}

	return txStore.Len()
}

func getTxs(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Get(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*transaction.Transaction)
		if !ok {
			continue
		}

		txHashes = append(txHashes, key)
		transactions = append(transactions, tx)
	}

	return transactions, txHashes, nil
}
