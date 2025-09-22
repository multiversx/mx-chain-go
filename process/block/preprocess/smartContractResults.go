package preprocess

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sliceUtil"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

var _ process.DataMarshalizer = (*smartContractResults)(nil)
var _ process.PreProcessor = (*smartContractResults)(nil)

// SmartContractResultsArgs is the struct which contains all the arguments needed for creating a smartContractResult preprocessor
type SmartContractResultsArgs struct {
	BasePreProcessorArgs
	ScrProcessor process.SmartContractResultProcessor
}

type smartContractResults struct {
	*basePreProcess
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	scrForBlock                  TxsForBlockHandler
	scrPool                      dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	scrProcessor                 process.SmartContractResultProcessor
}

// NewSmartContractResultPreprocessor creates a new smartContractResult preprocessor object
func NewSmartContractResultPreprocessor(args SmartContractResultsArgs) (*smartContractResults, error) {
	err := CheckBasePreProcessArgs(args.BasePreProcessorArgs)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.ScrProcessor) {
		return nil, process.ErrNilTxProcessor
	}

	if check.IfNil(args.ProcessedMiniBlocksTracker) {
		return nil, process.ErrNilProcessedMiniBlocksTracker
	}
	err = core.CheckHandlerCompatibility(args.EnableEpochsHandler, []core.EnableEpochFlag{
		common.OptimizeGasUsedInCrossMiniBlocksFlag,
		common.ScheduledMiniBlocksFlag,
		common.FrontRunningProtectionFlag,
	})
	if err != nil {
		return nil, err
	}

	bpp := &basePreProcess{
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
		accountsProposal:           args.AccountsProposal,
		pubkeyConverter:            args.PubkeyConverter,
		enableEpochsHandler:        args.EnableEpochsHandler,
		processedMiniBlocksTracker: args.ProcessedMiniBlocksTracker,
		txExecutionOrderHandler:    args.TxExecutionOrderHandler,
	}

	scr := &smartContractResults{
		basePreProcess:               bpp,
		storage:                      args.Store,
		scrPool:                      args.DataPool,
		onRequestSmartContractResult: args.OnRequestTransaction,
		scrProcessor:                 args.ScrProcessor,
	}

	scr.scrPool.RegisterOnAdded(scr.receivedSmartContractResult)
	scr.scrForBlock, err = NewTxsForBlock(args.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	return scr, nil
}

// IsDataPrepared returns non error if all the requested smartContractResults arrived and were saved into the pool
func (scr *smartContractResults) IsDataPrepared(requestedScrs int, haveTime func() time.Duration) error {
	if requestedScrs > 0 {
		log.Debug("requested missing scrs",
			"num scrs", requestedScrs)
		err := scr.scrForBlock.WaitForRequestedData(haveTime())
		missingScrs := scr.scrForBlock.GetMissingTxsCount()
		// TODO: previously the number of missing txs was cleared in scrForBlock - check if this is still needed
		log.Debug("received missing scrs",
			"num scrs", requestedScrs-missingScrs)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveBlockDataFromPools removes smart contract results and miniblocks from associated pools
func (scr *smartContractResults) RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	return scr.removeBlockDataFromPools(body, miniBlockPool, scr.scrPool, scr.isMiniBlockCorrect)
}

// RemoveTxsFromPools removes smart contract results from associated pools
func (scr *smartContractResults) RemoveTxsFromPools(body *block.Body) error {
	return scr.removeTxsFromPools(body, scr.scrPool, scr.isMiniBlockCorrect)
}

// RestoreBlockDataIntoPools restores the smart contract results and miniblocks to associated pools
func (scr *smartContractResults) RestoreBlockDataIntoPools(
	body *block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	if check.IfNil(body) {
		return 0, process.ErrNilBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return 0, process.ErrNilMiniBlockPool
	}

	scrRestored := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		err := scr.restoreSmartContractResultsIntoPool(miniBlock)
		if err != nil {
			return scrRestored, err
		}

		// TODO: Should be analyzed if restoring into pool only cross-shard miniblocks with destination in self shard,
		// would create problems or not
		if miniBlock.SenderShardID != scr.shardCoordinator.SelfId() {
			miniBlockHash, errHash := core.CalculateHash(scr.marshalizer, scr.hasher, miniBlock)
			if errHash != nil {
				return scrRestored, errHash
			}

			miniBlockPool.Put(miniBlockHash, miniBlock, miniBlock.Size())
		}

		scrRestored += len(miniBlock.TxHashes)
	}

	return scrRestored, nil
}

