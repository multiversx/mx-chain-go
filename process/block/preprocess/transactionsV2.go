package preprocess

import (
	"bytes"
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

func (txs *transactions) createAndProcessMiniBlocksFromMeV2(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	log.Debug("createAndProcessMiniBlocksFromMeV2 has been started")

	mbInfo := txs.initCreateAndProcessMiniBlocks()

	log.Debug("createAndProcessMiniBlocksFromMeV2", "totalGasConsumedInSelfShard", mbInfo.gasInfo.totalGasConsumedInSelfShard)

	defer func() {
		go txs.notifyTransactionProviderIfNeeded()
	}()

	for index := range sortedTxs {
		if !haveTime() {
			log.Debug("time is out in createAndProcessMiniBlocksFromMeV2")
			break
		}

		tx, miniBlock, shouldContinue := txs.shouldContinueProcessingTx(
			isShardStuck,
			sortedTxs[index],
			mbInfo)
		if !shouldContinue {
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
				"total txs", len(sortedTxs))
			break
		}

		err := txs.processTransaction(
			tx,
			txMbInfo.txType,
			txHash,
			senderShardID,
			receiverShardID,
			mbInfo)
		if err != nil {
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

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mbInfo.mapMiniBlocks, mbInfo.mapSCTxs)
	txs.displayProcessingResults(miniBlocks, len(sortedTxs), mbInfo)

	log.Debug("createAndProcessMiniBlocksFromMeV2 has been finished")

	return miniBlocks, mbInfo.mapSCTxs, nil
}

func (txs *transactions) initGasConsumed() map[uint32]map[txType]uint64 {
	mapGasConsumedByMiniBlockInReceiverShard := make(map[uint32]map[txType]uint64)
	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapGasConsumedByMiniBlockInReceiverShard[shardID] = make(map[txType]uint64)
	}

	mapGasConsumedByMiniBlockInReceiverShard[core.MetachainShardId] = make(map[txType]uint64)
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
		addressHasEnoughBalance = txs.balanceComputation.AddressHasEnoughBalance(tx.GetSndAddr(), txs.getTxMaxTotalCost(tx))
	}

	return addressHasEnoughBalance
}

func (txs *transactions) processTransaction(
	tx *transaction.Transaction,
	txType txType,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	mbInfo *createAndProcessMiniBlocksInfo,
) error {
	snapshot := txs.accounts.JournalLen()

	mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType]
	oldGasConsumedByMiniBlocksInSenderShard := mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := mbInfo.gasInfo.totalGasConsumedInSelfShard

	startTime := time.Now()
	err := txs.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		&mbInfo.gasInfo)
	elapsedTime := time.Since(startTime)
	mbInfo.processingInfo.totalTimeUsedForComputeGasConsumed += elapsedTime
	if err != nil {
		log.Trace("processTransaction.computeGasConsumed", "error", err)
		return err
	}

	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard

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

	txs.mutAccountsInfo.Lock()
	txs.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.mutAccountsInfo.Unlock()

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		if errors.Is(err, process.ErrHigherNonceInTransaction) {
			mbInfo.senderAddressToSkip = tx.GetSndAddr()
		}

		mbInfo.processingInfo.numBadTxs++
		log.Trace("bad tx", "error", err.Error(), "hash", txHash)

		errRevert := txs.accounts.RevertToSnapshot(snapshot)
		if errRevert != nil {
			log.Warn("revert to snapshot", "error", errRevert.Error())
		}

		txs.gasHandler.RemoveGasConsumed([][]byte{txHash})
		txs.gasHandler.RemoveGasRefunded([][]byte{txHash})

		mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = oldGasConsumedByMiniBlockInReceiverShard
		mbInfo.gasInfo.totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

		return err
	}

	if senderShardID == receiverShardID {
		gasRefunded := txs.gasHandler.GasRefunded(txHash)
		mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard -= gasRefunded
		mbInfo.gasInfo.totalGasConsumedInSelfShard -= gasRefunded
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] -= gasRefunded
	}

	if errors.Is(err, process.ErrFailedTransaction) {
		if !mbInfo.firstInvalidTxFound {
			mbInfo.firstInvalidTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		txs.blockSizeComputation.AddNumTxs(1)
		mbInfo.processingInfo.numTxsFailed++
	}

	return err
}

