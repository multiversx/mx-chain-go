package preprocess

import (
	"bytes"
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

type handleErrorInfo struct {
	err                                      error
	snapshot                                 int
	txHash                                   []byte
	tx                                       data.TransactionHandler
	senderShardID                            uint32
	receiverShardID                          uint32
	oldGasConsumedByMiniBlocksInSenderShard  uint64
	oldGasConsumedByMiniBlockInReceiverShard uint64
	oldTotalGasConsumedInSelfShard           uint64
	mbInfo                                   *createAndProcessMiniBlocksInfo
	txMbInfo                                 *txAndMbInfo
	isMaxBlockSizeReached                    func(int, int) bool
}

func (txs *transactions) createAndProcessMiniBlocksFromMeV2(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	log.Debug("createAndProcessMiniBlocksFromMeV2 has been started")

	mbInfo := txs.initCreateAndProcessMiniBlocks()

	log.Debug("createAndProcessMiniBlocksFromMeV2", "totalGasConsumedInSelfShard", mbInfo.gasInfo.TotalGasConsumedInSelfShard)

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

		txMbInfo := txs.getTxAndMbInfo(
			tx,
			miniBlock,
			mbInfo)

		if isMaxBlockSizeReached(txMbInfo.numNewMiniBlocks, txMbInfo.numNewTxs) {
			log.Debug("max txs accepted in one block is reached",
				"num txs added", mbInfo.processingInfo.numTxsAdded,
				"total txs", len(sortedTxs))
			break
		}

		err := txs.processTransaction(
			tx,
			txHash,
			senderShardID,
			receiverShardID,
			mbInfo,
			txMbInfo,
			isMaxBlockSizeReached,
		)
		if err != nil {
			log.Debug("createAndProcessMiniBlocksFromMeV2.processTransaction", "error", err)
			continue
		}

		txs.applyExecutedTransaction(
			tx,
			txHash,
			miniBlock,
			txMbInfo,
			mbInfo)

		txs.applyGeneratedSCRs(txHash, mbInfo)
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mbInfo.mapMiniBlocks)
	txs.applyTransactionsTypeForMiniBlocks(miniBlocks, mbInfo.mapTxHashTxType)
	txs.displayProcessingResults(miniBlocks, len(sortedTxs), mbInfo)

	log.Debug("createAndProcessMiniBlocksFromMeV2 has been finished")

	return miniBlocks, mbInfo.mapSCTxs, nil
}

func (txs *transactions) applyTransactionsTypeForMiniBlocks(miniBlocks block.MiniBlockSlice, mapTxHashTxType map[string]block.Type) {
	if !txs.flagMixedTxsInMiniBlocks.IsSet() {
		return
	}

	var txsType []block.Type
	var err error
	var mbr *block.MiniBlockReserved

	for _, miniBlock := range miniBlocks {
		txsType, err = getTxsTypeFromMiniBlock(miniBlock, mapTxHashTxType)
		if err != nil {
			log.Error("transactions.applyTransactionsTypeForMiniBlocks", "error", err)
			continue
		}

		mbr, err = miniBlock.GetMiniBlockReserved()
		if err != nil {
			log.Error("transactions.applyTransactionsTypeForMiniBlocks", "error", err)
			continue
		}
		if mbr == nil {
			mbr = &block.MiniBlockReserved{}
		}

		mbr.TransactionsType = generateTransactionsTypeBytes(txsType)
		err = miniBlock.SetMiniBlockReserved(mbr)
		if err != nil {
			log.Error("transactions.applyTransactionsTypeForMiniBlocks", "error", err)
			continue
		}
	}
}

func generateTransactionsTypeBytes(txsType []block.Type) []byte {
	numTxsType := len(txsType)
	transactionsTypeSize := numTxsType / 8
	if numTxsType%8 != 0 {
		transactionsTypeSize++
	}

	transactionsTypeBytes := make([]byte, transactionsTypeSize)
	for i := 0; i < numTxsType; i++ {
		if txsType[i] != block.TxBlock {
			transactionsTypeBytes[i/8] |= 1 << (uint16(i) % 8)
		}
	}

	return transactionsTypeBytes
}

