package preprocess

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ process.DataMarshalizer = (*transactions)(nil)
var _ process.PreProcessor = (*transactions)(nil)

var log = logger.GetOrCreate("process/block/preprocess")

// 200% bandwidth to allow 100% overshooting estimations
const selectionGasBandwidthIncreasePercent = 200

// 130% to allow 30% overshooting estimations for scheduled SC calls
const selectionGasBandwidthIncreaseScheduledPercent = 130

type accountTxsShards struct {
	accountsInfo map[string]*txShardInfo
	sync.RWMutex
}

// TODO: increase code coverage with unit test

var isShardStuckFalse = func(uint32) bool { return false }
var isMaxBlockSizeReachedFalse = func(int, int) bool { return false }
var haveAdditionalTimeFalse = func() bool { return false }

type transactions struct {
	*basePreProcess
	chRcvAllTxs                  chan bool
	onRequestTransaction         func(shardID uint32, txHashes [][]byte)
	txsForCurrBlock              txsForBlock
	txPool                       dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	txProcessor                  process.TransactionProcessor
	orderedTxs                   map[string][]data.TransactionHandler
	orderedTxHashes              map[string][][]byte
	mutOrderedTxs                sync.RWMutex
	blockTracker                 BlockTracker
	blockType                    block.Type
	accountTxsShards             accountTxsShards
	emptyAddress                 []byte
	txTypeHandler                process.TxTypeHandler
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	accntsTracker                *accountsTracker

	scheduledTXContinueFunc               func(isShardStuck func(uint32) bool, wrappedTx *txcache.WrappedTransaction, mapSCTxs map[string]struct{}, mbInfo *createScheduledMiniBlocksInfo) (*transaction.Transaction, *block.MiniBlock, bool)
	shouldSkipMiniBlockFunc               func(miniBlock *block.MiniBlock) bool
	isTransactionEligibleForExecutionFunc func(tx *transaction.Transaction, err error) bool
}

// ArgsTransactionPreProcessor holds the arguments to create a txs pre processor
type ArgsTransactionPreProcessor struct {
	TxDataPool                   dataRetriever.ShardedDataCacherNotifier
	Store                        dataRetriever.StorageService
	Hasher                       hashing.Hasher
	Marshalizer                  marshal.Marshalizer
	TxProcessor                  process.TransactionProcessor
	ShardCoordinator             sharding.Coordinator
	Accounts                     state.AccountsAdapter
	OnRequestTransaction         func(shardID uint32, txHashes [][]byte)
	EconomicsFee                 process.FeeHandler
	GasHandler                   process.GasHandler
	BlockTracker                 BlockTracker
	BlockType                    block.Type
	PubkeyConverter              core.PubkeyConverter
	BlockSizeComputation         BlockSizeComputationHandler
	BalanceComputation           BalanceComputationHandler
	EnableEpochsHandler          common.EnableEpochsHandler
	TxTypeHandler                process.TxTypeHandler
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	ProcessedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler      common.TxExecutionOrderHandler
}

