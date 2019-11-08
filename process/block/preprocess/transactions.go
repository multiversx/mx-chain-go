package preprocess

import (
	"bytes"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/block/preprocess")

// TODO: increase code coverage with unit tests

type transactions struct {
	*basePreProcess
	chRcvAllTxs          chan bool
	onRequestTransaction func(shardID uint32, txHashes [][]byte)
	txsForCurrBlock      txsForBlock
	txPool               dataRetriever.ShardedDataCacherNotifier
	storage              dataRetriever.StorageService
	txProcessor          process.TransactionProcessor
	accounts             state.AccountsAdapter
	orderedTxs           map[string][]*transaction.Transaction
	orderedTxHashes      map[string][][]byte
	mutOrderedTxs        sync.RWMutex
	economicsFee         process.FeeHandler
	miniBlocksCompacter  process.MiniBlocksCompacter
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
	economicsFee process.FeeHandler,
	miniBlocksCompacter process.MiniBlocksCompacter,
) (*transactions, error) {

	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if txDataPool == nil || txDataPool.IsInterfaceNil() {
		return nil, process.ErrNilTransactionPool
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilTxStorage
	}
	if txProcessor == nil || txProcessor.IsInterfaceNil() {
		return nil, process.ErrNilTxProcessor
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestTransaction == nil {
		return nil, process.ErrNilRequestHandler
	}
	if economicsFee == nil || economicsFee.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if miniBlocksCompacter == nil || miniBlocksCompacter.IsInterfaceNil() {
		return nil, process.ErrNilMiniBlocksCompacter
	}

	bpp := basePreProcess{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
	}

	txs := transactions{
		basePreProcess:       &bpp,
		storage:              store,
		txPool:               txDataPool,
		onRequestTransaction: onRequestTransaction,
		txProcessor:          txProcessor,
		accounts:             accounts,
		economicsFee:         economicsFee,
		miniBlocksCompacter:  miniBlocksCompacter,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterHandler(txs.receivedTransaction)

	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.orderedTxs = make(map[string][]*transaction.Transaction)
	txs.orderedTxHashes = make(map[string][][]byte)

	return &txs, nil
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
		log.Debug("requested missing txs",
			"num txs", requestedTxs)
		err := txs.waitForTxHashes(haveTime())
		txs.txsForCurrBlock.mutTxsForBlock.Lock()
		missingTxs := txs.txsForCurrBlock.missingTxs
		txs.txsForCurrBlock.missingTxs = 0
		txs.txsForCurrBlock.mutTxsForBlock.Unlock()
		log.Debug("received missing txs",
			"num txs", requestedTxs-missingTxs)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveTxBlockFromPools removes transactions and miniblocks from associated pools
func (txs *transactions) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil || body.IsInterfaceNil() {
		return process.ErrNilTxBlockBody
	}
	if miniBlockPool == nil || miniBlockPool.IsInterfaceNil() {
		return process.ErrNilMiniBlockPool
	}

	err := txs.removeDataFromPools(body, miniBlockPool, txs.txPool, block.TxBlock)

	return err
}

// RestoreTxBlockIntoPools restores the transactions and miniblocks to associated pools
func (txs *transactions) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
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

			err = txs.storage.GetStorer(dataRetriever.TransactionUnit).Remove([]byte(txHash))
			if err != nil {
				return txsRestored, err
			}
		}

		miniBlockHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, miniBlock)
		if err != nil {
			return txsRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock)

		err = txs.storage.GetStorer(dataRetriever.MiniBlockUnit).Remove(miniBlockHash)
		if err != nil {
			return txsRestored, err
		}

		txsRestored += len(miniBlock.TxHashes)
	}

	return txsRestored, nil
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(body block.Body, round uint64, haveTime func() bool) error {
	mapHashesAndTxs := txs.GetAllCurrentUsedTxs()
	expandedMiniBlocks, err := txs.miniBlocksCompacter.Expand(block.MiniBlockSlice(body), mapHashesAndTxs)
	if err != nil {
		return err
	}

	// basic validation already done in interceptors
	for i := 0; i < len(expandedMiniBlocks); i++ {
		miniBlock := expandedMiniBlocks[i]
		if miniBlock.Type != block.TxBlock {
			continue
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			txs.txsForCurrBlock.mutTxsForBlock.RLock()
			txInfo := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
			txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingTransaction
			}

			currTx, ok := txInfo.tx.(*transaction.Transaction)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err := txs.processAndRemoveBadTransaction(
				txHash,
				currTx,
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
		if miniBlock.Type != block.TxBlock {
			continue
		}

		err := txs.saveTxsToStorage(miniBlock.TxHashes, &txs.txsForCurrBlock, txs.storage, dataRetriever.TransactionUnit)
		if err != nil {
			return err
		}
	}

	return nil
}

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (txs *transactions) receivedTransaction(txHash []byte) {
	receivedAllMissing := txs.baseReceivedTransaction(txHash, &txs.txsForCurrBlock, txs.txPool)

	if receivedAllMissing {
		txs.chRcvAllTxs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created transactions at this round
func (txs *transactions) CreateBlockStarted() {
	_ = process.EmptyChannel(txs.chRcvAllTxs)

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.missingTxs = 0
	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	txs.mutOrderedTxs.Lock()
	txs.orderedTxs = make(map[string][]*transaction.Transaction)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.mutOrderedTxs.Unlock()
}

// RequestBlockTransactions request for transactions if missing from a block.Body
func (txs *transactions) RequestBlockTransactions(body block.Body) int {
	requestedTxs := 0
	missingTxsForShards := txs.computeMissingAndExistingTxsForShards(body)

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	for senderShardID, mbsTxHashes := range missingTxsForShards {
		for _, mbTxHashes := range mbsTxHashes {
			txs.setMissingTxsForShard(senderShardID, mbTxHashes)
		}
	}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	for senderShardID, mbsTxHashes := range missingTxsForShards {
		for _, mbTxHashes := range mbsTxHashes {
			requestedTxs += len(mbTxHashes.txHashes)
			txs.onRequestTransaction(senderShardID, mbTxHashes.txHashes)
		}
	}

	return requestedTxs
}

func (txs *transactions) setMissingTxsForShard(senderShardID uint32, mbTxHashes *txsHashesInfo) {
	txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: mbTxHashes.receiverShardID}
	for _, txHash := range mbTxHashes.txHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfo}
	}
}

