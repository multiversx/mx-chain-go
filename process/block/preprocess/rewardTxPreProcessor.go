package preprocess

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

var _ process.DataMarshalizer = (*rewardTxPreprocessor)(nil)
var _ process.PreProcessor = (*rewardTxPreprocessor)(nil)

// RewardsPreProcessorArgs is the struct that contains all the dependencies needed for creating a reward transaction preprocessor
type RewardsPreProcessorArgs struct {
	BasePreProcessorArgs
	RewardProcessor process.RewardTransactionProcessor
}

type rewardTxPreprocessor struct {
	*basePreProcess
	onRequestRewardTx func(shardID uint32, txHashes [][]byte)
	rewardTxsForBlock TxsForBlockHandler
	rewardTxPool      dataRetriever.ShardedDataCacherNotifier
	storage           dataRetriever.StorageService
	rewardsProcessor  process.RewardTransactionProcessor
}

// NewRewardTxPreprocessor creates a new reward transaction preprocessor object
func NewRewardTxPreprocessor(args RewardsPreProcessorArgs) (*rewardTxPreprocessor, error) {
	err := checkBasePreProcessArgs(args.BasePreProcessorArgs)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.RewardProcessor) {
		return nil, process.ErrNilRewardsTxProcessor
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
		processedMiniBlocksTracker: args.ProcessedMiniBlocksTracker,
		txExecutionOrderHandler:    args.TxExecutionOrderHandler,
		enableEpochsHandler:        args.EnableEpochsHandler,
	}

	rtp := &rewardTxPreprocessor{
		basePreProcess:    bpp,
		storage:           args.Store,
		rewardTxPool:      args.DataPool,
		onRequestRewardTx: args.OnRequestTransaction,
		rewardsProcessor:  args.RewardProcessor,
	}

	rtp.rewardTxPool.RegisterOnAdded(rtp.receivedRewardTransaction)
	rtp.rewardTxsForBlock, err = NewTxsForBlock(args.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	return rtp, nil
}

// IsDataPrepared returns non error if all the requested reward transactions arrived and were saved into the pool
func (rtp *rewardTxPreprocessor) IsDataPrepared(requestedRewardTxs int, haveTime func() time.Duration) error {
	if requestedRewardTxs > 0 {
		log.Debug("requested missing reward txs",
			"num reward txs", requestedRewardTxs)
		err := rtp.rewardTxsForBlock.WaitForRequestedData(haveTime())
		missingRewardTxs := rtp.rewardTxsForBlock.GetMissingTxsCount()
		// TODO: previously the number of missing reward txs was cleared in rewardTxsForBlock - check if this is still needed
		log.Debug("received reward txs",
			"num reward txs", requestedRewardTxs-missingRewardTxs)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveBlockDataFromPools removes reward transactions and miniblocks from associated pools
func (rtp *rewardTxPreprocessor) RemoveBlockDataFromPools(body *block.Body, miniBlockPool storage.Cacher) error {
	return rtp.removeBlockDataFromPools(body, miniBlockPool, rtp.rewardTxPool, rtp.isMiniBlockCorrect)
}

// RemoveTxsFromPools removes reward transactions from associated pools
func (rtp *rewardTxPreprocessor) RemoveTxsFromPools(body *block.Body) error {
	return rtp.removeTxsFromPools(body, rtp.rewardTxPool, rtp.isMiniBlockCorrect)
}

// RestoreBlockDataIntoPools restores the reward transactions and miniblocks to associated pools
func (rtp *rewardTxPreprocessor) RestoreBlockDataIntoPools(
	body *block.Body,
	miniBlockPool storage.Cacher,
) (int, error) {
	if check.IfNil(body) {
		return 0, process.ErrNilBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return 0, process.ErrNilMiniBlockPool
	}

	rewardTxsRestored := 0
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		err := rtp.restoreRewardTxsIntoPool(miniBlock)
		if err != nil {
			return rewardTxsRestored, err
		}

		miniBlockHash, err := core.CalculateHash(rtp.marshalizer, rtp.hasher, miniBlock)
		if err != nil {
			return rewardTxsRestored, err
		}

		miniBlockPool.Put(miniBlockHash, miniBlock, miniBlock.Size())

		rewardTxsRestored += len(miniBlock.TxHashes)
	}

	return rewardTxsRestored, nil
}

func (rtp *rewardTxPreprocessor) restoreRewardTxsIntoPool(miniBlock *block.MiniBlock) error {
	strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	rewardTxsBuff, err := rtp.storage.GetAll(dataRetriever.RewardTransactionUnit, miniBlock.TxHashes)
	if err != nil {
		log.Debug("reward txs from mini block were not found in RewardTransactionUnit",
			"sender shard ID", miniBlock.SenderShardID,
			"receiver shard ID", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes),
		)

		return err
	}

	for txHash, txBuff := range rewardTxsBuff {
		tx := rewardTx.RewardTx{}
		err = rtp.marshalizer.Unmarshal(&tx, txBuff)
		if err != nil {
			return err
		}

		rtp.rewardTxPool.AddData([]byte(txHash), &tx, tx.Size(), strCache)
	}

	return nil
}