// NewTransactionPreprocessor creates a new transaction preprocessor object
func NewTransactionPreprocessor(
	args ArgsTransactionPreProcessor,
) (*transactions, error) {
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.TxDataPool) {
		return nil, process.ErrNilTransactionPool
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilTxStorage
	}
	if check.IfNil(args.TxProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if args.OnRequestTransaction == nil {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.GasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(args.BlockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.BlockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(args.BalanceComputation) {
		return nil, process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.ScheduledTxsExecutionHandler) {
		return nil, process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(args.ProcessedMiniBlocksTracker) {
		return nil, process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(args.TxExecutionOrderHandler) {
		return nil, process.ErrNilTxExecutionOrderHandler
	}

	bpp := basePreProcess{
		hasher:      args.Hasher,
		marshalizer: args.Marshalizer,
		gasTracker: gasTracker{
			shardCoordinator: args.ShardCoordinator,
			gasHandler:       args.GasHandler,
			economicsFee:     args.EconomicsFee,
		},
		blockSizeComputation:       args.BlockSizeComputation,
		balanceComputation:         args.BalanceComputation,
		accounts:                   args.Accounts,
		pubkeyConverter:            args.PubkeyConverter,
		enableEpochsHandler:        args.EnableEpochsHandler,
		processedMiniBlocksTracker: args.ProcessedMiniBlocksTracker,
		txExecutionOrderHandler:    args.TxExecutionOrderHandler,
	}

	txs := &transactions{
		basePreProcess:               &bpp,
		storage:                      args.Store,
		txPool:                       args.TxDataPool,
		onRequestTransaction:         args.OnRequestTransaction,
		txProcessor:                  args.TxProcessor,
		blockTracker:                 args.BlockTracker,
		blockType:                    args.BlockType,
		txTypeHandler:                args.TxTypeHandler,
		scheduledTxsExecutionHandler: args.ScheduledTxsExecutionHandler,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterOnAdded(txs.receivedTransaction)

	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.accountTxsShards.accountsInfo = make(map[string]*txShardInfo)
	txs.accntsTracker = newAccountsTracker()

	txs.emptyAddress = make([]byte, txs.pubkeyConverter.Len())
	txs.scheduledTXContinueFunc = txs.shouldContinueProcessingScheduledTx
	txs.shouldSkipMiniBlockFunc = txs.shouldSkipMiniBlock
	txs.isTransactionEligibleForExecutionFunc = txs.isTransactionEligibleForExecution

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

// RemoveBlockDataFromPools removes transactions and miniblocks from associated pools
func (txs *transactions) RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	return txs.removeBlockDataFromPools(body, miniBlockPool, txs.txPool, txs.isMiniBlockCorrect)
}

// RemoveTxsFromPools removes transactions from associated pools
func (txs *transactions) RemoveTxsFromPools(body *block.Body) error {
	return txs.removeTxsFromPools(body, txs.txPool, txs.isMiniBlockCorrect)
}

// RestoreBlockDataIntoPools restores the transactions and miniblocks to associated pools
func (txs *transactions) RestoreBlockDataIntoPools(
	body *block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	if check.IfNil(body) {
		return 0, process.ErrNilBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return 0, process.ErrNilMiniBlockPool
	}

	txsRestored := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if !txs.isMiniBlockCorrect(miniBlock.Type) {
			continue
		}

		err := txs.restoreTxsIntoPool(miniBlock)
		if err != nil {
			return txsRestored, err
		}

		if miniBlock.SenderShardID != txs.shardCoordinator.SelfId() {
			miniBlockHash, errHash := core.CalculateHash(txs.marshalizer, txs.hasher, miniBlock)
			if errHash != nil {
				return txsRestored, errHash
			}

			miniBlockPool.Put(miniBlockHash, miniBlock, miniBlock.Size())
		}

		txsRestored += len(miniBlock.TxHashes)
	}

	return txsRestored, nil
}

func (txs *transactions) restoreTxsIntoPool(miniBlock *block.MiniBlock) error {
	miniBlockStrCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	txsBuff, err := txs.storage.GetAll(dataRetriever.TransactionUnit, miniBlock.TxHashes)
	if err != nil {
		log.Debug("txs from mini block were not found in TransactionUnit",
			"sender shard ID", miniBlock.SenderShardID,
			"receiver shard ID", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes),
		)

		return err
	}

	for txHash, txBuff := range txsBuff {
		tx := transaction.Transaction{}
		err = txs.marshalizer.Unmarshal(&tx, txBuff)
		if err != nil {
			return err
		}

		strCache := txs.computeCacheIdentifier(miniBlockStrCache, &tx, miniBlock.Type)
		txs.txPool.AddData([]byte(txHash), &tx, tx.Size(), strCache)
	}

	return nil
}

func (txs *transactions) computeCacheIdentifier(miniBlockStrCache string, tx *transaction.Transaction, miniBlockType block.Type) string {
	if miniBlockType != block.InvalidBlock {
		return miniBlockStrCache
	}
	if !txs.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		return miniBlockStrCache
	}

	// scheduled miniblocks feature requires that the transactions are properly restored in the correct cache, not the one
	// provided by the containing miniblock (think of how invalid transactions are executed and stored)

	senderShardID := txs.getShardFromAddress(tx.GetSndAddr())
	receiverShardID := txs.getShardFromAddress(tx.GetRcvAddr())

	return process.ShardCacherIdentifier(senderShardID, receiverShardID)
}

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(header data.HeaderHandler, body *block.Body, haveTime func() bool) (block.MiniBlockSlice, error) {
	if txs.isBodyToMe(body) {
		return txs.processTxsToMe(header, body, haveTime)
	}

	if txs.isBodyFromMe(body) {
		return txs.processTxsFromMeAndCreateScheduled(body, haveTime, header.GetPrevRandSeed())
	}

	return nil, process.ErrInvalidBody
}

func (txs *transactions) computeTxsToMe(
	headerHandler data.HeaderHandler,
	body *block.Body,
) ([]*txcache.WrappedTransaction, error) {
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

		pi, err := txs.getIndexesOfLastTxProcessed(miniBlock, headerHandler)
		if err != nil {
			return nil, err
		}

		txsFromMiniBlock, err := txs.computeTxsFromMiniBlock(miniBlock, pi)
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
		if txs.shouldSkipMiniBlockFunc(miniBlock) {
			continue
		}

		pi := &processedIndexes{
			indexOfLastTxProcessed:           -1,
			indexOfLastTxProcessedByProposer: int32(len(miniBlock.TxHashes)) - 1,
		}

		txsFromMiniBlock, err := txs.computeTxsFromMiniBlock(miniBlock, pi)
		if err != nil {
			return nil, err
		}

		allTxs = append(allTxs, txsFromMiniBlock...)
	}

	return allTxs, nil
}

func (txs *transactions) shouldSkipMiniBlock(miniBlock *block.MiniBlock) bool {
	shouldSkipMiniBlock := miniBlock.SenderShardID != txs.shardCoordinator.SelfId() ||
		!txs.isMiniBlockCorrect(miniBlock.Type) ||
		miniBlock.IsScheduledMiniBlock()

	return shouldSkipMiniBlock
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

		pi := &processedIndexes{
			indexOfLastTxProcessed:           -1,
			indexOfLastTxProcessedByProposer: int32(len(miniBlock.TxHashes)) - 1,
		}

		txsFromScheduledMiniBlock, err := txs.computeTxsFromMiniBlock(miniBlock, pi)
		if err != nil {
			return nil, err
		}

		allScheduledTxs = append(allScheduledTxs, txsFromScheduledMiniBlock...)
	}

	return allScheduledTxs, nil
}

func (txs *transactions) computeTxsFromMiniBlock(
	miniBlock *block.MiniBlock,
	pi *processedIndexes,
) ([]*txcache.WrappedTransaction, error) {

	txsFromMiniBlock := make([]*txcache.WrappedTransaction, 0, len(miniBlock.TxHashes))

	indexOfFirstTxToBeProcessed := pi.indexOfLastTxProcessed + 1
	err := process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, pi.indexOfLastTxProcessedByProposer, miniBlock)
	if err != nil {
		return nil, err
	}

	for i := indexOfFirstTxToBeProcessed; i <= pi.indexOfLastTxProcessedByProposer; i++ {
		txHash := miniBlock.TxHashes[i]
		txs.txsForCurrBlock.mutTxsForBlock.RLock()
		txInfoFromMap, ok := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
		txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

		if !ok || check.IfNil(txInfoFromMap.tx) {
			log.Warn("missing transaction in computeTxsFromMiniBlock", "type", miniBlock.Type, "txHash", txHash)
			return nil, process.ErrMissingTransaction
		}

		tx, ok := txInfoFromMap.tx.(*transaction.Transaction)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		calculatedSenderShardId := txs.getShardFromAddress(tx.GetSndAddr())
		calculatedReceiverShardId := txs.getShardFromAddress(tx.GetRcvAddr())

		wrappedTx := &txcache.WrappedTransaction{
			Tx:              tx,
			TxHash:          txHash,
			SenderShardID:   calculatedSenderShardId,
			ReceiverShardID: calculatedReceiverShardId,
		}

		txsFromMiniBlock = append(txsFromMiniBlock, wrappedTx)
	}

	return txsFromMiniBlock, nil
}

