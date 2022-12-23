package preprocess

import (
	"bytes"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const maxGasLimitPercentUsedForDestMeTxs = 50

type gasConsumedInfo struct {
	prevGasConsumedInReceiverShard        uint64
	gasConsumedByMiniBlocksInSenderShard  uint64
	gasConsumedByMiniBlockInReceiverShard uint64
	totalGasConsumedInSelfShard           uint64
}

type txAndMbInfo struct {
	numNewTxs                      int
	numNewMiniBlocks               int
	isReceiverSmartContractAddress bool
	isCrossShardScCallOrSpecialTx  bool
}

type scheduledTxAndMbInfo struct {
	numNewTxs            int
	numNewMiniBlocks     int
	isCrossShardScCallTx bool
}

type processedTxsInfo struct {
	numTxsAdded                        int
	numBadTxs                          int
	numTxsSkipped                      int
	numTxsFailed                       int
	numTxsWithInitialBalanceConsumed   int
	numCrossShardScCallsOrSpecialTxs   int
	numCrossShardTxsWithTooMuchGas     int
	totalTimeUsedForProcess            time.Duration
	totalTimeUsedForComputeGasProvided time.Duration
}

type createAndProcessMiniBlocksInfo struct {
	mapSCTxs                                 map[string]struct{}
	mapTxsForShard                           map[uint32]int
	mapScsForShard                           map[uint32]int
	mapCrossShardScCallsOrSpecialTxs         map[uint32]int
	mapGasConsumedByMiniBlockInReceiverShard map[uint32]uint64
	mapMiniBlocks                            map[uint32]*block.MiniBlock
	senderAddressToSkip                      []byte
	maxCrossShardScCallsOrSpecialTxsPerShard int
	firstInvalidTxFound                      bool
	firstCrossShardScCallOrSpecialTxFound    bool
	processingInfo                           processedTxsInfo
	gasInfo                                  gasConsumedInfo
}

type scheduledTxsInfo struct {
	numScheduledTxsAdded                        int
	numScheduledBadTxs                          int
	numScheduledTxsSkipped                      int
	numScheduledTxsWithInitialBalanceConsumed   int
	numScheduledCrossShardScCalls               int
	numCrossShardTxsWithTooMuchGas              int
	totalTimeUsedForScheduledVerify             time.Duration
	totalTimeUsedForScheduledComputeGasProvided time.Duration
}

type createScheduledMiniBlocksInfo struct {
	mapMiniBlocks                            map[uint32]*block.MiniBlock
	mapCrossShardScCallTxs                   map[uint32]int
	maxCrossShardScCallTxsPerShard           int
	schedulingInfo                           scheduledTxsInfo
	firstCrossShardScCallTxFound             bool
	mapGasConsumedByMiniBlockInReceiverShard map[uint32]uint64
	gasInfo                                  gasConsumedInfo
	senderAddressToSkip                      []byte
}

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx data.TransactionHandler
	*txShardInfo
}

type txsForBlock struct {
	missingTxs     int
	mutTxsForBlock sync.RWMutex
	txHashAndInfo  map[string]*txInfo
}

type processedIndexes struct {
	indexOfLastTxProcessed           int32
	indexOfLastTxProcessedByProposer int32
}

// basePreProcess is the base struct for all pre-processors
// beware of calling basePreProcess.epochConfirmed in all extensions of this struct if the flags from the basePreProcess are
// used in those extensions instances
type basePreProcess struct {
	gasTracker
	hasher                                      hashing.Hasher
	marshalizer                                 marshal.Marshalizer
	blockSizeComputation                        BlockSizeComputationHandler
	balanceComputation                          BalanceComputationHandler
	accounts                                    state.AccountsAdapter
	pubkeyConverter                             core.PubkeyConverter
	optimizeGasUsedInCrossMiniBlocksEnableEpoch uint32
	flagOptimizeGasUsedInCrossMiniBlocks        atomic.Flag
	frontRunningProtectionEnableEpoch           uint32
	flagFrontRunningProtection                  atomic.Flag
	processedMiniBlocksTracker                  process.ProcessedMiniBlocksTracker
}