// ProcessBlockTransactions processes all the reward transactions from the block.Body, updates the state
func (rtp *rewardTxPreprocessor) ProcessBlockTransactions(
	headerHandler data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		pi, err := rtp.getIndexesOfLastTxProcessed(miniBlock, headerHandler)
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

			txData, ok := rtp.rewardTxsForBlock.GetTxInfoByHash(txHash)
			if !ok || check.IfNil(txData.Tx) {
				log.Warn("missing rewardsTransaction in ProcessBlockTransactions ", "type", miniBlock.Type, "hash", txHash)
				return process.ErrMissingTransaction
			}

			rTx, ok := txData.Tx.(*rewardTx.RewardTx)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err = rtp.saveAccountBalanceForAddress(rTx.GetRcvAddr())
			if err != nil {
				return err
			}

			rtp.txExecutionOrderHandler.Add(txHash)
			err = rtp.rewardsProcessor.ProcessRewardTransaction(rTx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetCreatedMiniBlocksFromMe returns nil as this preprocessor does not create any mini blocks
func (rtp *rewardTxPreprocessor) GetCreatedMiniBlocksFromMe() block.MiniBlockSlice {
	return make(block.MiniBlockSlice, 0)
}

// GetUnExecutableTransactions returns an empty map as reward transactions are always executable
func (rtp *rewardTxPreprocessor) GetUnExecutableTransactions() map[string]struct{} {
	return make(map[string]struct{})
}

// SaveTxsToStorage saves the reward transactions from body into storage
func (rtp *rewardTxPreprocessor) SaveTxsToStorage(body *block.Body) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		rtp.saveTxsToStorage(
			miniBlock.TxHashes,
			rtp.rewardTxsForBlock,
			rtp.storage,
			dataRetriever.RewardTransactionUnit,
		)
	}

	return nil
}

// receivedRewardTransaction is a callback function called when a new reward transaction
// is added in the reward transactions pool
func (rtp *rewardTxPreprocessor) receivedRewardTransaction(key []byte, value interface{}) {
	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("rewardTxPreprocessor.receivedRewardTransaction", "error", process.ErrWrongTypeAssertion)
		return
	}

	rtp.baseReceivedTransaction(key, tx, rtp.rewardTxsForBlock)
}

// CreateBlockStarted cleans the local cache map for processed/created reward transactions at this round
func (rtp *rewardTxPreprocessor) CreateBlockStarted() {
	rtp.rewardTxsForBlock.Reset()
}

// RequestBlockTransactions request for reward transactions if missing from a block.Body
func (rtp *rewardTxPreprocessor) RequestBlockTransactions(body *block.Body) int {
	if check.IfNil(body) {
		return 0
	}

	return rtp.computeExistingAndRequestMissingRewardTxsForShards(body)
}