func (txs *transactions) getShardFromAddress(address []byte) uint32 {
	isEmptyAddress := bytes.Equal(address, txs.emptyAddress)
	if isEmptyAddress {
		return txs.shardCoordinator.SelfId()
	}

	return txs.shardCoordinator.ComputeId(address)
}

func (txs *transactions) processTxsToMe(header data.HeaderHandler, body *block.Body, haveTime func() bool) (block.MiniBlockSlice, error) {
	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}
	if check.IfNil(header) {
		return nil, process.ErrNilHeaderHandler
	}

	var err error
	scheduledMode := false
	if txs.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		scheduledMode, err = process.IsScheduledMode(header, body, txs.hasher, txs.marshalizer)
		if err != nil {
			return nil, err
		}
	}

	txsToMe, err := txs.computeTxsToMe(header, body)
	if err != nil {
		return nil, err
	}

	var totalGasConsumed uint64
	if scheduledMode {
		totalGasConsumed = txs.gasHandler.TotalGasProvidedAsScheduled()
	} else {
		totalGasConsumed = txs.getTotalGasConsumed()
	}

	numTXsProcessed := 0
	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlockInReceiverShard: uint64(0),
		gasConsumedByMiniBlocksInSenderShard:  uint64(0),
		totalGasConsumedInSelfShard:           totalGasConsumed,
	}

	log.Debug("transactions.processTxsToMe: before processing",
		"scheduled mode", scheduledMode,
		"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
		"total gas provided", txs.gasHandler.TotalGasProvided(),
		"total gas provided as scheduled", txs.gasHandler.TotalGasProvidedAsScheduled(),
		"total gas refunded", txs.gasHandler.TotalGasRefunded(),
		"total gas penalized", txs.gasHandler.TotalGasPenalized(),
	)
	defer func() {
		log.Debug("transactions.processTxsToMe after processing",
			"scheduled mode", scheduledMode,
			"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard,
			"num txs processed", numTXsProcessed,
			"total gas provided", txs.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", txs.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", txs.gasHandler.TotalGasRefunded(),
			"total gas penalized", txs.gasHandler.TotalGasPenalized(),
		)
	}()

	log.Debug("processTxsToMe", "scheduled mode", scheduledMode, "totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard)
	defer func() {
		log.Debug("processTxsToMe after processing", "totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard)
	}()

	for index := range txsToMe {
		if !haveTime() {
			return nil, process.ErrTimeIsOut
		}

		tx, ok := txsToMe[index].Tx.(*transaction.Transaction)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		txHash := txsToMe[index].TxHash
		senderShardID := txsToMe[index].SenderShardID
		receiverShardID := txsToMe[index].ReceiverShardID

		gasProvidedByTxInSelfShard, errComputeGas := txs.computeGasProvided(
			senderShardID,
			receiverShardID,
			tx,
			txHash,
			&gasInfo)

		if errComputeGas != nil {
			return nil, errComputeGas
		}

		if scheduledMode {
			txs.gasHandler.SetGasProvidedAsScheduled(gasProvidedByTxInSelfShard, txHash)
		} else {
			txs.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, txHash)
		}

		err = txs.saveAccountBalanceForAddress(tx.GetRcvAddr())
		if err != nil {
			return nil, err
		}

		if scheduledMode {
			txs.scheduledTxsExecutionHandler.AddScheduledTx(txHash, tx)
		} else {
			err = txs.processAndRemoveBadTransaction(
				txHash,
				tx,
				senderShardID,
				receiverShardID)
			if err != nil {
				return nil, err
			}

			txs.updateGasConsumedWithGasRefundedAndGasPenalized(txHash, &gasInfo)
		}

		numTXsProcessed++
	}

	return body.MiniBlocks, nil
}

func (txs *transactions) processTxsFromMe(body *block.Body,
	haveTime func() bool,
	randomness []byte,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	if check.IfNil(body) {
		return nil, nil, process.ErrNilBlockBody
	}

	txsFromMe, err := txs.computeTxsFromMe(body)
	if err != nil {
		return nil, nil, err
	}

	txs.sortTransactionsBySenderAndNonce(txsFromMe, randomness)

	isShardStuckFalse := func(uint32) bool {
		return false
	}
	isMaxBlockSizeReachedFalse := func(int, int) bool {
		return false
	}

	calculatedMiniBlocks, _, mapSCTxs, err := txs.createAndProcessMiniBlocksFromMe(
		haveTime,
		isShardStuckFalse,
		isMaxBlockSizeReachedFalse,
		txsFromMe,
	)
	if err != nil {
		return nil, nil, err
	}

	if !haveTime() {
		return nil, nil, process.ErrTimeIsOut
	}

	return calculatedMiniBlocks, mapSCTxs, nil
}

func (txs *transactions) processTxsFromMeAndCreateScheduled(body *block.Body, haveTime func() bool, randomness []byte) (block.MiniBlockSlice, error) {
	calculatedMiniBlocks, mapSCTxs, err := txs.processTxsFromMe(body, haveTime, randomness)
	if err != nil {
		return nil, err
	}

	scheduledMiniBlocks, err := txs.createScheduledMiniBlocksFromMeAsValidator(
		body,
		haveTime,
		haveAdditionalTimeFalse,
		isShardStuckFalse,
		isMaxBlockSizeReachedFalse,
		mapSCTxs,
		randomness,
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	calculatedBodyHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, &block.Body{MiniBlocks: calculatedMiniBlocks})
	if err != nil {
		return nil, err
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
		return nil, process.ErrBlockBodyHashMismatch
	}

	return calculatedMiniBlocks, nil
}

