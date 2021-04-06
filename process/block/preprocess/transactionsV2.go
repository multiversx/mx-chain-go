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

// TxType identifies the type of the tx
type TxType int32

const (
	nonScTx TxType = 0
	scTx    TxType = 1
)

const additionalTimeForCreatingScheduledMiniBlocks = 150 * time.Millisecond

type processedTxsInfo struct {
	numTxsAdded                        int
	numBadTxs                          int
	numTxsSkipped                      int
	numTxsFailed                       int
	numTxsWithInitialBalanceConsumed   int
	numCrossShardScCallsOrSpecialTxs   int
	totalTimeUsedForProcess            time.Duration
	totalTimeUsedForComputeGasConsumed time.Duration
}

type scheduledTxsInfo struct {
	numScheduledTxsAdded                        int
	numScheduledBadTxs                          int
	numScheduledTxsSkipped                      int
	numScheduledTxsWithInitialBalanceConsumed   int
	numScheduledCrossShardScCalls               int
	totalTimeUsedForScheduledVerify             time.Duration
	totalTimeUsedForScheduledComputeGasConsumed time.Duration
}

type createAndProcessMiniBlocksInfo struct {
	mapSCTxs                                 map[string]struct{}
	mapTxsForShard                           map[uint32]int
	mapScsForShard                           map[uint32]int
	mapCrossShardScCallsOrSpecialTxs         map[uint32]int
	mapGasConsumedByMiniBlockInReceiverShard map[uint32]map[TxType]uint64
	mapMiniBlocks                            map[uint32]*block.MiniBlock
	senderAddressToSkip                      []byte
	maxCrossShardScCallsOrSpecialTxsPerShard int
	firstInvalidTxFound                      bool
	firstCrossShardScCallOrSpecialTxFound    bool
	processingInfo                           processedTxsInfo
	gasInfo                                  gasConsumedInfo
}

type txAndMbInfo struct {
	numNewTxs                      int
	numNewMiniBlocks               int
	isReceiverSmartContractAddress bool
	isCrossShardScCallOrSpecialTx  bool
	txType                         TxType
}

type createScheduledMiniBlocksInfo struct {
	mapMiniBlocks                            map[uint32]*block.MiniBlock
	mapCrossShardScCallTxs                   map[uint32]int
	maxCrossShardScCallTxsPerShard           int
	schedulingInfo                           scheduledTxsInfo
	firstCrossShardScCallTxFound             bool
	mapGasConsumedByMiniBlockInReceiverShard map[uint32]map[TxType]uint64
	gasInfo                                  gasConsumedInfo
	senderAddressToSkip                      []byte
}

type scheduledTxAndMbInfo struct {
	numNewTxs            int
	numNewMiniBlocks     int
	isCrossShardScCallTx bool
}

