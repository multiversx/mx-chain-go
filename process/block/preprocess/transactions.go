package preprocess

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.DataMarshalizer = (*transactions)(nil)
var _ process.PreProcessor = (*transactions)(nil)

var log = logger.GetOrCreate("process/block/preprocess")

type accountTxsShards struct {
	accountsInfo map[string]*txShardInfo
	sync.RWMutex
}

// TODO: increase code coverage with unit test

type transactions struct {
	*basePreProcess
	chRcvAllTxs                     chan bool
	onRequestTransaction            func(shardID uint32, txHashes [][]byte)
	txsForCurrBlock                 txsForBlock
	txPool                          dataRetriever.ShardedDataCacherNotifier
	storage                         dataRetriever.StorageService
	txProcessor                     process.TransactionProcessor
	orderedTxs                      map[string][]data.TransactionHandler
	orderedTxHashes                 map[string][][]byte
	mutOrderedTxs                   sync.RWMutex
	blockTracker                    BlockTracker
	blockType                       block.Type
	accountTxsShards                accountTxsShards
	emptyAddress                    []byte
	scheduledMiniBlocksEnableEpoch  uint32
	flagScheduledMiniBlocks         atomic.Flag
	txTypeHandler                   process.TxTypeHandler
	scheduledTxsExecutionHandler    process.ScheduledTxsExecutionHandler
	mixedTxsInMiniBlocksEnableEpoch uint32
	postProcessorTxsHandler         process.PostProcessorTxsHandler
	flagMixedTxsInMiniBlocks        atomic.Flag
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
	pubkeyConverter core.PubkeyConverter,
	blockSizeComputation BlockSizeComputationHandler,
	balanceComputation BalanceComputationHandler,
	epochNotifier process.EpochNotifier,
	scheduledMiniBlocksEnableEpoch uint32,
	txTypeHandler process.TxTypeHandler,
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler,
	mixedTxsInMiniBlocksEnableEpoch uint32,
	postProcessorTxsHandler process.PostProcessorTxsHandler,
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
	if check.IfNil(pubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(blockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(balanceComputation) {
		return nil, process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(epochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(scheduledTxsExecutionHandler) {
		return nil, process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(postProcessorTxsHandler) {
		return nil, process.ErrNilPostProcessorTxsHandler
	}

	bpp := basePreProcess{
		hasher:      hasher,
		marshalizer: marshalizer,
		gasTracker: gasTracker{
			shardCoordinator: shardCoordinator,
			gasHandler:       gasHandler,
			economicsFee:     economicsFee,
		},
		blockSizeComputation: blockSizeComputation,
		balanceComputation:   balanceComputation,
		accounts:             accounts,
		pubkeyConverter:      pubkeyConverter,
	}

	txs := &transactions{
		basePreProcess:                  &bpp,
		storage:                         store,
		txPool:                          txDataPool,
		onRequestTransaction:            onRequestTransaction,
		txProcessor:                     txProcessor,
		blockTracker:                    blockTracker,
		blockType:                       blockType,
		scheduledMiniBlocksEnableEpoch:  scheduledMiniBlocksEnableEpoch,
		txTypeHandler:                   txTypeHandler,
		scheduledTxsExecutionHandler:    scheduledTxsExecutionHandler,
		mixedTxsInMiniBlocksEnableEpoch: mixedTxsInMiniBlocksEnableEpoch,
		postProcessorTxsHandler:         postProcessorTxsHandler,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterOnAdded(txs.receivedTransaction)

	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.accountTxsShards.accountsInfo = make(map[string]*txShardInfo)

	txs.emptyAddress = make([]byte, txs.pubkeyConverter.Len())

	epochNotifier.RegisterNotifyHandler(txs)

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
		log.Debug("requested missing txs", "num txs", requestedTxs)
		err := txs.waitForTxHashes(haveTime())
		txs.txsForCurrBlock.mutTxsForBlock.Lock()
		missingTxs := txs.txsForCurrBlock.missingTxs
		txs.txsForCurrBlock.missingTxs = 0
		txs.txsForCurrBlock.mutTxsForBlock.Unlock()
		log.Debug("received missing txs", "num txs", requestedTxs-missingTxs, "requested", requestedTxs, "missing", missingTxs)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveMiniBlocksFromPools removes tx mini blocks from associated pools
func (txs *transactions) RemoveMiniBlocksFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	return txs.removeMiniBlocksFromPools(body, miniBlockPool, txs.isMiniBlockCorrect)
}

// RemoveTxsFromPools removes transactions from associated pools
func (txs *transactions) RemoveTxsFromPools(body *block.Body) error {
	return txs.removeTxsFromPools(body, txs.txPool, txs.isMiniBlockCorrect)
}

// RestoreMiniBlocksIntoPools restores the miniblocks to associated pool
func (txs *transactions) RestoreMiniBlocksIntoPools(body *block.Body, miniBlockPool storage.Cacher) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if !txs.isMiniBlockCorrect(miniBlock.Type) {
			continue
		}

		//TODO: Should be analyzed if restoring into pool only cross-shard miniblocks with destination in self shard,
		//would create problems or not
		if miniBlock.SenderShardID != txs.shardCoordinator.SelfId() {
			miniBlockHash, errHash := core.CalculateHash(txs.marshalizer, txs.hasher, miniBlock)
			if errHash != nil {
				return errHash
			}

			miniBlockPool.Put(miniBlockHash, miniBlock, miniBlock.Size())
		}
	}

	return nil
}

// RestoreTxsIntoPools restores the transactions to associated pool
func (txs *transactions) RestoreTxsIntoPools(body *block.Body) (int, error) {
	if check.IfNil(body) {
		return 0, process.ErrNilBlockBody
	}

	txsRestored := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if !txs.isMiniBlockCorrect(miniBlock.Type) {
			continue
		}

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

			txs.txPool.AddData([]byte(txHash), &tx, tx.Size(), strCache)
		}

		txsRestored += len(miniBlock.TxHashes)
	}

	return txsRestored, nil
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
	scheduledMode bool,
	gasConsumedInfo *process.GasConsumedInfo,
) error {
	if txs.isBodyToMe(body) {
		return txs.processTxsToMe(header, body, haveTime, scheduledMode, gasConsumedInfo)
	}

	if txs.isBodyFromMe(body) {
		return txs.processTxsFromMe(body, haveTime)
	}

	return process.ErrInvalidBody
}

func (txs *transactions) computeTxsToMe(body *block.Body) ([]*txcache.WrappedTransaction, error) {
	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}

	allTxs := make([]*txcache.WrappedTransaction, 0)
	for _, miniBlock := range body.MiniBlocks {
		shouldSkipMiniblock := miniBlock.SenderShardID == txs.shardCoordinator.SelfId() || !txs.isMiniBlockCorrect(miniBlock.Type)
		if shouldSkipMiniblock {
			continue
		}
		if miniBlock.Type != txs.blockType {
			return nil, fmt.Errorf("%w: block type: %s, sender shard id: %d, receiver shard id: %d",
				process.ErrInvalidMiniBlockType,
				miniBlock.Type,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID)
		}

		txsFromMiniBlock, err := txs.computeTxsFromMiniBlock(miniBlock)
		if err != nil {
			return nil, err
		}

		allTxs = append(allTxs, txsFromMiniBlock...)
	}

	return allTxs, nil
}

func (txs *transactions) computeTxsFromMe(body *block.Body) ([]*txcache.WrappedTransaction, error) {
	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}

	allTxs := make([]*txcache.WrappedTransaction, 0)
	for _, miniBlock := range body.MiniBlocks {
		shouldSkipMiniBlock := miniBlock.SenderShardID != txs.shardCoordinator.SelfId() ||
			!txs.isMiniBlockCorrect(miniBlock.Type) ||
			miniBlock.IsScheduledMiniBlock()
		if shouldSkipMiniBlock {
			continue
		}

		txsFromMiniBlock, err := txs.computeTxsFromMiniBlock(miniBlock)
		if err != nil {
			return nil, err
		}

		allTxs = append(allTxs, txsFromMiniBlock...)
	}

	return allTxs, nil
}