func getTxsTypeFromMiniBlock(miniBlock *block.MiniBlock, mapTxHashTxType map[string]block.Type) ([]block.Type, error) {
	txsType := make([]block.Type, 0)
	for _, txHash := range miniBlock.TxHashes {
		txType, ok := mapTxHashTxType[string(txHash)]
		if !ok {
			return nil, process.ErrTxNotFound
		}

		txsType = append(txsType, txType)
	}

	return txsType, nil
}

func (txs *transactions) applyGeneratedSCRs(
	txHash []byte,
	mbInfo *createAndProcessMiniBlocksInfo,
) {
	if !txs.flagMixedTxsInMiniBlocks.IsSet() {
		return
	}

	mapAllIntermediateTxs := txs.postProcessorTxsHandler.GetAllIntermediateTxsForTxHash(txHash)
	for blockType, mapIntermediateTxsPerBlockType := range mapAllIntermediateTxs {
		if blockType != block.SmartContractResultBlock {
			continue
		}

		txs.addSCRsToMiniBlock(mapIntermediateTxsPerBlockType, mbInfo, blockType)
	}
}

func (txs *transactions) addSCRsToMiniBlock(
	mapIntermediateTxsPerBlockType map[uint32][]*process.TxInfo,
	mbInfo *createAndProcessMiniBlocksInfo,
	blockType block.Type,
) {
	for receiverShardID, intermediateTxsPerShard := range mapIntermediateTxsPerBlockType {
		if receiverShardID == txs.shardCoordinator.SelfId() {
			continue
		}

		miniBlock, ok := mbInfo.mapMiniBlocks[receiverShardID]
		if !ok {
			log.Debug("mini block is not created", "shard", receiverShardID)
			continue
		}

		if len(miniBlock.TxHashes) == 0 {
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		txs.blockSizeComputation.AddNumTxs(len(intermediateTxsPerShard))

		for _, txInfo := range intermediateTxsPerShard {
			miniBlock.TxHashes = append(miniBlock.TxHashes, txInfo.TxHash)
			txs.postProcessorTxsHandler.AddPostProcessorTx(txInfo.TxHash)
			mbInfo.mapTxHashTxType[string(txInfo.TxHash)] = blockType
		}
	}
}

func (txs *transactions) createEmptyMiniBlocks(blockType block.Type, isScheduledMiniBlock bool) map[uint32]*block.MiniBlock {
	mapMiniBlocks := make(map[uint32]*block.MiniBlock)
	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapMiniBlocks[shardID] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), shardID, blockType, isScheduledMiniBlock)
	}

	mapMiniBlocks[core.MetachainShardId] = txs.createEmptyMiniBlock(txs.shardCoordinator.SelfId(), core.MetachainShardId, blockType, isScheduledMiniBlock)
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
	txMbInfo *txAndMbInfo,
	isMaxBlockSizeReached func(int, int) bool,
) error {
	snapshot := txs.accounts.JournalLen()

	mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
	oldGasConsumedByMiniBlocksInSenderShard := mbInfo.gasInfo.GasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := mbInfo.gasInfo.TotalGasConsumedInSelfShard

	startTime := time.Now()
	gasConsumedByTxInSelfShard, err := txs.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		mbInfo.gasInfo)
	elapsedTime := time.Since(startTime)
	mbInfo.processingInfo.totalTimeUsedForComputeGasConsumed += elapsedTime
	if err != nil {
		log.Trace("processTransaction.computeGasConsumed", "error", err)
		return err
	}

	txs.gasHandler.SetGasConsumed(gasConsumedByTxInSelfShard, txHash)
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard

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

	hei := &handleErrorInfo{
		err:                                      err,
		snapshot:                                 snapshot,
		txHash:                                   txHash,
		tx:                                       tx,
		receiverShardID:                          receiverShardID,
		senderShardID:                            senderShardID,
		oldGasConsumedByMiniBlocksInSenderShard:  oldGasConsumedByMiniBlocksInSenderShard,
		oldGasConsumedByMiniBlockInReceiverShard: oldGasConsumedByMiniBlockInReceiverShard,
		oldTotalGasConsumedInSelfShard:           oldTotalGasConsumedInSelfShard,
		mbInfo:                                   mbInfo,
		txMbInfo:                                 txMbInfo,
		isMaxBlockSizeReached:                    isMaxBlockSizeReached,
	}

	return txs.handleErrorIfNeeded(hei)
}