func (txs *transactions) getMiniBlockSliceFromMapV2(
	mapMiniBlocks map[uint32]*block.MiniBlock,
	mapSCTxs map[string]struct{},
) block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		if miniBlock, ok := mapMiniBlocks[shardID]; ok {
			if len(miniBlock.TxHashes) > 0 {
				miniBlocks = append(miniBlocks, splitMiniBlockIfNeeded(miniBlock, mapSCTxs)...)
			}
		}
	}

	if miniBlock, ok := mapMiniBlocks[core.MetachainShardId]; ok {
		if len(miniBlock.TxHashes) > 0 {
			miniBlocks = append(miniBlocks, splitMiniBlockIfNeeded(miniBlock, mapSCTxs)...)
		}
	}

	return miniBlocks
}

func splitMiniBlockIfNeeded(miniBlock *block.MiniBlock, mapSCTxs map[string]struct{}) block.MiniBlockSlice {
	splitMiniBlocks := make(block.MiniBlockSlice, 0)
	nonScTxHashes := make([][]byte, 0)
	scTxHashes := make([][]byte, 0)

	for _, txHash := range miniBlock.TxHashes {
		_, isSCTx := mapSCTxs[string(txHash)]
		if !isSCTx {
			nonScTxHashes = append(nonScTxHashes, txHash)
			continue
		}

		scTxHashes = append(scTxHashes, txHash)
	}

	if len(nonScTxHashes) > 0 {
		nonScMiniBlock := &block.MiniBlock{
			TxHashes:        nonScTxHashes,
			SenderShardID:   miniBlock.SenderShardID,
			ReceiverShardID: miniBlock.ReceiverShardID,
			Type:            miniBlock.Type,
			Reserved:        miniBlock.Reserved,
		}

		splitMiniBlocks = append(splitMiniBlocks, nonScMiniBlock)
	}

	if len(scTxHashes) > 0 {
		scMiniBlock := &block.MiniBlock{
			TxHashes:        scTxHashes,
			SenderShardID:   miniBlock.SenderShardID,
			ReceiverShardID: miniBlock.ReceiverShardID,
			Type:            miniBlock.Type,
			Reserved:        miniBlock.Reserved,
		}

		splitMiniBlocks = append(splitMiniBlocks, scMiniBlock)
	}

	return splitMiniBlocks
}