// computeMissingAndExistingTxsForShards calculates what transactions are available and what are missing from block.Body
func (txs *transactions) computeMissingAndExistingTxsForShards(body block.Body) map[uint32][]*txsHashesInfo {
	missingTxsForShard := txs.computeExistingAndMissing(
		body,
		&txs.txsForCurrBlock,
		txs.chRcvAllTxs,
		block.TxBlock,
		txs.txPool)

	return missingTxsForShard
}

// processAndRemoveBadTransactions processed transactions, if txs are with error it removes them from pool
func (txs *transactions) processAndRemoveBadTransaction(
	transactionHash []byte,
	transaction *transaction.Transaction,
	round uint64,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := txs.txProcessor.ProcessTransaction(transaction, round)
	if err == process.ErrLowerNonceInTransaction ||
		err == process.ErrInsufficientFunds {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txs.txPool.RemoveData(transactionHash, strCache)
	}

	if err != nil {
		return err
	}

	txShardInfo := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(transactionHash)] = &txInfo{tx: transaction, txShardInfo: txShardInfo}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return nil
}

// RequestTransactionsForMiniBlock requests missing transactions for a certain miniblock
func (txs *transactions) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if miniBlock == nil {
		return 0
	}

	missingTxsForMiniBlock := txs.computeMissingTxsForMiniBlock(miniBlock)
	if len(missingTxsForMiniBlock) > 0 {
		txs.onRequestTransaction(miniBlock.SenderShardID, missingTxsForMiniBlock)
	}

	return len(missingTxsForMiniBlock)
}