func (txs *transactions) createScheduledMiniBlocksFromMeAsValidator(
	body *block.Body,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	mapSCTxs map[string]struct{},
	randomness []byte,
) (block.MiniBlockSlice, error) {

	if !txs.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		return make(block.MiniBlockSlice, 0), nil
	}

	scheduledTxsFromMe, err := txs.computeScheduledTxsFromMe(body)
	if err != nil {
		return nil, err
	}

	txs.sortTransactionsBySenderAndNonce(scheduledTxsFromMe, randomness)

	scheduledMiniBlocks, err := txs.createScheduledMiniBlocks(
		haveTime,
		haveAdditionalTime,
		isShardStuck,
		isMaxBlockSizeReached,
		scheduledTxsFromMe,
		mapSCTxs,
	)
	if err != nil {
		return nil, err
	}

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

		txs.saveTxsToStorage(miniBlock.TxHashes, &txs.txsForCurrBlock, txs.storage, dataRetriever.TransactionUnit)
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

	txs.accntsTracker.init()
	txs.scheduledTxsExecutionHandler.Init()
}

// AddTxsFromMiniBlocks will add the transactions from the provided miniblocks into the internal cache
func (txs *transactions) AddTxsFromMiniBlocks(miniBlocks block.MiniBlockSlice) {
	for _, mb := range miniBlocks {
		if !txs.isMiniBlockCorrect(mb.Type) {
			log.Warn("transactions.addTxsFromScheduledMiniBlocks: mini block type is not correct",
				"type", mb.Type,
				"sender", mb.SenderShardID,
				"receiver", mb.ReceiverShardID,
				"num txs", len(mb.TxHashes))
			continue
		}

		txShardInfoToSet := &txShardInfo{senderShardID: mb.SenderShardID, receiverShardID: mb.ReceiverShardID}
		method := process.SearchMethodJustPeek
		if mb.Type == block.InvalidBlock {
			method = process.SearchMethodSearchFirst
		}

		for _, txHash := range mb.TxHashes {
			tx, err := process.GetTransactionHandler(
				mb.SenderShardID,
				mb.ReceiverShardID,
				txHash,
				txs.txPool,
				txs.storage,
				txs.marshalizer,
				method,
			)
			if err != nil {
				log.Debug("transactions.AddTxsFromMiniBlocks: GetTransactionHandler", "tx hash", txHash, "error", err.Error())
				continue
			}

			txs.txsForCurrBlock.mutTxsForBlock.Lock()
			txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoToSet}
			txs.txsForCurrBlock.mutTxsForBlock.Unlock()
		}
	}
}

// AddTransactions adds the given transactions to the current block transactions
func (txs *transactions) AddTransactions(txHandlers []data.TransactionHandler) {
	for i, tx := range txHandlers {
		senderShardID := txs.getShardFromAddress(tx.GetSndAddr())
		receiverShardID := txs.getShardFromAddress(tx.GetRcvAddr())
		txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
		txHash, err := core.CalculateHash(txs.marshalizer, txs.hasher, tx)
		if err != nil {
			log.Warn("transactions.AddTransactions CalculateHash", "error", err.Error())
			continue
		}
		txs.txsForCurrBlock.mutTxsForBlock.Lock()
		txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: txHandlers[i], txShardInfo: txShardInfoToSet}
		txs.txsForCurrBlock.mutTxsForBlock.Unlock()
	}
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

	txs.txExecutionOrderHandler.Add(txHash)
	_, err := txs.txProcessor.ProcessTransaction(tx)
	isTxTargetedForDeletion := errors.Is(err, process.ErrLowerNonceInTransaction) || errors.Is(err, process.ErrInsufficientFee) || errors.Is(err, process.ErrTransactionNotExecutable)
	if isTxTargetedForDeletion {
		txs.txExecutionOrderHandler.Remove(txHash)
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txs.txPool.RemoveData(txHash, strCache)
	}

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		txs.txExecutionOrderHandler.Remove(txHash)

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

	missingTxsHashesForMiniBlock := txs.computeMissingTxsHashesForMiniBlock(miniBlock)
	if len(missingTxsHashesForMiniBlock) > 0 {
		txs.onRequestTransaction(miniBlock.SenderShardID, missingTxsHashesForMiniBlock)
	}

	return len(missingTxsHashesForMiniBlock)
}

// computeMissingTxsHashesForMiniBlock computes missing transactions hashes for a certain miniblock
func (txs *transactions) computeMissingTxsHashesForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
	missingTransactionsHashes := make([][]byte, 0)

	if miniBlock.Type != txs.blockType {
		return missingTransactionsHashes
	}

	method := process.SearchMethodJustPeek
	if txs.blockType == block.InvalidBlock {
		method = process.SearchMethodSearchFirst
	}

	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			txs.txPool,
			method)

		if check.IfNil(tx) {
			missingTransactionsHashes = append(missingTransactionsHashes, txHash)
		}
	}

	return missingTransactionsHashes
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

func (txs *transactions) getRemainingGasPerBlock() uint64 {
	gasConsumed := txs.getTotalGasConsumed()
	maxGasPerBlock := txs.economicsFee.MaxGasLimitPerBlock(txs.shardCoordinator.SelfId())
	gasBandwidth := uint64(0)
	if gasConsumed < maxGasPerBlock {
		gasBandwidth = maxGasPerBlock - gasConsumed
	}
	return gasBandwidth
}

func (txs *transactions) getRemainingGasPerBlockAsScheduled() uint64 {
	gasProvided := txs.gasHandler.TotalGasProvidedAsScheduled()
	maxGasPerBlock := txs.economicsFee.MaxGasLimitPerBlock(txs.shardCoordinator.SelfId())
	gasBandwidth := uint64(0)
	if gasProvided < maxGasPerBlock {
		gasBandwidth = maxGasPerBlock - gasProvided
	}
	return gasBandwidth
}