func (scr *smartContractResults) restoreSmartContractResultsIntoPool(miniBlock *block.MiniBlock) error {
	strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	scrsBuff, err := scr.storage.GetAll(dataRetriever.UnsignedTransactionUnit, miniBlock.TxHashes)
	if err != nil {
		log.Debug("smart contract results from mini block were not found in UnsignedTransactionUnit",
			"sender shard ID", miniBlock.SenderShardID,
			"receiver shard ID", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes),
		)

		return err
	}

	for txHash, txBuff := range scrsBuff {
		tx := smartContractResult.SmartContractResult{}
		err = scr.marshalizer.Unmarshal(&tx, txBuff)
		if err != nil {
			return err
		}

		scr.scrPool.AddData([]byte(txHash), &tx, tx.Size(), strCache)
	}

	return nil
}

// ProcessBlockTransactions processes all the smartContractResult from the block.Body, updates the state
func (scr *smartContractResults) ProcessBlockTransactions(
	headerHandler data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	numSCRsProcessed := 0
	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard:  uint64(0),
		gasConsumedByMiniBlockInReceiverShard: uint64(0),
		totalGasConsumedInSelfShard:           scr.getTotalGasConsumed(),
	}

	log.Debug("smartContractResults.ProcessBlockTransactions: before processing",
		"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
		"total gas provided", scr.gasHandler.TotalGasProvided(),
		"total gas provided as scheduled", scr.gasHandler.TotalGasProvidedAsScheduled(),
		"total gas refunded", scr.gasHandler.TotalGasRefunded(),
		"total gas penalized", scr.gasHandler.TotalGasPenalized(),
	)
	defer func() {
		log.Debug("smartContractResults.ProcessBlockTransactions after processing",
			"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard,
			"num scrs processed", numSCRsProcessed,
			"total gas provided", scr.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", scr.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", scr.gasHandler.TotalGasRefunded(),
			"total gas penalized", scr.gasHandler.TotalGasPenalized(),
		)
	}()

	// basic validation already done in interceptors
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		// smart contract results are needed to be processed only at destination and only if they are cross shard
		if miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}
		if miniBlock.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		pi, err := scr.getIndexesOfLastTxProcessed(miniBlock, headerHandler)
		if err != nil {
			return err
		}

		indexOfFirstTxToBeProcessed := pi.indexOfLastTxProcessed + 1
		err = process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, pi.indexOfLastTxProcessedByProposer, miniBlock)
		if err != nil {
			return err
		}

		for j := indexOfFirstTxToBeProcessed; j <= pi.indexOfLastTxProcessedByProposer; j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]

			txInfoFromMap, ok := scr.scrForBlock.GetTxInfoByHash(txHash)
			if !ok || check.IfNil(txInfoFromMap.Tx) {
				log.Warn("missing transaction in ProcessBlockTransactions ", "type", miniBlock.Type, "txHash", txHash)
				return process.ErrMissingTransaction
			}

			currScr, ok := txInfoFromMap.Tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			if scr.enableEpochsHandler.IsFlagEnabled(common.OptimizeGasUsedInCrossMiniBlocksFlag) {
				gasProvidedByTxInSelfShard, err := scr.computeGasProvided(
					miniBlock.SenderShardID,
					miniBlock.ReceiverShardID,
					currScr,
					txHash,
					&gasInfo)

				if err != nil {
					return err
				}

				scr.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, txHash)
			}

			err = scr.saveAccountBalanceForAddress(currScr.GetRcvAddr())
			if err != nil {
				return err
			}

			scr.txExecutionOrderHandler.Add(txHash)
			_, err = scr.scrProcessor.ProcessSmartContractResult(currScr)
			if err != nil {
				return err
			}

			scr.updateGasConsumedWithGasRefundedAndGasPenalized(txHash, &gasInfo)
			numSCRsProcessed++
		}
	}

	return nil
}

