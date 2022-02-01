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

// 200% bandwidth to allow 100% overshooting estimations
const selectionGasBandwidthIncreasePercent = 200

// 130% to allow 30% overshooting estimations for scheduled SC calls
const selectionGasBandwidthIncreaseScheduledPercent = 130

type accountTxsShards struct {
	accountsInfo map[string]*txShardInfo
	sync.RWMutex
}

// TODO: increase code coverage with unit test

type transactions struct {
	*basePreProcess
	chRcvAllTxs                    chan bool
	onRequestTransaction           func(shardID uint32, txHashes [][]byte)
	txsForCurrBlock                txsForBlock
	txPool                         dataRetriever.ShardedDataCacherNotifier
	storage                        dataRetriever.StorageService
	txProcessor                    process.TransactionProcessor
	orderedTxs                     map[string][]data.TransactionHandler
	orderedTxHashes                map[string][][]byte
	mutOrderedTxs                  sync.RWMutex
	blockTracker                   BlockTracker
	blockType                      block.Type
	accountTxsShards               accountTxsShards
	emptyAddress                   []byte
	scheduledMiniBlocksEnableEpoch uint32
	flagScheduledMiniBlocks        atomic.Flag
	txTypeHandler                  process.TxTypeHandler
	scheduledTxsExecutionHandler   process.ScheduledTxsExecutionHandler
}

