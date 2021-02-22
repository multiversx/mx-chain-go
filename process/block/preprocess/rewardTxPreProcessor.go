package preprocess

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.DataMarshalizer = (*rewardTxPreprocessor)(nil)
var _ process.PreProcessor = (*rewardTxPreprocessor)(nil)

type rewardTxPreprocessor struct {
	*basePreProcess
	chReceivedAllRewardTxs chan bool
	onRequestRewardTx      func(shardID uint32, txHashes [][]byte)
	rewardTxsForBlock      txsForBlock
	rewardTxPool           dataRetriever.ShardedDataCacherNotifier
	storage                dataRetriever.StorageService
	rewardsProcessor       process.RewardTransactionProcessor
}

// NewRewardTxPreprocessor creates a new reward transaction preprocessor object
func NewRewardTxPreprocessor(
	rewardTxDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	rewardProcessor process.RewardTransactionProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestRewardTransaction func(shardID uint32, txHashes [][]byte),
	gasHandler process.GasHandler,
	pubkeyConverter core.PubkeyConverter,
	blockSizeComputation BlockSizeComputationHandler,
	balanceComputation BalanceComputationHandler,
) (*rewardTxPreprocessor, error) {

	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(rewardTxDataPool) {
		return nil, process.ErrNilRewardTxDataPool
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(rewardProcessor) {
		return nil, process.ErrNilRewardsTxProcessor
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestRewardTransaction == nil {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
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
		blockSizeComputation: blockSizeComputation,
		balanceComputation:   balanceComputation,
		accounts:             accounts,
		pubkeyConverter:      pubkeyConverter,
	}

	rtp := &rewardTxPreprocessor{
		basePreProcess:    bpp,
		storage:           store,
		rewardTxPool:      rewardTxDataPool,
		onRequestRewardTx: onRequestRewardTransaction,
		rewardsProcessor:  rewardProcessor,
	}

	rtp.chReceivedAllRewardTxs = make(chan bool)
	rtp.rewardTxPool.RegisterOnAdded(rtp.receivedRewardTransaction)
	rtp.rewardTxsForBlock.txHashAndInfo = make(map[string]*txInfo)

	return rtp, nil
}

// waitForRewardTxHashes waits for a call whether all the requested smartContractResults appeared
func (rtp *rewardTxPreprocessor) waitForRewardTxHashes(waitTime time.Duration) error {
	select {
	case <-rtp.chReceivedAllRewardTxs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested reward transactions arrived and were saved into the pool
func (rtp *rewardTxPreprocessor) IsDataPrepared(requestedRewardTxs int, haveTime func() time.Duration) error {
	if requestedRewardTxs > 0 {
		log.Debug("requested missing reward txs",
			"num reward txs", requestedRewardTxs)
		err := rtp.waitForRewardTxHashes(haveTime())
		rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
		missingRewardTxs := rtp.rewardTxsForBlock.missingTxs
		rtp.rewardTxsForBlock.missingTxs = 0
		rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()
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

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		rewardTxBuff, err := rtp.storage.GetAll(dataRetriever.RewardTransactionUnit, miniBlock.TxHashes)
		if err != nil {
			log.Debug("reward tx from mini block was not found in RewardTransactionUnit",
				"sender shard ID", miniBlock.SenderShardID,
				"receiver shard ID", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes),
			)

			return rewardTxsRestored, err
		}

		for txHash, txBuff := range rewardTxBuff {
			tx := rewardTx.RewardTx{}
			err = rtp.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return rewardTxsRestored, err
			}

			rtp.rewardTxPool.AddData([]byte(txHash), &tx, tx.Size(), strCache)
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

// ProcessBlockTransactions processes all the reward transactions from the block.Body, updates the state
func (rtp *rewardTxPreprocessor) ProcessBlockTransactions(
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

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			rtp.rewardTxsForBlock.mutTxsForBlock.RLock()
			txData, ok := rtp.rewardTxsForBlock.txHashAndInfo[string(txHash)]
			rtp.rewardTxsForBlock.mutTxsForBlock.RUnlock()
			if !ok || check.IfNil(txData.tx) {
				log.Warn("missing rewardsTransaction in ProcessBlockTransactions ", "type", miniBlock.Type, "hash", txHash)
				return process.ErrMissingTransaction
			}

			rTx, ok := txData.tx.(*rewardTx.RewardTx)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			rtp.saveAccountBalanceForAddress(rTx.GetRcvAddr())

			err := rtp.rewardsProcessor.ProcessRewardTransaction(rTx)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

		err := rtp.saveTxsToStorage(
			miniBlock.TxHashes,
			&rtp.rewardTxsForBlock,
			rtp.storage,
			dataRetriever.RewardTransactionUnit,
		)
		if err != nil {
			return err
		}
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

	receivedAllMissing := rtp.baseReceivedTransaction(key, tx, &rtp.rewardTxsForBlock)

	if receivedAllMissing {
		rtp.chReceivedAllRewardTxs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created reward transactions at this round
func (rtp *rewardTxPreprocessor) CreateBlockStarted() {
	_ = core.EmptyChannel(rtp.chReceivedAllRewardTxs)

	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
	rtp.rewardTxsForBlock.missingTxs = 0
	rtp.rewardTxsForBlock.txHashAndInfo = make(map[string]*txInfo)
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()
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
	rewardTxs := block.Body{}
	for _, mb := range body.MiniBlocks {
		if mb.Type != block.RewardsBlock {
			continue
		}
		if mb.SenderShardID != core.MetachainShardId {
			continue
		}

		rewardTxs.MiniBlocks = append(rewardTxs.MiniBlocks, mb)
	}

	numMissingTxsForShards := rtp.computeExistingAndRequestMissing(
		&rewardTxs,
		&rtp.rewardTxsForBlock,
		rtp.chReceivedAllRewardTxs,
		rtp.isMiniBlockCorrect,
		rtp.rewardTxPool,
		rtp.onRequestRewardTx,
	)

	return numMissingTxsForShards
}

// RequestTransactionsForMiniBlock requests missing reward transactions for a certain miniblock
func (rtp *rewardTxPreprocessor) RequestTransactionsForMiniBlock(miniBlock *block.MiniBlock) int {
	if miniBlock == nil {
		return 0
	}

	missingRewardTxsForMiniBlock := rtp.computeMissingRewardTxsForMiniBlock(miniBlock)
	if len(missingRewardTxsForMiniBlock) > 0 {
		rtp.onRequestRewardTx(miniBlock.SenderShardID, missingRewardTxsForMiniBlock)
	}

	return len(missingRewardTxsForMiniBlock)
}

// computeMissingRewardTxsForMiniBlock computes missing reward transactions for a certain miniblock
func (rtp *rewardTxPreprocessor) computeMissingRewardTxsForMiniBlock(miniBlock *block.MiniBlock) [][]byte {
	missingRewardTxs := make([][]byte, 0, len(miniBlock.TxHashes))
	if miniBlock.Type != block.RewardsBlock {
		return missingRewardTxs
	}

	for _, txHash := range miniBlock.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
			txHash,
			rtp.rewardTxPool,
			false,
		)

		if tx == nil {
			missingRewardTxs = append(missingRewardTxs, txHash)
		}
	}

	return sliceUtil.TrimSliceSliceByte(missingRewardTxs)
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

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (rtp *rewardTxPreprocessor) CreateAndProcessMiniBlocks(
	_ func() bool,
) (block.MiniBlockSlice, error) {
	// rewards are created only by meta
	return make(block.MiniBlockSlice, 0), nil
}

// ProcessMiniBlock processes all the reward transactions from a miniblock and saves the processed reward transactions
// in local cache
func (rtp *rewardTxPreprocessor) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, _ func() (int, int)) ([][]byte, int, error) {

	if miniBlock.Type != block.RewardsBlock {
		return nil, 0, process.ErrWrongTypeInMiniBlock
	}
	if miniBlock.SenderShardID != core.MetachainShardId {
		return nil, 0, process.ErrRewardMiniBlockNotFromMeta
	}

	miniBlockRewardTxs, miniBlockTxHashes, err := rtp.getAllRewardTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return nil, 0, err
	}

	if rtp.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, len(miniBlockRewardTxs)) {
		return nil, 0, process.ErrMaxBlockSizeReached
	}

	processedTxHashes := make([][]byte, 0)
	for index := range miniBlockRewardTxs {
		if !haveTime() {
			return processedTxHashes, index, process.ErrTimeIsOut
		}

		rtp.saveAccountBalanceForAddress(miniBlockRewardTxs[index].GetRcvAddr())

		err = rtp.rewardsProcessor.ProcessRewardTransaction(miniBlockRewardTxs[index])
		if err != nil {
			return processedTxHashes, index, err
		}

		processedTxHashes = append(processedTxHashes, miniBlockTxHashes[index])
	}

	txShardData := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		rtp.rewardTxsForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockRewardTxs[index], txShardInfo: txShardData}
	}
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()

	rtp.blockSizeComputation.AddNumMiniBlocks(1)
	rtp.blockSizeComputation.AddNumTxs(len(miniBlockRewardTxs))

	return nil, len(processedTxHashes), nil
}

// CreateMarshalizedData marshalizes reward transaction hashes and and saves them into a new structure
func (rtp *rewardTxPreprocessor) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	marshaledRewardTxs, err := rtp.createMarshalizedData(txHashes, &rtp.rewardTxsForBlock)
	if err != nil {
		return nil, err
	}

	return marshaledRewardTxs, nil
}

// GetAllCurrentUsedTxs returns all the reward transactions used at current creation / processing
func (rtp *rewardTxPreprocessor) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	rtp.rewardTxsForBlock.mutTxsForBlock.RLock()
	rewardTxPool := make(map[string]data.TransactionHandler, len(rtp.rewardTxsForBlock.txHashAndInfo))
	for txHash, txData := range rtp.rewardTxsForBlock.txHashAndInfo {
		rewardTxPool[txHash] = txData.tx
	}
	rtp.rewardTxsForBlock.mutTxsForBlock.RUnlock()

	return rewardTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *rewardTxPreprocessor) IsInterfaceNil() bool {
	return rtp == nil
}

func (rtp *rewardTxPreprocessor) isMiniBlockCorrect(mbType block.Type) bool {
	return mbType == block.RewardsBlock
}
