package preprocess

import (
	"bytes"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
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
	orderedTxs           map[string][]data.TransactionHandler
	orderedTxHashes      map[string][][]byte
	mutOrderedTxs        sync.RWMutex
	blockTracker         BlockTracker
	blockType            block.Type
	addressConverter     state.AddressConverter
	accountsInfo         map[string]*txShardInfo
	mutAccountsInfo      sync.RWMutex
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
	gasHandler process.GasHandler,
	blockTracker BlockTracker,
	blockType block.Type,
	addressConverter state.AddressConverter,
) (*transactions, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(txDataPool) {
		return nil, process.ErrNilTransactionPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilTxStorage
	}
	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestTransaction == nil {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(addressConverter) {
		return nil, process.ErrNilAddressConverter
	}

	blockSizeComputation := NewBlockSizeComputation()

	bpp := basePreProcess{
		hasher:               hasher,
		marshalizer:          marshalizer,
		shardCoordinator:     shardCoordinator,
		gasHandler:           gasHandler,
		economicsFee:         economicsFee,
		blockSizeComputation: blockSizeComputation,
	}

	txs := transactions{
		basePreProcess:       &bpp,
		storage:              store,
		txPool:               txDataPool,
		onRequestTransaction: onRequestTransaction,
		txProcessor:          txProcessor,
		accounts:             accounts,
		blockTracker:         blockTracker,
		blockType:            blockType,
		addressConverter:     addressConverter,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterHandler(txs.receivedTransaction)

	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.accountsInfo = make(map[string]*txShardInfo)

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

	err := txs.removeDataFromPools(body, miniBlockPool, txs.txPool, txs.blockType)

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
			log.Debug("tx from mini block was not found in TransactionUnit",
				"sender shard ID", miniBlock.SenderShardID,
				"receiver shard ID", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes),
			)

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

		miniBlockHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, miniBlock)
		if err != nil {
			return txsRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock)

		txsRestored += len(miniBlock.TxHashes)
	}

	return txsRestored, nil
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(
	body block.Body,
	haveTime func() bool,
) error {
	//TODO: Should be analyzed if next commented check could be added here
	//if txs.blockType != block.TxBlock {
	//	return nil
	//}

	sortedTxsAndHashes, err := txs.computeTxsToMe(body)
	if err != nil {
		return err
	}

	if len(sortedTxsAndHashes) == 0 {
		sortedTxsAndHashes, err = txs.computeSortedTxsFromMe(body)
		if err != nil {
			return err
		}
	}

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	//TODO: Should be verified if gas computation should be done here (before the next for) and should be done
	//separately per miniblock, exactly as in the initial implementation. Also should be checked if with the last
	//change, which call this method two times, first for mini blocks with destination in self shard and afterward
	//with miniblocks from self shard, the gas computation is still correct.
	mapGasConsumedByMiniBlockInSenderShard := make(map[uint32]map[uint32]uint64)
	mapGasConsumedByMiniBlockInReceiverShard := make(map[uint32]map[uint32]uint64)
	totalGasConsumedInSelfShard := txs.gasHandler.TotalGasConsumed()

	for index := range sortedTxsAndHashes {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		txHandler := sortedTxsAndHashes[index].Tx
		tx := txHandler.(*transaction.Transaction)
		txHash := sortedTxsAndHashes[index].TxHash

		txs.txsForCurrBlock.mutTxsForBlock.RLock()
		txInfoFromMap := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
		if txInfoFromMap == nil || txInfoFromMap.tx == nil {
			txs.txsForCurrBlock.mutTxsForBlock.RUnlock()
			log.Debug("missing transaction in ProcessBlockTransactions", "txHash", txHash)
			return process.ErrMissingTransaction
		}
		senderShardID := txInfoFromMap.senderShardID
		receiverShardID := txInfoFromMap.receiverShardID
		txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

		err = txs.processAndRemoveBadTransaction(
			txHash,
			tx,
			senderShardID,
			receiverShardID,
		)

		if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
			return err
		}

		if _, ok := mapGasConsumedByMiniBlockInSenderShard[senderShardID]; !ok {
			mapGasConsumedByMiniBlockInSenderShard[senderShardID] = make(map[uint32]uint64)
		}
		if _, ok := mapGasConsumedByMiniBlockInReceiverShard[senderShardID]; !ok {
			mapGasConsumedByMiniBlockInReceiverShard[senderShardID] = make(map[uint32]uint64)
		}

		gasConsumedByMiniBlockInSenderShard := mapGasConsumedByMiniBlockInSenderShard[senderShardID][receiverShardID]
		gasConsumedByMiniBlockInReceiverShard := mapGasConsumedByMiniBlockInReceiverShard[senderShardID][receiverShardID]

		err = txs.computeGasConsumed(
			senderShardID,
			receiverShardID,
			tx,
			txHash,
			&gasConsumedByMiniBlockInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard,
			&totalGasConsumedInSelfShard)

		if err != nil {
			return err
		}

		mapGasConsumedByMiniBlockInSenderShard[senderShardID][receiverShardID] = gasConsumedByMiniBlockInSenderShard
		mapGasConsumedByMiniBlockInReceiverShard[senderShardID][receiverShardID] = gasConsumedByMiniBlockInReceiverShard
	}

	return nil
}