func (txs *transactions) initGasConsumed() map[uint32]map[TxType]uint64 {
	mapGasConsumedByMiniBlockInReceiverShard := make(map[uint32]map[TxType]uint64)
	for shardID := uint32(0); shardID < txs.shardCoordinator.NumberOfShards(); shardID++ {
		mapGasConsumedByMiniBlockInReceiverShard[shardID] = make(map[TxType]uint64)
	}

	mapGasConsumedByMiniBlockInReceiverShard[core.MetachainShardId] = make(map[TxType]uint64)
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

func getAdditionalTimeFunc() func() bool {
	startTime := time.Now()
	return func() bool {
		return additionalTimeForCreatingScheduledMiniBlocks > time.Since(startTime)
	}
}

func (txs *transactions) createAndProcessMiniBlocksFromMeV2(
	haveTime func() bool,
	isShardStuck func(uint32) bool,
	isMaxBlockSizeReached func(int, int) bool,
	sortedTxs []*txcache.WrappedTransaction,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	log.Debug("createAndProcessMiniBlocksFromMeV2 has been started")

	cpmbi := txs.initCreateAndProcessMiniBlocks()

	log.Debug("createAndProcessMiniBlocksFromMeV2", "totalGasConsumedInSelfShard", cpmbi.gasInfo.totalGasConsumedInSelfShard)

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
			cpmbi)
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
			cpmbi)

		if isMaxBlockSizeReached(txMbInfo.numNewMiniBlocks, txMbInfo.numNewTxs) {
			log.Debug("max txs accepted in one block is reached",
				"num txs added", cpmbi.processingInfo.numTxsAdded,
				"total txs", len(sortedTxs))
			break
		}

		err := txs.processTransaction(
			tx,
			txMbInfo.txType,
			txHash,
			senderShardID,
			receiverShardID,
			cpmbi)
		if err != nil {
			continue
		}

		txs.applyExecutedTransaction(
			tx,
			txHash,
			miniBlock,
			receiverShardID,
			txMbInfo,
			cpmbi)
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(cpmbi.mapMiniBlocks, cpmbi.mapSCTxs)
	txs.displayProcessingResults(miniBlocks, len(sortedTxs), cpmbi)

	log.Debug("createAndProcessMiniBlocksFromMeV2 has been finished")

	return miniBlocks, cpmbi.mapSCTxs, nil
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
	cpmbi *createAndProcessMiniBlocksInfo,
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

	if len(cpmbi.senderAddressToSkip) > 0 {
		if bytes.Equal(cpmbi.senderAddressToSkip, tx.GetSndAddr()) {
			cpmbi.processingInfo.numTxsSkipped++
			return nil, nil, false
		}
	}

	miniBlock, ok := cpmbi.mapMiniBlocks[receiverShardID]
	if !ok {
		log.Debug("mini block is not created", "shard", receiverShardID)
		return nil, nil, false
	}

	addressHasEnoughBalance := txs.hasAddressEnoughInitialBalance(tx)
	if !addressHasEnoughBalance {
		cpmbi.processingInfo.numTxsWithInitialBalanceConsumed++
		return nil, nil, false
	}

	return tx, miniBlock, true
}