// computeExistingAndRequestMissingRewardTxsForShards calculates what reward transactions are available and requests
// what are missing from block.Body
func (rtp *rewardTxPreprocessor) computeExistingAndRequestMissingRewardTxsForShards(body *block.Body) int {
	rewardTxsBody := block.Body{}
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.RewardsBlock {
			continue
		}
		if mb.SenderShardID != core.MetachainShardId {
			continue
		}

		rewardTxsBody.MiniBlocks = append(rewardTxsBody.MiniBlocks, mb)
	}

	numMissingTxsForShards := rtp.computeExistingAndRequestMissing(
		&rewardTxsBody,
		rtp.rewardTxsForBlock,
		rtp.isMiniBlockCorrect,
		rtp.rewardTxPool,
		rtp.onRequestRewardTx,
	)

	return numMissingTxsForShards
}

// GetTransactionsAndRequestMissingForMiniBlock returns the reward transactions from pool and requests missing for a certain miniblock
func (rtp *rewardTxPreprocessor) GetTransactionsAndRequestMissingForMiniBlock(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
	if miniBlock == nil {
		return nil, 0
	}

	existingTxs, missingRewardTxsHashesForMiniBlock := rtp.computeMissingRewardTxsHashesForMiniBlock(miniBlock)
	if len(missingRewardTxsHashesForMiniBlock) > 0 {
		rtp.onRequestRewardTx(miniBlock.SenderShardID, missingRewardTxsHashesForMiniBlock)
	}

	return existingTxs, len(missingRewardTxsHashesForMiniBlock)
}

// computeMissingRewardTxsHashesForMiniBlock computes missing reward transactions hashes for a certain miniblock
func (rtp *rewardTxPreprocessor) computeMissingRewardTxsHashesForMiniBlock(miniBlock *block.MiniBlock) ([]data.TransactionHandler, [][]byte) {
	missingRewardTxsHashes := make([][]byte, 0)
	existingTxs := make([]data.TransactionHandler, 0)

	if miniBlock.Type != block.RewardsBlock {
		return existingTxs, missingRewardTxsHashes
	}

	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			rtp.rewardTxPool,
			process.SearchMethodJustPeek,
		)

		if check.IfNil(tx) {
			missingRewardTxsHashes = append(missingRewardTxsHashes, txHash)
			continue
		}

		existingTxs = append(existingTxs, tx)
	}

	return existingTxs, missingRewardTxsHashes
}

// getAllRewardTxsFromMiniBlock gets all the reward transactions from a miniblock into a new structure
func (rtp *rewardTxPreprocessor) getAllRewardTxsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*rewardTx.RewardTx, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := rtp.rewardTxPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilRewardTxDataPool
	}

	// verify if all reward transactions exists
	rewardTxs := make([]*rewardTx.RewardTx, len(mb.TxHashes))
	txHashes := make([][]byte, len(mb.TxHashes))
	for idx, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, ok := txCache.Peek(txHash)
		if !ok {
			return nil, nil, process.ErrNilRewardTransaction
		}

		tx, ok := tmp.(*rewardTx.RewardTx)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		txHashes[idx] = txHash
		rewardTxs[idx] = tx
	}

	return rewardTxs, txHashes, nil
}

