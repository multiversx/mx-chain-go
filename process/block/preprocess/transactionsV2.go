package preprocess

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

func (txs *transactions) createAndProcessMiniBlocksFromMeV2(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, []*txcache.WrappedTransaction, map[string]struct{}, error) {
	log.Debug("createAndProcessMiniBlocksFromMeV2 has been started")

	mbInfo := txs.initCreateAndProcessMiniBlocks()

	log.Debug("createAndProcessMiniBlocksFromMeV2", "totalGasConsumedInSelfShard", mbInfo.gasInfo.totalGasConsumedInSelfShard)

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	f, err := os.Create(fmt.Sprintf("cpu-profile-%d-%d.pprof", time.Now().Unix(), len(sortedTxs)))
	if err != nil {
		log.Error("could not create CPU profile", "error", err)
	}
	var index int
	if len(sortedTxs) > 0 {
		debug.SetGCPercent(-1)
		pprof.StartCPUProfile(f)

		defer func() {
			pprof.StopCPUProfile()
			runtime.GC()

			log.Debug("createAndProcessMiniBlocksFromMeV2 has been finished", "num txs", len(sortedTxs), "last index", index)
		}()
	}
	remainingTxs := make([]*txcache.WrappedTransaction, 0)

	for index = range sortedTxs {
		if !haveTime() {
			log.Debug("time is out in createAndProcessMiniBlocksFromMeV2", "num txs", len(sortedTxs), "last index", index)
			remainingTxs = append(remainingTxs, sortedTxs[index:]...)
			break
		}

		tx, miniBlock, shouldContinue := txs.shouldContinueProcessingTx(
			isShardStuck,
			sortedTxs[index],
			mbInfo)
		if !shouldContinue {
			log.Debug("should not continue createAndProcessMiniBlocksFromMeV2", "num txs", len(sortedTxs), "last index", index)
			continue
		}

		txHash := sortedTxs[index].TxHash
		senderShardID := sortedTxs[index].SenderShardID
		receiverShardID := sortedTxs[index].ReceiverShardID

		isMiniBlockEmpty := len(miniBlock.TxHashes) == 0
		txMbInfo := txs.getTxAndMbInfo(
			tx,
			isMiniBlockEmpty,
			receiverShardID,
			mbInfo)

		if isMaxBlockSizeReached(txMbInfo.numNewMiniBlocks, txMbInfo.numNewTxs) {
			log.Debug("max txs accepted in one block is reached",
				"num txs added", mbInfo.processingInfo.numTxsAdded,
				"total txs", len(sortedTxs),
				"last index", index)
			break
		}

		shouldAddToRemaining, err := txs.processTransaction(
			tx,
			txHash,
			senderShardID,
			receiverShardID,
			mbInfo)
		if err != nil {
			if core.IsGetNodeFromDBError(err) {
				return nil, nil, nil, err
			}
			if shouldAddToRemaining {
				remainingTxs = append(remainingTxs, sortedTxs[index])
			}
			continue
		}

		txs.applyExecutedTransaction(
			tx,
			txHash,
			miniBlock,
			receiverShardID,
			txMbInfo,
			mbInfo)
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mbInfo.mapMiniBlocks)
	txs.displayProcessingResults(miniBlocks, len(sortedTxs), mbInfo)

	log.Debug("createAndProcessMiniBlocksFromMeV2 has been finished")

	return miniBlocks, remainingTxs, mbInfo.mapSCTxs, nil
}

func (txs *transactions) initGasConsumed() map[uint32]uint64 {
	mapGasConsumedByMiniBlockInReceiverShard := make(map[uint32]uint64)
	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapGasConsumedByMiniBlockInReceiverShard[shardID] = 0
	}

	mapGasConsumedByMiniBlockInReceiverShard[core.MetachainShardId] = 0
	return mapGasConsumedByMiniBlockInReceiverShard
}

func (txs *transactions) createEmptyMiniBlocks(blockType block.Type, reserved []byte) map[uint32]*block.MiniBlock {
	mapMiniBlocks := make(map[uint32]*block.MiniBlock)
	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapMiniBlocks[shardID] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), shardID, blockType, reserved)
	}

	mapMiniBlocks[core.MetachainShardId] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), core.MetachainShardId, blockType, reserved)
	return mapMiniBlocks
}