func (bpp *basePreProcess) removeBlockDataFromPools(
	body *block.Body,
	miniBlockPool storage.Cacher,
	txPool dataRetriever.ShardedDataCacherNotifier,
	isMiniBlockCorrect func(block.Type) bool,
) error {
	err := bpp.removeTxsFromPools(body, txPool, isMiniBlockCorrect)
	if err != nil {
		return err
	}

	err = bpp.removeMiniBlocksFromPools(body, miniBlockPool, isMiniBlockCorrect)
	if err != nil {
		return err
	}

	return nil
}

func (bpp *basePreProcess) removeTxsFromPools(
	body *block.Body,
	txPool dataRetriever.ShardedDataCacherNotifier,
	isMiniBlockCorrect func(block.Type) bool,
) error {
	if check.IfNil(body) {
		return process.ErrNilTxBlockBody
	}
	if check.IfNil(txPool) {
		return process.ErrNilTransactionPool
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		currentMiniBlock := body.MiniBlocks[i]
		if !isMiniBlockCorrect(currentMiniBlock.Type) {
			log.Trace("removeTxsFromPools.isMiniBlockCorrect: false",
				"miniblock type", currentMiniBlock.Type)
			continue
		}

		strCache := process.ShardCacherIdentifier(currentMiniBlock.SenderShardID, currentMiniBlock.ReceiverShardID)
		txPool.RemoveSetOfDataFromPool(currentMiniBlock.TxHashes, strCache)
	}

	return nil
}

func (bpp *basePreProcess) removeMiniBlocksFromPools(
	body *block.Body,
	miniBlockPool storage.Cacher,
	isMiniBlockCorrect func(block.Type) bool,
) error {
	if check.IfNil(body) {
		return process.ErrNilTxBlockBody
	}
	if check.IfNil(miniBlockPool) {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		currentMiniBlock := body.MiniBlocks[i]
		if !isMiniBlockCorrect(currentMiniBlock.Type) {
			log.Trace("removeMiniBlocksFromPools.isMiniBlockCorrect: false",
				"miniblock type", currentMiniBlock.Type)
			continue
		}

		miniBlockHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, currentMiniBlock)
		if err != nil {
			return err
		}

		miniBlockPool.Remove(miniBlockHash)
	}

	return nil
}

func (bpp *basePreProcess) createMarshalizedData(txHashes [][]byte, forBlock *txsForBlock) ([][]byte, error) {
	mrsTxs := make([][]byte, 0, len(txHashes))
	for _, txHash := range txHashes {
		forBlock.mutTxsForBlock.RLock()
		txInfoFromMap := forBlock.txHashAndInfo[string(txHash)]
		forBlock.mutTxsForBlock.RUnlock()

		if txInfoFromMap == nil || check.IfNil(txInfoFromMap.tx) {
			log.Warn("basePreProcess.createMarshalizedData: tx not found", "hash", txHash)
			continue
		}

		txMrs, err := bpp.marshalizer.Marshal(txInfoFromMap.tx)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsTxs = append(mrsTxs, txMrs)
	}

	log.Trace("basePreProcess.createMarshalizedData",
		"num txs", len(mrsTxs),
	)

	return mrsTxs, nil
}

func (bpp *basePreProcess) saveTxsToStorage(
	txHashes [][]byte,
	forBlock *txsForBlock,
	store dataRetriever.StorageService,
	dataUnit dataRetriever.UnitType,
) {
	for i := 0; i < len(txHashes); i++ {
		txHash := txHashes[i]
		bpp.saveTransactionToStorage(txHash, forBlock, store, dataUnit)
	}
}