// SaveTxsToStorage saves smart contract results from body into storage
func (scr *smartContractResults) SaveTxsToStorage(body *block.Body) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}
		if miniBlock.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		scr.saveTxsToStorage(miniBlock.TxHashes, scr.scrForBlock, scr.storage, dataRetriever.UnsignedTransactionUnit)
	}

	return nil
}

// receivedSmartContractResult is a call back function which is called when a new smartContractResult
// is added in the smartContractResult pool
func (scr *smartContractResults) receivedSmartContractResult(key []byte, value interface{}) {
	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("smartContractResults.receivedSmartContractResult", "error", process.ErrWrongTypeAssertion)
		return
	}

	scr.baseReceivedTransaction(key, tx, scr.scrForBlock)
}

// CreateBlockStarted cleans the local cache map for processed/created smartContractResults at this round
func (scr *smartContractResults) CreateBlockStarted() {
	scr.scrForBlock.Reset()
}

// RequestBlockTransactions request for smartContractResults if missing from a block.Body
func (scr *smartContractResults) RequestBlockTransactions(body *block.Body) int {
	if check.IfNil(body) {
		return 0
	}

	return scr.computeExistingAndRequestMissingSCResultsForShards(body)
}

// computeExistingAndRequestMissingSCResultsForShards calculates what smartContractResults are available and requests
// what are missing from block.Body
func (scr *smartContractResults) computeExistingAndRequestMissingSCResultsForShards(body *block.Body) int {
	scrTxs := block.Body{}
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.SmartContractResultBlock {
			continue
		}
		if mb.SenderShardID == scr.shardCoordinator.SelfId() {
			continue
		}

		scrTxs.MiniBlocks = append(scrTxs.MiniBlocks, mb)
	}

	numMissingTxsForShard := scr.computeExistingAndRequestMissing(
		&scrTxs,
		scr.scrForBlock,
		scr.isMiniBlockCorrect,
		scr.scrPool,
		scr.onRequestSmartContractResult,
	)

	return numMissingTxsForShard
}

// GetTransactionsAndRequestMissingForMiniBlock requests missing smartContractResults for a certain miniblock
func (scr *smartContractResults) GetTransactionsAndRequestMissingForMiniBlock(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
	if miniBlock == nil {
		return nil, 0
	}

	existingTxs, missingScrsHashesForMiniBlock := scr.computeMissingScrsHashesForMiniBlock(miniBlock)
	if len(missingScrsHashesForMiniBlock) > 0 {
		scr.onRequestSmartContractResult(miniBlock.SenderShardID, missingScrsHashesForMiniBlock)
	}

	return existingTxs, len(missingScrsHashesForMiniBlock)
}

// computeMissingScrsHashesForMiniBlock computes missing smart contract results hashes for a certain miniblock
func (scr *smartContractResults) computeMissingScrsHashesForMiniBlock(miniBlock *block.MiniBlock) ([]data.TransactionHandler, [][]byte) {
	missingSmartContractResultsHashes := make([][]byte, 0)
	existingTxs := make([]data.TransactionHandler, 0)

	if miniBlock.Type != block.SmartContractResultBlock {
		return existingTxs, missingSmartContractResultsHashes
	}

	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			scr.scrPool,
			process.SearchMethodPeekWithFallbackSearchFirst)

		if check.IfNil(tx) {
			missingSmartContractResultsHashes = append(missingSmartContractResultsHashes, txHash)
			continue
		}

		existingTxs = append(existingTxs, tx)
	}

	return existingTxs, missingSmartContractResultsHashes
}

