package preprocess

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.DefaultLogger()

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx *transaction.Transaction
	*txShardInfo
	has bool
}

type txsHashesInfo struct {
	txHashes        [][]byte
	receiverShardID uint32
}

type transactions struct {
	chRcvAllTxs          chan bool
	onRequestTransaction func(shardID uint32, txHashes [][]byte)
	missingTxs           int
	mutTxsForBlock       sync.RWMutex
	txsForBlock          map[string]*txInfo
	txPool               dataRetriever.ShardedDataCacherNotifier
	storage              dataRetriever.StorageService
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	txProcessor          process.TransactionProcessor
	shardCoordinator     sharding.Coordinator
	accounts             state.AccountsAdapter
}

// NewTransactionPreprocessor creates a new transaction preprocessor object
func NewTransactionPreprocessor(
	txDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestTransaction func(shardID uint32, txHashes [][]byte),
) (*transactions, error) {

	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if txDataPool == nil {
		return nil, process.ErrNilTransactionPool
	}
	if store == nil {
		return nil, process.ErrNilTxStorage
	}
	if txProcessor == nil {
		return nil, process.ErrNilTxProcessor
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestTransaction == nil {
		return nil, process.ErrNilRequestHandler
	}

	txs := &transactions{
		hasher:               hasher,
		marshalizer:          marshalizer,
		shardCoordinator:     shardCoordinator,
		storage:              store,
		txPool:               txDataPool,
		onRequestTransaction: onRequestTransaction,
		txProcessor:          txProcessor,
		accounts:             accounts,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterHandler(txs.receivedTransaction)
	txs.txsForBlock = make(map[string]*txInfo)

	return txs, nil
}

// waitForTxHashes waits for a call whether all the requested transactions appeared
func (txs *transactions) waitForTxHashes(waitTime time.Duration) error {
	select {
	case <-txs.chRcvAllTxs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested transactions arrived and were saved into the pool
func (txs *transactions) IsDataPrepared(requestedTxs int, haveTime func() time.Duration) error {
	if requestedTxs > 0 {
		log.Info(fmt.Sprintf("requested %d missing txs\n", requestedTxs))
		err := txs.waitForTxHashes(haveTime())
		txs.mutTxsForBlock.Lock()
		missingTxs := txs.missingTxs
		txs.missingTxs = 0
		txs.mutTxsForBlock.Unlock()
		log.Info(fmt.Sprintf("received %d missing txs\n", requestedTxs-missingTxs))
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveTxBlockFromPools removes transactions and miniblocks from associated pools
func (txs *transactions) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil {
		return process.ErrNilTxBlockBody
	}
	if miniBlockPool == nil {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body); i++ {
		currentMiniBlock := body[i]
		strCache := process.ShardCacherIdentifier(currentMiniBlock.SenderShardID, currentMiniBlock.ReceiverShardID)
		txs.txPool.RemoveSetOfDataFromPool(currentMiniBlock.TxHashes, strCache)

		buff, err := txs.marshalizer.Marshal(currentMiniBlock)
		if err != nil {
			return err
		}

		miniBlockHash := txs.hasher.Compute(string(buff))
		miniBlockPool.Remove(miniBlockHash)
	}

	return nil
}

// RestoreTxBlockIntoPools restores the transactions and miniblocks to associated pools
func (txs *transactions) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockHashes map[int][]byte,
	miniBlockPool storage.Cacher,
) (int, error) {
	if miniBlockPool == nil {
		return 0, process.ErrNilMiniBlockPool
	}

	txsRestored := 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		txsBuff, err := txs.storage.GetAll(dataRetriever.TransactionUnit, miniBlock.TxHashes)
		if err != nil {
			return txsRestored, err
		}

		for txHash, txBuff := range txsBuff {
			tx := transaction.Transaction{}
			err = txs.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return txsRestored, err
			}

			txs.txPool.AddData([]byte(txHash), &tx, strCache)
		}

		buff, err := txs.marshalizer.Marshal(miniBlock)
		if err != nil {
			return txsRestored, err
		}

		miniBlockHash := txs.hasher.Compute(string(buff))
		miniBlockPool.Put(miniBlockHash, miniBlock)
		if miniBlock.SenderShardID != txs.shardCoordinator.SelfId() {
			miniBlockHashes[i] = miniBlockHash
		}

		txsRestored += len(miniBlock.TxHashes)
	}

	return txsRestored, nil
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(body block.Body, round uint32, haveTime func() time.Duration) error {
	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			txs.mutTxsForBlock.RLock()
			txInfo := txs.txsForBlock[string(txHash)]
			txs.mutTxsForBlock.RUnlock()
			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingTransaction
			}

			err := txs.processAndRemoveBadTransaction(
				txHash,
				txInfo.tx,
				round,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveTxBlockToStorage saves transactions from body into storage
func (txs *transactions) SaveTxBlockToStorage(body block.Body) error {
	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]

			txs.mutTxsForBlock.RLock()
			txInfo := txs.txsForBlock[string(txHash)]
			txs.mutTxsForBlock.RUnlock()

			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingTransaction
			}

			buff, err := txs.marshalizer.Marshal(txInfo.tx)
			if err != nil {
				return err
			}

			errNotCritical := txs.storage.Put(dataRetriever.TransactionUnit, txHash, buff)
			if errNotCritical != nil {
				log.Error(errNotCritical.Error())
			}
		}
	}

	return nil
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (txs *transactions) getTransactionFromPool(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
) *transaction.Transaction {
	strCache := process.ShardCacherIdentifier(senderShardID, destShardID)
	txStore := txs.txPool.ShardDataStore(strCache)
	if txStore == nil {
		log.Error(process.ErrNilStorage.Error())
		return nil
	}

	val, ok := txStore.Peek(txHash)
	if !ok {
		log.Debug(process.ErrTxNotFound.Error())
		return nil
	}

	tx, ok := val.(*transaction.Transaction)
	if !ok {
		log.Error(process.ErrInvalidTxInPool.Error())
		return nil
	}

	return tx
}

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (txs *transactions) receivedTransaction(txHash []byte) {
	txs.mutTxsForBlock.Lock()
	if txs.missingTxs > 0 {
		txInfoForHash := txs.txsForBlock[string(txHash)]
		if txInfoForHash != nil &&
			txInfoForHash.txShardInfo != nil &&
			!txInfoForHash.has {
			tx := txs.getTransactionFromPool(txInfoForHash.senderShardID, txInfoForHash.receiverShardID, txHash)
			if tx != nil {
				txs.txsForBlock[string(txHash)].tx = tx
				txs.txsForBlock[string(txHash)].has = true
				txs.missingTxs--
			}
		}

		missingTxs := txs.missingTxs
		txs.mutTxsForBlock.Unlock()

		if missingTxs == 0 {
			txs.chRcvAllTxs <- true
		}
	} else {
		txs.mutTxsForBlock.Unlock()
	}
}

// CreateBlockStarted cleans the local cache map for processed/created transactions at this round
func (txs *transactions) CreateBlockStarted() {
	txs.mutTxsForBlock.Lock()
	txs.txsForBlock = make(map[string]*txInfo)
	txs.mutTxsForBlock.Unlock()
}

// RequestBlockTransactions request for transactions if missing from a block.Body
func (txs *transactions) RequestBlockTransactions(body block.Body) int {
	txs.CreateBlockStarted()

	requestedTxs := 0
	missingTxsForShards := txs.computeMissingAndExistingTxsForShards(body)

	txs.mutTxsForBlock.Lock()
	for senderShardID, txsHashesInfo := range missingTxsForShards {
		txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: txsHashesInfo.receiverShardID}
		for _, txHash := range txsHashesInfo.txHashes {
			txs.txsForBlock[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfo, has: false}
		}
	}
	txs.mutTxsForBlock.Unlock()

	for senderShardID, txsHashesInfo := range missingTxsForShards {
		requestedTxs += len(txsHashesInfo.txHashes)
		txs.onRequestTransaction(senderShardID, txsHashesInfo.txHashes)
	}

	return requestedTxs
}