func (txs *transactions) createScheduledMiniBlocks(
	haveTime func() bool,
	haveAdditionalTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
	mapSCTxs map[string]struct{},
) block.MiniBlockSlice {
	log.Debug("createScheduledMiniBlocks has been started")

	mbInfo := txs.initCreateScheduledMiniBlocks()

	for index := range sortedTxs {
		if !haveTime() && !haveAdditionalTime() {
			log.Debug("time is out in createScheduledMiniBlocks")
			break
		}

		tx, miniBlock, shouldContinue := txs.shouldContinueProcessingScheduledTx(
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
			scTx,
			txHash,
			senderShardID,
			receiverShardID,
			mbInfo)
		if err != nil {
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

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mbInfo.mapMiniBlocks, mapSCTxs)

	txs.displayProcessingResultsOfScheduledMiniBlocks(miniBlocks, len(sortedTxs), mbInfo)

	log.Debug("createScheduledMiniBlocks has been finished")

	return miniBlocks
}

func (txs *transactions) verifyTransaction(
	tx *transaction.Transaction,
	txType txType,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	mbInfo *createScheduledMiniBlocksInfo,
) error {
	mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType]
	oldGasConsumedByMiniBlocksInSenderShard := mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := mbInfo.gasInfo.totalGasConsumedInSelfShard

	startTime := time.Now()
	err := txs.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		&mbInfo.gasInfo)
	elapsedTime := time.Since(startTime)
	mbInfo.schedulingInfo.totalTimeUsedForScheduledComputeGasConsumed += elapsedTime
	if err != nil {
		log.Trace("verifyTransaction.computeGasConsumed", "error", err)
		return err
	}

	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = mbInfo.gasInfo.gasConsumedByMiniBlockInReceiverShard

	startTime = time.Now()
	err = txs.txProcessor.VerifyTransaction(tx)
	elapsedTime = time.Since(startTime)
	mbInfo.schedulingInfo.totalTimeUsedForScheduledVerify += elapsedTime

	txs.mutAccountsInfo.Lock()
	txs.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.mutAccountsInfo.Unlock()

	if err != nil {
		isTxTargetedForDeletion := errors.Is(err, process.ErrLowerNonceInTransaction) || errors.Is(err, process.ErrInsufficientFee)
		if isTxTargetedForDeletion {
			strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
			txs.txPool.RemoveData(txHash, strCache)
		}

		mbInfo.schedulingInfo.numScheduledBadTxs++
		log.Trace("bad tx", "error", err.Error(), "hash", txHash)

		txs.gasHandler.RemoveGasConsumed([][]byte{txHash})

		mbInfo.gasInfo.gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = oldGasConsumedByMiniBlockInReceiverShard
		mbInfo.gasInfo.totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

		return err
	}

	txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoToSet}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return nil
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
			"gas consumed in receiver shard for non sc txs", mbInfo.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID][nonScTx],
			"gas consumed in receiver shard for sc txs", mbInfo.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID][scTx],
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
		"used time for computeGasConsumed", mbInfo.processingInfo.totalTimeUsedForComputeGasConsumed,
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
			"gas consumed in receiver shard", mbInfo.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID][scTx],
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("scheduled transactions info", "total txs", nbSortedTxs,
		"num scheduled txs added", mbInfo.schedulingInfo.numScheduledTxsAdded,
		"num scheduled bad txs", mbInfo.schedulingInfo.numScheduledBadTxs,
		"num scheduled txs skipped", mbInfo.schedulingInfo.numScheduledTxsSkipped,
		"num scheduled txs with initial balance consumed", mbInfo.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed,
		"num scheduled cross shard sc calls", mbInfo.schedulingInfo.numScheduledCrossShardScCalls,
		"used time for scheduled computeGasConsumed", mbInfo.schedulingInfo.totalTimeUsedForScheduledComputeGasConsumed,
		"used time for scheduled VerifyTransaction", mbInfo.schedulingInfo.totalTimeUsedForScheduledVerify,
	)
}