func (txs *transactions) getTxAndMbInfo(
	tx *transaction.Transaction,
	isMiniBlockEmpty bool,
	receiverShardID uint32,
	cpmbi *createAndProcessMiniBlocksInfo,
) *txAndMbInfo {
	_, txTypeDstShard := txs.txTypeHandler.ComputeTransactionType(tx)
	isReceiverSmartContractAddress := txTypeDstShard == process.SCDeployment || txTypeDstShard == process.SCInvoking
	txType := nonScTx
	if isReceiverSmartContractAddress {
		txType = scTx
	}

	numNewMiniBlocks := 0
	if isMiniBlockEmpty || cpmbi.firstCrossShardScCallOrSpecialTxFound {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	isCrossShardScCallOrSpecialTx := receiverShardID != txs.shardCoordinator.SelfId() &&
		(isReceiverSmartContractAddress || len(tx.RcvUserName) > 0)
	if isCrossShardScCallOrSpecialTx {
		if !cpmbi.firstCrossShardScCallOrSpecialTxFound {
			numNewMiniBlocks++
		}
		if cpmbi.mapCrossShardScCallsOrSpecialTxs[receiverShardID] >= cpmbi.maxCrossShardScCallsOrSpecialTxsPerShard {
			numNewTxs += core.AdditionalScrForEachScCallOrSpecialTx
		}
	}

	return &txAndMbInfo{
		numNewTxs:                      numNewTxs,
		numNewMiniBlocks:               numNewMiniBlocks,
		isReceiverSmartContractAddress: isReceiverSmartContractAddress,
		isCrossShardScCallOrSpecialTx:  isCrossShardScCallOrSpecialTx,
		txType:                         txType,
	}
}

func (txs *transactions) processTransaction(
	tx *transaction.Transaction,
	txType TxType,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	cpmbi *createAndProcessMiniBlocksInfo,
) error {
	snapshot := txs.accounts.JournalLen()

	cpmbi.gasInfo.gasConsumedByMiniBlockInReceiverShard = cpmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType]
	oldGasConsumedByMiniBlocksInSenderShard := cpmbi.gasInfo.gasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := cpmbi.gasInfo.gasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := cpmbi.gasInfo.totalGasConsumedInSelfShard

	startTime := time.Now()
	err := txs.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		&cpmbi.gasInfo)
	elapsedTime := time.Since(startTime)
	cpmbi.processingInfo.totalTimeUsedForComputeGasConsumed += elapsedTime
	if err != nil {
		log.Trace("processTransaction.computeGasConsumed", "error", err)
		return err
	}

	cpmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = cpmbi.gasInfo.gasConsumedByMiniBlockInReceiverShard

	// execute transaction to change the trie root hash
	startTime = time.Now()
	err = txs.processAndRemoveBadTransaction(
		txHash,
		tx,
		senderShardID,
		receiverShardID,
	)
	elapsedTime = time.Since(startTime)
	cpmbi.processingInfo.totalTimeUsedForProcess += elapsedTime

	txs.mutAccountsInfo.Lock()
	txs.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.mutAccountsInfo.Unlock()

	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		if errors.Is(err, process.ErrHigherNonceInTransaction) {
			cpmbi.senderAddressToSkip = tx.GetSndAddr()
		}

		cpmbi.processingInfo.numBadTxs++
		log.Trace("bad tx", "error", err.Error(), "hash", txHash)

		errRevert := txs.accounts.RevertToSnapshot(snapshot)
		if errRevert != nil {
			log.Warn("revert to snapshot", "error", errRevert.Error())
		}

		txs.gasHandler.RemoveGasConsumed([][]byte{txHash})
		txs.gasHandler.RemoveGasRefunded([][]byte{txHash})

		cpmbi.gasInfo.gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		cpmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = oldGasConsumedByMiniBlockInReceiverShard
		cpmbi.gasInfo.totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

		return err
	}

	gasRefunded := txs.gasHandler.GasRefunded(txHash)
	cpmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] -= gasRefunded
	if senderShardID == receiverShardID {
		cpmbi.gasInfo.gasConsumedByMiniBlocksInSenderShard -= gasRefunded
		cpmbi.gasInfo.totalGasConsumedInSelfShard -= gasRefunded
	}

	if errors.Is(err, process.ErrFailedTransaction) {
		if !cpmbi.firstInvalidTxFound {
			cpmbi.firstInvalidTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}

		txs.blockSizeComputation.AddNumTxs(1)
		cpmbi.processingInfo.numTxsFailed++
	}

	return err
}

func (txs *transactions) applyExecutedTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	miniBlock *block.MiniBlock,
	receiverShardID uint32,
	txMbInfo *txAndMbInfo,
	cpmbi *createAndProcessMiniBlocksInfo,
) {
	cpmbi.senderAddressToSkip = []byte("")

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
	firstMiniBlockSplitForReceiverShardFound := (txMbInfo.isReceiverSmartContractAddress && cpmbi.mapScsForShard[receiverShardID] == 0 && cpmbi.mapTxsForShard[receiverShardID] > 0) ||
		(!txMbInfo.isReceiverSmartContractAddress && cpmbi.mapTxsForShard[receiverShardID] == 0 && cpmbi.mapScsForShard[receiverShardID] > 0)
	if isMiniBlockEmpty || firstMiniBlockSplitForReceiverShardFound {
		txs.blockSizeComputation.AddNumMiniBlocks(1)
	}

	miniBlock.TxHashes = append(miniBlock.TxHashes, txHash)
	txs.blockSizeComputation.AddNumTxs(1)
	if txMbInfo.isCrossShardScCallOrSpecialTx {
		if !cpmbi.firstCrossShardScCallOrSpecialTxFound {
			cpmbi.firstCrossShardScCallOrSpecialTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}
		cpmbi.mapCrossShardScCallsOrSpecialTxs[receiverShardID]++
		if cpmbi.mapCrossShardScCallsOrSpecialTxs[receiverShardID] > cpmbi.maxCrossShardScCallsOrSpecialTxsPerShard {
			cpmbi.maxCrossShardScCallsOrSpecialTxsPerShard = cpmbi.mapCrossShardScCallsOrSpecialTxs[receiverShardID]
			//we need to increment this as to account for the corresponding SCR hash
			txs.blockSizeComputation.AddNumTxs(core.AdditionalScrForEachScCallOrSpecialTx)
		}
		cpmbi.processingInfo.numCrossShardScCallsOrSpecialTxs++
	}

	if txMbInfo.isReceiverSmartContractAddress {
		cpmbi.mapSCTxs[string(txHash)] = struct{}{}
		cpmbi.mapScsForShard[receiverShardID]++
	} else {
		cpmbi.mapTxsForShard[receiverShardID]++
	}

	cpmbi.processingInfo.numTxsAdded++
}

