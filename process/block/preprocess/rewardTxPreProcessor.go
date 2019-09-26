package preprocess

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
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

type rewardTxPreprocessor struct {
	*basePreProcess
	chReceivedAllRewardTxs chan bool
	onRequestRewardTx      func(shardID uint32, txHashes [][]byte)
	rewardTxsForBlock      txsForBlock
	rewardTxPool           dataRetriever.ShardedDataCacherNotifier
	storage                dataRetriever.StorageService
	rewardsProcessor       process.RewardTransactionProcessor
	rewardsProducer        process.InternalTransactionProducer
	accounts               state.AccountsAdapter
}

// NewRewardTxPreprocessor creates a new reward transaction preprocessor object
func NewRewardTxPreprocessor(
	rewardTxDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	rewardProcessor process.RewardTransactionProcessor,
	rewardProducer process.InternalTransactionProducer,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestRewardTransaction func(shardID uint32, txHashes [][]byte),
) (*rewardTxPreprocessor, error) {

	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if rewardTxDataPool == nil || rewardTxDataPool.IsInterfaceNil() {
		return nil, process.ErrNilRewardTxDataPool
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilStorage
	}
	if rewardProcessor == nil || rewardProcessor.IsInterfaceNil() {
		return nil, process.ErrNilRewardsTxProcessor
	}
	if rewardProducer == nil || rewardProcessor.IsInterfaceNil() {
		return nil, process.ErrNilInternalTransactionProducer
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestRewardTransaction == nil {
		return nil, process.ErrNilRequestHandler
	}

	bpp := &basePreProcess{
		hasher:           hasher,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
	}

	rtp := &rewardTxPreprocessor{
		basePreProcess:    bpp,
		storage:           store,
		rewardTxPool:      rewardTxDataPool,
		onRequestRewardTx: onRequestRewardTransaction,
		rewardsProcessor:  rewardProcessor,
		rewardsProducer:   rewardProducer,
		accounts:          accounts,
	}

	rtp.chReceivedAllRewardTxs = make(chan bool)
	rtp.rewardTxPool.RegisterHandler(rtp.receivedRewardTransaction)
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
		log.Info(fmt.Sprintf("requested %d missing reward Txs\n", requestedRewardTxs))
		err := rtp.waitForRewardTxHashes(haveTime())
		rtp.rewardTxsForBlock.mutTxsForBlock.RLock()
		missingRewardTxs := rtp.rewardTxsForBlock.missingTxs
		rtp.rewardTxsForBlock.mutTxsForBlock.RUnlock()
		log.Info(fmt.Sprintf("received %d missing reward Txs\n", requestedRewardTxs-missingRewardTxs))
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveTxBlockFromPools removes reward transactions and miniblocks from associated pools
func (rtp *rewardTxPreprocessor) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil {
		return process.ErrNilTxBlockBody
	}

	return rtp.removeDataFromPools(body, miniBlockPool, rtp.rewardTxPool, block.RewardsBlock)
}

// RestoreTxBlockIntoPools restores the reward transactions and miniblocks to associated pools
func (rtp *rewardTxPreprocessor) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockPool storage.Cacher,
) (int, map[int][]byte, error) {
	if miniBlockPool == nil {
		return 0, nil, process.ErrNilMiniBlockPool
	}

	miniBlockHashes := make(map[int][]byte)

	rewardTxsRestored := 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		rewardTxBuff, err := rtp.storage.GetAll(dataRetriever.RewardTransactionUnit, miniBlock.TxHashes)
		if err != nil {
			return rewardTxsRestored, miniBlockHashes, err
		}

		for txHash, txBuff := range rewardTxBuff {
			tx := rewardTx.RewardTx{}
			err = rtp.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return rewardTxsRestored, miniBlockHashes, err
			}

			rtp.rewardTxPool.AddData([]byte(txHash), &tx, strCache)
		}

		miniBlockHash, err := core.CalculateHash(rtp.marshalizer, rtp.hasher, miniBlock)
		if err != nil {
			return rewardTxsRestored, miniBlockHashes, err
		}

		restoredHash := rtp.restoreMiniBlock(miniBlock, miniBlockHash, miniBlockPool)

		err = rtp.storage.GetStorer(dataRetriever.MiniBlockUnit).Remove(miniBlockHash)
		if err != nil {
			return rewardTxsRestored, miniBlockHashes, err
		}

		miniBlockHashes[i] = restoredHash
		rewardTxsRestored += len(miniBlock.TxHashes)
	}

	return rewardTxsRestored, miniBlockHashes, nil
}

