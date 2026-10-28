package track

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
)

type confirmedMiniBlockInfo struct {
	cacheID string
	mbType  block.Type
	nonce   uint64
}

type miniBlockTrack struct {
	blockTransactionsPool    dataRetriever.ShardedDataCacherNotifier
	rewardTransactionsPool   dataRetriever.ShardedDataCacherNotifier
	unsignedTransactionsPool dataRetriever.ShardedDataCacherNotifier
	miniBlocksPool           storage.Cacher
	shardCoordinator         sharding.Coordinator
	whitelistHandler         process.WhiteListHandler
	mutConfirmedMiniBlocks   sync.RWMutex
	confirmedMiniBlocks      map[string]confirmedMiniBlockInfo
}

// NewMiniBlockTrack creates an object for tracking the received mini blocks
func NewMiniBlockTrack(
	dataPool dataRetriever.PoolsHolder,
	blockTracker process.BlockTracker,
	shardCoordinator sharding.Coordinator,
	whitelistHandler process.WhiteListHandler,
) (*miniBlockTrack, error) {

	if check.IfNil(dataPool) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(dataPool.Transactions()) {
		return nil, process.ErrNilTransactionPool
	}
	if check.IfNil(dataPool.RewardTransactions()) {
		return nil, process.ErrNilRewardTxDataPool
	}
	if check.IfNil(dataPool.UnsignedTransactions()) {
		return nil, process.ErrNilUnsignedTxDataPool
	}
	if check.IfNil(dataPool.MiniBlocks()) {
		return nil, process.ErrNilMiniBlockPool
	}
	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(whitelistHandler) {
		return nil, process.ErrNilWhiteListHandler
	}

	mbt := miniBlockTrack{
		blockTransactionsPool:    dataPool.Transactions(),
		rewardTransactionsPool:   dataPool.RewardTransactions(),
		unsignedTransactionsPool: dataPool.UnsignedTransactions(),
		miniBlocksPool:           dataPool.MiniBlocks(),
		shardCoordinator:         shardCoordinator,
		whitelistHandler:         whitelistHandler,
		confirmedMiniBlocks:      make(map[string]confirmedMiniBlockInfo),
	}

	mbt.miniBlocksPool.RegisterHandler(mbt.receivedMiniBlock, core.UniqueIdentifier())
	mbt.registerBlockTrackerHandlers(blockTracker)

	return &mbt, nil
}

func (mbt *miniBlockTrack) receivedMiniBlock(key []byte, value interface{}) {
	if key == nil {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		log.Warn("miniBlockTrack.receivedMiniBlock", "error", process.ErrWrongTypeAssertion)
		return
	}

	log.Debug("received miniblock from network in block tracker",
		"hash", key,
		"sender", miniBlock.SenderShardID,
		"receiver", miniBlock.ReceiverShardID,
		"type", miniBlock.Type,
		"num txs", len(miniBlock.TxHashes))

	if miniBlock.SenderShardID == mbt.shardCoordinator.SelfId() {
		return
	}

	mbt.immunizeMiniBlock(key, miniBlock)
}

func (mbt *miniBlockTrack) getTransactionPool(mbType block.Type) dataRetriever.ShardedDataCacherNotifier {
	switch mbType {
	case block.TxBlock:
		return mbt.blockTransactionsPool
	case block.RewardsBlock:
		return mbt.rewardTransactionsPool
	case block.SmartContractResultBlock:
		return mbt.unsignedTransactionsPool
	}

	return nil
}

func (mbt *miniBlockTrack) registerBlockTrackerHandlers(blockTracker process.BlockTracker) {
	if mbt.shardCoordinator.SelfId() == core.MetachainShardId {
		blockTracker.RegisterCrossNotarizedHeadersHandler(func(_ uint32, headers []data.HeaderHandler, _ [][]byte) {
			mbt.registerConfirmedMiniBlocks(headers)
		})
		return
	}

	blockTracker.RegisterFinalMetachainHeadersHandler(func(_ uint32, headers []data.HeaderHandler, _ [][]byte) {
		mbt.registerConfirmedMiniBlocks(headers)
	})
}

func (mbt *miniBlockTrack) registerConfirmedMiniBlocks(headers []data.HeaderHandler) {
	for _, header := range headers {
		mbt.registerConfirmedMiniBlocksForHeader(header)
	}
}

