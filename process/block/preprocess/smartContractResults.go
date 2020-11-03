package preprocess

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.DataMarshalizer = (*smartContractResults)(nil)
var _ process.PreProcessor = (*smartContractResults)(nil)

type smartContractResults struct {
	*basePreProcess
	chRcvAllScrs                 chan bool
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	scrForBlock                  txsForBlock
	scrPool                      dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	scrProcessor                 process.SmartContractResultProcessor
}

// NewSmartContractResultPreprocessor creates a new smartContractResult preprocessor object
func NewSmartContractResultPreprocessor(
	scrDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	scrProcessor process.SmartContractResultProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte),
	gasHandler process.GasHandler,
	economicsFee process.FeeHandler,
	pubkeyConverter core.PubkeyConverter,
	blockSizeComputation BlockSizeComputationHandler,
	balanceComputation BalanceComputationHandler,
) (*smartContractResults, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(scrDataPool) {
		return nil, process.ErrNilUTxDataPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilUTxStorage
	}
	if check.IfNil(scrProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestSmartContractResult == nil {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
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

	bpp := &basePreProcess{
		hasher:               hasher,
		marshalizer:          marshalizer,
		shardCoordinator:     shardCoordinator,
		gasHandler:           gasHandler,
		economicsFee:         economicsFee,
		blockSizeComputation: blockSizeComputation,
		balanceComputation:   balanceComputation,
		accounts:             accounts,
		pubkeyConverter:      pubkeyConverter,
	}

	scr := &smartContractResults{
		basePreProcess:               bpp,
		storage:                      store,
		scrPool:                      scrDataPool,
		onRequestSmartContractResult: onRequestSmartContractResult,
		scrProcessor:                 scrProcessor,
	}

	scr.chRcvAllScrs = make(chan bool)
	scr.scrPool.RegisterOnAdded(scr.receivedSmartContractResult)
	scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)

	return scr, nil
}

// waitForScrHashes waits for a call whether all the requested smartContractResults appeared
func (scr *smartContractResults) waitForScrHashes(waitTime time.Duration) error {
	select {
	case <-scr.chRcvAllScrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested smartContractResults arrived and were saved into the pool
func (scr *smartContractResults) IsDataPrepared(requestedScrs int, haveTime func() time.Duration) error {
	if requestedScrs > 0 {
		log.Debug("requested missing scrs",
			"num scrs", requestedScrs)
		err := scr.waitForScrHashes(haveTime())
		scr.scrForBlock.mutTxsForBlock.Lock()
		missingScrs := scr.scrForBlock.missingTxs
		scr.scrForBlock.missingTxs = 0
		scr.scrForBlock.mutTxsForBlock.Unlock()
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

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		scrBuff, err := scr.storage.GetAll(dataRetriever.UnsignedTransactionUnit, miniBlock.TxHashes)
		if err != nil {
			log.Debug("unsigned tx from mini block was not found in UnsignedTransactionUnit",
				"sender shard ID", miniBlock.SenderShardID,
				"receiver shard ID", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes),
			)

			return scrRestored, err
		}

		for txHash, txBuff := range scrBuff {
			tx := smartContractResult.SmartContractResult{}
			err = scr.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return scrRestored, err
			}

			scr.scrPool.AddData([]byte(txHash), &tx, tx.Size(), strCache)
		}

		//TODO: Should be analyzed if restoring into pool only cross-shard miniblocks with destination in self shard,
		//would create problems or not
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

// ProcessBlockTransactions processes all the smartContractResult from the block.Body, updates the state
func (scr *smartContractResults) ProcessBlockTransactions(
	body *block.Body,
	haveTime func() bool,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

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

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			scr.scrForBlock.mutTxsForBlock.RLock()
			txInfoFromMap, ok := scr.scrForBlock.txHashAndInfo[string(txHash)]
			scr.scrForBlock.mutTxsForBlock.RUnlock()
			if !ok || check.IfNil(txInfoFromMap.tx) {
				log.Warn("missing transaction in ProcessBlockTransactions ", "type", miniBlock.Type, "txHash", txHash)
				return process.ErrMissingTransaction
			}

			currScr, ok := txInfoFromMap.tx.(*smartContractResult.SmartContractResult)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			scr.saveAccountBalanceForAddress(currScr.GetRcvAddr())

			_, err := scr.scrProcessor.ProcessSmartContractResult(currScr)
			if err != nil {
				return err
			}
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

		err := scr.saveTxsToStorage(miniBlock.TxHashes, &scr.scrForBlock, scr.storage, dataRetriever.UnsignedTransactionUnit)
		if err != nil {
			return err
		}
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

	receivedAllMissing := scr.baseReceivedTransaction(key, tx, &scr.scrForBlock)

	if receivedAllMissing {
		scr.chRcvAllScrs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created smartContractResults at this round
func (scr *smartContractResults) CreateBlockStarted() {
	_ = core.EmptyChannel(scr.chRcvAllScrs)

	scr.scrForBlock.mutTxsForBlock.Lock()
	scr.scrForBlock.missingTxs = 0
	scr.scrForBlock.txHashAndInfo = make(map[string]*txInfo)
	scr.scrForBlock.mutTxsForBlock.Unlock()
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
		&scr.scrForBlock,
		scr.chRcvAllScrs,
		scr.isMiniBlockCorrect,
		scr.scrPool,
		scr.onRequestSmartContractResult,
	)

	return numMissingTxsForShard
}

// RequestTransactionsForMiniBlock requests missing smartContractResults for a certain miniblock
func (scr *smartContractResults) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if miniBlock == nil {
		return 0
	}

	missingScrsForMiniBlock := scr.computeMissingScrsForMiniBlock(miniBlock)
	if len(missingScrsForMiniBlock) > 0 {
		scr.onRequestSmartContractResult(miniBlock.SenderShardID, missingScrsForMiniBlock)
	}

	return len(missingScrsForMiniBlock)
}

// computeMissingScrsForMiniBlock computes missing smartContractResults for a certain miniblock
func (scr *smartContractResults) computeMissingScrsForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
	if miniBlock.Type != block.SmartContractResultBlock {
		return [][]byte{}
	}

	missingSmartContractResults := make([][]byte, 0, len(miniBlock.TxHashes))
	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			scr.scrPool,
			false)

		if check.IfNil(tx) {
			missingSmartContractResults = append(missingSmartContractResults, txHash)
		}
	}

	return sliceUtil.TrimSliceSliceByte(missingSmartContractResults)
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
			return nil, nil, process.ErrNilSmartContractResult
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

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (scr *smartContractResults) CreateAndProcessMiniBlocks(_ func() bool) (block.MiniBlockSlice, error) {
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock processes all the smartContractResults from a and saves the processed smartContractResults in local cache complete miniblock
func (scr *smartContractResults) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, _ func() (int, int)) ([][]byte, error) {

	if miniBlock.Type != block.SmartContractResultBlock {
		return nil, process.ErrWrongTypeInMiniBlock
	}

	miniBlockScrs, miniBlockTxHashes, err := scr.getAllScrsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return nil, err
	}

	if scr.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlockScrs)) {
		return nil, process.ErrMaxBlockSizeReached
	}

	processedTxHashes := make([][]byte, 0)

	defer func() {
		if err != nil {
			scr.gasHandler.RemoveGasConsumed(processedTxHashes)
			scr.gasHandler.RemoveGasRefunded(processedTxHashes)
		}
	}()

	gasConsumedByMiniBlockInSenderShard := uint64(0)
	gasConsumedByMiniBlockInReceiverShard := uint64(0)
	totalGasConsumedInSelfShard := scr.gasHandler.TotalGasConsumed()

	log.Trace("smartContractResults.ProcessMiniBlock", "totalGasConsumedInSelfShard", totalGasConsumedInSelfShard)

	for index := range miniBlockScrs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return processedTxHashes, err
		}

		err = scr.computeGasConsumed(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			miniBlockScrs[index],
			miniBlockTxHashes[index],
			&gasConsumedByMiniBlockInSenderShard,
			&gasConsumedByMiniBlockInReceiverShard,
			&totalGasConsumedInSelfShard)

		if err != nil {
			return processedTxHashes, err
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])
	}

	for index := range miniBlockScrs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return processedTxHashes, err
		}

		scr.saveAccountBalanceForAddress(miniBlockScrs[index].GetRcvAddr())

		_, err = scr.scrProcessor.ProcessSmartContractResult(miniBlockScrs[index])
		if err != nil {
			return processedTxHashes, err
		}
	}

	txShardInfoToSet := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	scr.scrForBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		scr.scrForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockScrs[index], txShardInfo: txShardInfoToSet}
	}
	scr.scrForBlock.mutTxsForBlock.Unlock()

	scr.blockSizeComputation.AddNumMiniBlocks(1)
	scr.blockSizeComputation.AddNumTxs(len(miniBlockScrs))

	return nil, nil
}

// CreateMarshalizedData marshalizes smartContractResults and creates and saves them into a new structure
func (scr *smartContractResults) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs, err := scr.createMarshalizedData(txHashes, &scr.scrForBlock)
	if err != nil {
		return nil, err
	}

	return mrsScrs, nil
}

// GetAllCurrentUsedTxs returns all the smartContractResults used at current creation / processing
func (scr *smartContractResults) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	scr.scrForBlock.mutTxsForBlock.RLock()
	scrPool := make(map[string]data.TransactionHandler, len(scr.scrForBlock.txHashAndInfo))
	for txHash, txInfoFromMap := range scr.scrForBlock.txHashAndInfo {
		scrPool[txHash] = txInfoFromMap.tx
	}
	scr.scrForBlock.mutTxsForBlock.RUnlock()

	return scrPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (scr *smartContractResults) IsInterfaceNil() bool {
	return scr == nil
}

func (scr *smartContractResults) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.SmartContractResultBlock
}