// ArgsTransactionPreProcessor holds the arguments to create a txs pre processor
type ArgsTransactionPreProcessor struct {
	TxDataPool                                  dataRetriever.ShardedDataCacherNotifier
	Store                                       dataRetriever.StorageService
	Hasher                                      hashing.Hasher
	Marshalizer                                 marshal.Marshalizer
	TxProcessor                                 process.TransactionProcessor
	ShardCoordinator                            sharding.Coordinator
	Accounts                                    state.AccountsAdapter
	OnRequestTransaction                        func(shardID uint32, txHashes [][]byte)
	EconomicsFee                                process.FeeHandler
	GasHandler                                  process.GasHandler
	BlockTracker                                BlockTracker
	BlockType                                   block.Type
	PubkeyConverter                             core.PubkeyConverter
	BlockSizeComputation                        BlockSizeComputationHandler
	BalanceComputation                          BalanceComputationHandler
	EpochNotifier                               process.EpochNotifier
	OptimizeGasUsedInCrossMiniBlocksEnableEpoch uint32
	FrontRunningProtectionEnableEpoch           uint32
	ScheduledMiniBlocksEnableEpoch              uint32
	TxTypeHandler                               process.TxTypeHandler
	ScheduledTxsExecutionHandler                process.ScheduledTxsExecutionHandler
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
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(args.TxTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(args.ScheduledTxsExecutionHandler) {
		return nil, process.ErrNilScheduledTxsExecutionHandler
	}

	bpp := basePreProcess{
		hasher:      args.Hasher,
		marshalizer: args.Marshalizer,
		gasTracker: gasTracker{
			shardCoordinator: args.ShardCoordinator,
			gasHandler:       args.GasHandler,
			economicsFee:     args.EconomicsFee,
		},
		blockSizeComputation: args.BlockSizeComputation,
		balanceComputation:   args.BalanceComputation,
		accounts:             args.Accounts,
		pubkeyConverter:      args.PubkeyConverter,
		optimizeGasUsedInCrossMiniBlocksEnableEpoch: args.OptimizeGasUsedInCrossMiniBlocksEnableEpoch,
		frontRunningProtectionEnableEpoch:           args.FrontRunningProtectionEnableEpoch,
	}

	txs := &transactions{
		basePreProcess:                 &bpp,
		storage:                        args.Store,
		txPool:                         args.TxDataPool,
		onRequestTransaction:           args.OnRequestTransaction,
		txProcessor:                    args.TxProcessor,
		blockTracker:                   args.BlockTracker,
		blockType:                      args.BlockType,
		scheduledMiniBlocksEnableEpoch: args.ScheduledMiniBlocksEnableEpoch,
		txTypeHandler:                  args.TxTypeHandler,
		scheduledTxsExecutionHandler:   args.ScheduledTxsExecutionHandler,
	}

	txs.chRcvAllTxs = make(chan bool)
	txs.txPool.RegisterOnAdded(txs.receivedTransaction)

	txs.txsForCurrBlock.txHashAndInfo = make(map[string]*txInfo)
	txs.orderedTxs = make(map[string][]data.TransactionHandler)
	txs.orderedTxHashes = make(map[string][][]byte)
	txs.accountTxsShards.accountsInfo = make(map[string]*txShardInfo)

	txs.emptyAddress = make([]byte, txs.pubkeyConverter.Len())

	log.Debug("transactions: enable epoch for optimize gas used in cross shard mini blocks", "epoch", txs.optimizeGasUsedInCrossMiniBlocksEnableEpoch)
	log.Debug("transactions: enable epoch for front running protection", "epoch", txs.frontRunningProtectionEnableEpoch)
	log.Debug("transactions: enable epoch for scheduled mini blocks", "epoch", txs.scheduledMiniBlocksEnableEpoch)

	args.EpochNotifier.RegisterNotifyHandler(txs)

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

// ProcessBlockTransactions processes all the transaction from the block.Body, updates the state
func (txs *transactions) ProcessBlockTransactions(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) error {
	if txs.isBodyToMe(body) {
		return txs.processTxsToMe(header, body, haveTime)
	}

	if txs.isBodyFromMe(body) {
		return txs.processTxsFromMe(body, haveTime, header.GetPrevRandSeed())
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

		tx, ok := txInfoFromMap.tx.(*transaction.Transaction)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		calculatedSenderShardId, err := txs.getShardFromAddress(tx.GetSndAddr())
		if err != nil {
			return nil, err
		}

		calculatedReceiverShardId, err := txs.getShardFromAddress(tx.GetRcvAddr())
		if err != nil {
			return nil, err
		}

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
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}
	if check.IfNil(header) {
		return process.ErrNilHeaderHandler
	}

	var err error
	scheduledMode := false
	if txs.flagScheduledMiniBlocks.IsSet() {
		scheduledMode, err = process.IsScheduledMode(header, body, txs.hasher, txs.marshalizer)
		if err != nil {
			return err
		}
	}

	txsToMe, err := txs.computeTxsToMe(body)
	if err != nil {
		return err
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
		"total gas refunded", txs.gasHandler.TotalGasRefunded(),
		"total gas penalized", txs.gasHandler.TotalGasPenalized(),
	)
	defer func() {
		log.Debug("transactions.processTxsToMe after processing",
			"scheduled mode", scheduledMode,
			"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard,
			"num scrs processed", numTXsProcessed,
			"total gas provided", txs.gasHandler.TotalGasProvided(),
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
			return process.ErrTimeIsOut
		}

		tx, ok := txsToMe[index].Tx.(*transaction.Transaction)
		if !ok {
			return process.ErrWrongTypeAssertion
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
			return errComputeGas
		}

		if scheduledMode {
			txs.gasHandler.SetGasProvidedAsScheduled(gasProvidedByTxInSelfShard, txHash)
		} else {
			txs.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, txHash)
		}

		txs.saveAccountBalanceForAddress(tx.GetRcvAddr())

		if scheduledMode {
			txs.scheduledTxsExecutionHandler.Add(txHash, tx)
		} else {
			err = txs.processAndRemoveBadTransaction(
				txHash,
				tx,
				senderShardID,
				receiverShardID)
			if err != nil {
				return err
			}

			txs.updateGasConsumedWithGasRefundedAndGasPenalized(txHash, &gasInfo)
		}

		numTXsProcessed++
	}

	return nil
}

func (txs *transactions) processTxsFromMe(
	body *block.Body,
	haveTime func() bool,
	randomness []byte,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	txsFromMe, err := txs.computeTxsFromMe(body)
	if err != nil {
		return err
	}

	txs.sortTransactionsBySenderAndNonce(txsFromMe, randomness)

	isShardStuckFalse := func(uint32) bool {
		return false
	}
	isMaxBlockSizeReachedFalse := func(int, int) bool {
		return false
	}
	haveAdditionalTimeFalse := func() bool {
		return false
	}

	calculatedMiniBlocks, _, mapSCTxs, err := txs.createAndProcessMiniBlocksFromMe(
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
		randomness,
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
	randomness []byte,
) (block.MiniBlockSlice, error) {

	if !txs.flagScheduledMiniBlocks.IsSet() {
		return make(block.MiniBlockSlice, 0), nil
	}

	scheduledTxsFromMe, err := txs.computeScheduledTxsFromMe(body)
	if err != nil {
		return nil, err
	}

	txs.sortTransactionsBySenderAndNonce(scheduledTxsFromMe, randomness)

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

func (txs *transactions) getRemainingGasPerBlock() uint64 {
	gasConsumed := txs.getTotalGasConsumed()
	maxGasPerBlock := txs.economicsFee.MaxGasLimitPerBlock(txs.shardCoordinator.SelfId())
	gasBandwidth := uint64(0)
	if gasConsumed < maxGasPerBlock {
		gasBandwidth = maxGasPerBlock - gasConsumed
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
	if txs.flagScheduledMiniBlocks.IsSet() {
		gasBandwidthForScheduled = txs.economicsFee.MaxGasLimitPerBlock(txs.shardCoordinator.SelfId()) * selectionGasBandwidthIncreaseScheduledPercent / 100
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
	scheduledMiniBlocks, err := txs.createAndProcessScheduledMiniBlocksFromMeAsProposer(
		haveTime,
		haveAdditionalTime,
		sortedTxsForScheduled,
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
	if !txs.flagOptimizeGasUsedInCrossMiniBlocks.IsSet() {
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
		txInfo, ok := txs.txsForCurrBlock.txHashAndInfo[string(txHash)]
		if !ok {
			log.Warn("transactions.splitMiniBlockIfNeeded: missing tx", "hash", txHash)
			currentMiniBlock.TxHashes = append(currentMiniBlock.TxHashes, txHash)
			continue
		}

		_, gasProvidedByTxInReceiverShard, err := txs.computeGasProvidedByTx(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txInfo.tx,
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

// ProcessMiniBlock processes all the transactions from a and saves the processed transactions in local cache complete miniblock
func (txs *transactions) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	getNumOfCrossInterMbsAndTxs func() (int, int),
	scheduledMode bool,
) (processedTxHashes [][]byte, numProcessedTxs int, err error) {

	if miniBlock.Type != block.TxBlock {
		return nil, 0, process.ErrWrongTypeInMiniBlock
	}

	numTXsProcessed := 0
	var gasProvidedByTxInSelfShard uint64
	processedTxHashes = make([][]byte, 0)
	miniBlockTxs, miniBlockTxHashes, err := txs.getAllTxsFromMiniBlock(miniBlock, haveTime, haveAdditionalTime)
	if err != nil {
		return nil, 0, err
	}

	if txs.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlockTxs)) {
		return nil, 0, process.ErrMaxBlockSizeReached
	}

	defer func() {
		if err != nil {
			for _, hash := range processedTxHashes {
				log.Trace("transactions.ProcessMiniBlock: defer func()", "tx hash", hash)
			}

			txs.gasHandler.RestoreGasSinceLastReset()
		}
	}()

	var totalGasConsumed uint64
	if scheduledMode {
		totalGasConsumed = txs.gasHandler.TotalGasProvidedAsScheduled()
	} else {
		totalGasConsumed = txs.getTotalGasConsumed()
	}

	var maxGasLimitUsedForDestMeTxs uint64
	isFirstMiniBlockDestMe := totalGasConsumed == 0
	if isFirstMiniBlockDestMe {
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
			"total gas refunded", txs.gasHandler.TotalGasRefunded(),
			"total gas penalized", txs.gasHandler.TotalGasPenalized(),
		)
	}()

	numOfOldCrossInterMbs, numOfOldCrossInterTxs := getNumOfCrossInterMbsAndTxs()

	for index := range miniBlockTxs {
		if !haveTime() && !haveAdditionalTime() {
			return processedTxHashes, index, process.ErrTimeIsOut
		}

		gasProvidedByTxInSelfShard, err = txs.computeGasProvided(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockTxs[index],
			miniBlockTxHashes[index],
			&gasInfo)

		if err != nil {
			return processedTxHashes, index, err
		}

		if scheduledMode {
			txs.gasHandler.SetGasProvidedAsScheduled(gasProvidedByTxInSelfShard, miniBlockTxHashes[index])
		} else {
			txs.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, miniBlockTxHashes[index])
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])

		if txs.flagOptimizeGasUsedInCrossMiniBlocks.IsSet() {
			if gasInfo.totalGasConsumedInSelfShard > maxGasLimitUsedForDestMeTxs {
				return processedTxHashes, index, process.ErrMaxGasLimitUsedForDestMeTxsIsReached
			}
		}

		txs.saveAccountBalanceForAddress(miniBlockTxs[index].GetRcvAddr())

		if !scheduledMode {
			_, err = txs.txProcessor.ProcessTransaction(miniBlockTxs[index])
			if err != nil {
				return processedTxHashes, index, err
			}

			txs.updateGasConsumedWithGasRefundedAndGasPenalized(miniBlockTxHashes[index], &gasInfo)
		}

		numTXsProcessed++
	}

	numOfCrtCrossInterMbs, numOfCrtCrossInterTxs := getNumOfCrossInterMbsAndTxs()
	numOfNewCrossInterMbs := numOfCrtCrossInterMbs - numOfOldCrossInterMbs
	numOfNewCrossInterTxs := numOfCrtCrossInterTxs - numOfOldCrossInterTxs

	log.Debug("transactions.ProcessMiniBlock",
		"scheduled mode", scheduledMode,
		"numOfOldCrossInterMbs", numOfOldCrossInterMbs, "numOfOldCrossInterTxs", numOfOldCrossInterTxs,
		"numOfCrtCrossInterMbs", numOfCrtCrossInterMbs, "numOfCrtCrossInterTxs", numOfCrtCrossInterTxs,
		"numOfNewCrossInterMbs", numOfNewCrossInterMbs, "numOfNewCrossInterTxs", numOfNewCrossInterTxs,
	)

	numMiniBlocks := 1 + numOfNewCrossInterMbs
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
	txs.flagOptimizeGasUsedInCrossMiniBlocks.SetValue(epoch >= txs.optimizeGasUsedInCrossMiniBlocksEnableEpoch)
	log.Debug("transactions: optimize gas used in cross mini blocks", "enabled", txs.flagOptimizeGasUsedInCrossMiniBlocks.IsSet())

	txs.flagScheduledMiniBlocks.SetValue(epoch >= txs.scheduledMiniBlocksEnableEpoch)
	log.Debug("transactions: scheduled mini blocks", "enabled", txs.flagScheduledMiniBlocks.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (txs *transactions) IsInterfaceNil() bool {
	return txs == nil
}

// sortTransactionsBySenderAndNonce sorts the provided transactions and hashes simultaneously
func (txs *transactions) sortTransactionsBySenderAndNonce(transactions []*txcache.WrappedTransaction, randomness []byte) {
	if !txs.flagFrontRunningProtection.IsSet() {
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

	if txs.flagScheduledMiniBlocks.IsSet() {
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