func (mbt *miniBlockTrack) registerConfirmedMiniBlocksForHeader(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	switch typedHeader := header.(type) {
	case data.MetaHeaderHandler:
		mbt.registerFromMiniBlockHeaders(typedHeader.GetNonce(), core.MetachainShardId, typedHeader.GetMiniBlockHeaderHandlers())
		for _, shardInfo := range typedHeader.GetShardInfoHandlers() {
			mbt.registerFromMiniBlockHeaders(typedHeader.GetNonce(), shardInfo.GetShardID(), shardInfo.GetShardMiniBlockHeaderHandlers())
		}
	case data.ShardHeaderHandler:
		mbt.registerFromMiniBlockHeaders(typedHeader.GetNonce(), typedHeader.GetShardID(), typedHeader.GetMiniBlockHeaderHandlers())
	}
}

func (mbt *miniBlockTrack) registerFromMiniBlockHeaders(
	nonce uint64,
	processingShard uint32,
	miniBlockHeaders []data.MiniBlockHeaderHandler,
) {
	selfShardID := mbt.shardCoordinator.SelfId()
	for _, miniBlockHeader := range miniBlockHeaders {
		receiverShard := miniBlockHeader.GetReceiverShardID()
		// AllShardId from a metaheader (e.g. rewards) is treated as receiver = self.
		receiverIsAllShardsMiniBlockFromMetaHeader := receiverShard == core.AllShardId && processingShard == core.MetachainShardId
		receiverIsRelevantForCurrentShard := receiverShard == selfShardID || receiverIsAllShardsMiniBlockFromMetaHeader
		senderShard := miniBlockHeader.GetSenderShardID()
		senderIsSelfShard := senderShard == selfShardID
		if !receiverIsRelevantForCurrentShard || senderIsSelfShard {
			continue
		}

		cacheID := process.ShardCacherIdentifier(senderShard, receiverShard)
		mbInfo := confirmedMiniBlockInfo{
			cacheID: cacheID,
			mbType:  block.Type(miniBlockHeader.GetTypeInt32()),
			nonce:   nonce,
		}

		transactionPool := mbt.getTransactionPool(mbInfo.mbType)
		if check.IfNil(transactionPool) {
			continue
		}

		// Threshold advance is deferred to commit (see ReleaseImmunityForCommittedMetaBlocks).
		// Advancing here would release items from older metablocks before this shard executes them.
		mbt.storeConfirmedMiniBlockInfo(miniBlockHeader.GetHash(), mbInfo)
		mbt.tryProcessStoredMiniBlock(miniBlockHeader.GetHash())
	}
}

func (mbt *miniBlockTrack) tryProcessStoredMiniBlock(miniBlockHash []byte) {
	value, ok := mbt.miniBlocksPool.Peek(miniBlockHash)
	if !ok {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		return
	}

	mbt.immunizeMiniBlock(miniBlockHash, miniBlock)
}

func (mbt *miniBlockTrack) immunizeMiniBlock(miniBlockHash []byte, miniBlock *block.MiniBlock) {
	// TODO - stop reusing miniBlock.TxHashes for peer changes, add new fields
	transactionPool := mbt.getTransactionPool(miniBlock.Type)
	if check.IfNil(transactionPool) {
		return
	}

	confirmationInfo, ok := mbt.getConfirmedMiniBlockInfo(miniBlockHash)
	if !ok {
		return
	}

	mbt.whitelistHandler.Add(miniBlock.TxHashes)
	transactionPool.ImmunizeSetOfDataAgainstEviction(miniBlock.TxHashes, confirmationInfo.cacheID, confirmationInfo.nonce)
	mbt.removeConfirmedMiniBlockInfo(miniBlockHash, confirmationInfo.nonce)
}

func (mbt *miniBlockTrack) storeConfirmedMiniBlockInfo(miniBlockHash []byte, info confirmedMiniBlockInfo) {
	mbt.mutConfirmedMiniBlocks.Lock()
	defer mbt.mutConfirmedMiniBlocks.Unlock()

	key := string(miniBlockHash)
	existingInfo, exists := mbt.confirmedMiniBlocks[key]
	if exists && existingInfo.nonce >= info.nonce {
		return
	}

	mbt.confirmedMiniBlocks[key] = info
}

func (mbt *miniBlockTrack) getConfirmedMiniBlockInfo(miniBlockHash []byte) (confirmedMiniBlockInfo, bool) {
	mbt.mutConfirmedMiniBlocks.RLock()
	defer mbt.mutConfirmedMiniBlocks.RUnlock()

	info, ok := mbt.confirmedMiniBlocks[string(miniBlockHash)]
	return info, ok
}