func (txs *transactions) computeScheduledTxsFromMe(body *block.Body) ([]*txcache.WrappedTransaction, error) {
	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}

	allScheduledTxs := make([]*txcache.WrappedTransaction, 0)
	for _, miniBlock := range body.MiniBlocks {
		shouldSkipMiniBlock := miniBlock.SenderShardID != txs.shardCoordinator.SelfId() ||
			miniBlock.Type != block.TxBlock ||
			!miniBlock.IsScheduledMiniBlock()
		if shouldSkipMiniBlock {
			continue
		}

		txsFromScheduledMiniBlock, err := txs.computeTxsFromMiniBlock(miniBlock)
		if err != nil {
			return nil, err
		}

		allScheduledTxs = append(allScheduledTxs, txsFromScheduledMiniBlock...)
	}

	return allScheduledTxs, nil
}

func (txs *transactions) computeTxsFromMiniBlock(miniBlock *block.MiniBlock) ([]*txcache.WrappedTransaction, error) {
	txsFromMiniBlock := make([]*txcache.WrappedTransaction, 0, len(miniBlock.TxHashes))
	for i := 0; i < len(miniBlock.TxHashes); i++ {
		txHash := miniBlock.TxHashes[i]
		txs.txsForCurrBlock.mutTxsForBlock.RLock()
		txInfoFromMap, ok := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
		txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

		if !ok || check.IfNil(txInfoFromMap.tx) {
			log.Warn("missing transaction in computeTxsFromMiniBlock", "type", miniBlock.Type, "txHash", txHash)
			return nil, process.ErrMissingTransaction
		}

		_, ok = txInfoFromMap.tx.(*transaction.Transaction)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		calculatedSenderShardId, err := txs.getShardFromAddress(txInfoFromMap.tx.GetSndAddr())
		if err != nil {
			return nil, err
		}

		calculatedReceiverShardId, err := txs.getShardFromAddress(txInfoFromMap.tx.GetRcvAddr())
		if err != nil {
			return nil, err
		}

		wrappedTx := &txcache.WrappedTransaction{
			Tx:              txInfoFromMap.tx,
			TxHash:          txHash,
			SenderShardID:   calculatedSenderShardId,
			ReceiverShardID: calculatedReceiverShardId,
		}

		txsFromMiniBlock = append(txsFromMiniBlock, wrappedTx)
	}

	return txsFromMiniBlock, nil
}