func (txs *transactions) hasAddressEnoughInitialBalance(tx *transaction.Transaction) bool {
	addressHasEnoughBalance := true
	isAddressSet := txs.balanceComputation.IsAddressSet(tx.GetSndAddr())
	if isAddressSet {
		addressHasEnoughBalance = txs.balanceComputation.AddressHasEnoughBalance(tx.GetSndAddr(), getTxMaxTotalCost(tx))
	}

	return addressHasEnoughBalance
}

func (txs *transactions) processTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	mbInfo *createAndProcessMiniBlocksInfo,
) (bool, error) {
	snapshot := txs.accounts.JournalLen()

	mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
	oldGasConsumedByMiniBlocksInSenderShard := mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := mbInfo.gasInfo.totalGasConsumedInSelfShard

	startTime := time.Now()
	gasProvidedByTxInSelfShard, err := txs.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		&mbInfo.gasInfo)
	elapsedTime := time.Since(startTime)
	mbInfo.processingInfo.totalTimeUsedForComputeGasProvided += elapsedTime
	if err != nil {
		log.Trace("processTransaction.computeGasProvided", "error", err)
		isTxTargetedForDeletion := errors.Is(err, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached)
		if isTxTargetedForDeletion {
			mbInfo.processingInfo.numCrossShardTxsWithTooMuchGas++
			strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
			txs.txPool.RemoveData(txHash, strCache)
			return false, err
		}
		return true, err
	}

	txs.gasHandler.SetGasProvided(gasProvidedByTxInSelfShard, txHash)
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard

	// execute transaction to change the trie root hash
	startTime = time.Now()
	err = txs.processAndRemoveBadTransaction(
		txHash,
		tx,
		senderShardID,
		receiverShardID,
	)
	elapsedTime = time.Since(startTime)
	mbInfo.processingInfo.totalTimeUsedForProcess += elapsedTime

	txs.accountTxsShards.Lock()
	txs.accountTxsShards.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.accountTxsShards.Unlock()

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		if errors.Is(err, process.ErrHigherNonceInTransaction) {
			mbInfo.senderAddressToSkip = tx.GetSndAddr()
		}

		mbInfo.processingInfo.numBadTxs++
		log.Trace("bad tx", "error", err.Error(), "hash", txHash)

		errRevert := txs.accounts.RevertToSnapshot(snapshot)
		if errRevert != nil && !core.IsClosingError(errRevert) {
			log.Warn("revert to snapshot", "error", errRevert.Error())
		}

		txs.gasHandler.RemoveGasProvided([][]byte{txHash})
		txs.gasHandler.RemoveGasRefunded([][]byte{txHash})
		txs.gasHandler.RemoveGasPenalized([][]byte{txHash})

		mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = oldGasConsumedByMiniBlockInReceiverShard
		mbInfo.gasInfo.totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

		return false, err
	}

	if senderShardID == receiverShardID {
		gasRefunded := txs.gasHandler.GasRefunded(txHash)
		gasPenalized := txs.gasHandler.GasPenalized(txHash)
		gasToBeSubtracted := gasRefunded + gasPenalized
		shouldDoTheSubtraction := gasToBeSubtracted <= mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard &&
			gasToBeSubtracted <= mbInfo.gasInfo.totalGasConsumedInSelfShard &&
			gasToBeSubtracted <= mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
		if shouldDoTheSubtraction {
			mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard -= gasToBeSubtracted
			mbInfo.gasInfo.totalGasConsumedInSelfShard -= gasToBeSubtracted
			mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] -= gasToBeSubtracted
		}
	}

	if errors.Is(err, process.ErrFailedTransaction) {
		log.Debug("transactions.processTransaction",
			"txHash", txHash,
			"nonce", tx.Nonce,
			"value", tx.Value,
			"gas price", tx.GasPrice,
			"gas limit", tx.GasLimit,
			"sender", senderShardID,
			"receiver", receiverShardID,
			"senderAddr", tx.GetSndAddr(),
			"senderUsername", tx.GetSndUserName(),
			"receiverAddr", tx.GetRcvAddr(),
			"receiverUsername", tx.GetRcvUserName(),
			"data", string(tx.Data),
			"err", err.Error(),
		)
		if !mbInfo.firstInvalidTxFound {
			mbInfo.firstInvalidTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		txs.blockSizeComputation.AddNumTxs(1)
		mbInfo.processingInfo.numTxsFailed++
	}

	return false, err
}