// computeMissingTxsForMiniBlock computes missing transactions for a certain miniblock
func (txs *transactions) computeMissingTxsForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
	if miniBlock.Type != block.TxBlock {
		return nil
	}

	missingTransactions := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			txs.txPool)

		if tx == nil || tx.IsInterfaceNil() {
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

//TODO move this to smart contract address calculation component
func isSmartContractAddress(rcvAddress []byte) bool {
	isEmptyAddress := bytes.Equal(rcvAddress, make([]byte, len(rcvAddress)))
	if isEmptyAddress {
		return true
	}

	isSCAddress := bytes.Equal(rcvAddress[:(hooks.NumInitCharactersForScAddress-hooks.VMTypeLen)],
		make([]byte, hooks.NumInitCharactersForScAddress-hooks.VMTypeLen))
	if isSCAddress {
		return true
	}

	return false
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the transactions added into the miniblocks
// as long as it has time
func (txs *transactions) CreateAndProcessMiniBlocks(
	maxTxSpaceRemained uint32,
	maxMbSpaceRemained uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	newMBAdded := true
	txSpaceRemained := int(maxTxSpaceRemained)

	for newMBAdded {
		newMBAdded = false
		for shardId := uint32(0); shardId < txs.shardCoordinator.NumberOfShards(); shardId++ {
			if maxTxSpaceRemained <= 0 {
				break
			}

			mbSpaceRemained := int(maxMbSpaceRemained) - len(miniBlocks)
			if mbSpaceRemained <= 0 {
				break
			}

			miniBlock, err := txs.CreateAndProcessMiniBlock(
				txs.shardCoordinator.SelfId(),
				shardId,
				txSpaceRemained,
				haveTime,
				round)
			if err != nil {
				continue
			}

			if len(miniBlock.TxHashes) > 0 {
				txSpaceRemained -= len(miniBlock.TxHashes)
				miniBlocks = append(miniBlocks, miniBlock)
				newMBAdded = true
			}
		}
	}

	mapHashesAndTxs := txs.GetAllCurrentUsedTxs()
	compactedMiniBlocks := txs.miniBlocksCompacter.Compact(miniBlocks, mapHashesAndTxs)

	return compactedMiniBlocks, nil
}

// CreateAndProcessMiniBlock creates the miniblock from storage and processes the transactions added into the miniblock
func (txs *transactions) CreateAndProcessMiniBlock(
	sndShardId uint32,
	dstShardId uint32,
	spaceRemained int,
	haveTime func() bool,
	round uint64,
) (*block.MiniBlock, error) {

	var orderedTxs []*transaction.Transaction
	var orderedTxHashes [][]byte

	timeBefore := time.Now()
	orderedTxs, orderedTxHashes, err := txs.computeOrderedTxs(sndShardId, dstShardId)
	timeAfter := time.Now()

	if err != nil {
		log.Trace("computeOrderedTxs", "error", err.Error())
		return nil, err
	}

	if !haveTime() {
		log.Debug("time is up ordering txs",
			"num txs", len(orderedTxs),
			"time [s]", timeAfter.Sub(timeBefore).Seconds(),
		)
		return nil, process.ErrTimeIsOut
	}

	log.Trace("time elapsed to ordered txs,"+
		"num txs", len(orderedTxs),
		"time [s]", timeAfter.Sub(timeBefore).Seconds(),
	)

	miniBlock := &block.MiniBlock{}
	miniBlock.SenderShardID = sndShardId
	miniBlock.ReceiverShardID = dstShardId
	miniBlock.TxHashes = make([][]byte, 0)
	miniBlock.Type = block.TxBlock

	addedTxs := 0
	addedGasLimitPerCrossShardMiniblock := uint64(0)
	for index := range orderedTxs {
		if !haveTime() {
			break
		}

		if txs.isTxAlreadyProcessed(orderedTxHashes[index], &txs.txsForCurrBlock) {
			continue
		}

		currTxGasLimit := txs.economicsFee.ComputeGasLimit(orderedTxs[index])
		if isSmartContractAddress(orderedTxs[index].RcvAddr) {
			currTxGasLimit = orderedTxs[index].GasLimit
		}

		isGasLimitReached := addedGasLimitPerCrossShardMiniblock+currTxGasLimit > process.MaxGasLimitPerMiniBlock
		if isGasLimitReached {
			log.Trace("max gas limit per mini block is reached",
				"num added txs", len(miniBlock.TxHashes),
				"ordered txs", len(orderedTxs),
			)
			continue
		}

		snapshot := txs.accounts.JournalLen()

		// execute transaction to change the trie root hash
		err := txs.processAndRemoveBadTransaction(
			orderedTxHashes[index],
			orderedTxs[index],
			round,
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
		)

		if err != nil {
			log.Trace("bad tx",
				"error", err.Error(),
				"hash", display.ConvertHash(orderedTxHashes[index]),
			)
			err = txs.accounts.RevertToSnapshot(snapshot)
			if err != nil {
				log.Debug("revert to snapshot", "error", err.Error())
			}
			continue
		}

		miniBlock.TxHashes = append(miniBlock.TxHashes, orderedTxHashes[index])
		addedTxs++
		addedGasLimitPerCrossShardMiniblock += currTxGasLimit

		if addedTxs >= spaceRemained { // max transactions count in one block was reached
			log.Debug("max txs accepted in one block is reached",
				"num added txs", len(miniBlock.TxHashes),
				"total txs", len(orderedTxs),
			)
			return miniBlock, nil
		}
	}

	return miniBlock, nil
}

func (txs *transactions) computeOrderedTxs(
	sndShardId uint32,
	dstShardId uint32,
) ([]*transaction.Transaction, [][]byte, error) {

	var err error

	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txShardPool := txs.txPool.ShardDataStore(strCache)

	if txShardPool == nil {
		return nil, nil, process.ErrNilTxDataPool
	}
	if txShardPool.Len() == 0 {
		return nil, nil, process.ErrEmptyTxDataPool
	}

	txs.mutOrderedTxs.RLock()
	orderedTxs := txs.orderedTxs[strCache]
	orderedTxHashes := txs.orderedTxHashes[strCache]
	txs.mutOrderedTxs.RUnlock()

	alreadyOrdered := len(orderedTxs) > 0
	if !alreadyOrdered {
		orderedTxs, orderedTxHashes, err = SortTxByNonce(txShardPool)
		if err != nil {
			return nil, nil, err
		}

		log.Debug("creating mini blocks has been started",
			"have num txs", len(orderedTxs),
			"snd shard", sndShardId,
			"dest shard", dstShardId,
		)

		txs.mutOrderedTxs.Lock()
		txs.orderedTxs[strCache] = orderedTxs
		txs.orderedTxHashes[strCache] = orderedTxHashes
		txs.mutOrderedTxs.Unlock()
	}

	return orderedTxs, orderedTxHashes, nil
}

// ProcessMiniBlock processes all the transactions from a and saves the processed transactions in local cache complete miniblock
func (txs *transactions) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint64) error {
	if miniBlock.Type != block.TxBlock {
		return process.ErrWrongTypeInMiniBlock
	}

	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	for index := range miniBlockTxs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		err = txs.txProcessor.ProcessTransaction(miniBlockTxs[index], round)
		if err != nil {
			return err
		}
	}

	txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockTxs[index], txShardInfo: txShardInfo}
	}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return nil
}

// SortTxByNonce sort transactions according to nonces
func SortTxByNonce(txShardPool storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardPool == nil {
		return nil, nil, process.ErrNilTxDataPool
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	mTxHashes := make(map[uint64][][]byte)
	mTransactions := make(map[uint64][]*transaction.Transaction)

	nonces := make([]uint64, 0)

	for _, key := range txShardPool.Keys() {
		val, _ := txShardPool.Peek(key)
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
	mrsScrs, err := txs.createMarshalizedData(txHashes, &txs.txsForCurrBlock)
	if err != nil {
		return nil, err
	}

	return mrsScrs, nil
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
func (txs *transactions) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	txPool := make(map[string]data.TransactionHandler)

	txs.txsForCurrBlock.mutTxsForBlock.RLock()
	for txHash, txInfo := range txs.txsForCurrBlock.txHashAndInfo {
		txPool[txHash] = txInfo.tx
	}
	txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

	return txPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (txs *transactions) IsInterfaceNil() bool {
	if txs == nil {
		return true
	}
	return false
}