func (txs *transactions) computeTxsToMe(body block.Body) ([]*txcache.WrappedTransaction, error) {
	txsAndHashes := make([]*txcache.WrappedTransaction, 0, process.MaxItemsInBlock)

	for _, miniBlock := range block.MiniBlockSlice(body) {
		if miniBlock.Type != txs.blockType {
			continue
		}
		if miniBlock.SenderShardID == txs.shardCoordinator.SelfId() {
			continue
		}

		for i := 0; i < len(miniBlock.TxHashes); i++ {
			txHash := miniBlock.TxHashes[i]
			txs.txsForCurrBlock.mutTxsForBlock.RLock()
			txInfoFromMap := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
			txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

			if txInfoFromMap == nil || txInfoFromMap.tx == nil {
				log.Debug("missing transaction in computeTxsToMe", "type", miniBlock.Type, "txHash", txHash)
				return nil, process.ErrMissingTransaction
			}

			tx, ok := txInfoFromMap.tx.(*transaction.Transaction)
			if !ok {
				return nil, process.ErrWrongTypeAssertion
			}

			txsAndHashes = append(txsAndHashes, &txcache.WrappedTransaction{Tx: tx, TxHash: txHash})
		}
	}

	return txsAndHashes, nil
}

func (txs *transactions) computeSortedTxsFromMe(body block.Body) ([]*txcache.WrappedTransaction, error) {
	txsAndHashes := make([]*txcache.WrappedTransaction, 0, process.MaxItemsInBlock)

	for _, miniBlock := range block.MiniBlockSlice(body) {
		if miniBlock.Type != txs.blockType {
			continue
		}
		if miniBlock.SenderShardID != txs.shardCoordinator.SelfId() {
			continue
		}

		for i := 0; i < len(miniBlock.TxHashes); i++ {
			txHash := miniBlock.TxHashes[i]
			txs.txsForCurrBlock.mutTxsForBlock.RLock()
			txInfoFromMap := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
			txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

			if txInfoFromMap == nil || txInfoFromMap.tx == nil {
				log.Debug("missing transaction in computeSortedTxsFromMe", "type", miniBlock.Type, "txHash", txHash)
				return nil, process.ErrMissingTransaction
			}

			tx, ok := txInfoFromMap.tx.(*transaction.Transaction)
			if !ok {
				return nil, process.ErrWrongTypeAssertion
			}

			txsAndHashes = append(txsAndHashes, &txcache.WrappedTransaction{Tx: tx, TxHash: txHash})
		}
	}

	sortTransactionsBySenderAndNonce(txsAndHashes)
	return txsAndHashes, nil
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
	receivedAllMissing := txs.baseReceivedTransaction(txHash, &txs.txsForCurrBlock, txs.txPool, txs.blockType)

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
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.mutOrderedTxs.Unlock()

	txs.mutAccountsInfo.Lock()
	txs.accountsInfo = make(map[string]*txShardInfo)
	txs.mutAccountsInfo.Unlock()

	txs.blockSizeComputation.Init()
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
	txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: mbTxHashes.receiverShardID}
	for _, txHash := range mbTxHashes.txHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfoToSet}
	}
}