func (bpp *basePreProcess) saveTransactionToStorage(
	txHash []byte,
	forBlock *txsForBlock,
	store dataRetriever.StorageService,
	dataUnit dataRetriever.UnitType,
) {
	forBlock.mutTxsForBlock.RLock()
	txInfoFromMap := forBlock.txHashAndInfo[string(txHash)]
	forBlock.mutTxsForBlock.RUnlock()

	if txInfoFromMap == nil || txInfoFromMap.tx == nil {
		log.Warn("basePreProcess.saveTransactionToStorage", "type", dataUnit, "txHash", txHash, "error", process.ErrMissingTransaction.Error())
		return
	}

	buff, err := bpp.marshalizer.Marshal(txInfoFromMap.tx)
	if err != nil {
		log.Warn("basePreProcess.saveTransactionToStorage", "txHash", txHash, "error", err.Error())
		return
	}

	errNotCritical := store.Put(dataUnit, txHash, buff)
	if errNotCritical != nil {
		log.Debug("store.Put",
			"error", errNotCritical.Error(),
			"dataUnit", dataUnit,
		)
	}
}

func (bpp *basePreProcess) baseReceivedTransaction(
	txHash []byte,
	tx data.TransactionHandler,
	forBlock *txsForBlock,
) bool {

	forBlock.mutTxsForBlock.Lock()
	defer forBlock.mutTxsForBlock.Unlock()

	if forBlock.missingTxs > 0 {
		txInfoForHash := forBlock.txHashAndInfo[string(txHash)]
		if txInfoForHash != nil && txInfoForHash.txShardInfo != nil &&
			(txInfoForHash.tx == nil || txInfoForHash.tx.IsInterfaceNil()) {
			forBlock.txHashAndInfo[string(txHash)].tx = tx
			forBlock.missingTxs--
		}

		return forBlock.missingTxs == 0
	}

	return false
}

func (bpp *basePreProcess) computeExistingAndRequestMissing(
	body *block.Body,
	forBlock *txsForBlock,
	_ chan bool,
	isMiniBlockCorrect func(block.Type) bool,
	txPool dataRetriever.ShardedDataCacherNotifier,
	onRequestTxs func(shardID uint32, txHashes [][]byte),
) int {

	if check.IfNil(body) {
		return 0
	}

	forBlock.mutTxsForBlock.Lock()
	defer forBlock.mutTxsForBlock.Unlock()

	missingTxsForShard := make(map[uint32][][]byte, bpp.shardCoordinator.NumberOfShards())
	txHashes := make([][]byte, 0)
	uniqueTxHashes := make(map[string]struct{})
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if !isMiniBlockCorrect(miniBlock.Type) {
			continue
		}

		txShardInfoObject := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}
		// TODO refactor this section
		method := process.SearchMethodJustPeek
		if miniBlock.Type == block.InvalidBlock {
			method = process.SearchMethodSearchFirst
		}
		if miniBlock.Type == block.SmartContractResultBlock {
			method = process.SearchMethodPeekWithFallbackSearchFirst
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]

			_, isAlreadyEvaluated := uniqueTxHashes[string(txHash)]
			if isAlreadyEvaluated {
				continue
			}
			uniqueTxHashes[string(txHash)] = struct{}{}

			tx, err := process.GetTransactionHandlerFromPool(
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
				txHash,
				txPool,
				method)

			if err != nil {
				txHashes = append(txHashes, txHash)
				forBlock.missingTxs++
				log.Trace("missing tx",
					"miniblock type", miniBlock.Type,
					"sender", miniBlock.SenderShardID,
					"receiver", miniBlock.ReceiverShardID,
					"hash", txHash,
				)
				continue
			}

			forBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoObject}
		}

		if len(txHashes) > 0 {
			bpp.setMissingTxsForShard(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHashes, forBlock)
			missingTxsForShard[miniBlock.SenderShardID] = append(missingTxsForShard[miniBlock.SenderShardID], txHashes...)
		}

		txHashes = make([][]byte, 0)
	}

	return bpp.requestMissingTxsForShard(missingTxsForShard, onRequestTxs)
}