func (txs *transactions) displayProcessingResults(
	miniBlocks block.MiniBlockSlice,
	nbSortedTxs int,
	cpmbi *createAndProcessMiniBlocksInfo,
) {
	if len(miniBlocks) > 0 {
		log.Debug("mini blocks from me created", "num", len(miniBlocks))
	}

	log.Debug("displayProcessingResults",
		"self shard", txs.shardCoordinator.SelfId(),
		"gas consumed in self shard for txs from me", cpmbi.gasInfo.gasConsumedByMiniBlocksInSenderShard,
		"total gas consumed in self shard", cpmbi.gasInfo.totalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard for non sc txs", cpmbi.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID][nonScTx],
			"gas consumed in receiver shard for sc txs", cpmbi.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID][scTx],
			"txs added", len(miniBlock.TxHashes),
			"non sc txs", cpmbi.mapTxsForShard[miniBlock.ReceiverShardID],
			"sc txs", cpmbi.mapScsForShard[miniBlock.ReceiverShardID])
	}

	log.Debug("transactions info", "total txs", nbSortedTxs,
		"num txs added", cpmbi.processingInfo.numTxsAdded,
		"num bad txs", cpmbi.processingInfo.numBadTxs,
		"num txs failed", cpmbi.processingInfo.numTxsFailed,
		"num txs skipped", cpmbi.processingInfo.numTxsSkipped,
		"num txs with initial balance consumed", cpmbi.processingInfo.numTxsWithInitialBalanceConsumed,
		"num cross shard sc calls or special txs", cpmbi.processingInfo.numCrossShardScCallsOrSpecialTxs,
		"used time for computeGasConsumed", cpmbi.processingInfo.totalTimeUsedForComputeGasConsumed,
		"used time for processAndRemoveBadTransaction", cpmbi.processingInfo.totalTimeUsedForProcess,
	)
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

	csmbi := txs.initCreateScheduledMiniBlocks()

	for index := range sortedTxs {
		if !haveTime() && !haveAdditionalTime() {
			log.Debug("time is out in createScheduledMiniBlocks")
			break
		}

		tx, miniBlock, shouldContinue := txs.shouldContinueProcessingScheduledTx(
			isShardStuck,
			sortedTxs[index],
			mapSCTxs,
			csmbi)
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
			csmbi)

		if isMaxBlockSizeReached(scheduledTxMbInfo.numNewMiniBlocks, scheduledTxMbInfo.numNewTxs) {
			log.Debug("max txs accepted in one block is reached",
				"num scheduled txs added", csmbi.schedulingInfo.numScheduledTxsAdded,
				"total txs", len(sortedTxs))
			break
		}

		err := txs.verifyTransaction(
			tx,
			scTx,
			txHash,
			senderShardID,
			receiverShardID,
			csmbi)
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
			csmbi)
	}

	miniBlocks := txs.getMiniBlockSliceFromMapV2(csmbi.mapMiniBlocks, mapSCTxs)

	txs.displayProcessingResultsOfScheduledMiniBlocks(miniBlocks, len(sortedTxs), csmbi)

	log.Debug("createScheduledMiniBlocks has been finished")

	return miniBlocks
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
	csmbi *createScheduledMiniBlocksInfo,
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

	if len(csmbi.senderAddressToSkip) > 0 {
		if bytes.Equal(csmbi.senderAddressToSkip, tx.GetSndAddr()) {
			csmbi.schedulingInfo.numScheduledTxsSkipped++
			return nil, nil, false
		}
	}

	csmbi.senderAddressToSkip = tx.GetSndAddr()

	_, txTypeDstShard := txs.txTypeHandler.ComputeTransactionType(tx)
	isReceiverSmartContractAddress := txTypeDstShard == process.SCDeployment || txTypeDstShard == process.SCInvoking
	if !isReceiverSmartContractAddress {
		return nil, nil, false
	}

	miniBlock, ok := csmbi.mapMiniBlocks[receiverShardID]
	if !ok {
		log.Debug("scheduled mini block is not created", "shard", receiverShardID)
		return nil, nil, false
	}

	addressHasEnoughBalance := txs.hasAddressEnoughInitialBalance(tx)
	if !addressHasEnoughBalance {
		csmbi.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed++
		return nil, nil, false
	}

	return tx, miniBlock, true
}