// computeMissingAndExistingTxsForShards calculates what transactions are available and what are missing from block.Body
func (txs *transactions) computeMissingAndExistingTxsForShards(body block.Body) map[uint32]*txsHashesInfo {
	missingTxsForShard := make(map[uint32]*txsHashesInfo)

	txs.mutTxsForBlock.Lock()

	txs.missingTxs = 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}
		txHashes := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := txs.getTransactionFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)

			if tx == nil {
				txHashes = append(txHashes, txHash)
				txs.missingTxs++
			} else {
				txs.txsForBlock[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfo, has: true}
			}
		}

		if len(txHashes) > 0 {
			missingTxsForShard[miniBlock.SenderShardID] = &txsHashesInfo{
				txHashes:        txHashes,
				receiverShardID: miniBlock.ReceiverShardID,
			}
		}
	}

	if txs.missingTxs > 0 {
		process.EmptyChannel(txs.chRcvAllTxs)
	}

	txs.mutTxsForBlock.Unlock()

	return missingTxsForShard
}

// processAndRemoveBadTransactions processed transactions, if txs are with error it removes them from pool
func (txs *transactions) processAndRemoveBadTransaction(
	transactionHash []byte,
	transaction *transaction.Transaction,
	round uint32,
	sndShardId uint32,
	dstShardId uint32,
) error {

	_, err := txs.txProcessor.ProcessTransaction(transaction, round)
	if err == process.ErrLowerNonceInTransaction ||
		err == process.ErrInsufficientFunds {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txs.txPool.RemoveData(transactionHash, strCache)
	}

	if err != nil {
		return err
	}

	txShardInfo := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	txs.mutTxsForBlock.Lock()
	txs.txsForBlock[string(transactionHash)] = &txInfo{tx: transaction, txShardInfo: txShardInfo, has: true}
	txs.mutTxsForBlock.Unlock()

	return nil
}

// RequestTransactionsForMiniBlock requests missing transactions for a certain miniblock
func (txs *transactions) RequestTransactionsForMiniBlock(mb block.MiniBlock) int {
	missingTxsForMiniBlock := txs.computeMissingTxsForMiniBlock(mb)
	txs.onRequestTransaction(mb.SenderShardID, missingTxsForMiniBlock)

	return len(missingTxsForMiniBlock)
}