func (txs *transactions) getMiniBlockSliceFromMapV2(mapMiniBlocks map[uint32]*block.MiniBlock) block.MiniBlockSlice {
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

func (txs *transactions) createScheduledMiniBlocks(
	haveTime func() bool,
	haveAdditionalTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
) (block.MiniBlockSlice, error) {
	log.Debug("createScheduledMiniBlocks has been started")

	mbInfo := txs.initCreateScheduledMiniBlocks()
	for index := range sortedTxs {
		if !haveTime() && !haveAdditionalTime() {
			log.Debug("time is out in createScheduledMiniBlocks")
			break
		}

		tx, miniBlock, shouldContinue := txs.scheduledTXContinueFunc(
			isShardStuck,
			sortedTxs[index],
			mapSCTxs,
			mbInfo)
		if !shouldContinue {
			continue
		}

		txHash := sortedTxs[index].TxHash
		senderShardID := sortedTxs[index].SenderShardID
		receiverShardID := sortedTxs[index].ReceiverShardID

		isMiniBlockEmpty := len(miniBlock.TxHashes) == 0
		scheduledTxMbInfo := txs.getScheduledTxAndMbInfo(
			isMiniBlockEmpty,
			receiverShardID,
			mbInfo)

		if isMaxBlockSizeReached(scheduledTxMbInfo.numNewMiniBlocks, scheduledTxMbInfo.numNewTxs) {
			log.Debug("max txs accepted in one block is reached",
				"num scheduled txs added", mbInfo.schedulingInfo.numScheduledTxsAdded,
				"total txs", len(sortedTxs))
			break
		}

		err := txs.verifyTransaction(
			tx,
			txHash,
			senderShardID,
			receiverShardID,
			mbInfo)
		if err != nil {
			if core.IsGetNodeFromDBError(err) {
				return nil, err
			}
			continue
		}

		txs.applyVerifiedTransaction(
			tx,
			txHash,
			miniBlock,
			receiverShardID,
			scheduledTxMbInfo,
			mapSCTxs,
			mbInfo)
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mbInfo.mapMiniBlocks)
	txs.displayProcessingResultsOfScheduledMiniBlocks(miniBlocks, len(sortedTxs), mbInfo)

	log.Debug("createScheduledMiniBlocks has been finished")

	return miniBlocks, nil
}

func (txs *transactions) verifyTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	mbInfo *createScheduledMiniBlocksInfo,
) error {
	mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
	oldGasConsumedByMiniBlocksInSenderShard := mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := mbInfo.gasInfo.totalGasConsumedInSelfShard

	startTime := time.Now()
	gasProvidedByTxInSelfShard, err := txs.computeGasProvided(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		&mbInfo.gasInfo)
	elapsedTime := time.Since(startTime)
	mbInfo.schedulingInfo.totalTimeUsedForScheduledComputeGasProvided += elapsedTime
	if err != nil {
		log.Trace("verifyTransaction.computeGasProvided", "error", err)
		isTxTargetedForDeletion := errors.Is(err, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached)
		if isTxTargetedForDeletion {
			mbInfo.schedulingInfo.numCrossShardTxsWithTooMuchGas++
			strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
			txs.txPool.RemoveData(txHash, strCache)
		}
		return err
	}

	txs.gasHandler.SetGasProvidedAsScheduled(gasProvidedByTxInSelfShard, txHash)
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard

	startTime = time.Now()
	err = txs.txProcessor.VerifyTransaction(tx)
	elapsedTime = time.Since(startTime)
	mbInfo.schedulingInfo.totalTimeUsedForScheduledVerify += elapsedTime

	txs.accountTxsShards.Lock()
	txs.accountTxsShards.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.accountTxsShards.Unlock()

	executionErr, canExecute := txs.isTransactionEligibleForExecutionFunc(tx, err)
	if !canExecute {
		isTxTargetedForDeletion := errors.Is(executionErr, process.ErrLowerNonceInTransaction) || errors.Is(executionErr, process.ErrInsufficientFee) || errors.Is(executionErr, process.ErrTransactionNotExecutable)
		if isTxTargetedForDeletion {
			strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
			txs.txPool.RemoveData(txHash, strCache)
		}

		mbInfo.schedulingInfo.numScheduledBadTxs++
		log.Trace("bad tx", "error", executionErr, "hash", txHash)

		txs.gasHandler.RemoveGasProvidedAsScheduled([][]byte{txHash})

		mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = oldGasConsumedByMiniBlockInReceiverShard
		mbInfo.gasInfo.totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

		return executionErr
	}

	txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoToSet}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return nil
}

