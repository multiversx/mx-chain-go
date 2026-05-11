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

	confirmationInfo, ok := mbt.getConfirmedMiniBlockInfo(key)
	if !ok {
		return
	}

	mbt.immunizeMiniBlock(key, miniBlock, confirmationInfo)
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
		receiverIsAllShardsMiniBlockFromMetaHeader := receiverShard == core.AllShardId && processingShard == core.MetachainShardId
		receiverIsRelevantForCurrentShard := receiverShard == selfShardID || receiverIsAllShardsMiniBlockFromMetaHeader
		senderShard := miniBlockHeader.GetSenderShardID()
		senderIsSelfShard := senderShard == selfShardID
		// Track only miniblocks that are relevant for this shard and come from another shard.
		// This includes direct cross-shard miniblocks addressed to this shard and the
		// special metachain-header case where the receiver is AllShardId.
		// Intra-shard miniblocks are produced and processed locally, so they are skipped here.
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

		mbt.storeConfirmedMiniBlockInfo(miniBlockHeader.GetHash(), mbInfo)
		transactionPool.SetOldestImmuneNonce(cacheID, nonce)
		mbt.cleanupConfirmedMiniBlocks(cacheID, nonce)
		mbt.tryProcessStoredMiniBlock(miniBlockHeader.GetHash(), mbInfo)
	}
}

func (mbt *miniBlockTrack) tryProcessStoredMiniBlock(miniBlockHash []byte, confirmationInfo confirmedMiniBlockInfo) {
	value, ok := mbt.miniBlocksPool.Peek(miniBlockHash)
	if !ok {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		return
	}

	mbt.immunizeMiniBlock(miniBlockHash, miniBlock, confirmationInfo)
}

func (mbt *miniBlockTrack) immunizeMiniBlock(miniBlockHash []byte, miniBlock *block.MiniBlock, confirmationInfo confirmedMiniBlockInfo) {
	// TODO - stop reusing miniBlock.TxHashes for peer changes, add new fields
	transactionPool := mbt.getTransactionPool(miniBlock.Type)
	if check.IfNil(transactionPool) {
		return
	}

	mbt.whitelistHandler.Add(miniBlock.TxHashes)
	transactionPool.SetOldestImmuneNonce(confirmationInfo.cacheID, confirmationInfo.nonce)
	transactionPool.ImmunizeSetOfDataAgainstEviction(miniBlock.TxHashes, confirmationInfo.cacheID, confirmationInfo.nonce)
	mbt.removeConfirmedMiniBlockInfo(miniBlockHash)
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

func (mbt *miniBlockTrack) removeConfirmedMiniBlockInfo(miniBlockHash []byte) {
	mbt.mutConfirmedMiniBlocks.Lock()
	delete(mbt.confirmedMiniBlocks, string(miniBlockHash))
	mbt.mutConfirmedMiniBlocks.Unlock()
}

func (mbt *miniBlockTrack) cleanupConfirmedMiniBlocks(cacheID string, nonce uint64) {
	mbt.mutConfirmedMiniBlocks.Lock()
	defer mbt.mutConfirmedMiniBlocks.Unlock()

	for key, info := range mbt.confirmedMiniBlocks {
		if info.cacheID != cacheID || info.nonce >= nonce {
			continue
		}

		delete(mbt.confirmedMiniBlocks, key)
	}
}