// this method should be called only under the mutex protection: forBlock.mutTxsForBlock
func (bpp *basePreProcess) setMissingTxsForShard(
	senderShardID uint32,
	receiverShardID uint32,
	txHashes [][]byte,
	forBlock *txsForBlock,
) {
	txShardInfoToSet := &txShardInfo{
		senderShardID:   senderShardID,
		receiverShardID: receiverShardID,
	}

	for _, txHash := range txHashes {
		forBlock.txHashAndInfo[string(txHash)] = &txInfo{
			tx:          nil,
			txShardInfo: txShardInfoToSet,
		}
	}
}

// this method should be called only under the mutex protection: forBlock.mutTxsForBlock
func (bpp *basePreProcess) requestMissingTxsForShard(
	missingTxsForShard map[uint32][][]byte,
	onRequestTxs func(shardID uint32, txHashes [][]byte),
) int {
	requestedTxs := 0
	for shardID, txHashes := range missingTxsForShard {
		requestedTxs += len(txHashes)
		go func(providedsShardID uint32, providedTxHashes [][]byte) {
			onRequestTxs(providedsShardID, providedTxHashes)
		}(shardID, txHashes)
	}

	return requestedTxs
}

func (bpp *basePreProcess) saveAccountBalanceForAddress(address []byte) {
	if bpp.balanceComputation.IsAddressSet(address) {
		return
	}

	balance, err := bpp.getBalanceForAddress(address)
	if err != nil {
		balance = big.NewInt(0)
	}

	bpp.balanceComputation.SetBalanceToAddress(address, balance)
}