func (mbt *miniBlockTrack) removeConfirmedMiniBlockInfo(miniBlockHash []byte, nonce uint64) {
	mbt.mutConfirmedMiniBlocks.Lock()
	defer mbt.mutConfirmedMiniBlocks.Unlock()

	key := string(miniBlockHash)
	info, ok := mbt.confirmedMiniBlocks[key]
	if !ok {
		return
	}
	if info.nonce > nonce {
		return
	}

	delete(mbt.confirmedMiniBlocks, key)
}

// CleanupConfirmedMiniBlocksBelow drops every tracked confirmation whose nonce
// is strictly below `threshold`. Called from the shard's commit path alongside
// SetOldestImmuneNonceForAllCaches so that the local registry doesn't accumulate
// stale entries for miniblocks that never arrived in the pool.
func (mbt *miniBlockTrack) CleanupConfirmedMiniBlocksBelow(threshold uint64) {
	mbt.mutConfirmedMiniBlocks.Lock()
	defer mbt.mutConfirmedMiniBlocks.Unlock()

	for key, info := range mbt.confirmedMiniBlocks {
		if info.nonce >= threshold {
			continue
		}

		delete(mbt.confirmedMiniBlocks, key)
	}
}

// CleanupConfirmedMiniBlocksBelowForCacheID drops every tracked confirmation whose
// cacheID matches and nonce is strictly below `threshold`. Used by the meta commit
// path where the threshold is per-sender-shard rather than uniform.
func (mbt *miniBlockTrack) CleanupConfirmedMiniBlocksBelowForCacheID(cacheID string, threshold uint64) {
	mbt.mutConfirmedMiniBlocks.Lock()
	defer mbt.mutConfirmedMiniBlocks.Unlock()

	for key, info := range mbt.confirmedMiniBlocks {
		if info.cacheID != cacheID || info.nonce >= threshold {
			continue
		}

		delete(mbt.confirmedMiniBlocks, key)
	}
}

// ReleaseImmunityForCommittedMetaBlocks advances the immunity threshold uniformly
// across every tx-pool cache and prunes the local registry for entries below
// `threshold`. Called from the shard's commit path once the cross-notarized
// metablock has advanced past (threshold-1).
func (mbt *miniBlockTrack) ReleaseImmunityForCommittedMetaBlocks(threshold uint64) {
	if !check.IfNil(mbt.blockTransactionsPool) {
		mbt.blockTransactionsPool.SetOldestImmuneNonceForAllCaches(threshold)
	}
	if !check.IfNil(mbt.rewardTransactionsPool) {
		mbt.rewardTransactionsPool.SetOldestImmuneNonceForAllCaches(threshold)
	}
	if !check.IfNil(mbt.unsignedTransactionsPool) {
		mbt.unsignedTransactionsPool.SetOldestImmuneNonceForAllCaches(threshold)
	}
	mbt.CleanupConfirmedMiniBlocksBelow(threshold)
}

// ReleaseImmunityForCommittedShardBlocks advances the immunity threshold only on
// caches with senderShardID = `senderShard` and receiver = metachain, and prunes
// the local registry for matching entries below `threshold`. Called from the
// meta processor after its cross-notarized shard header has advanced for `senderShard`.
func (mbt *miniBlockTrack) ReleaseImmunityForCommittedShardBlocks(senderShard uint32, threshold uint64) {
	cacheID := process.ShardCacherIdentifier(senderShard, core.MetachainShardId)
	if !check.IfNil(mbt.blockTransactionsPool) {
		mbt.blockTransactionsPool.SetOldestImmuneNonce(cacheID, threshold)
	}
	if !check.IfNil(mbt.rewardTransactionsPool) {
		mbt.rewardTransactionsPool.SetOldestImmuneNonce(cacheID, threshold)
	}
	if !check.IfNil(mbt.unsignedTransactionsPool) {
		mbt.unsignedTransactionsPool.SetOldestImmuneNonce(cacheID, threshold)
	}
	mbt.CleanupConfirmedMiniBlocksBelowForCacheID(cacheID, threshold)
}

// IsInterfaceNil returns true if the receiver is a nil interface
func (mbt *miniBlockTrack) IsInterfaceNil() bool {
	return mbt == nil
}