// computeMissingAndExistingTxsForShards calculates what transactions are available and what are missing from block.Body
func (txs *transactions) computeMissingAndExistingTxsForShards(body block.Body) map[uint32][]*txsHashesInfo {
	missingTxsForShard := txs.computeExistingAndMissing(
		body,
		&txs.txsForCurrBlock,
		txs.chRcvAllTxs,
		txs.blockType,
		txs.txPool)

	return missingTxsForShard
}

// processAndRemoveBadTransactions processed transactions, if txs are with error it removes them from pool
func (txs *transactions) processAndRemoveBadTransaction(
	transactionHash []byte,
	transaction *transaction.Transaction,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := txs.txProcessor.ProcessTransaction(transaction)
	isTxTargetedForDeletion := err == process.ErrLowerNonceInTransaction || errors.Is(err, process.ErrInsufficientFee)
	if isTxTargetedForDeletion {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txs.txPool.RemoveData(transactionHash, strCache)
	}

	txs.mutAccountsInfo.Lock()
	txs.accountsInfo[string(transaction.GetSndAddress())] = &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	txs.mutAccountsInfo.Unlock()

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		return err
	}

	txShardInfoToSet := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(transactionHash)] = &txInfo{tx: transaction, txShardInfo: txShardInfoToSet}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return err
}

func (txs *transactions) notifyTransactionProviderIfNeeded() {
	txs.mutAccountsInfo.RLock()
	for senderAddress, txShardInfo := range txs.accountsInfo {
		if txShardInfo.senderShardID != txs.shardCoordinator.SelfId() {
			continue
		}

		account, err := txs.getAccountForAddress([]byte(senderAddress))
		if err != nil {
			log.Debug("notifyTransactionProviderIfNeeded.getAccountForAddress", "error", err)
			continue
		}

		strCache := process.ShardCacherIdentifier(txShardInfo.senderShardID, txShardInfo.receiverShardID)
		txShardPool := txs.txPool.ShardDataStore(strCache)
		if txShardPool == nil {
			log.Trace("notifyTransactionProviderIfNeeded", "error", process.ErrNilTxDataPool)
			continue
		}
		if txShardPool.Len() == 0 {
			log.Trace("notifyTransactionProviderIfNeeded", "error", process.ErrEmptyTxDataPool)
			continue
		}

		sortedTransactionsProvider := createSortedTransactionsProvider(txs, txShardPool, strCache)
		sortedTransactionsProvider.NotifyAccountNonce([]byte(senderAddress), account.GetNonce())
	}
	txs.mutAccountsInfo.RUnlock()
}