func (txs *transactions) getShardFromAddress(address []byte) (uint32, error) {
	isEmptyAddress := bytes.Equal(address, txs.emptyAddress)
	if isEmptyAddress {
		return txs.shardCoordinator.SelfId(), nil
	}

	return txs.shardCoordinator.ComputeId(address), nil
}

func (txs *transactions) processTxsToMe(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
	scheduledMode bool,
	gasConsumedInfo *process.GasConsumedInfo,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}
	if check.IfNil(header) {
		return process.ErrNilHeaderHandler
	}

	txsToMe, err := txs.computeTxsToMe(body)
	if err != nil {
		return err
	}

	log.Trace("processTxsToMe", "scheduled mode", scheduledMode, "totalGasConsumedInSelfShard", gasConsumedInfo.TotalGasConsumedInSelfShard)

	for index := range txsToMe {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		tx, ok := txsToMe[index].Tx.(*transaction.Transaction)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		txHash := txsToMe[index].TxHash
		senderShardID := txsToMe[index].SenderShardID
		receiverShardID := txsToMe[index].ReceiverShardID

		txs.saveAccountBalanceForAddress(tx.GetRcvAddr())

		if !scheduledMode {
			err = txs.processAndRemoveBadTransaction(
				txHash,
				tx,
				senderShardID,
				receiverShardID)
			if err != nil {
				return err
			}
		}

		gasConsumedByTxInSelfShard, errComputeGas := txs.computeGasConsumed(
			senderShardID,
			receiverShardID,
			tx,
			txHash,
			gasConsumedInfo)
		if errComputeGas != nil {
			return errComputeGas
		}

		if scheduledMode {
			txs.gasHandler.SetGasConsumedAsScheduled(gasConsumedByTxInSelfShard, txHash)
			txs.scheduledTxsExecutionHandler.Add(txHash, tx)
		} else {
			txs.gasHandler.SetGasConsumed(gasConsumedByTxInSelfShard, txHash)
		}
	}

	return nil
}