// CreateAndProcessMiniBlocks creates miniBlocks from storage and processes the transactions added into the miniblocks
// as long as it has time
// TODO: check if possible for transaction pre processor to receive a blockChainHook and use it to get the randomness instead
func (txs *transactions) CreateAndProcessMiniBlocks(haveTime func() bool, randomness []byte) (block.MiniBlockSlice, error) {
	startTime := time.Now()

	gasBandwidth := txs.getRemainingGasPerBlock() * selectionGasBandwidthIncreasePercent / 100
	gasBandwidthForScheduled := uint64(0)
	if txs.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		gasBandwidthForScheduled = txs.getRemainingGasPerBlockAsScheduled() * selectionGasBandwidthIncreaseScheduledPercent / 100
		gasBandwidth += gasBandwidthForScheduled
	}

	sortedTxs, remainingTxsForScheduled, err := txs.computeSortedTxs(txs.shardCoordinator.SelfId(), txs.shardCoordinator.SelfId(), gasBandwidth, randomness)
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

	if txs.blockTracker.ShouldSkipMiniBlocksCreationFromSelf() {
		log.Debug("CreateAndProcessMiniBlocks global stuck")
		return make(block.MiniBlockSlice, 0), nil
	}

	startTime = time.Now()
	miniBlocks, remainingTxs, mapSCTxs, err := txs.createAndProcessMiniBlocksFromMe(
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

	sortedTxsForScheduled := append(remainingTxs, remainingTxsForScheduled...)
	sortedTxsForScheduled, _ = txs.prefilterTransactions(nil, sortedTxsForScheduled, 0, gasBandwidthForScheduled)
	txs.sortTransactionsBySenderAndNonce(sortedTxsForScheduled, randomness)

	haveAdditionalTime := process.HaveAdditionalTime()
	scheduledMiniBlocks, err := txs.createScheduledMiniBlocksFromMeAsProposer(
		haveTime,
		haveAdditionalTime,
		sortedTxsForScheduled,
		mapSCTxs,
	)
	if err != nil {
		log.Debug("createScheduledMiniBlocksFromMeAsProposer", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	miniBlocks = append(miniBlocks, scheduledMiniBlocks...)

	return miniBlocks, nil
}

func (txs *transactions) createScheduledMiniBlocksFromMeAsProposer(
	haveTime func() bool,
	haveAdditionalTime func() bool,
	sortedTxs []*txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
) (block.MiniBlockSlice, error) {

	if !txs.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		return make(block.MiniBlockSlice, 0), nil
	}

	startTime := time.Now()
	scheduledMiniBlocks, err := txs.createScheduledMiniBlocks(
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
	if err != nil {
		return nil, err
	}

	return scheduledMiniBlocks, nil
}

type processingActions struct {
	canAddTx             bool
	canAddMore           bool
	shouldAddToRemaining bool
}

func (txs *transactions) createAndProcessMiniBlocksFromMeV1(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, []*txcache.WrappedTransaction, error) {
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
		getTotalGasConsumed:       txs.getTotalGasConsumed,
		txPool:                    txs.txPool,
	}

	mbBuilder, err := newMiniBlockBuilder(args)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	remainingTxs := make([]*txcache.WrappedTransaction, 0)
	for idx, wtx := range sortedTxs {
		actions, tx := mbBuilder.checkAddTransaction(wtx)
		if !actions.canAddMore {
			if actions.shouldAddToRemaining {
				remainingTxs = append(remainingTxs, sortedTxs[idx:]...)
			}
			break
		}

		if !actions.canAddTx {
			if actions.shouldAddToRemaining {
				remainingTxs = append(remainingTxs, sortedTxs[idx])
			}
			continue
		}

		err = txs.processMiniBlockBuilderTx(mbBuilder, wtx, tx)
		if err != nil {
			if core.IsGetNodeFromDBError(err) {
				return nil, nil, err
			}
			continue
		}

		mbBuilder.addTxAndUpdateBlockSize(tx, wtx)
	}

	miniBlocks := txs.getMiniBlockSliceFromMap(mbBuilder.miniBlocks)

	logCreatedMiniBlocksStats(miniBlocks, txs.shardCoordinator.SelfId(), mbBuilder, len(sortedTxs))

	return miniBlocks, remainingTxs, nil
}

func logCreatedMiniBlocksStats(
	miniBlocks block.MiniBlockSlice,
	selfShardID uint32,
	mbb *miniBlocksBuilder,
	nbSortedTxs int,
) {
	log.Debug("createAndProcessMiniBlocksFromMeV1",
		"self shard", selfShardID,
		"gas consumed in sender shard", mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard,
		"total gas consumed in self shard", mbb.gasInfo.totalGasConsumedInSelfShard)

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
		"num cross shard txs with too much gas", mbb.stats.numCrossShardTxsWithTooMuchGas,
		"used time for computeGasProvided", mbb.stats.totalGasComputeTime,
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
	if errRevert != nil && !core.IsClosingError(errRevert) {
		log.Warn("revert to snapshot", "error", err.Error())
	}

	mbb.handleBadTransaction(err, wtx, tx)
}

func (txs *transactions) refundGas(
	wtx *txcache.WrappedTransaction,
	mbb *miniBlocksBuilder,
) {
	gasRefunded := txs.gasHandler.GasRefunded(wtx.TxHash)
	gasPenalized := txs.gasHandler.GasPenalized(wtx.TxHash)
	mbb.handleGasRefund(wtx, gasRefunded, gasPenalized)
}

func (txs *transactions) createEmptyMiniBlock(
	senderShardID uint32,
	receiverShardID uint32,
	blockType block.Type,
	reserved []byte,
) *block.MiniBlock {

	miniBlock := &block.MiniBlock{
		Type:            blockType,
		SenderShardID:   senderShardID,
		ReceiverShardID: receiverShardID,
		TxHashes:        make([][]byte, 0),
		Reserved:        reserved,
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

	return txs.splitMiniBlocksBasedOnMaxGasLimitIfNeeded(miniBlocks)
}

func (txs *transactions) splitMiniBlocksBasedOnMaxGasLimitIfNeeded(miniBlocks block.MiniBlockSlice) block.MiniBlockSlice {
	if !txs.enableEpochsHandler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() {
		return miniBlocks
	}

	splitMiniBlocks := make(block.MiniBlockSlice, 0)
	for _, miniBlock := range miniBlocks {
		if miniBlock.ReceiverShardID == txs.shardCoordinator.SelfId() {
			splitMiniBlocks = append(splitMiniBlocks, miniBlock)
			continue
		}

		splitMiniBlocks = append(splitMiniBlocks, txs.splitMiniBlockBasedOnMaxGasLimitIfNeeded(miniBlock)...)
	}

	return splitMiniBlocks
}

func (txs *transactions) splitMiniBlockBasedOnMaxGasLimitIfNeeded(miniBlock *block.MiniBlock) block.MiniBlockSlice {
	splitMiniBlocks := make(block.MiniBlockSlice, 0)
	currentMiniBlock := createEmptyMiniBlockFromMiniBlock(miniBlock)
	gasLimitInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txInfoInstance, ok := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
		if !ok {
			log.Warn("transactions.splitMiniBlockIfNeeded: missing tx", "hash", txHash)
			currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
			continue
		}

		_, gasProvidedByTxInReceiverShard, err := txs.computeGasProvidedByTx(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txInfoInstance.tx,
			txHash)
		if err != nil {
			log.Warn("transactions.splitMiniBlockIfNeeded: failed to compute gas consumed by tx", "hash", txHash, "error", err.Error())
			currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
			continue
		}

		isGasLimitExceeded := gasLimitInReceiverShard+gasProvidedByTxInReceiverShard > txs.economicsFee.MaxGasLimitPerMiniBlockForSafeCrossShard()
		if isGasLimitExceeded {
			log.Debug("transactions.splitMiniBlockIfNeeded: gas limit exceeded",
				"mb type", currentMiniBlock.Type,
				"sender shard", currentMiniBlock.SenderShardID,
				"receiver shard", currentMiniBlock.ReceiverShardID,
				"initial num txs", len(miniBlock.TxHashes),
				"adjusted num txs", len(currentMiniBlock.TxHashes),
			)

			if len(currentMiniBlock.TxHashes) > 0 {
				splitMiniBlocks = append(splitMiniBlocks, currentMiniBlock)
			}

			currentMiniBlock = createEmptyMiniBlockFromMiniBlock(miniBlock)
			gasLimitInReceiverShard = 0
		}

		gasLimitInReceiverShard += gasProvidedByTxInReceiverShard
		currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
	}

	if len(currentMiniBlock.TxHashes) > 0 {
		splitMiniBlocks = append(splitMiniBlocks, currentMiniBlock)
	}

	return splitMiniBlocks
}

func createEmptyMiniBlockFromMiniBlock(miniBlock *block.MiniBlock) *block.MiniBlock {
	return &block.MiniBlock{
		SenderShardID:   miniBlock.SenderShardID,
		ReceiverShardID: miniBlock.ReceiverShardID,
		Type:            miniBlock.Type,
		Reserved:        miniBlock.Reserved,
		TxHashes:        make([][]byte, 0),
	}
}

func (txs *transactions) computeSortedTxs(
	sndShardId uint32,
	dstShardId uint32,
	gasBandwidth uint64,
	randomness []byte,
) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction, error) {
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txShardPool := txs.txPool.ShardDataStore(strCache)

	if check.IfNil(txShardPool) {
		return nil, nil, process.ErrNilTxDataPool
	}

	sortedTransactionsProvider := createSortedTransactionsProvider(txShardPool)
	log.Debug("computeSortedTxs.GetSortedTransactions")
	sortedTxs := sortedTransactionsProvider.GetSortedTransactions()

	// TODO: this could be moved to SortedTransactionsProvider
	selectedTxs, remainingTxs := txs.preFilterTransactionsWithMoveBalancePriority(sortedTxs, gasBandwidth)
	txs.sortTransactionsBySenderAndNonce(selectedTxs, randomness)

	return selectedTxs, remainingTxs, nil
}