func getAdditionalTimeFunc() func() bool {
	startTime := time.Now()
	return func() bool {
		return additionalTimeForCreatingScheduledMiniBlocks > time.Since(startTime)
	}
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
			totalGasConsumedInSelfShard:           txs.gasHandler.TotalGasConsumed(),
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

	if isShardStuck != nil && isShardStuck(receiverShardID) {
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
	_, txTypeDstShard := txs.txTypeHandler.ComputeTransactionType(tx)
	isReceiverSmartContractAddress := txTypeDstShard == process.SCDeployment || txTypeDstShard == process.SCInvoking
	scTxType := nonScTx
	if isReceiverSmartContractAddress {
		scTxType = scTx
	}

	numNewMiniBlocks := 0
	if isMiniBlockEmpty || mbInfo.firstCrossShardScCallOrSpecialTxFound {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	isCrossShardScCallOrSpecialTx := receiverShardID != txs.shardCoordinator.SelfId() &&
		(isReceiverSmartContractAddress || len(tx.RcvUserName) > 0)
	if isCrossShardScCallOrSpecialTx {
		if !mbInfo.firstCrossShardScCallOrSpecialTxFound {
			numNewMiniBlocks++
		}
		if mbInfo.mapCrossShardScCallsOrSpecialTxs[receiverShardID] >= mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard {
			numNewTxs += core.AdditionalScrForEachScCallOrSpecialTx
		}
	}

	return &txAndMbInfo{
		numNewTxs:                      numNewTxs,
		numNewMiniBlocks:               numNewMiniBlocks,
		isReceiverSmartContractAddress: isReceiverSmartContractAddress,
		isCrossShardScCallOrSpecialTx:  isCrossShardScCallOrSpecialTx,
		txType:                         scTxType,
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
		txMaxTotalCost := txs.getTxMaxTotalCost(tx)
		ok := txs.balanceComputation.SubBalanceFromAddress(tx.GetSndAddr(), txMaxTotalCost)
		if !ok {
			log.Error("applyExecutedTransaction.SubBalanceFromAddress",
				"sender address", tx.GetSndAddr(),
				"tx max total cost", txMaxTotalCost,
				"err", process.ErrInsufficientFunds)
		}
	}

	isMiniBlockEmpty := len(miniBlock.TxHashes) == 0
	isFirstSplitFound := txs.isFirstMiniBlockSplitForReceiverShardFound(receiverShardID, txMbInfo, mbInfo)
	if isMiniBlockEmpty || isFirstSplitFound {
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
			txs.blockSizeComputation.AddNumTxs(core.AdditionalScrForEachScCallOrSpecialTx)
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

func (txs *transactions) isFirstMiniBlockSplitForReceiverShardFound(
	receiverShardID uint32,
	txMbInfo *txAndMbInfo,
	mbInfo *createAndProcessMiniBlocksInfo,
) bool {
	numTxsForShard := mbInfo.mapTxsForShard[receiverShardID]
	numScsForShard := mbInfo.mapScsForShard[receiverShardID]

	if numTxsForShard == 0 && numScsForShard == 0 {
		return false
	}
	if numTxsForShard > 0 && numScsForShard > 0 {
		return false
	}

	if txMbInfo.isReceiverSmartContractAddress {
		return numScsForShard == 0
	}

	return numTxsForShard == 0
}

func (txs *transactions) initCreateScheduledMiniBlocks() *createScheduledMiniBlocksInfo {
	return &createScheduledMiniBlocksInfo{
		mapCrossShardScCallTxs:                   make(map[uint32]int),
		mapGasConsumedByMiniBlockInReceiverShard: txs.initGasConsumed(),
		mapMiniBlocks:                            txs.createEmptyMiniBlocks(block.TxBlock, []byte{byte(block.ScheduledBlock)}),
		senderAddressToSkip:                      []byte(""),
		maxCrossShardScCallTxsPerShard:           0,
		firstCrossShardScCallTxFound:             false,
		schedulingInfo:                           scheduledTxsInfo{},
		gasInfo: gasConsumedInfo{
			gasConsumedByMiniBlockInReceiverShard: uint64(0),
			gasConsumedByMiniBlocksInSenderShard:  uint64(0),
			totalGasConsumedInSelfShard:           uint64(0),
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

	if isShardStuck != nil && isShardStuck(receiverShardID) {
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
			numNewTxs += core.AdditionalScrForEachScCallOrSpecialTx
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
		txMaxTotalCost := txs.getTxMaxTotalCost(tx)
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
			txs.blockSizeComputation.AddNumTxs(core.AdditionalScrForEachScCallOrSpecialTx)
		}
		mbInfo.schedulingInfo.numScheduledCrossShardScCalls++
	}

	mapSCTxs[string(txHash)] = struct{}{}
	mbInfo.schedulingInfo.numScheduledTxsAdded++
	txs.scheduledTxsExecutionHandler.Add(txHash, tx)
}