func (txs *transactions) processTxsFromMe(
	body *block.Body,
	haveTime func() bool,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	txsFromMe, err := txs.computeTxsFromMe(body)
	if err != nil {
		return err
	}

	sortTransactionsBySenderAndNonce(txsFromMe)

	isShardStuckFalse := func(uint32) bool {
		return false
	}
	isMaxBlockSizeReachedFalse := func(int, int) bool {
		return false
	}
	haveAdditionalTimeFalse := func() bool {
		return false
	}

	calculatedMiniBlocks, mapSCTxs, err := txs.createAndProcessMiniBlocksFromMe(
		haveTime,
		isShardStuckFalse,
		isMaxBlockSizeReachedFalse,
		txsFromMe,
	)
	if err != nil {
		return err
	}

	if !haveTime() {
		return process.ErrTimeIsOut
	}

	scheduledMiniBlocks, err := txs.createAndProcessScheduledMiniBlocksFromMeAsValidator(
		body,
		haveTime,
		haveAdditionalTimeFalse,
		isShardStuckFalse,
		isMaxBlockSizeReachedFalse,
		mapSCTxs,
	)
	if err != nil {
		return err
	}

	calculatedMiniBlocks = append(calculatedMiniBlocks, scheduledMiniBlocks...)

	receivedMiniBlocks := make(block.MiniBlockSlice, 0)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type == block.InvalidBlock {
			continue
		}

		receivedMiniBlocks = append(receivedMiniBlocks, miniBlock)
	}

	receivedBodyHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, &block.Body{MiniBlocks: receivedMiniBlocks})
	if err != nil {
		return err
	}

	calculatedBodyHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, &block.Body{MiniBlocks: calculatedMiniBlocks})
	if err != nil {
		return err
	}

	if !bytes.Equal(receivedBodyHash, calculatedBodyHash) {
		for _, mb := range receivedMiniBlocks {
			log.Debug("received miniblock", "type", mb.Type, "sender", mb.SenderShardID, "receiver", mb.ReceiverShardID, "numTxs", len(mb.TxHashes))
		}

		for _, mb := range calculatedMiniBlocks {
			log.Debug("calculated miniblock", "type", mb.Type, "sender", mb.SenderShardID, "receiver", mb.ReceiverShardID, "numTxs", len(mb.TxHashes))
		}

		log.Debug("block body missmatch",
			"received body hash", receivedBodyHash,
			"calculated body hash", calculatedBodyHash)
		return process.ErrBlockBodyHashMismatch
	}

	return nil
}

func (txs *transactions) createAndProcessScheduledMiniBlocksFromMeAsValidator(
	body *block.Body,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	mapSCTxs map[string]struct{},
) (block.MiniBlockSlice, error) {

	if !txs.flagScheduledMiniBlocks.IsSet() {
		return make(block.MiniBlockSlice, 0), nil
	}

	scheduledTxsFromMe, err := txs.computeScheduledTxsFromMe(body)
	if err != nil {
		return nil, err
	}

	sortTransactionsBySenderAndNonce(scheduledTxsFromMe)

	scheduledMiniBlocks := txs.createScheduledMiniBlocks(
		haveTime,
		haveAdditionalTime,
		isShardStuck,
		isMaxBlockSizeReached,
		scheduledTxsFromMe,
		mapSCTxs,
	)

	if !haveTime() && !haveAdditionalTime() {
		return nil, process.ErrTimeIsOut
	}

	return scheduledMiniBlocks, nil
}