// ProcessMiniBlock processes all the transactions from the given miniblock and saves the processed ones in a local cache
func (txs *transactions) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
	partialMbExecutionMode bool,
	indexOfLastTxProcessed int,
	preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler,
) ([][]byte, int, bool, error) {

	if miniBlock.Type != block.TxBlock {
		return nil, indexOfLastTxProcessed, false, process.ErrWrongTypeInMiniBlock
	}

	numTXsProcessed := 0
	var gasProvidedByTxInSelfShard uint64
	var err error
	var txIndex int
	processedTxHashes := make([][]byte, 0)

	indexOfFirstTxToBeProcessed := indexOfLastTxProcessed + 1
	err = process.CheckIfIndexesAreOutOfBound(int32(indexOfFirstTxToBeProcessed), int32(len(miniBlock.TxHashes))-1, miniBlock)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime, haveAdditionalTime)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	if txs.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlock.TxHashes)) {
		return nil, indexOfLastTxProcessed, false, process.ErrMaxBlockSizeReached
	}

	var totalGasConsumed uint64
	if scheduledMode {
		totalGasConsumed = txs.gasHandler.TotalGasProvidedAsScheduled()
	} else {
		totalGasConsumed = txs.getTotalGasConsumed()
	}

	isSelfShardStuck := txs.blockTracker.IsShardStuck(txs.shardCoordinator.SelfId())

	var maxGasLimitUsedForDestMeTxs uint64
	isFirstMiniBlockDestMe := totalGasConsumed == 0
	if isFirstMiniBlockDestMe || isSelfShardStuck {
		maxGasLimitUsedForDestMeTxs = txs.economicsFee.MaxGasLimitPerBlock(txs.shardCoordinator.SelfId())
	} else {
		maxGasLimitUsedForDestMeTxs = txs.economicsFee.MaxGasLimitPerBlock(txs.shardCoordinator.SelfId()) * maxGasLimitPercentUsedForDestMeTxs / 100
	}

	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlockInReceiverShard: uint64(0),
		gasConsumedByMiniBlocksInSenderShard:  uint64(0),
		totalGasConsumedInSelfShard:           totalGasConsumed,
	}

	log.Debug("transactions.ProcessMiniBlock: before processing",
		"scheduled mode", scheduledMode,
		"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
		"total gas provided", txs.gasHandler.TotalGasProvided(),
		"total gas provided as scheduled", txs.gasHandler.TotalGasProvidedAsScheduled(),
		"total gas refunded", txs.gasHandler.TotalGasRefunded(),
		"total gas penalized", txs.gasHandler.TotalGasPenalized(),
	)
	defer func() {
		log.Debug("transactions.ProcessMiniBlock after processing",
			"scheduled mode", scheduledMode,
			"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard,
			"num txs processed", numTXsProcessed,
			"total gas provided", txs.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", txs.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", txs.gasHandler.TotalGasRefunded(),
			"total gas penalized", txs.gasHandler.TotalGasPenalized(),
		)
	}()

	numOfOldCrossInterMbs, numOfOldCrossInterTxs := preProcessorExecutionInfoHandler.GetNumOfCrossInterMbsAndTxs()

	for txIndex = indexOfFirstTxToBeProcessed; txIndex < len(miniBlockTxs); txIndex++ {
		if !haveTime() && !haveAdditionalTime() {
			err = process.ErrTimeIsOut
			break
		}

		gasProvidedByTxInSelfShard, err = txs.computeGasProvided(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockTxs[txIndex],
			miniBlockTxHashes[txIndex],
			&gasInfo)

		if err != nil {
			break
		}

		if txs.enableEpochsHandler.IsOptimizeGasUsedInCrossMiniBlocksFlagEnabled() {
			if gasInfo.totalGasConsumedInSelfShard > maxGasLimitUsedForDestMeTxs {
				err = process.ErrMaxGasLimitUsedForDestMeTxsIsReached
				break
			}
		}

		err = txs.saveAccountBalanceForAddress(miniBlockTxs[txIndex].GetRcvAddr())
		if err != nil {
			break
		}

		if !scheduledMode {
			err = txs.processInNormalMode(
				preProcessorExecutionInfoHandler,
				miniBlockTxs[txIndex],
				miniBlockTxHashes[txIndex],
				&gasInfo,
				gasProvidedByTxInSelfShard)
			if err != nil {
				break
			}
		} else {
			txs.gasHandler.SetGasProvidedAsScheduled(gasProvidedByTxInSelfShard, miniBlockTxHashes[txIndex])
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[txIndex])
		numTXsProcessed++
	}

	if err != nil && (!partialMbExecutionMode || isSelfShardStuck) {
		return processedTxHashes, txIndex - 1, true, err
	}

	numOfCrtCrossInterMbs, numOfCrtCrossInterTxs := preProcessorExecutionInfoHandler.GetNumOfCrossInterMbsAndTxs()
	numOfNewCrossInterMbs := numOfCrtCrossInterMbs - numOfOldCrossInterMbs
	numOfNewCrossInterTxs := numOfCrtCrossInterTxs - numOfOldCrossInterTxs

	log.Debug("transactions.ProcessMiniBlock",
		"scheduled mode", scheduledMode,
		"numOfOldCrossInterMbs", numOfOldCrossInterMbs, "numOfOldCrossInterTxs", numOfOldCrossInterTxs,
		"numOfCrtCrossInterMbs", numOfCrtCrossInterMbs, "numOfCrtCrossInterTxs", numOfCrtCrossInterTxs,
		"numOfNewCrossInterMbs", numOfNewCrossInterMbs, "numOfNewCrossInterTxs", numOfNewCrossInterTxs,
	)

	numMiniBlocks := 1 + numOfNewCrossInterMbs
	numTxs := len(miniBlock.TxHashes) + numOfNewCrossInterTxs
	if txs.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(numMiniBlocks, numTxs) {
		return processedTxHashes, txIndex - 1, true, process.ErrMaxBlockSizeReached
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
		for index := indexOfFirstTxToBeProcessed; index <= txIndex-1; index++ {
			txs.scheduledTxsExecutionHandler.AddScheduledTx(miniBlockTxHashes[index], miniBlockTxs[index])
		}
	}

	return nil, txIndex - 1, false, err
}