func (txs *transactions) handleErrorIfNeeded(hei *handleErrorInfo) error {
	var shouldRevertGasAndAccountState bool

	defer func() {
		if shouldRevertGasAndAccountState {
			txs.revertGasAndAccountsState(hei)
		}
	}()

	mbInfo := hei.mbInfo
	txMbInfo := hei.txMbInfo
	tx := hei.tx

	isErrorAsideOfFailedTransaction := hei.err != nil && !errors.Is(hei.err, process.ErrFailedTransaction)
	if isErrorAsideOfFailedTransaction {
		if errors.Is(hei.err, process.ErrHigherNonceInTransaction) {
			mbInfo.senderAddressToSkip = tx.GetSndAddr()
		}

		mbInfo.processingInfo.numBadTxs++
		log.Trace("bad tx", "error", hei.err, "hash", hei.txHash)

		shouldRevertGasAndAccountState = true
		return hei.err
	}

	isErrorFailedTransaction := hei.err != nil && errors.Is(hei.err, process.ErrFailedTransaction)
	if isErrorFailedTransaction {
		if !mbInfo.firstInvalidTxFound {
			txMbInfo.numNewMiniBlocks++
		}
		txMbInfo.numNewTxs++

		if hei.isMaxBlockSizeReached(txMbInfo.numNewMiniBlocks, txMbInfo.numNewTxs) {
			shouldRevertGasAndAccountState = true
			return process.ErrMaxBlockSizeReached
		}
	}

	errCompute := txs.computeBlockSizeAndGasConsumedBySCRsFromTx(hei.txHash, mbInfo, txMbInfo, hei.isMaxBlockSizeReached)
	if errCompute != nil {
		shouldRevertGasAndAccountState = true
		return errCompute
	}

	txs.applyGeneratedSCRs(hei.txHash, mbInfo)

	if hei.senderShardID == hei.receiverShardID {
		gasRefunded := txs.gasHandler.GasRefunded(hei.txHash)
		mbInfo.gasInfo.GasConsumedByMiniBlocksInSenderShard -= gasRefunded
		mbInfo.gasInfo.TotalGasConsumedInSelfShard -= gasRefunded
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[hei.receiverShardID] -= gasRefunded
	}

	if isErrorFailedTransaction {
		if !mbInfo.firstInvalidTxFound {
			mbInfo.firstInvalidTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		txs.blockSizeComputation.AddNumTxs(1)
		mbInfo.processingInfo.numTxsFailed++
	}

	return hei.err
}

func (txs *transactions) revertGasAndAccountsState(hei *handleErrorInfo) {
	err := txs.accounts.RevertToSnapshot(hei.snapshot)
	if err != nil {
		log.Warn("revert to snapshot", "error", err.Error())
	}

	txs.gasHandler.RemoveGasConsumed([][]byte{hei.txHash})
	txs.gasHandler.RemoveGasRefunded([][]byte{hei.txHash})

	mbInfo := hei.mbInfo
	gasInfo := mbInfo.gasInfo

	gasInfo.GasConsumedByMiniBlocksInSenderShard = hei.oldGasConsumedByMiniBlocksInSenderShard
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[hei.receiverShardID] = hei.oldGasConsumedByMiniBlockInReceiverShard
	gasInfo.TotalGasConsumedInSelfShard = hei.oldTotalGasConsumedInSelfShard
}

func (txs *transactions) computeBlockSizeAndGasConsumedBySCRsFromTx(
	txHash []byte,
	mbInfo *createAndProcessMiniBlocksInfo,
	txMbInfo *txAndMbInfo,
	isMaxBlockSizeReached func(int, int) bool,
) error {
	if !txs.flagMixedTxsInMiniBlocks.IsSet() {
		return nil
	}

	var err error
	var oldMapGasConsumedByMiniBlockInReceiverShard map[uint32]uint64

	defer func() {
		if err != nil {
			for receiverShardID, gasConsumed := range oldMapGasConsumedByMiniBlockInReceiverShard {
				mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = gasConsumed
			}
		}
	}()

	oldMapGasConsumedByMiniBlockInReceiverShard = make(map[uint32]uint64)
	for receiverShardID, gasConsumed := range mbInfo.mapGasConsumedByMiniBlockInReceiverShard {
		oldMapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = gasConsumed
	}

	mapAllIntermediateTxsInfo := txs.postProcessorTxsHandler.GetAllIntermediateTxsForTxHash(txHash)
	for blockType, mapIntermediateTxsInfoPerBlockType := range mapAllIntermediateTxsInfo {
		if blockType != block.SmartContractResultBlock {
			continue
		}

		err = txs.computeBlockSizeAndGasConsumedBySCRs(mbInfo, txMbInfo, isMaxBlockSizeReached, mapIntermediateTxsInfoPerBlockType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txs *transactions) computeBlockSizeAndGasConsumedBySCRs(
	mbInfo *createAndProcessMiniBlocksInfo,
	txMbInfo *txAndMbInfo,
	isMaxBlockSizeReached func(int, int) bool,
	mapIntermediateTxsInfo map[uint32][]*process.TxInfo,
) error {
	var err error
	for receiverShardID, intermediateTxsInfoPerReceiverShard := range mapIntermediateTxsInfo {
		if receiverShardID == txs.shardCoordinator.SelfId() {
			continue
		}

		err = txs.computeBlockSizeGeneratedBySCRsInShard(mbInfo, txMbInfo, len(intermediateTxsInfoPerReceiverShard), receiverShardID, isMaxBlockSizeReached)
		if err != nil {
			return err
		}

		err = txs.computeGasConsumedBySCRsInShard(mbInfo, intermediateTxsInfoPerReceiverShard, receiverShardID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (txs *transactions) computeBlockSizeGeneratedBySCRsInShard(
	mbInfo *createAndProcessMiniBlocksInfo,
	txMbInfo *txAndMbInfo,
	numIntermediateTxs int,
	receiverShardID uint32,
	isMaxBlockSizeReached func(int, int) bool,
) error {
	miniBlock, ok := mbInfo.mapMiniBlocks[receiverShardID]
	if !ok {
		log.Debug("mini block is not created", "shard", receiverShardID)
		return process.ErrNilMiniBlock
	}

	if len(miniBlock.TxHashes) == 0 {
		txMbInfo.numNewMiniBlocks++
	}
	txMbInfo.numNewTxs += numIntermediateTxs

	if isMaxBlockSizeReached(txMbInfo.numNewMiniBlocks, txMbInfo.numNewTxs) {
		return process.ErrMaxBlockSizeReached
	}

	return nil
}

func (txs *transactions) computeGasConsumedBySCRsInShard(
	mbInfo *createAndProcessMiniBlocksInfo,
	intermediateTxsInfo []*process.TxInfo,
	receiverShardID uint32,
) error {
	mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
	for _, intermediateTxInfo := range intermediateTxsInfo {
		err := txs.computeGasConsumedByCrossScrInReceiverShard(mbInfo.gasInfo, intermediateTxInfo.Tx)
		if err != nil {
			return err
		}
	}
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard
	return nil
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

	return miniBlocks
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

	miniBlocks := txs.getMiniBlockSliceFromMapV2(mbInfo.mapMiniBlocks)
	txs.applyTransactionsTypeForMiniBlocks(miniBlocks, mbInfo.mapTxHashTxType)
	txs.displayProcessingResultsOfScheduledMiniBlocks(miniBlocks, len(sortedTxs), mbInfo)

	log.Debug("createScheduledMiniBlocks has been finished")

	return miniBlocks
}

func (txs *transactions) verifyTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	mbInfo *createScheduledMiniBlocksInfo,
) error {
	mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard = mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID]
	oldGasConsumedByMiniBlocksInSenderShard := mbInfo.gasInfo.GasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := mbInfo.gasInfo.TotalGasConsumedInSelfShard

	startTime := time.Now()
	gasConsumedByTxInSelfShard, err := txs.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		mbInfo.gasInfo)
	elapsedTime := time.Since(startTime)
	mbInfo.schedulingInfo.totalTimeUsedForScheduledComputeGasConsumed += elapsedTime
	if err != nil {
		log.Trace("verifyTransaction.computeGasConsumed", "error", err)
		return err
	}

	txs.gasHandler.SetGasConsumedAsScheduled(gasConsumedByTxInSelfShard, txHash)
	mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = mbInfo.gasInfo.GasConsumedByMiniBlockInReceiverShard

	startTime = time.Now()
	err = txs.txProcessor.VerifyTransaction(tx)
	elapsedTime = time.Since(startTime)
	mbInfo.schedulingInfo.totalTimeUsedForScheduledVerify += elapsedTime

	txs.accountTxsShards.Lock()
	txs.accountTxsShards.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.accountTxsShards.Unlock()

	if err != nil {
		isTxTargetedForDeletion := errors.Is(err, process.ErrLowerNonceInTransaction) || errors.Is(err, process.ErrInsufficientFee)
		if isTxTargetedForDeletion {
			strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
			txs.txPool.RemoveData(txHash, strCache)
		}

		mbInfo.schedulingInfo.numScheduledBadTxs++
		log.Trace("bad tx", "error", err.Error(), "hash", txHash)

		txs.gasHandler.RemoveGasConsumedAsScheduled([][]byte{txHash})

		mbInfo.gasInfo.GasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		mbInfo.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID] = oldGasConsumedByMiniBlockInReceiverShard
		mbInfo.gasInfo.TotalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

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
		"gas consumed in self shard for txs from me", mbInfo.gasInfo.GasConsumedByMiniBlocksInSenderShard,
		"total gas consumed in self shard", mbInfo.gasInfo.TotalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard", mbInfo.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID],
			"txs added", len(miniBlock.TxHashes),
			"txs", mbInfo.mapTxsForShard[miniBlock.ReceiverShardID])
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
		"total gas consumed in self shard", mbInfo.gasInfo.TotalGasConsumedInSelfShard)

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
		"used time for scheduled computeGasConsumed", mbInfo.schedulingInfo.totalTimeUsedForScheduledComputeGasConsumed,
		"used time for scheduled VerifyTransaction", mbInfo.schedulingInfo.totalTimeUsedForScheduledVerify,
	)
}