func (bpp *basePreProcess) getBalanceForAddress(address []byte) (*big.Int, error) {
	accountHandler, err := bpp.accounts.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	account, ok := accountHandler.(state.UserAccountHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return account.GetBalance(), nil
}

func getTxMaxTotalCost(txHandler data.TransactionHandler) *big.Int {
	cost := big.NewInt(0)
	cost.Mul(big.NewInt(0).SetUint64(txHandler.GetGasPrice()), big.NewInt(0).SetUint64(txHandler.GetGasLimit()))

	if txHandler.GetValue() != nil {
		cost.Add(cost, txHandler.GetValue())
	}

	return cost
}

func (bpp *basePreProcess) getTotalGasConsumed() uint64 {
	if !bpp.flagOptimizeGasUsedInCrossMiniBlocks.IsSet() {
		return bpp.gasHandler.TotalGasProvided()
	}

	totalGasToBeSubtracted := bpp.gasHandler.TotalGasRefunded() + bpp.gasHandler.TotalGasPenalized()
	totalGasProvided := bpp.gasHandler.TotalGasProvided()
	if totalGasToBeSubtracted > totalGasProvided {
		log.Warn("basePreProcess.getTotalGasConsumed: too much gas to be subtracted",
			"totalGasRefunded", bpp.gasHandler.TotalGasRefunded(),
			"totalGasPenalized", bpp.gasHandler.TotalGasPenalized(),
			"totalGasToBeSubtracted", totalGasToBeSubtracted,
			"totalGasProvided", totalGasProvided,
		)
		return totalGasProvided
	}

	return totalGasProvided - totalGasToBeSubtracted
}

func (bpp *basePreProcess) updateGasConsumedWithGasRefundedAndGasPenalized(
	txHash []byte,
	gasInfo *gasConsumedInfo,
) {
	if !bpp.flagOptimizeGasUsedInCrossMiniBlocks.IsSet() {
		return
	}

	gasRefunded := bpp.gasHandler.GasRefunded(txHash)
	gasPenalized := bpp.gasHandler.GasPenalized(txHash)
	gasToBeSubtracted := gasRefunded + gasPenalized
	couldUpdateGasConsumedWithGasSubtracted := gasToBeSubtracted <= gasInfo.gasConsumedByMiniBlockInReceiverShard &&
		gasToBeSubtracted <= gasInfo.totalGasConsumedInSelfShard
	if !couldUpdateGasConsumedWithGasSubtracted {
		log.Warn("basePreProcess.updateGasConsumedWithGasRefundedAndGasPenalized: too much gas to be subtracted",
			"gasRefunded", gasRefunded,
			"gasPenalized", gasPenalized,
			"gasToBeSubtracted", gasToBeSubtracted,
			"gasConsumedByMiniBlockInReceiverShard", gasInfo.gasConsumedByMiniBlockInReceiverShard,
			"totalGasConsumedInSelfShard", gasInfo.totalGasConsumedInSelfShard,
		)
		return
	}

	gasInfo.gasConsumedByMiniBlockInReceiverShard -= gasToBeSubtracted
	gasInfo.totalGasConsumedInSelfShard -= gasToBeSubtracted
}

func (bpp *basePreProcess) handleProcessTransactionInit(preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler, txHash []byte) int {
	snapshot := bpp.accounts.JournalLen()
	preProcessorExecutionInfoHandler.InitProcessedTxsResults(txHash)
	return snapshot
}

func (bpp *basePreProcess) handleProcessTransactionError(preProcessorExecutionInfoHandler process.PreProcessorExecutionInfoHandler, snapshot int, txHash []byte) {
	bpp.gasHandler.RestoreGasSinceLastReset(txHash)

	errRevert := bpp.accounts.RevertToSnapshot(snapshot)
	if errRevert != nil {
		log.Debug("basePreProcess.handleProcessError: RevertToSnapshot", "error", errRevert.Error())
	}

	preProcessorExecutionInfoHandler.RevertProcessedTxsResults([][]byte{txHash}, txHash)
}

func getMiniBlockHeaderOfMiniBlock(headerHandler data.HeaderHandler, miniBlockHash []byte) (data.MiniBlockHeaderHandler, error) {
	for _, miniBlockHeader := range headerHandler.GetMiniBlockHeaderHandlers() {
		if bytes.Equal(miniBlockHeader.GetHash(), miniBlockHash) {
			return miniBlockHeader, nil
		}
	}

	return nil, process.ErrMissingMiniBlockHeader
}

// epochConfirmed is called whenever a new epoch is confirmed from the structs that extend this instance
func (bpp *basePreProcess) epochConfirmed(epoch uint32, _ uint64) {
	bpp.flagOptimizeGasUsedInCrossMiniBlocks.SetValue(epoch >= bpp.optimizeGasUsedInCrossMiniBlocksEnableEpoch)
	log.Debug("basePreProcess: optimize gas used in cross mini blocks", "enabled", bpp.flagOptimizeGasUsedInCrossMiniBlocks.IsSet())
	bpp.flagFrontRunningProtection.SetValue(epoch >= bpp.frontRunningProtectionEnableEpoch)
	log.Debug("basePreProcess: front running protection", "enabled", bpp.flagFrontRunningProtection.IsSet())
}

func (bpp *basePreProcess) getIndexesOfLastTxProcessed(
	miniBlock *block.MiniBlock,
	headerHandler data.HeaderHandler,
) (*processedIndexes, error) {

	miniBlockHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, miniBlock)
	if err != nil {
		return nil, err
	}

	pi := &processedIndexes{}

	processedMiniBlockInfo, _ := bpp.processedMiniBlocksTracker.GetProcessedMiniBlockInfo(miniBlockHash)
	pi.indexOfLastTxProcessed = processedMiniBlockInfo.IndexOfLastTxProcessed

	miniBlockHeader, err := getMiniBlockHeaderOfMiniBlock(headerHandler, miniBlockHash)
	if err != nil {
		return nil, err
	}

	pi.indexOfLastTxProcessedByProposer = miniBlockHeader.GetIndexOfLastTxProcessed()

	return pi, nil
}