// getAllScrsFromMiniBlock gets all the smartContractResults from a miniblock into a new structure
func (scr *smartContractResults) getAllScrsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*smartContractResult.SmartContractResult, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := scr.scrPool.ShardDataStore(strCache)
	if check.IfNil(txCache) {
		return nil, nil, process.ErrNilUTxDataPool
	}

	// verify if all smartContractResult exists
	scResSlice := make([]*smartContractResult.SmartContractResult, 0, len(mb.TxHashes))
	txHashes := make([][]byte, 0, len(mb.TxHashes))
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			tmp, _ = scr.scrPool.SearchFirstData(txHash)
			if tmp == nil {
				return nil, nil, process.ErrNilSmartContractResult
			}

			log.Debug("scr hash not found with Peek method but found with SearchFirstData",
				"scr hash", txHash,
				"strCache", strCache)
		}

		tx, ok := tmp.(*smartContractResult.SmartContractResult)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		txHashes = append(txHashes, txHash)
		scResSlice = append(scResSlice, tx)
	}

	return smartContractResult.TrimSlicePtr(scResSlice), sliceUtil.TrimSliceSliceByte(txHashes), nil
}

// SelectOutgoingTransactions returns an empty slice of byte slices, as this preprocessor does not handle outgoing transactions
func (scr *smartContractResults) SelectOutgoingTransactions(_ uint64) ([][]byte, []data.TransactionHandler, error) {
	return make([][]byte, 0), make([]data.TransactionHandler, 0), nil
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (scr *smartContractResults) CreateAndProcessMiniBlocks(_ func() bool, _ []byte) (block.MiniBlockSlice, error) {
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock processes all the smart contract results from the given miniblock and saves the processed ones in a local cache
func (scr *smartContractResults) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
	_ func() bool,
	_ bool,
	partialMbExecutionMode bool,
	indexOfLastTxProcessed int,
	preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler,
) ([][]byte, int, bool, error) {

	if miniBlock.Type != block.SmartContractResultBlock {
		return nil, indexOfLastTxProcessed, false, process.ErrWrongTypeInMiniBlock
	}

	numSCRsProcessed := 0
	var gasProvidedByTxInSelfShard uint64
	var err error
	var txIndex int
	processedTxHashes := make([][]byte, 0)

	indexOfFirstTxToBeProcessed := indexOfLastTxProcessed + 1
	err = process.CheckIfIndexesAreOutOfBound(int32(indexOfFirstTxToBeProcessed), int32(len(miniBlock.TxHashes))-1, miniBlock)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	miniBlockScrs, miniBlockTxHashes, err := scr.getAllScrsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	if scr.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlock.TxHashes)) {
		return nil, indexOfLastTxProcessed, false, process.ErrMaxBlockSizeReached
	}

	miniBlockHash, err := core.CalculateHash(scr.marshalizer, scr.hasher, miniBlock)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlockInReceiverShard: uint64(0),
		gasConsumedByMiniBlocksInSenderShard:  uint64(0),
		totalGasConsumedInSelfShard:           scr.getTotalGasConsumed(),
	}

	var maxGasLimitUsedForDestMeTxs uint64
	isFirstMiniBlockDestMe := gasInfo.totalGasConsumedInSelfShard == 0
	if isFirstMiniBlockDestMe {
		maxGasLimitUsedForDestMeTxs = scr.economicsFee.MaxGasLimitPerBlock(scr.shardCoordinator.SelfId())
	} else {
		maxGasLimitUsedForDestMeTxs = scr.economicsFee.MaxGasLimitPerBlock(scr.shardCoordinator.SelfId()) * maxGasLimitPercentUsedForDestMeTxs / 100
	}

	log.Debug("smartContractResults.ProcessMiniBlock: before processing",
		"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
		"total gas provided", scr.gasHandler.TotalGasProvided(),
		"total gas provided as scheduled", scr.gasHandler.TotalGasProvidedAsScheduled(),
		"total gas refunded", scr.gasHandler.TotalGasRefunded(),
		"total gas penalized", scr.gasHandler.TotalGasPenalized(),
	)
	defer func() {
		log.Debug("smartContractResults.ProcessMiniBlock after processing",
			"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard,
			"num scrs processed", numSCRsProcessed,
			"total gas provided", scr.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", scr.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", scr.gasHandler.TotalGasRefunded(),
			"total gas penalized", scr.gasHandler.TotalGasPenalized(),
		)
	}()

	for txIndex = indexOfFirstTxToBeProcessed; txIndex < len(miniBlockScrs); txIndex++ {
		if !haveTime() {
			err = process.ErrTimeIsOut
			break
		}

		gasProvidedByTxInSelfShard, err = scr.computeGasProvided(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockScrs[txIndex],
			miniBlockTxHashes[txIndex],
			&gasInfo)

		if err != nil {
			break
		}

		if scr.enableEpochsHandler.IsFlagEnabled(common.OptimizeGasUsedInCrossMiniBlocksFlag) {
			if gasInfo.totalGasConsumedInSelfShard > maxGasLimitUsedForDestMeTxs {
				err = process.ErrMaxGasLimitUsedForDestMeTxsIsReached
				break
			}
		}

		err = scr.saveAccountBalanceForAddress(miniBlockScrs[txIndex].GetRcvAddr())
		if err != nil {
			break
		}

		snapshot := scr.handleProcessTransactionInit(preProcessorExecutionInfoHandler, miniBlockTxHashes[txIndex], miniBlockHash)

		scr.txExecutionOrderHandler.Add(miniBlockTxHashes[txIndex])
		_, err = scr.scrProcessor.ProcessSmartContractResult(miniBlockScrs[txIndex])
		if err != nil {
			scr.handleProcessTransactionError(preProcessorExecutionInfoHandler, snapshot, miniBlockTxHashes[txIndex])
			break
		}

		scr.updateGasConsumedWithGasRefundedAndGasPenalized(miniBlockTxHashes[txIndex], &gasInfo)
		scr.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, miniBlockTxHashes[txIndex])
		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[txIndex])
		numSCRsProcessed++
	}

	if err != nil && !partialMbExecutionMode {
		return processedTxHashes, txIndex - 1, true, err
	}

	for index, txHash := range miniBlockTxHashes {
		scr.scrForBlock.AddTransaction(txHash, miniBlockScrs[index], miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	}

	scr.blockSizeComputation.AddNumMiniBlocks(1)
	scr.blockSizeComputation.AddNumTxs(len(miniBlock.TxHashes))

	return nil, txIndex - 1, false, err
}

// CreateMarshalledData marshals smart contract results hashes and saves them into a new structure
func (scr *smartContractResults) CreateMarshalledData(txHashes [][]byte) ([][]byte, error) {
	marshalledScrs, err := scr.createMarshalledData(txHashes, scr.scrForBlock)
	if err != nil {
		return nil, err
	}

	return marshalledScrs, nil
}

// GetAllCurrentUsedTxs returns all the smartContractResults used at current creation / processing
func (scr *smartContractResults) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	return scr.scrForBlock.GetAllCurrentUsedTxs()
}

// AddTxsFromMiniBlocks does nothing
func (scr *smartContractResults) AddTxsFromMiniBlocks(_ block.MiniBlockSlice) {
}

// AddTransactions does nothing
func (scr *smartContractResults) AddTransactions(_ []data.TransactionHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (scr *smartContractResults) IsInterfaceNil() bool {
	return scr == nil
}

func (scr *smartContractResults) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.SmartContractResultBlock
}