func (txs *transactions) getScheduledTxAndMbInfo(
	isMiniBlockEmpty bool,
	receiverShardID uint32,
	csmbi *createScheduledMiniBlocksInfo,
) *scheduledTxAndMbInfo {
	numNewMiniBlocks := 0
	if isMiniBlockEmpty {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	isCrossShardScCallTx := receiverShardID != txs.shardCoordinator.SelfId()
	if isCrossShardScCallTx {
		if !csmbi.firstCrossShardScCallTxFound {
			numNewMiniBlocks++
		}
		if csmbi.mapCrossShardScCallTxs[receiverShardID] >= csmbi.maxCrossShardScCallTxsPerShard {
			numNewTxs += core.AdditionalScrForEachScCallOrSpecialTx
		}
	}

	return &scheduledTxAndMbInfo{
		numNewTxs:            numNewTxs,
		numNewMiniBlocks:     numNewMiniBlocks,
		isCrossShardScCallTx: isCrossShardScCallTx,
	}
}

func (txs *transactions) verifyTransaction(
	tx *transaction.Transaction,
	txType TxType,
	txHash []byte,
	senderShardID uint32,
	receiverShardID uint32,
	csmbi *createScheduledMiniBlocksInfo,
) error {
	csmbi.gasInfo.gasConsumedByMiniBlockInReceiverShard = csmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType]
	oldGasConsumedByMiniBlocksInSenderShard := csmbi.gasInfo.gasConsumedByMiniBlocksInSenderShard
	oldGasConsumedByMiniBlockInReceiverShard := csmbi.gasInfo.gasConsumedByMiniBlockInReceiverShard
	oldTotalGasConsumedInSelfShard := csmbi.gasInfo.totalGasConsumedInSelfShard

	startTime := time.Now()
	err := txs.computeGasConsumed(
		senderShardID,
		receiverShardID,
		tx,
		txHash,
		&csmbi.gasInfo)
	elapsedTime := time.Since(startTime)
	csmbi.schedulingInfo.totalTimeUsedForScheduledComputeGasConsumed += elapsedTime
	if err != nil {
		log.Trace("verifyTransaction.computeGasConsumed", "error", err)
		return err
	}

	csmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = csmbi.gasInfo.gasConsumedByMiniBlockInReceiverShard

	startTime = time.Now()
	err = txs.txProcessor.VerifyTransaction(tx)
	elapsedTime = time.Since(startTime)
	csmbi.schedulingInfo.totalTimeUsedForScheduledVerify += elapsedTime

	txs.mutAccountsInfo.Lock()
	txs.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.mutAccountsInfo.Unlock()

	if err != nil {
		isTxTargetedForDeletion := errors.Is(err, process.ErrLowerNonceInTransaction) || errors.Is(err, process.ErrInsufficientFee)
		if isTxTargetedForDeletion {
			strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
			txs.txPool.RemoveData(txHash, strCache)
		}

		csmbi.schedulingInfo.numScheduledBadTxs++
		log.Trace("bad tx", "error", err.Error(), "hash", txHash)

		txs.gasHandler.RemoveGasConsumed([][]byte{txHash})

		csmbi.gasInfo.gasConsumedByMiniBlocksInSenderShard = oldGasConsumedByMiniBlocksInSenderShard
		csmbi.mapGasConsumedByMiniBlockInReceiverShard[receiverShardID][txType] = oldGasConsumedByMiniBlockInReceiverShard
		csmbi.gasInfo.totalGasConsumedInSelfShard = oldTotalGasConsumedInSelfShard

		return err
	}

	txShardInfoToSet := &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	txs.txsForCurrBlock.mutTxsForBlock.Lock()
	txs.txsForCurrBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoToSet}
	txs.txsForCurrBlock.mutTxsForBlock.Unlock()

	return nil
}