func (txs *transactions) isTransactionEligibleForExecution(_ *transaction.Transaction, err error) (error, bool) {
	return err, err == nil
}

func (txs *transactions) displayProcessingResults(
	miniBlocks block.MiniBlockSlice,
	nbSortedTxs int,
	mbInfo *createAndProcessMiniBlocksInfo,
) {
	if len(miniBlocks) > 0 {
		log.Debug("mini blocks from me created", "num", len(miniBlocks))
	}

	log.Debug("displayProcessingResults",
		"self shard", txs.shardCoordinator.SelfId(),
		"gas consumed in self shard for txs from me", mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard,
		"total gas consumed in self shard", mbInfo.gasInfo.totalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard", mbInfo.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID],
			"txs added", len(miniBlock.TxHashes),
			"non sc txs", mbInfo.mapTxsForShard[miniBlock.ReceiverShardID],
			"sc txs", mbInfo.mapScsForShard[miniBlock.ReceiverShardID])
	}

	log.Debug("transactions info", "total txs", nbSortedTxs,
		"num txs added", mbInfo.processingInfo.numTxsAdded,
		"num bad txs", mbInfo.processingInfo.numBadTxs,
		"num txs failed", mbInfo.processingInfo.numTxsFailed,
		"num txs skipped", mbInfo.processingInfo.numTxsSkipped,
		"num txs with initial balance consumed", mbInfo.processingInfo.numTxsWithInitialBalanceConsumed,
		"num cross shard sc calls or special txs", mbInfo.processingInfo.numCrossShardScCallsOrSpecialTxs,
		"num cross shard txs with too much gas", mbInfo.processingInfo.numCrossShardTxsWithTooMuchGas,
		"used time for computeGasProvided", mbInfo.processingInfo.totalTimeUsedForComputeGasProvided,
		"used time for processAndRemoveBadTransaction", mbInfo.processingInfo.totalTimeUsedForProcess,
	)
}