func (txs *transactions) processInNormalMode(
	preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler,
	tx *transaction.Transaction,
	txHash []byte,
	gasInfo *gasConsumedInfo,
	gasProvidedByTxInSelfShard uint64,
) error {

	snapshot := txs.handleProcessTransactionInit(preProcessorExecutionInfoHandler, txHash)

	txs.txExecutionOrderHandler.Add(txHash)
	_, err := txs.txProcessor.ProcessTransaction(tx)
	if err != nil {
		txs.handleProcessTransactionError(preProcessorExecutionInfoHandler, snapshot, txHash)
		return err
	}

	txs.updateGasConsumedWithGasRefundedAndGasPenalized(txHash, gasInfo)
	txs.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, txHash)

	return nil
}

// CreateMarshalledData marshals transactions hashes and saves them into a new structure
func (txs *transactions) CreateMarshalledData(txHashes [][]byte) ([][]byte, error) {
	marshalledTxs, err := txs.createMarshalledData(txHashes, &txs.txsForCurrBlock)
	if err != nil {
		return nil, err
	}

	return marshalledTxs, nil
}

// GetAllCurrentUsedTxs returns all the transactions used at current creation / processing
func (txs *transactions) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	txs.txsForCurrBlock.mutTxsForBlock.RLock()
	txsPool := make(map[string]data.TransactionHandler, len(txs.txsForCurrBlock.txHashAndInfo))
	for txHash, txInfoFromMap := range txs.txsForCurrBlock.txHashAndInfo {
		txsPool[txHash] = txInfoFromMap.tx
	}
	txs.txsForCurrBlock.mutTxsForBlock.RUnlock()

	return txsPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (txs *transactions) IsInterfaceNil() bool {
	return txs == nil
}

// sortTransactionsBySenderAndNonce sorts the provided transactions and hashes simultaneously
func (txs *transactions) sortTransactionsBySenderAndNonce(transactions []*txcache.WrappedTransaction, randomness []byte) {
	if !txs.enableEpochsHandler.IsFrontRunningProtectionFlagEnabled() {
		sortTransactionsBySenderAndNonceLegacy(transactions)
		return
	}

	txs.sortTransactionsBySenderAndNonceWithFrontRunningProtection(transactions, randomness)
}