// SelectOutgoingTransactions does nothing as rewards transactions are created by meta chain
func (rtp *rewardTxPreprocessor) SelectOutgoingTransactions(_ uint64) ([][]byte, []data.TransactionHandler, error) {
	return make([][]byte, 0), make([]data.TransactionHandler, 0), nil
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (rtp *rewardTxPreprocessor) CreateAndProcessMiniBlocks(
	_ func() bool,
	_ []byte,
) (block.MiniBlockSlice, error) {
	// rewards are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock processes all the reward transactions from the given miniblock and saves the processed ones in a local cache
func (rtp *rewardTxPreprocessor) ProcessMiniBlock(
	miniBlock *block.MiniBlock,
	haveTime func() bool,
	_ func() bool,
	_ bool,
	partialMbExecutionMode bool,
	indexOfLastTxProcessed int,
	preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler,
) ([][]byte, int, bool, error) {

	var err error
	var txIndex int

	if miniBlock.Type != block.RewardsBlock {
		return nil, indexOfLastTxProcessed, false, process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return nil, indexOfLastTxProcessed, false, process.ErrRewardMiniBlockNotFromMeta
	}

	indexOfFirstTxToBeProcessed := indexOfLastTxProcessed + 1
	err = process.CheckIfIndexesAreOutOfBound(int32(indexOfFirstTxToBeProcessed), int32(len(miniBlock.TxHashes))-1, miniBlock)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	miniBlockRewardTxs, miniBlockTxHashes, err := rtp.getAllRewardTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	if rtp.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlock.TxHashes)) {
		return nil, indexOfLastTxProcessed, false, process.ErrMaxBlockSizeReached
	}

	miniBlockHash, err := core.CalculateHash(rtp.marshalizer, rtp.hasher, miniBlock)
	if err != nil {
		return nil, indexOfLastTxProcessed, false, err
	}

	processedTxHashes := make([][]byte, 0)
	for txIndex = indexOfFirstTxToBeProcessed; txIndex < len(miniBlockRewardTxs); txIndex++ {
		if !haveTime() {
			err = process.ErrTimeIsOut
			break
		}

		err = rtp.saveAccountBalanceForAddress(miniBlockRewardTxs[txIndex].GetRcvAddr())
		if err != nil {
			break
		}

		snapshot := rtp.handleProcessTransactionInit(preProcessorExecutionInfoHandler, miniBlockTxHashes[txIndex], miniBlockHash)

		rtp.txExecutionOrderHandler.Add(miniBlockTxHashes[txIndex])
		err = rtp.rewardsProcessor.ProcessRewardTransaction(miniBlockRewardTxs[txIndex])
		if err != nil {
			rtp.handleProcessTransactionError(preProcessorExecutionInfoHandler, snapshot, miniBlockTxHashes[txIndex])
			break
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[txIndex])
	}

	if err != nil && !partialMbExecutionMode {
		return processedTxHashes, txIndex - 1, true, err
	}

	for index, txHash := range miniBlockTxHashes {
		rtp.rewardTxsForBlock.AddTransaction(txHash, miniBlockRewardTxs[index], miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	}

	rtp.blockSizeComputation.AddNumMiniBlocks(1)
	rtp.blockSizeComputation.AddNumTxs(len(miniBlock.TxHashes))

	return nil, txIndex - 1, false, err
}

// CreateMarshalledData marshals reward transactions hashes and saves them into a new structure
func (rtp *rewardTxPreprocessor) CreateMarshalledData(txHashes [][]byte) ([][]byte, error) {
	marshalledRewardTxs, err := rtp.createMarshalledData(txHashes, rtp.rewardTxsForBlock)
	if err != nil {
		return nil, err
	}

	return marshalledRewardTxs, nil
}

// GetAllCurrentUsedTxs returns all the reward transactions used at current creation / processing
func (rtp *rewardTxPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	return rtp.rewardTxsForBlock.GetAllCurrentUsedTxs()
}

// AddTxsFromMiniBlocks does nothing
func (rtp *rewardTxPreprocessor) AddTxsFromMiniBlocks(_ block.MiniBlockSlice) {
}

// AddTransactions does nothing
func (rtp *rewardTxPreprocessor) AddTransactions(_ []data.TransactionHandler) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *rewardTxPreprocessor) IsInterfaceNil() bool {
	return rtp == nil
}

func (rtp *rewardTxPreprocessor) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.RewardsBlock
}