func (txs *transactions) displayProcessingResultsOfScheduledMiniBlocks(
	miniBlocks block.MiniBlockSlice,
	nbSortedTxs int,
	mbInfo *createScheduledMiniBlocksInfo,
) {
	if len(miniBlocks) > 0 {
		log.Debug("scheduled mini blocks from me created", "num", len(miniBlocks))
	}

	log.Debug("displayProcessingResultsOfScheduledMiniBlocks",
		"self shard", txs.shardCoordinator.SelfId(),
		"total gas consumed in self shard", mbInfo.gasInfo.totalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("scheduled mini block info",
			"type", "ScheduledBlock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard", mbInfo.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID],
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("scheduled transactions info", "total txs", nbSortedTxs,
		"num scheduled txs added", mbInfo.schedulingInfo.numScheduledTxsAdded,
		"num scheduled bad txs", mbInfo.schedulingInfo.numScheduledBadTxs,
		"num scheduled txs skipped", mbInfo.schedulingInfo.numScheduledTxsSkipped,
		"num scheduled txs with initial balance consumed", mbInfo.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed,
		"num scheduled cross shard sc calls", mbInfo.schedulingInfo.numScheduledCrossShardScCalls,
		"num cross shard txs with too much gas", mbInfo.schedulingInfo.numCrossShardTxsWithTooMuchGas,
		"used time for scheduled computeGasProvided", mbInfo.schedulingInfo.totalTimeUsedForScheduledComputeGasProvided,
		"used time for scheduled VerifyTransaction", mbInfo.schedulingInfo.totalTimeUsedForScheduledVerify,
	)
}

func (txs *transactions) initCreateAndProcessMiniBlocks() *createAndProcessMiniBlocksInfo {
	return &createAndProcessMiniBlocksInfo{
		mapSCTxs:                                 make(map[string]struct{}),
		mapTxsForShard:                           make(map[uint32]int),
		mapScsForShard:                           make(map[uint32]int),
		mapCrossShardScCallsOrSpecialTxs:         make(map[uint32]int),
		mapGasConsumedByMiniBlockInReceiverShard: txs.initGasConsumed(),
		mapMiniBlocks:                            txs.createEmptyMiniBlocks(block.TxBlock, nil),
		senderAddressToSkip:                      []byte(""),
		maxCrossShardScCallsOrSpecialTxsPerShard: 0,
		firstInvalidTxFound:                      false,
		firstCrossShardScCallOrSpecialTxFound:    false,
		processingInfo:                           processedTxsInfo{},
		gasInfo: gasConsumedInfo{
			gasConsumedByMiniBlockInReceiverShard: uint64(0),
			gasConsumedByMiniBlocksInSenderShard:  uint64(0),
			totalGasConsumedInSelfShard:           txs.getTotalGasConsumed(),
		},
	}
}

func (txs *transactions) shouldContinueProcessingTx(
	isShardStuck func(uint32) bool,
	wrappedTx *txcache.WrappedTransaction,
	mbInfo *createAndProcessMiniBlocksInfo,
) (*transaction.Transaction, *block.MiniBlock, bool) {

	txHash := wrappedTx.TxHash
	senderShardID := wrappedTx.SenderShardID
	receiverShardID := wrappedTx.ReceiverShardID

	if senderShardID != receiverShardID && isShardStuck != nil && isShardStuck(receiverShardID) {
		log.Trace("shard is stuck", "shard", receiverShardID)
		return nil, nil, false
	}

	tx, ok := wrappedTx.Tx.(*transaction.Transaction)
	if !ok {
		log.Debug("wrong type assertion",
			"hash", txHash,
			"sender shard", senderShardID,
			"receiver shard", receiverShardID)
		return nil, nil, false
	}

	if len(mbInfo.senderAddressToSkip) > 0 {
		if bytes.Equal(mbInfo.senderAddressToSkip, tx.GetSndAddr()) {
			mbInfo.processingInfo.numTxsSkipped++
			return nil, nil, false
		}
	}

	miniBlock, ok := mbInfo.mapMiniBlocks[receiverShardID]
	if !ok {
		log.Debug("mini block is not created", "shard", receiverShardID)
		return nil, nil, false
	}

	addressHasEnoughBalance := txs.hasAddressEnoughInitialBalance(tx)
	if !addressHasEnoughBalance {
		mbInfo.processingInfo.numTxsWithInitialBalanceConsumed++
		return nil, nil, false
	}

	return tx, miniBlock, true
}

func (txs *transactions) getTxAndMbInfo(
	tx *transaction.Transaction,
	isMiniBlockEmpty bool,
	receiverShardID uint32,
	mbInfo *createAndProcessMiniBlocksInfo,
) *txAndMbInfo {
	numNewMiniBlocks := 0
	if isMiniBlockEmpty {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	_, txTypeDstShard := txs.txTypeHandler.ComputeTransactionType(tx)
	isReceiverSmartContractAddress := txTypeDstShard == process.SCDeployment || txTypeDstShard == process.SCInvoking
	isCrossShardScCallOrSpecialTx := receiverShardID != txs.shardCoordinator.SelfId() &&
		(isReceiverSmartContractAddress || len(tx.RcvUserName) > 0)
	if isCrossShardScCallOrSpecialTx {
		if !mbInfo.firstCrossShardScCallOrSpecialTxFound {
			numNewMiniBlocks++
		}
		if mbInfo.mapCrossShardScCallsOrSpecialTxs[receiverShardID] >= mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard {
			numNewTxs += common.AdditionalScrForEachScCallOrSpecialTx
		}
	}

	return &txAndMbInfo{
		numNewTxs:                      numNewTxs,
		numNewMiniBlocks:               numNewMiniBlocks,
		isReceiverSmartContractAddress: isReceiverSmartContractAddress,
		isCrossShardScCallOrSpecialTx:  isCrossShardScCallOrSpecialTx,
	}
}

func (txs *transactions) applyExecutedTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	miniBlock *block.MiniBlock,
	receiverShardID uint32,
	txMbInfo *txAndMbInfo,
	mbInfo *createAndProcessMiniBlocksInfo,
) {
	mbInfo.senderAddressToSkip = []byte("")

	if txs.balanceComputation.IsAddressSet(tx.GetSndAddr()) {
		txMaxTotalCost := getTxMaxTotalCost(tx)
		ok := txs.balanceComputation.SubBalanceFromAddress(tx.GetSndAddr(), txMaxTotalCost)
		if !ok {
			log.Error("applyExecutedTransaction.SubBalanceFromAddress",
				"sender address", tx.GetSndAddr(),
				"tx max total cost", txMaxTotalCost,
				"err", process.ErrInsufficientFunds)
		}
	}

	if len(miniBlock.TxHashes) == 0 {
		txs.blockSizeComputation.AddNumMiniBlocks(1)
	}

	miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
	txs.blockSizeComputation.AddNumTxs(1)
	if txMbInfo.isCrossShardScCallOrSpecialTx {
		if !mbInfo.firstCrossShardScCallOrSpecialTxFound {
			mbInfo.firstCrossShardScCallOrSpecialTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}
		mbInfo.mapCrossShardScCallsOrSpecialTxs[receiverShardID]++
		crossShardScCallsOrSpecialTxs := mbInfo.mapCrossShardScCallsOrSpecialTxs[receiverShardID]
		if crossShardScCallsOrSpecialTxs > mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard {
			mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard = crossShardScCallsOrSpecialTxs
			//we need to increment this as to account for the corresponding SCR hash
			txs.blockSizeComputation.AddNumTxs(common.AdditionalScrForEachScCallOrSpecialTx)
		}
		mbInfo.processingInfo.numCrossShardScCallsOrSpecialTxs++
	}

	if txMbInfo.isReceiverSmartContractAddress {
		mbInfo.mapSCTxs[string(txHash)] = struct{}{}
		mbInfo.mapScsForShard[receiverShardID]++
	} else {
		mbInfo.mapTxsForShard[receiverShardID]++
	}

	mbInfo.processingInfo.numTxsAdded++
}

func (txs *transactions) initCreateScheduledMiniBlocks() *createScheduledMiniBlocksInfo {
	reserved, _ := (&block.MiniBlockReserved{
		ExecutionType:    block.Scheduled,
		TransactionsType: nil,
	}).Marshal()

	return &createScheduledMiniBlocksInfo{
		mapCrossShardScCallTxs:                   make(map[uint32]int),
		mapGasConsumedByMiniBlockInReceiverShard: txs.initGasConsumed(),
		mapMiniBlocks:                            txs.createEmptyMiniBlocks(block.TxBlock, reserved),
		senderAddressToSkip:                      []byte(""),
		maxCrossShardScCallTxsPerShard:           0,
		firstCrossShardScCallTxFound:             false,
		schedulingInfo:                           scheduledTxsInfo{},
		gasInfo: gasConsumedInfo{
			gasConsumedByMiniBlockInReceiverShard: uint64(0),
			gasConsumedByMiniBlocksInSenderShard:  uint64(0),
			totalGasConsumedInSelfShard:           txs.gasHandler.TotalGasProvidedAsScheduled(),
		},
	}
}

func (txs *transactions) shouldContinueProcessingScheduledTx(
	isShardStuck func(uint32) bool,
	wrappedTx *txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
	mbInfo *createScheduledMiniBlocksInfo,
) (*transaction.Transaction, *block.MiniBlock, bool) {

	txHash := wrappedTx.TxHash
	senderShardID := wrappedTx.SenderShardID
	receiverShardID := wrappedTx.ReceiverShardID

	_, alreadyAdded := mapSCTxs[string(txHash)]
	if alreadyAdded {
		return nil, nil, false
	}

	if senderShardID != receiverShardID && isShardStuck != nil && isShardStuck(receiverShardID) {
		log.Trace("shard is stuck", "shard", receiverShardID)
		return nil, nil, false
	}

	tx, ok := wrappedTx.Tx.(*transaction.Transaction)
	if !ok {
		log.Debug("wrong type assertion",
			"hash", txHash,
			"sender shard", senderShardID,
			"receiver shard", receiverShardID)
		return nil, nil, false
	}

	if len(mbInfo.senderAddressToSkip) > 0 {
		if bytes.Equal(mbInfo.senderAddressToSkip, tx.GetSndAddr()) {
			mbInfo.schedulingInfo.numScheduledTxsSkipped++
			return nil, nil, false
		}
	}

	mbInfo.senderAddressToSkip = tx.GetSndAddr()

	_, txTypeDstShard := txs.txTypeHandler.ComputeTransactionType(tx)
	isReceiverSmartContractAddress := txTypeDstShard == process.SCDeployment || txTypeDstShard == process.SCInvoking
	if !isReceiverSmartContractAddress {
		return nil, nil, false
	}

	miniBlock, ok := mbInfo.mapMiniBlocks[receiverShardID]
	if !ok {
		log.Debug("scheduled mini block is not created", "shard", receiverShardID)
		return nil, nil, false
	}

	addressHasEnoughBalance := txs.hasAddressEnoughInitialBalance(tx)
	if !addressHasEnoughBalance {
		mbInfo.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed++
		return nil, nil, false
	}

	return tx, miniBlock, true
}

func (txs *transactions) getScheduledTxAndMbInfo(
	isMiniBlockEmpty bool,
	receiverShardID uint32,
	mbInfo *createScheduledMiniBlocksInfo,
) *scheduledTxAndMbInfo {
	numNewMiniBlocks := 0
	if isMiniBlockEmpty {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	isCrossShardScCallTx := receiverShardID != txs.shardCoordinator.SelfId()
	if isCrossShardScCallTx {
		if !mbInfo.firstCrossShardScCallTxFound {
			numNewMiniBlocks++
		}
		if mbInfo.mapCrossShardScCallTxs[receiverShardID] >= mbInfo.maxCrossShardScCallTxsPerShard {
			numNewTxs += common.AdditionalScrForEachScCallOrSpecialTx
		}
	}

	return &scheduledTxAndMbInfo{
		numNewTxs:            numNewTxs,
		numNewMiniBlocks:     numNewMiniBlocks,
		isCrossShardScCallTx: isCrossShardScCallTx,
	}
}

func (txs *transactions) applyVerifiedTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	miniBlock *block.MiniBlock,
	receiverShardID uint32,
	scheduledTxMbInfo *scheduledTxAndMbInfo,
	mapSCTxs map[string]struct{},
	mbInfo *createScheduledMiniBlocksInfo,
) {
	if txs.balanceComputation.IsAddressSet(tx.GetSndAddr()) {
		txMaxTotalCost := getTxMaxTotalCost(tx)
		ok := txs.balanceComputation.SubBalanceFromAddress(tx.GetSndAddr(), txMaxTotalCost)
		if !ok {
			log.Error("applyVerifiedTransaction.SubBalanceFromAddress",
				"sender address", tx.GetSndAddr(),
				"tx max total cost", txMaxTotalCost,
				"err", process.ErrInsufficientFunds)
		}
	}

	if len(miniBlock.TxHashes) == 0 {
		txs.blockSizeComputation.AddNumMiniBlocks(1)
	}

	miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
	txs.blockSizeComputation.AddNumTxs(1)
	if scheduledTxMbInfo.isCrossShardScCallTx {
		if !mbInfo.firstCrossShardScCallTxFound {
			mbInfo.firstCrossShardScCallTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}
		mbInfo.mapCrossShardScCallTxs[receiverShardID]++
		crossShardScCallTxs := mbInfo.mapCrossShardScCallTxs[receiverShardID]
		if crossShardScCallTxs > mbInfo.maxCrossShardScCallTxsPerShard {
			mbInfo.maxCrossShardScCallTxsPerShard = crossShardScCallTxs
			//we need to increment this as to account for the corresponding SCR hash
			txs.blockSizeComputation.AddNumTxs(common.AdditionalScrForEachScCallOrSpecialTx)
		}
		mbInfo.schedulingInfo.numScheduledCrossShardScCalls++
	}

	mapSCTxs[string(txHash)] = struct{}{}
	mbInfo.schedulingInfo.numScheduledTxsAdded++
	txs.scheduledTxsExecutionHandler.AddScheduledTx(txHash, tx)
}