func (txs *transactions) getAccountForAddress(address []byte) (state.AccountHandler, error) {
	addressContainer, err := txs.addressConverter.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	account, err := txs.accounts.GetAccountWithJournal(addressContainer)
	if err != nil {
		return nil, err
	}

	return account, nil
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
	if miniBlock.Type != txs.blockType {
		return nil
	}

	missingTransactions := make([][]byte, 0, len(miniBlock.TxHashes))
	searchFirst := txs.blockType == block.InvalidBlock

	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			txs.txPool,
			searchFirst)

		if tx == nil || tx.IsInterfaceNil() {
			missingTransactions = append(missingTransactions, txHash)
		}
	}

	return sliceUtil.TrimSliceSliceByte(missingTransactions)
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
	txsSlice := make([]*transaction.Transaction, 0, len(mb.TxHashes))
	txHashes := make([][]byte, 0, len(mb.TxHashes))
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
		txsSlice = append(txsSlice, tx)
	}

	return txsSlice, txHashes, nil
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the transactions added into the miniblocks
// as long as it has time
func (txs *transactions) CreateAndProcessMiniBlocks(haveTime func() bool) (block.MiniBlockSlice, error) {
	if txs.blockType != block.TxBlock {
		return make(block.MiniBlockSlice, 0), nil
	}

	timeBefore := time.Now()
	sortedTxs, err := txs.computeSortedTxs(txs.shardCoordinator.SelfId(), txs.shardCoordinator.SelfId())
	timeAfter := time.Now()
	if err != nil {
		log.Debug("computeSortedTxs", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	if !haveTime() {
		log.Debug("time is up ordering txs",
			"num txs", len(sortedTxs),
			"time [s]", timeAfter.Sub(timeBefore).Seconds(),
		)
		return make(block.MiniBlockSlice, 0), nil
	}

	log.Debug("time elapsed to ordered txs",
		"num txs", len(sortedTxs),
		"time [s]", timeAfter.Sub(timeBefore).Seconds(),
	)

	startTime := time.Now()
	miniBlocks, err := txs.createAndProcessMiniBlock(haveTime, txs.blockTracker.IsShardStuck, sortedTxs)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to createAndProcessMiniBlock",
		"time [s]", elapsedTime,
	)

	if err != nil {
		log.Debug("createAndProcessMiniBlock", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	return miniBlocks, nil
}

// CreateAndProcessMiniBlock creates and processes miniblocks
func (txs *transactions) createAndProcessMiniBlock(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, error) {
	log.Debug("createAndProcessMiniBlock has been started")

	mapMiniBlocks := make(map[uint32]*block.MiniBlock)
	num_txsAdded := 0
	num_txsBad := 0
	num_txsSkipped := 0
	totalTimeUsedForProcesss := time.Duration(0)
	totalTimeUsedForComputeGasConsumed := time.Duration(0)
	gasConsumedByMiniBlocksInSenderShard := uint64(0)
	mapGasConsumedByMiniBlockInReceiverShard := make(map[uint32]uint64)
	totalGasConsumedInSelfShard := txs.gasHandler.TotalGasConsumed()
	senderAddressToSkip := []byte("")

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	for index := range sortedTxs {
		if !haveTime() {
			log.Debug("time is out in createAndProcessMiniBlock")
			break
		}

		txHandler := sortedTxs[index].Tx
		tx := txHandler.(*transaction.Transaction)
		txHash := sortedTxs[index].TxHash
		senderShardID := sortedTxs[index].SenderShardID
		receiverShardID := sortedTxs[index].ReceiverShardID

		numNewMiniBlocks := 0
		if _, ok := mapMiniBlocks[receiverShardID]; !ok {
			numNewMiniBlocks = 1
		}
		if txs.blockSizeComputation.IsMaxBlockSizeReached(numNewMiniBlocks, 1) {
			log.Debug("max txs accepted in one block is reached",
				"num txs added", num_txsAdded,
				"total txs", len(sortedTxs))
			break
		}

		if isShardStuck != nil && isShardStuck(receiverShardID) {
			log.Trace("shard is stuck", "shard", receiverShardID)
			continue
		}

		if len(senderAddressToSkip) > 0 {
			if bytes.Equal(senderAddressToSkip, tx.GetSndAddress()) {
				num_txsSkipped++
				continue
			}
		}

		snapshot := txs.accounts.JournalLen()

		gasConsumedByMiniBlockInReceiverShard := mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
		oldGasConsumedByMiniBlocksInSenderShard := gasConsumedByMiniBlocksInSenderShard
		oldGasConsumedByMiniBlockInReceiverShard := gasConsumedByMiniBlockInReceiverShard
		oldTotalGasConsumedInSelfShard := totalGasConsumedInSelfShard
		startTime := time.Now()
		err := txs.computeGasConsumed(
			senderShardID,
			receiverShardID,
			tx,
			txHash,
			&gasConsumedByMiniBlocksInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard,
			&totalGasConsumedInSelfShard)
		elapsedTime := time.Since(startTime)
		totalTimeUsedForComputeGasConsumed += elapsedTime
		if err != nil {
			log.Debug("createAndProcessMiniBlock.computeGasConsumed", "error", err)
			continue
		}

		mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = gasConsumedByMiniBlockInReceiverShard

		// execute transaction to change the trie root hash
		startTime = time.Now()
		err = txs.processAndRemoveBadTransaction(
			txHash,
			tx,
			senderShardID,
			receiverShardID,
		)
		elapsedTime = time.Since(startTime)
		totalTimeUsedForProcesss += elapsedTime

		if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
			if err == process.ErrHigherNonceInTransaction {
				senderAddressToSkip = tx.GetSndAddress()
			}

			num_txsBad++
			log.Trace("bad tx",
				"error", err.Error(),
				"hash", txHash,
			)

			err = txs.accounts.RevertToSnapshot(snapshot)
			if err != nil {
				log.Debug("revert to snapshot", "error", err.Error())
			}

			txs.gasHandler.RemoveGasConsumed([][]byte{txHash})
			txs.gasHandler.RemoveGasRefunded([][]byte{txHash})

			gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
			mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = oldGasConsumedByMiniBlockInReceiverShard
			totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard
			continue
		}

		senderAddressToSkip = []byte("")

		gasRefunded := txs.gasHandler.GasRefunded(txHash)
		mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] -= gasRefunded
		if senderShardID == receiverShardID {
			gasConsumedByMiniBlocksInSenderShard -= gasRefunded
			totalGasConsumedInSelfShard -= gasRefunded
		}

		if errors.Is(err, process.ErrFailedTransaction) {
			continue
		}

		if _, ok := mapMiniBlocks[receiverShardID]; !ok {
			mapMiniBlocks[receiverShardID] = txs.createEmptyMiniBlock(senderShardID, receiverShardID, block.TxBlock)
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		mapMiniBlocks[receiverShardID].TxHashes = append(mapMiniBlocks[receiverShardID].TxHashes, txHash)
		txs.blockSizeComputation.AddNumTxs(1)
		num_txsAdded++
	}

	miniBlocks := txs.getMiniBlockSliceFromMap(mapMiniBlocks)

	for _, miniBlock := range miniBlocks {
		log.Debug("mini block info",
			"sender shard", miniBlock.SenderShardID,
			"gas consumed", gasConsumedByMiniBlocksInSenderShard,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed", mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID],
			"self shard", txs.shardCoordinator.SelfId(),
			"gas consumed", totalGasConsumedInSelfShard,
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("createAndProcessMiniBlock has been finished",
		"total txs", len(sortedTxs),
		"num txs added", num_txsAdded,
		"num txs bad", num_txsBad,
		"num txs skipped", num_txsSkipped,
		"computeGasConsumed used time", totalTimeUsedForComputeGasConsumed,
		"processAndRemoveBadTransaction used time", totalTimeUsedForProcesss)

	return miniBlocks, nil
}

func (txs *transactions) createEmptyMiniBlock(
	senderShardID uint32,
	receiverShardID uint32,
	blockType block.Type,
) *block.MiniBlock {

	miniBlock := &block.MiniBlock{}
	miniBlock.SenderShardID = senderShardID
	miniBlock.ReceiverShardID = receiverShardID
	miniBlock.TxHashes = make([][]byte, 0)
	miniBlock.Type = blockType

	return miniBlock
}

func (txs *transactions) getMiniBlockSliceFromMap(mapMiniBlocks map[uint32]*block.MiniBlock) block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		if miniBlock, ok := mapMiniBlocks[shardID]; ok {
			if len(miniBlock.TxHashes) > 0 {
				miniBlocks = append(miniBlocks, miniBlock)
			}
		}
	}

	if miniBlock, ok := mapMiniBlocks[sharding.MetachainShardId]; ok {
		if len(miniBlock.TxHashes) > 0 {
			miniBlocks = append(miniBlocks, miniBlock)
		}
	}

	return miniBlocks
}

func (txs *transactions) computeSortedTxs(
	sndShardId uint32,
	dstShardId uint32,
) ([]*txcache.WrappedTransaction, error) {
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txShardPool := txs.txPool.ShardDataStore(strCache)

	if txShardPool == nil {
		return nil, process.ErrNilTxDataPool
	}
	if txShardPool.Len() == 0 {
		return nil, process.ErrEmptyTxDataPool
	}

	sortedTransactionsProvider := createSortedTransactionsProvider(txs, txShardPool, strCache)
	log.Debug("computeSortedTxs.GetSortedTransactions")
	sortedTxs := sortedTransactionsProvider.GetSortedTransactions()

	sortTransactionsBySenderAndNonce(sortedTxs)
	return sortedTxs, nil
}

// ProcessMiniBlock processes all the transactions from a and saves the processed transactions in local cache complete miniblock
func (txs *transactions) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
) error {
	if txs.blockType != block.TxBlock {
		return nil
	}

	if miniBlock.Type != block.TxBlock {
		return process.ErrWrongTypeInMiniBlock
	}

	var err error

	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	if txs.blockSizeComputation.IsMaxBlockSizeReached(1, len(miniBlockTxs)) {
		return process.ErrMaxBlockSizeReached
	}

	processedTxHashes := make([][]byte, 0)

	defer func() {
		if err != nil {
			txs.gasHandler.RemoveGasConsumed(processedTxHashes)
			txs.gasHandler.RemoveGasRefunded(processedTxHashes)
		}
	}()

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)
	totalGasConsumedInSelfShard := txs.gasHandler.TotalGasConsumed()

	for index := range miniBlockTxs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		err = txs.computeGasConsumed(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockTxs[index],
			miniBlockTxHashes[index],
			&gasConsumedByMiniBlockInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard,
			&totalGasConsumedInSelfShard)

		if err != nil {
			return err
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])
	}

	for index := range miniBlockTxs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		err = txs.txProcessor.ProcessTransaction(miniBlockTxs[index])
		if err != nil {
			return err
		}
	}

	txShardInfoToSet := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockTxs[index], txShardInfo: txShardInfoToSet}
	}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	txs.blockSizeComputation.AddNumMiniBlocks(1)
	txs.blockSizeComputation.AddNumTxs(len(miniBlockTxs))

	return nil
}