// SaveTxsToStorage saves transactions from body into storage
func (txs *transactions) SaveTxsToStorage(body *block.Body) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if !txs.isMiniBlockCorrect(miniBlock.Type) {
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
func (txs *transactions) receivedTransaction(key []byte, value interface{}) {
	wrappedTx, ok := value.(*txcache.WrappedTransaction)
	if !ok {
		log.Warn("transactions.receivedTransaction", "error", process.ErrWrongTypeAssertion)
		return
	}

	receivedAllMissing := txs.baseReceivedTransaction(key, wrappedTx.Tx, &txs.txsForCurrBlock)

	if receivedAllMissing {
		txs.chRcvAllTxs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created transactions at this round
func (txs *transactions) CreateBlockStarted() {
	_ = core.EmptyChannel(txs.chRcvAllTxs)

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.missingTxs = 0
	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	txs.mutOrderedTxs.Lock()
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.mutOrderedTxs.Unlock()

	txs.accountTxsShards.Lock()
	txs.accountTxsShards.accountsInfo = make(map[string]*txShardInfo)
	txs.accountTxsShards.Unlock()

	txs.scheduledTxsExecutionHandler.Init()
}

// RequestBlockTransactions request for transactions if missing from a block.Body
func (txs *transactions) RequestBlockTransactions(body *block.Body) int {
	if check.IfNil(body) {
		return 0
	}

	return txs.computeExistingAndRequestMissingTxsForShards(body)
}

// computeExistingAndRequestMissingTxsForShards calculates what transactions are available and requests
// what are missing from block.Body
func (txs *transactions) computeExistingAndRequestMissingTxsForShards(body *block.Body) int {
	numMissingTxsForShard := txs.computeExistingAndRequestMissing(
		body,
		&txs.txsForCurrBlock,
		txs.chRcvAllTxs,
		txs.isMiniBlockCorrect,
		txs.txPool,
		txs.onRequestTransaction,
	)

	return numMissingTxsForShard
}

// processAndRemoveBadTransactions processed transactions, if txs are with error it removes them from pool
func (txs *transactions) processAndRemoveBadTransaction(
	txHash []byte,
	tx *transaction.Transaction,
	sndShardId uint32,
	dstShardId uint32,
) error {

	_, err := txs.txProcessor.ProcessTransaction(tx)
	isTxTargetedForDeletion := errors.Is(err, process.ErrLowerNonceInTransaction) || errors.Is(err, process.ErrInsufficientFee)
	if isTxTargetedForDeletion {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txs.txPool.RemoveData(txHash, strCache)
	}

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		return err
	}

	txShardInfoToSet := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoToSet}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return err
}

func (txs *transactions) notifyTransactionProviderIfNeeded() {
	txs.accountTxsShards.RLock()
	for senderAddress, txShardInfoValue := range txs.accountTxsShards.accountsInfo {
		if txShardInfoValue.senderShardID != txs.shardCoordinator.SelfId() {
			continue
		}

		account, err := txs.getAccountForAddress([]byte(senderAddress))
		if err != nil {
			log.Debug("notifyTransactionProviderIfNeeded.getAccountForAddress", "error", err)
			continue
		}

		strCache := process.ShardCacherIdentifier(txShardInfoValue.senderShardID, txShardInfoValue.receiverShardID)
		txShardPool := txs.txPool.ShardDataStore(strCache)
		if check.IfNil(txShardPool) {
			log.Trace("notifyTransactionProviderIfNeeded", "error", process.ErrNilTxDataPool)
			continue
		}

		sortedTransactionsProvider := createSortedTransactionsProvider(txShardPool)
		sortedTransactionsProvider.NotifyAccountNonce([]byte(senderAddress), account.GetNonce())
	}
	txs.accountTxsShards.RUnlock()
}

func (txs *transactions) getAccountForAddress(address []byte) (vmcommon.AccountHandler, error) {
	account, err := txs.accounts.GetExistingAccount(address)
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
	haveAdditionalTime func() bool,
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
		if !haveTime() && !haveAdditionalTime() {
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
	startTime := time.Now()
	sortedTxs, err := txs.computeSortedTxs(txs.shardCoordinator.SelfId(), txs.shardCoordinator.SelfId())
	elapsedTime := time.Since(startTime)
	if err != nil {
		log.Debug("computeSortedTxs", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	if len(sortedTxs) == 0 {
		log.Trace("no transaction found after computeSortedTxs",
			"time [s]", elapsedTime,
		)
		return make(block.MiniBlockSlice, 0), nil
	}

	if !haveTime() {
		log.Debug("time is up after computeSortedTxs",
			"num txs", len(sortedTxs),
			"time [s]", elapsedTime,
		)
		return make(block.MiniBlockSlice, 0), nil
	}

	log.Debug("elapsed time to computeSortedTxs",
		"num txs", len(sortedTxs),
		"time [s]", elapsedTime,
	)

	startTime = time.Now()
	miniBlocks, mapSCTxs, err := txs.createAndProcessMiniBlocksFromMe(
		haveTime,
		txs.blockTracker.IsShardStuck,
		txs.blockSizeComputation.IsMaxBlockSizeReached,
		sortedTxs,
	)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to createAndProcessMiniBlocksFromMe",
		"time [s]", elapsedTime,
	)

	if err != nil {
		log.Debug("createAndProcessMiniBlocksFromMe", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	haveAdditionalTime := process.HaveAdditionalTime()
	scheduledMiniBlocks, err := txs.createAndProcessScheduledMiniBlocksFromMeAsProposer(
		haveTime,
		haveAdditionalTime,
		sortedTxs,
		mapSCTxs,
	)
	if err != nil {
		log.Debug("createAndProcessScheduledMiniBlocksFromMeAsProposer", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	miniBlocks = append(miniBlocks, scheduledMiniBlocks...)

	return miniBlocks, nil
}

func (txs *transactions) createAndProcessScheduledMiniBlocksFromMeAsProposer(
	haveTime func() bool,
	haveAdditionalTime func() bool,
	sortedTxs []*txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
) (block.MiniBlockSlice, error) {

	if !txs.flagScheduledMiniBlocks.IsSet() {
		return make(block.MiniBlockSlice, 0), nil
	}

	startTime := time.Now()
	scheduledMiniBlocks := txs.createScheduledMiniBlocks(
		haveTime,
		haveAdditionalTime,
		txs.blockTracker.IsShardStuck,
		txs.blockSizeComputation.IsMaxBlockSizeReached,
		sortedTxs,
		mapSCTxs,
	)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to createScheduledMiniBlocks",
		"time [s]", elapsedTime,
	)

	return scheduledMiniBlocks, nil
}

func (txs *transactions) createAndProcessMiniBlocksFromMeV1(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, error) {
	log.Debug("createAndProcessMiniBlocksFromMeV1 has been started")

	args := miniBlocksBuilderArgs{
		gasTracker:                txs.gasTracker,
		accounts:                  txs.accounts,
		accountTxsShards:          &txs.accountTxsShards,
		balanceComputationHandler: txs.balanceComputation,
		blockSizeComputation:      txs.blockSizeComputation,
		haveTime:                  haveTime,
		isShardStuck:              isShardStuck,
		isMaxBlockSizeReached:     isMaxBlockSizeReached,
		getTxMaxTotalCost:         getTxMaxTotalCost,
	}

	mbBuilder, err := newMiniBlockBuilder(args)
	if err != nil {
		return nil, err
	}

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	for _, wtx := range sortedTxs {
		canAddTx, canContinue, tx := mbBuilder.checkAddTransaction(wtx)
		if !canContinue {
			break
		}
		if !canAddTx {
			continue
		}

		err = txs.processMiniBlockBuilderTx(mbBuilder, wtx, tx)
		if err != nil {
			continue
		}

		mbBuilder.addTxAndUpdateBlockSize(tx, wtx)
	}

	miniBlocks := txs.getMiniBlockSliceFromMap(mbBuilder.miniBlocks)

	logCreatedMiniBlocksStats(miniBlocks, txs.shardCoordinator.SelfId(), mbBuilder, len(sortedTxs))

	return miniBlocks, nil
}

func logCreatedMiniBlocksStats(
	miniBlocks block.MiniBlockSlice,
	selfShardID uint32,
	mbb *miniBlocksBuilder,
	nbSortedTxs int,
) {
	log.Debug("createAndProcessMiniBlocksFromMeV1",
		"self shard", selfShardID,
		"gas consumed in sender shard", mbb.gasInfo.GasConsumedByMiniBlocksInSenderShard,
		"total gas consumed in self shard", mbb.gasInfo.TotalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard", mbb.gasConsumedInReceiverShard[miniBlock.ReceiverShardID],
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("createAndProcessMiniBlocksFromMeV1 has been finished",
		"total txs", nbSortedTxs,
		"num txs added", mbb.stats.numTxsAdded,
		"num txs bad", mbb.stats.numTxsBad,
		"num txs failed", mbb.stats.numTxsFailed,
		"num txs skipped", mbb.stats.numTxsSkipped,
		"num txs with initial balance consumed", mbb.stats.numTxsWithInitialBalanceConsumed,
		"num cross shard sc calls or special txs", mbb.stats.numCrossShardSCCallsOrSpecialTxs,
		"used time for computeGasConsumed", mbb.stats.totalGasComputeTime,
		"used time for processAndRemoveBadTransaction", mbb.stats.totalProcessingTime)
}

func (txs *transactions) processMiniBlockBuilderTx(
	mb *miniBlocksBuilder,
	wtx *txcache.WrappedTransaction,
	tx *transaction.Transaction,
) error {
	snapshot := mb.accounts.JournalLen()
	startTime := time.Now()
	err := txs.processAndRemoveBadTransaction(
		wtx.TxHash,
		tx,
		wtx.SenderShardID,
		wtx.ReceiverShardID,
	)
	elapsedTime := time.Since(startTime)
	mb.stats.totalProcessingTime += elapsedTime
	mb.updateAccountShardsInfo(tx, wtx)

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		txs.handleBadTransaction(err, wtx, tx, mb, snapshot)
		return err
	}

	mb.senderToSkip = []byte("")

	txs.refundGas(wtx, mb)

	if errors.Is(err, process.ErrFailedTransaction) {
		mb.handleFailedTransaction()
		return err
	}

	return nil
}

func (txs *transactions) handleBadTransaction(
	err error,
	wtx *txcache.WrappedTransaction,
	tx *transaction.Transaction,
	mbb *miniBlocksBuilder,
	snapshot int,
) {
	log.Trace("bad tx", "error", err.Error(), "hash", wtx.TxHash)
	errRevert := txs.accounts.RevertToSnapshot(snapshot)
	if errRevert != nil {
		log.Warn("revert to snapshot", "error", err.Error())
	}

	mbb.handleBadTransaction(err, wtx, tx)
}

func (txs *transactions) refundGas(
	wtx *txcache.WrappedTransaction,
	mbb *miniBlocksBuilder,
) {
	gasRefunded := txs.gasHandler.GasRefunded(wtx.TxHash)
	mbb.handleGasRefund(wtx, gasRefunded)
}

func (txs *transactions) createEmptyMiniBlock(
	senderShardID uint32,
	receiverShardID uint32,
	blockType block.Type,
	isScheduledMiniBlock bool,
) *block.MiniBlock {

	miniBlock := &block.MiniBlock{
		Type:            blockType,
		SenderShardID:   senderShardID,
		ReceiverShardID: receiverShardID,
		TxHashes:        make([][]byte, 0),
	}

	if txs.flagScheduledMiniBlocks.IsSet() {
		executionType := block.Normal
		if isScheduledMiniBlock {
			executionType = block.Scheduled
		}
		_ = miniBlock.SetMiniBlockReserved(&block.MiniBlockReserved{ExecutionType: executionType})
	}

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

	if miniBlock, ok := mapMiniBlocks[core.MetachainShardId]; ok {
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

	if check.IfNil(txShardPool) {
		return nil, process.ErrNilTxDataPool
	}

	sortedTransactionsProvider := createSortedTransactionsProvider(txShardPool)
	log.Debug("computeSortedTxs.GetSortedTransactions")
	sortedTxs := sortedTransactionsProvider.GetSortedTransactions()

	sortTransactionsBySenderAndNonce(sortedTxs)
	return sortedTxs, nil
}

// ProcessMiniBlock processes all the transactions from a and saves the processed transactions in local cache complete miniblock
func (txs *transactions) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	getNumOfCrossInterMbsAndTxs func() (int, int),
	scheduledMode bool,
	isNewMiniBlock bool,
	gasConsumedInfo *process.GasConsumedInfo,
) ([][]byte, int, error) {

	if miniBlock.Type != block.TxBlock {
		return nil, 0, process.ErrWrongTypeInMiniBlock
	}

	numOfNewMiniBlocks := 0
	if isNewMiniBlock {
		numOfNewMiniBlocks = 1
	}

	var err error
	processedTxHashes := make([][]byte, 0)
	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime, haveAdditionalTime)
	if err != nil {
		return nil, 0, err
	}

	if txs.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(numOfNewMiniBlocks, len(miniBlockTxs)) {
		return nil, 0, process.ErrMaxBlockSizeReached
	}

	defer func() {
		if err != nil {
			if scheduledMode {
				txs.gasHandler.RemoveGasConsumedAsScheduled(processedTxHashes)
			} else {
				txs.gasHandler.RemoveGasConsumed(processedTxHashes)
				txs.gasHandler.RemoveGasRefunded(processedTxHashes)
			}
		}
	}()

	log.Trace("transactions.ProcessMiniBlock", "scheduled mode", scheduledMode,
		"isNewMiniBlock", isNewMiniBlock,
		"totalGasConsumedInSelfShard", gasConsumedInfo.TotalGasConsumedInSelfShard)

	for index := range miniBlockTxs {
		if !haveTime() && !haveAdditionalTime() {
			err = process.ErrTimeIsOut
			return processedTxHashes, 0, err
		}

		gasConsumedByTxInSelfShard, errComputeGas := txs.computeGasConsumed(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockTxs[index],
			miniBlockTxHashes[index],
			gasConsumedInfo)
		if errComputeGas != nil {
			return processedTxHashes, 0, errComputeGas
		}

		if scheduledMode {
			txs.gasHandler.SetGasConsumedAsScheduled(gasConsumedByTxInSelfShard, miniBlockTxHashes[index])
		} else {
			txs.gasHandler.SetGasConsumed(gasConsumedByTxInSelfShard, miniBlockTxHashes[index])
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])
	}

	numOfOldCrossInterMbs, numOfOldCrossInterTxs := getNumOfCrossInterMbsAndTxs()

	for index := range miniBlockTxs {
		if !haveTime() && !haveAdditionalTime() {
			err = process.ErrTimeIsOut
			return processedTxHashes, index, err
		}

		txs.saveAccountBalanceForAddress(miniBlockTxs[index].GetRcvAddr())

		if scheduledMode {
			continue
		}

		_, err = txs.txProcessor.ProcessTransaction(miniBlockTxs[index])
		if err != nil {
			return processedTxHashes, index, err
		}
	}

	numOfCrtCrossInterMbs, numOfCrtCrossInterTxs := getNumOfCrossInterMbsAndTxs()
	numOfNewCrossInterMbs := numOfCrtCrossInterMbs - numOfOldCrossInterMbs
	numOfNewCrossInterTxs := numOfCrtCrossInterTxs - numOfOldCrossInterTxs

	log.Trace("transactions.ProcessMiniBlock",
		"scheduled mode", scheduledMode,
		"numOfOldCrossInterMbs", numOfOldCrossInterMbs, "numOfOldCrossInterTxs", numOfOldCrossInterTxs,
		"numOfCrtCrossInterMbs", numOfCrtCrossInterMbs, "numOfCrtCrossInterTxs", numOfCrtCrossInterTxs,
		"numOfNewCrossInterMbs", numOfNewCrossInterMbs, "numOfNewCrossInterTxs", numOfNewCrossInterTxs,
	)

	numMiniBlocks := numOfNewMiniBlocks + numOfNewCrossInterMbs
	numTxs := len(miniBlockTxs) + numOfNewCrossInterTxs
	if txs.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(numMiniBlocks, numTxs) {
		return processedTxHashes, len(processedTxHashes), process.ErrMaxBlockSizeReached
	}

	txShardInfoToSet := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockTxs[index], txShardInfo: txShardInfoToSet}
	}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	txs.blockSizeComputation.AddNumMiniBlocks(numMiniBlocks)
	txs.blockSizeComputation.AddNumTxs(numTxs)

	if scheduledMode {
		for index := range miniBlockTxs {
			txs.scheduledTxsExecutionHandler.Add(miniBlockTxHashes[index], miniBlockTxs[index])
		}
	}

	return nil, len(processedTxHashes), nil
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

// EpochConfirmed is called whenever a new epoch is confirmed
func (txs *transactions) EpochConfirmed(epoch uint32, _ uint64) {
	txs.flagScheduledMiniBlocks.Toggle(epoch >= txs.scheduledMiniBlocksEnableEpoch)
	log.Debug("transactions: scheduled mini blocks", "enabled", txs.flagScheduledMiniBlocks.IsSet())
	txs.flagMixedTxsInMiniBlocks.Toggle(epoch >= txs.mixedTxsInMiniBlocksEnableEpoch)
	log.Debug("transactions: mixed txs in mini blocks", "enabled", txs.flagMixedTxsInMiniBlocks.IsSet())
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

		delta := bytes.Compare(txI.GetSndAddr(), txJ.GetSndAddr())
		if delta == 0 {
			delta = int(txI.GetNonce()) - int(txJ.GetNonce())
		}

		return delta < 0
	}

	sort.Slice(transactions, sorter)
}

func (txs *transactions) isBodyToMe(body *block.Body) bool {
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.SenderShardID == txs.shardCoordinator.SelfId() {
			return false
		}
	}
	return true
}

func (txs *transactions) isBodyFromMe(body *block.Body) bool {
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.SenderShardID != txs.shardCoordinator.SelfId() {
			return false
		}
	}
	return true
}

func (txs *transactions) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.TxBlock || mbType == block.InvalidBlock
}

func (txs *transactions) createAndProcessMiniBlocksFromMe(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	var miniBlocks block.MiniBlockSlice
	var err error
	var mapSCTxs map[string]struct{}

	if txs.flagScheduledMiniBlocks.IsSet() {
		miniBlocks, mapSCTxs, err = txs.createAndProcessMiniBlocksFromMeV2(
			haveTime,
			isShardStuck,
			isMaxBlockSizeReached,
			sortedTxs,
		)
	} else {
		miniBlocks, err = txs.createAndProcessMiniBlocksFromMeV1(
			haveTime,
			isShardStuck,
			isMaxBlockSizeReached,
			sortedTxs,
		)
	}

	return miniBlocks, mapSCTxs, err
}