// computeMissingTxsForMiniBlock computes missing transactions for a certain miniblock
func (txs *transactions) computeMissingTxsForMiniBlock(mb block.MiniBlock) [][]byte {
	missingTransactions := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
		tx := txs.getTransactionFromPool(mb.SenderShardID, mb.ReceiverShardID, txHash)
		if tx == nil {
			missingTransactions = append(missingTransactions, txHash)
		}
	}

	return missingTransactions
}

// getAllTxsFromMiniBlock gets all the transactions from a miniblock into a new structure
func (txs *transactions) getAllTxsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*transaction.Transaction, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := txs.txPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilTransactionPool
	}

	// verify if all transaction exists
	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			return nil, nil, process.ErrNilTransaction
		}

		tx, ok := tmp.(*transaction.Transaction)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
		txHashes = append(txHashes, txHash)
		transactions = append(transactions, tx)
	}

	return transactions, txHashes, nil
}

// CreateAndProcessMiniBlock creates the miniblock from storage and processes the transactions added into the miniblock
func (txs *transactions) CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint32) (*block.MiniBlock, error) {
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txStore := txs.txPool.ShardDataStore(strCache)

	timeBefore := time.Now()
	orderedTxes, orderedTxHashes, err := txs.getTxs(txStore)
	timeAfter := time.Now()

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after ordered %d txs in %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))
		return nil, process.ErrTimeIsOut
	}

	log.Info(fmt.Sprintf("time elapsed to ordered %d txs: %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))

	if err != nil {
		log.Debug(fmt.Sprintf("when trying to order txs: %s", err.Error()))
	}

	miniBlock := &block.MiniBlock{}
	miniBlock.SenderShardID = sndShardId
	miniBlock.ReceiverShardID = dstShardId
	miniBlock.TxHashes = make([][]byte, 0)
	log.Info(fmt.Sprintf("creating mini blocks has been started: have %d txs in pool for shard id %d\n", len(orderedTxes), miniBlock.ReceiverShardID))

	addedTxs := 0
	for index := range orderedTxes {
		if !haveTime() {
			break
		}

		snapshot := txs.accounts.JournalLen()

		// execute transaction to change the trie root hash
		err := txs.processAndRemoveBadTransaction(
			orderedTxHashes[index],
			orderedTxes[index],
			round,
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
		)

		if err != nil {
			log.Error(err.Error())
			err = txs.accounts.RevertToSnapshot(snapshot)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		miniBlock.TxHashes = append(miniBlock.TxHashes, orderedTxHashes[index])
		addedTxs++

		if addedTxs >= spaceRemained { // max transactions count in one block was reached
			log.Info(fmt.Sprintf("max txs accepted in one block is reached: added %d txs from %d txs\n", len(miniBlock.TxHashes), len(orderedTxes)))
			return miniBlock, nil
		}
	}

	return miniBlock, nil
}

// ProcessMiniBlock processes all the transactions from a and saves the processed transactions in local cache complete miniblock
func (txs *transactions) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint32) error {
	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	for index := range miniBlockTxs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		_, err = txs.txProcessor.ProcessTransaction(miniBlockTxs[index], round)
		if err != nil {
			return err
		}
	}

	txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	txs.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		txs.txsForBlock[string(txHash)] = &txInfo{tx: miniBlockTxs[index], txShardInfo: txShardInfo, has: true}
	}
	txs.mutTxsForBlock.Unlock()

	return nil
}

// SortTxByNonce sort transactions according to nonces
func SortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	mTxHashes := make(map[uint64][][]byte)
	mTransactions := make(map[uint64][]*transaction.Transaction)

	nonces := make([]uint64, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Peek(key)
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

// CreateMarshalizedData marshalizes transactions and creates and saves them into a new structure
func (txs *transactions) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsTxs := make([][]byte, 0)
	for _, txHash := range txHashes {
		txs.mutTxsForBlock.RLock()
		txInfo := txs.txsForBlock[string(txHash)]
		txs.mutTxsForBlock.RUnlock()

		if txInfo == nil || txInfo.tx == nil {
			continue
		}

		txMrs, err := txs.marshalizer.Marshal(txInfo.tx)
		if err != nil {
			log.Debug(process.ErrMarshalWithoutSuccess.Error())
			continue
		}
		mrsTxs = append(mrsTxs, txMrs)
	}

	return mrsTxs, nil
}

// getTxs gets all the available transactions from the pool
func (txs *transactions) getTxs(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Peek(key)
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

// GetAllCurrentUsedTxs returns all the transactions used at current creation / processing
func (txs *transactions) GetAllCurrentUsedTxs() map[string]*transaction.Transaction {
	txPool := make(map[string]*transaction.Transaction)

	txs.mutTxsForBlock.RLock()
	for txHash, txInfo := range txs.txsForBlock {
		txPool[txHash] = txInfo.tx
	}
	txs.mutTxsForBlock.RUnlock()

	return txPool
}