// CreateMarshalizedData marshalizes transactions and creates and saves them into a new structure
func (txs *transactions) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs, err := txs.createMarshalizedData(txHashes, &txs.txsForCurrBlock)
	if err != nil {
		return nil, err
	}

	return mrsScrs, nil
}

// GetAllCurrentUsedTxs returns all the transactions used at current creation / processing
func (txs *transactions) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	txPool := make(map[string]data.TransactionHandler, len(txs.txsForCurrBlock.txHashAndInfo))

	txs.txsForCurrBlock.mutTxsForBlock.RLock()
	for txHash, txInfoFromMap := range txs.txsForCurrBlock.txHashAndInfo {
		txPool[txHash] = txInfoFromMap.tx
	}
	txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

	return txPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (txs *transactions) IsInterfaceNil() bool {
	return txs == nil
}

// sortTransactionsBySenderAndNonce sorts the provided transactions and hashes simultaneously
func sortTransactionsBySenderAndNonce(transactions []*txcache.WrappedTransaction) {
	sorter := func(i, j int) bool {
		txI := transactions[i].Tx
		txJ := transactions[j].Tx

		delta := bytes.Compare(txI.GetSndAddress(), txJ.GetSndAddress())
		if delta == 0 {
			delta = int(txI.GetNonce()) - int(txJ.GetNonce())
		}

		return delta < 0
	}

	sort.Slice(transactions, sorter)
}