// ProcessBlockTransactions processes all the reward transactions from the block.Body, updates the state
func (rtp *rewardTxPreprocessor) ProcessBlockTransactions(body block.Body, round uint64, haveTime func() bool) error {
	rewardMiniBlocksSlice := make(block.MiniBlockSlice, 0)
	computedRewardsMbsMap := rtp.rewardsProducer.CreateAllInterMiniBlocks()
	for _, mb := range computedRewardsMbsMap {
		rewardMiniBlocksSlice = append(rewardMiniBlocksSlice, mb)
	}
	rtp.AddComputedRewardMiniBlocks(rewardMiniBlocksSlice)

	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.RewardsBlock {
			continue
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if !haveTime() {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			rtp.rewardTxsForBlock.mutTxsForBlock.RLock()
			txInfo := rtp.rewardTxsForBlock.txHashAndInfo[string(txHash)]
			rtp.rewardTxsForBlock.mutTxsForBlock.RUnlock()
			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingTransaction
			}

			rTx, ok := txInfo.tx.(*rewardTx.RewardTx)
			if !ok {
				return process.ErrWrongTypeAssertion
			}

			err := rtp.processRewardTransaction(
				txHash,
				rTx,
				round,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddComputedRewardMiniBlocks adds to the local cache the reward transactions from the given miniblocks
func (rtp *rewardTxPreprocessor) AddComputedRewardMiniBlocks(computedRewardMiniblocks block.MiniBlockSlice) {
	for _, rewardMb := range computedRewardMiniblocks {
		txShardInfo := &txShardInfo{senderShardID: rewardMb.SenderShardID, receiverShardID: rewardMb.ReceiverShardID}
		for _, txHash := range rewardMb.TxHashes {
			tx, ok := rtp.rewardTxPool.SearchFirstData(txHash)
			if !ok {
				log.Error(process.ErrRewardTransactionNotFound.Error())
				continue
			}

			rTx, ok := tx.(*rewardTx.RewardTx)
			if !ok {
				log.Error(process.ErrWrongTypeAssertion.Error())
			}

			rtp.rewardTxsForBlock.txHashAndInfo[string(txHash)] = &txInfo{
				tx:          rTx,
				txShardInfo: txShardInfo,
			}
		}
	}
}

// SaveTxBlockToStorage saves the reward transactions from body into storage
func (rtp *rewardTxPreprocessor) SaveTxBlockToStorage(body block.Body) error {
	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
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
func (rtp *rewardTxPreprocessor) receivedRewardTransaction(txHash []byte) {
	receivedAllMissing := rtp.baseReceivedTransaction(txHash, &rtp.rewardTxsForBlock, rtp.rewardTxPool)

	if receivedAllMissing {
		rtp.chReceivedAllRewardTxs <- true
	}
}

// CreateBlockStarted cleans the local cache map for processed/created reward transactions at this round
func (rtp *rewardTxPreprocessor) CreateBlockStarted() {
	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
	rtp.rewardTxsForBlock.txHashAndInfo = make(map[string]*txInfo)
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()
}

// RequestBlockTransactions request for reward transactions if missing from a block.Body
func (rtp *rewardTxPreprocessor) RequestBlockTransactions(body block.Body) int {
	requestedRewardTxs := 0
	missingRewardTxsForShards := rtp.computeMissingAndExistingRewardTxsForShards(body)

	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
	for senderShardID, rewardTxHashes := range missingRewardTxsForShards {
		for _, txHash := range rewardTxHashes {
			rtp.setMissingTxsForShard(senderShardID, txHash)
		}
	}
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()

	for senderShardID, mbsRewardTxHashes := range missingRewardTxsForShards {
		for _, mbRewardTxHashes := range mbsRewardTxHashes {
			requestedRewardTxs += len(mbRewardTxHashes.txHashes)
			rtp.onRequestRewardTx(senderShardID, mbRewardTxHashes.txHashes)
		}
	}

	return requestedRewardTxs
}

func (rtp *rewardTxPreprocessor) setMissingTxsForShard(senderShardID uint32, mbTxHashes *txsHashesInfo) {
	txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: mbTxHashes.receiverShardID}
	for _, txHash := range mbTxHashes.txHashes {
		rtp.rewardTxsForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfo}
	}
}

// computeMissingAndExistingRewardTxsForShards calculates what reward transactions are available and what are missing
// from block.Body
func (rtp *rewardTxPreprocessor) computeMissingAndExistingRewardTxsForShards(body block.Body) map[uint32][]*txsHashesInfo {
	rewardTxs := block.Body{}
	for _, mb := range body {
		if mb.Type != block.RewardsBlock {
			continue
		}
		if mb.SenderShardID == rtp.shardCoordinator.SelfId() {
			continue
		}

		rewardTxs = append(rewardTxs, mb)
	}

	missingTxsForShards := rtp.computeExistingAndMissing(
		rewardTxs,
		&rtp.rewardTxsForBlock,
		rtp.chReceivedAllRewardTxs,
		block.RewardsBlock,
		rtp.rewardTxPool,
	)

	return missingTxsForShards
}

// processRewardTransaction processes a reward transaction, if the transactions has an error it removes it from pool
func (rtp *rewardTxPreprocessor) processRewardTransaction(
	rewardTxHash []byte,
	rewardTx *rewardTx.RewardTx,
	round uint64,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := rtp.rewardsProcessor.ProcessRewardTransaction(rewardTx)
	if err != nil {
		return err
	}

	txShardInfo := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
	rtp.rewardTxsForBlock.txHashAndInfo[string(rewardTxHash)] = &txInfo{tx: rewardTx, txShardInfo: txShardInfo}
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()

	return nil
}

// RequestTransactionsForMiniBlock requests missing reward transactions for a certain miniblock
func (rtp *rewardTxPreprocessor) RequestTransactionsForMiniBlock(mb block.MiniBlock) int {
	missingRewardTxsForMiniBlock := rtp.computeMissingRewardTxsForMiniBlock(mb)
	rtp.onRequestRewardTx(mb.SenderShardID, missingRewardTxsForMiniBlock)

	return len(missingRewardTxsForMiniBlock)
}

// computeMissingRewardTxsForMiniBlock computes missing reward transactions for a certain miniblock
func (rtp *rewardTxPreprocessor) computeMissingRewardTxsForMiniBlock(mb block.MiniBlock) [][]byte {
	missingRewardTxs := make([][]byte, 0)
	if mb.Type != block.RewardsBlock {
		return missingRewardTxs
	}

	for _, txHash := range mb.TxHashes {
		tx, _ := process.GetTransactionHandlerFromPool(
			mb.SenderShardID,
			mb.ReceiverShardID,
			txHash,
			rtp.rewardTxPool,
		)

		if tx == nil {
			missingRewardTxs = append(missingRewardTxs, txHash)
		}
	}

	return missingRewardTxs
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
	rewardTxs := make([]*rewardTx.RewardTx, 0)
	txHashes := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
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

		txHashes = append(txHashes, txHash)
		rewardTxs = append(rewardTxs, tx)
	}

	return rewardTxs, txHashes, nil
}

// CreateAndProcessMiniBlock creates the miniblock from storage and processes the reward transactions added into the miniblock
func (rtp *rewardTxPreprocessor) CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint64) (*block.MiniBlock, error) {
	return nil, nil
}