func (txs *transactions) applyVerifiedTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	miniBlock *block.MiniBlock,
	receiverShardID uint32,
	scheduledTxMbInfo *scheduledTxAndMbInfo,
	mapSCTxs map[string]struct{},
	csmbi *createScheduledMiniBlocksInfo,
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
		if !csmbi.firstCrossShardScCallTxFound {
			csmbi.firstCrossShardScCallTxFound = true
			txs.blockSizeComputation.AddNumMiniBlocks(1)
		}
		csmbi.mapCrossShardScCallTxs[receiverShardID]++
		if csmbi.mapCrossShardScCallTxs[receiverShardID] > csmbi.maxCrossShardScCallTxsPerShard {
			csmbi.maxCrossShardScCallTxsPerShard = csmbi.mapCrossShardScCallTxs[receiverShardID]
			//we need to increment this as to account for the corresponding SCR hash
			txs.blockSizeComputation.AddNumTxs(core.AdditionalScrForEachScCallOrSpecialTx)
		}
		csmbi.schedulingInfo.numScheduledCrossShardScCalls++
	}

	mapSCTxs[string(txHash)] = struct{}{}
	csmbi.schedulingInfo.numScheduledTxsAdded++
	txs.scheduledTxsExecutionHandler.Add(txHash, tx)
}

func (txs *transactions) displayProcessingResultsOfScheduledMiniBlocks(
	miniBlocks block.MiniBlockSlice,
	nbSortedTxs int,
	csmbi *createScheduledMiniBlocksInfo,
) {
	if len(miniBlocks) > 0 {
		log.Debug("scheduled mini blocks from me created", "num", len(miniBlocks))
	}

	log.Debug("displayProcessingResultsOfScheduledMiniBlocks",
		"self shard", txs.shardCoordinator.SelfId(),
		"total gas consumed in self shard", csmbi.gasInfo.totalGasConsumedInSelfShard)

	for _, miniBlock := range miniBlocks {
		log.Debug("scheduled mini block info",
			"type", "ScheduledBlock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"gas consumed in receiver shard", csmbi.mapGasConsumedByMiniBlockInReceiverShard[miniBlock.ReceiverShardID][scTx],
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("scheduled transactions info", "total txs", nbSortedTxs,
		"num scheduled txs added", csmbi.schedulingInfo.numScheduledTxsAdded,
		"num scheduled bad txs", csmbi.schedulingInfo.numScheduledBadTxs,
		"num scheduled txs skipped", csmbi.schedulingInfo.numScheduledTxsSkipped,
		"num scheduled txs with initial balance consumed", csmbi.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed,
		"num scheduled cross shard sc calls", csmbi.schedulingInfo.numScheduledCrossShardScCalls,
		"used time for scheduled computeGasConsumed", csmbi.schedulingInfo.totalTimeUsedForScheduledComputeGasConsumed,
		"used time for scheduled VerifyTransaction", csmbi.schedulingInfo.totalTimeUsedForScheduledVerify,
	)
}