func (txs *transactions) initCreateAndProcessMiniBlocks() *createAndProcessMiniBlocksInfo {
	return &createAndProcessMiniBlocksInfo{
		mapSCTxs:                                 make(map[string]struct{}),
		mapTxsForShard:                           make(map[uint32]int),
		mapCrossShardScCallsOrSpecialTxs:         make(map[uint32]int),
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
		mapMiniBlocks:                            txs.createEmptyMiniBlocks(block.TxBlock, false),
		mapTxHashTxType:                          make(map[string]block.Type),
		senderAddressToSkip:                      []byte(""),
		maxCrossShardScCallsOrSpecialTxsPerShard: 0,
		firstInvalidTxFound:                      false,
		firstCrossShardScCallOrSpecialTxFound:    false,
		processingInfo:                           &processedTxsInfo{},
		gasInfo: &process.GasConsumedInfo{
			GasConsumedByMiniBlockInReceiverShard: uint64(0),
			GasConsumedByMiniBlocksInSenderShard:  uint64(0),
			TotalGasConsumedInSelfShard:           txs.gasHandler.TotalGasConsumed(),
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
	miniBlock *block.MiniBlock,
	mbInfo *createAndProcessMiniBlocksInfo,
) *txAndMbInfo {
	numNewMiniBlocks := 0
	if len(miniBlock.TxHashes) == 0 {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	_, txTypeDstShard := txs.txTypeHandler.ComputeTransactionType(tx)
	isReceiverSmartContractAddress := txTypeDstShard == process.SCDeployment || txTypeDstShard == process.SCInvoking
	isScCallOrSpecialTx := isReceiverSmartContractAddress || len(tx.RcvUserName) > 0
	isCrossShardScCallOrSpecialTx := isScCallOrSpecialTx && miniBlock.ReceiverShardID != txs.shardCoordinator.SelfId()
	if isCrossShardScCallOrSpecialTx {
		if !mbInfo.firstCrossShardScCallOrSpecialTxFound {
			numNewMiniBlocks++
		}
		if mbInfo.mapCrossShardScCallsOrSpecialTxs[miniBlock.ReceiverShardID] >= mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard {
			numNewTxs += common.AdditionalScrForEachScCallOrSpecialTx
		}
	}

	return &txAndMbInfo{
		numNewTxs:                      numNewTxs,
		numNewMiniBlocks:               numNewMiniBlocks,
		isReceiverSmartContractAddress: isReceiverSmartContractAddress,
		isScCallOrSpecialTx:            isScCallOrSpecialTx,
		isCrossShardScCallOrSpecialTx:  isCrossShardScCallOrSpecialTx,
	}
}

func (txs *transactions) applyExecutedTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	miniBlock *block.MiniBlock,
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

	isMiniBlockEmpty := len(miniBlock.TxHashes) == 0
	if isMiniBlockEmpty {
		txs.blockSizeComputation.AddNumMiniBlocks(1)
	}

	miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
	mbInfo.mapTxHashTxType[string(txHash)] = miniBlock.Type
	txs.blockSizeComputation.AddNumTxs(1)
	if txMbInfo.isCrossShardScCallOrSpecialTx {
		if !mbInfo.firstCrossShardScCallOrSpecialTxFound {
			mbInfo.firstCrossShardScCallOrSpecialTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}
		mbInfo.mapCrossShardScCallsOrSpecialTxs[miniBlock.ReceiverShardID]++
		crossShardScCallsOrSpecialTxs := mbInfo.mapCrossShardScCallsOrSpecialTxs[miniBlock.ReceiverShardID]
		if crossShardScCallsOrSpecialTxs > mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard {
			mbInfo.maxCrossShardScCallsOrSpecialTxsPerShard = crossShardScCallsOrSpecialTxs
			//we need to increment this as to account for the corresponding SCR hash
			txs.blockSizeComputation.AddNumTxs(common.AdditionalScrForEachScCallOrSpecialTx)
		}
		mbInfo.processingInfo.numCrossShardScCallsOrSpecialTxs++
	}

	if txMbInfo.isReceiverSmartContractAddress {
		mbInfo.mapSCTxs[string(txHash)] = struct{}{}
	}

	mbInfo.mapTxsForShard[miniBlock.ReceiverShardID]++
	mbInfo.processingInfo.numTxsAdded++
}

func (txs *transactions) initCreateScheduledMiniBlocks() *createScheduledMiniBlocksInfo {
	return &createScheduledMiniBlocksInfo{
		mapCrossShardScCallTxs:                   make(map[uint32]int),
		mapGasConsumedByMiniBlockInReceiverShard: make(map[uint32]uint64),
		mapMiniBlocks:                            txs.createEmptyMiniBlocks(block.TxBlock, true),
		mapTxHashTxType:                          make(map[string]block.Type),
		senderAddressToSkip:                      []byte(""),
		maxCrossShardScCallTxsPerShard:           0,
		firstCrossShardScCallTxFound:             false,
		schedulingInfo:                           &scheduledTxsInfo{},
		gasInfo: &process.GasConsumedInfo{
			GasConsumedByMiniBlockInReceiverShard: uint64(0),
			GasConsumedByMiniBlocksInSenderShard:  uint64(0),
			TotalGasConsumedInSelfShard:           txs.gasHandler.TotalGasConsumedAsScheduled(),
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
	mbInfo.mapTxHashTxType[string(txHash)] = miniBlock.Type
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
	txs.scheduledTxsExecutionHandler.Add(txHash, tx)
}