func (txs *transactions) sortTransactionsBySenderAndNonceWithFrontRunningProtection(transactions []*txcache.WrappedTransaction, randomness []byte) {
	// make sure randomness is 32bytes and uniform
	randSeed := txs.hasher.Compute(string(randomness))
	xoredAddresses := make(map[string][]byte)

	for _, tx := range transactions {
		xoredBytes := xorBytes(tx.Tx.GetSndAddr(), randSeed)
		xoredAddresses[string(tx.Tx.GetSndAddr())] = txs.hasher.Compute(string(xoredBytes))
	}

	sorter := func(i, j int) bool {
		txI := transactions[i].Tx
		txJ := transactions[j].Tx

		delta := bytes.Compare(xoredAddresses[string(txI.GetSndAddr())], xoredAddresses[string(txJ.GetSndAddr())])
		if delta == 0 {
			delta = int(txI.GetNonce()) - int(txJ.GetNonce())
		}

		return delta < 0
	}

	sort.Slice(transactions, sorter)
}

func sortTransactionsBySenderAndNonceLegacy(transactions []*txcache.WrappedTransaction) {
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

// parameters need to be of the same len, otherwise it will panic (if second slice shorter)
func xorBytes(a, b []byte) []byte {
	res := make([]byte, len(a))
	for i := range a {
		res[i] = a[i] ^ b[i]
	}
	return res
}

func (txs *transactions) filterMoveBalance(transactions []*txcache.WrappedTransaction) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction, uint64) {
	selectedTxs := make([]*txcache.WrappedTransaction, 0, len(transactions))
	skippedTxs := make([]*txcache.WrappedTransaction, 0, len(transactions))
	skippedAddresses := make(map[string]struct{})

	skipped := 0
	gasEstimation := uint64(0)
	for i, tx := range transactions {
		// prioritize move balance operations
		// and don't care about gas cost for them
		if len(tx.Tx.GetData()) > 0 {
			skippedTxs = append(skippedTxs, transactions[i])
			skippedAddresses[string(tx.Tx.GetSndAddr())] = struct{}{}
			skipped++
			continue
		}

		if shouldSkipTransactionIfMarkedAddress(tx, skippedAddresses) {
			skippedTxs = append(skippedTxs, transactions[i])
			continue
		}

		selectedTxs = append(selectedTxs, transactions[i])
		gasEstimation += txs.economicsFee.MinGasLimit()
	}
	return selectedTxs, skippedTxs, gasEstimation
}

// preFilterTransactions filters the transactions prioritising the move balance operations
func (txs *transactions) preFilterTransactionsWithMoveBalancePriority(
	transactions []*txcache.WrappedTransaction,
	gasBandwidth uint64,
) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction) {
	selectedTxs, skippedTxs, gasEstimation := txs.filterMoveBalance(transactions)

	return txs.prefilterTransactions(selectedTxs, skippedTxs, gasEstimation, gasBandwidth)
}

func (txs *transactions) prefilterTransactions(
	initialTxs []*txcache.WrappedTransaction,
	additionalTxs []*txcache.WrappedTransaction,
	initialTxsGasEstimation uint64,
	gasBandwidth uint64,
) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction) {

	selectedTxs, remainingTxs, gasEstimation := txs.addTxsWithinBandwidth(initialTxs, additionalTxs, initialTxsGasEstimation, gasBandwidth)

	log.Debug("preFilterTransactions estimation",
		"initialTxs", len(initialTxs),
		"gasCost initialTxs", initialTxsGasEstimation,
		"additionalTxs", len(additionalTxs),
		"gasCostEstimation", gasEstimation,
		"selected", len(selectedTxs),
		"skipped", len(remainingTxs))

	return selectedTxs, remainingTxs
}

func (txs *transactions) addTxsWithinBandwidth(
	initialTxs []*txcache.WrappedTransaction,
	additionalTxs []*txcache.WrappedTransaction,
	initialTxsGasEstimation uint64,
	totalGasBandwidth uint64,
) ([]*txcache.WrappedTransaction, []*txcache.WrappedTransaction, uint64) {
	remainingTxs := make([]*txcache.WrappedTransaction, 0, len(additionalTxs))
	resultedTxs := make([]*txcache.WrappedTransaction, 0, len(initialTxs))
	resultedTxs = append(resultedTxs, initialTxs...)

	gasEstimation := initialTxsGasEstimation
	for i, tx := range additionalTxs {
		gasInShard, _, err := txs.gasHandler.ComputeGasProvidedByTx(tx.SenderShardID, tx.ReceiverShardID, tx.Tx)
		if err != nil {
			continue
		}
		if gasEstimation+gasInShard > totalGasBandwidth {
			remainingTxs = append(remainingTxs, additionalTxs[i:]...)
			break
		}

		resultedTxs = append(resultedTxs, additionalTxs[i])
		gasEstimation += gasInShard
	}
	return resultedTxs, remainingTxs, gasEstimation
}

func shouldSkipTransactionIfMarkedAddress(tx *txcache.WrappedTransaction, addressesToSkip map[string]struct{}) bool {
	_, ok := addressesToSkip[string(tx.Tx.GetSndAddr())]
	return ok
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
) (block.MiniBlockSlice, []*txcache.WrappedTransaction, map[string]struct{}, error) {
	var miniBlocks block.MiniBlockSlice
	var err error
	var mapSCTxs map[string]struct{}
	var remainingTxs []*txcache.WrappedTransaction

	if txs.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		miniBlocks, remainingTxs, mapSCTxs, err = txs.createAndProcessMiniBlocksFromMeV2(
			haveTime,
			isShardStuck,
			isMaxBlockSizeReached,
			sortedTxs,
		)
	} else {
		miniBlocks, remainingTxs, err = txs.createAndProcessMiniBlocksFromMeV1(
			haveTime,
			isShardStuck,
			isMaxBlockSizeReached,
			sortedTxs,
		)
	}

	return miniBlocks, remainingTxs, mapSCTxs, err
}