// CreateAndProcessMiniBlocks creates miniblocks from storage and processes the reward transactions added into the miniblocks
// as long as it has time
func (rtp *rewardTxPreprocessor) CreateAndProcessMiniBlocks(
	maxTxSpaceRemained uint32,
	maxMbSpaceRemained uint32,
	round uint64,
	_ func() bool,
) (block.MiniBlockSlice, error) {

	// always have time for rewards
	haveTime := func() bool {
		return true
	}

	rewardMiniBlocksSlice := make(block.MiniBlockSlice, 0)
	computedRewardsMbsMap := rtp.rewardsProducer.CreateAllInterMiniBlocks()
	for _, mb := range computedRewardsMbsMap {
		rewardMiniBlocksSlice = append(rewardMiniBlocksSlice, mb)
	}

	snapshot := rtp.accounts.JournalLen()

	for _, mb := range rewardMiniBlocksSlice {
		err := rtp.ProcessMiniBlock(mb, haveTime, round)

		if err != nil {
			log.Error(err.Error())
			errAccountState := rtp.accounts.RevertToSnapshot(snapshot)
			if errAccountState != nil {
				// TODO: evaluate if reloading the trie from disk will might solve the problem
				log.Error(errAccountState.Error())
			}
			return nil, err
		}
	}

	return rewardMiniBlocksSlice, nil
}

// ProcessMiniBlock processes all the reward transactions from a miniblock and saves the processed reward transactions
// in local cache
func (rtp *rewardTxPreprocessor) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint64) error {
	if miniBlock.Type != block.RewardsBlock {
		return process.ErrWrongTypeInMiniBlock
	}

	miniBlockRewardTxs, miniBlockTxHashes, err := rtp.getAllRewardTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	for index := range miniBlockRewardTxs {
		if !haveTime() {
			return process.ErrTimeIsOut
		}

		err = rtp.rewardsProcessor.ProcessRewardTransaction(miniBlockRewardTxs[index])
		if err != nil {
			return err
		}
	}

	txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	rtp.rewardTxsForBlock.mutTxsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		rtp.rewardTxsForBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: miniBlockRewardTxs[index], txShardInfo: txShardInfo}
	}
	rtp.rewardTxsForBlock.mutTxsForBlock.Unlock()

	return nil
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
	rewardTxPool := make(map[string]data.TransactionHandler)

	rtp.rewardTxsForBlock.mutTxsForBlock.RLock()
	for txHash, txInfo := range rtp.rewardTxsForBlock.txHashAndInfo {
		rewardTxPool[txHash] = txInfo.tx
	}
	rtp.rewardTxsForBlock.mutTxsForBlock.RUnlock()

	return rewardTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtp *rewardTxPreprocessor) IsInterfaceNil() bool {
	if rtp == nil {
		return true
	}
	return false
}
