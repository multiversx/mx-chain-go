package preprocess

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// txsForBlock holds the information about the missing and existing transactions required for a block.
type txsForBlock struct {
	shardCoordinator sharding.Coordinator
	numMissingTxs    int
	mutTxsForBlock   sync.RWMutex
	txHashAndInfo    map[string]*process.TxInfo
	chRcvAllTxs      chan bool
}

// NewTxsForBlock creates a new instance of txsForBlock
func NewTxsForBlock(shardCoordinator sharding.Coordinator) (*txsForBlock, error) {
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	return &txsForBlock{
		shardCoordinator: shardCoordinator,
		numMissingTxs:    0,
		mutTxsForBlock:   sync.RWMutex{},
		txHashAndInfo:    make(map[string]*process.TxInfo),
		chRcvAllTxs:      make(chan bool),
	}, nil
}

// Reset resets the state of txsForBlock, clearing the missing transactions count and the transaction hash map.
func (tfb *txsForBlock) Reset() {
	_ = core.EmptyChannel(tfb.chRcvAllTxs)

	tfb.mutTxsForBlock.Lock()
	defer tfb.mutTxsForBlock.Unlock()
	tfb.numMissingTxs = 0
	tfb.txHashAndInfo = make(map[string]*process.TxInfo)
}

// GetTxInfoByHash retrieves the transaction information by its hash.
func (tfb *txsForBlock) GetTxInfoByHash(hash []byte) (*process.TxInfo, bool) {
	tfb.mutTxsForBlock.RLock()
	defer tfb.mutTxsForBlock.RUnlock()

	value, ok := tfb.txHashAndInfo[string(hash)]
	return value, ok
}

// GetAllCurrentUsedTxs returns all the transactions used at current creation / processing
func (tfb *txsForBlock) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	tfb.mutTxsForBlock.RLock()
	txsPool := make(map[string]data.TransactionHandler, len(tfb.txHashAndInfo))
	for txHash, txInfoFromMap := range tfb.txHashAndInfo {
		txsPool[txHash] = txInfoFromMap.Tx
	}
	tfb.mutTxsForBlock.RUnlock()

	return txsPool
}

// ReceivedTransaction updates the transaction information in the txsForBlock instance when a transaction is received.
func (tfb *txsForBlock) ReceivedTransaction(
	txHash []byte,
	tx data.TransactionHandler,
) {
	tfb.mutTxsForBlock.Lock()
	defer tfb.mutTxsForBlock.Unlock()

	if tfb.numMissingTxs <= 0 {
		return
	}

	txInfoForHash := tfb.txHashAndInfo[string(txHash)]
	if txInfoForHash != nil && txInfoForHash.TxShardInfo != nil &&
		(txInfoForHash.Tx == nil || txInfoForHash.Tx.IsInterfaceNil()) {
		tfb.txHashAndInfo[string(txHash)].Tx = tx
		tfb.numMissingTxs--
	}

	if tfb.numMissingTxs == 0 {
		go func() {
			tfb.chRcvAllTxs <- true
		}()
	}
}

// AddTransaction adds a transaction to the txsForBlock instance with its associated sender and receiver shard IDs.
func (tfb *txsForBlock) AddTransaction(
	txHash []byte,
	tx data.TransactionHandler,
	senderShardID uint32,
	receiverShardID uint32,
) {
	if check.IfNil(tx) {
		return
	}
	txShardInfoToSet := &process.TxShardInfo{SenderShardID: senderShardID, ReceiverShardID: receiverShardID}
	tfb.mutTxsForBlock.Lock()
	tfb.txHashAndInfo[string(txHash)] = &process.TxInfo{Tx: tx, TxShardInfo: txShardInfoToSet}
	tfb.mutTxsForBlock.Unlock()
}

// HasMissingTransactions checks if there are any missing transactions in the txsForBlock instance.
func (tfb *txsForBlock) HasMissingTransactions() bool {
	tfb.mutTxsForBlock.RLock()
	defer tfb.mutTxsForBlock.RUnlock()

	return tfb.numMissingTxs > 0
}

// GetMissingTxsCount returns the count of missing transactions in the txsForBlock instance.
func (tfb *txsForBlock) GetMissingTxsCount() int {
	tfb.mutTxsForBlock.RLock()
	defer tfb.mutTxsForBlock.RUnlock()

	return tfb.numMissingTxs
}

// ComputeExistingAndRequestMissing processes the block body to compute existing transactions (in the node tx pool) and request missing ones.
func (tfb *txsForBlock) ComputeExistingAndRequestMissing(
	body *block.Body,
	isMiniBlockCorrect func(block.Type) bool,
	txPool dataRetriever.ShardedDataCacherNotifier,
	onRequestTxs func(shardID uint32, txHashes [][]byte),
) int {
	if check.IfNil(body) {
		return 0
	}

	tfb.mutTxsForBlock.Lock()
	defer tfb.mutTxsForBlock.Unlock()

	missingTxsForShard := make(map[uint32][][]byte, tfb.shardCoordinator.NumberOfShards())
	missingTxHashes := make([][]byte, 0)
	uniqueTxHashes := make(map[string]struct{})
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if !isMiniBlockCorrect(miniBlock.Type) {
			continue
		}

		txShardInfoObject := &process.TxShardInfo{SenderShardID: miniBlock.SenderShardID, ReceiverShardID: miniBlock.ReceiverShardID}
		// TODO refactor this section
		method := process.SearchMethodJustPeek
		if miniBlock.Type == block.InvalidBlock {
			method = process.SearchMethodSearchFirst
		}
		if miniBlock.Type == block.SmartContractResultBlock {
			method = process.SearchMethodPeekWithFallbackSearchFirst
		}

		missingTxHashesInMiniBlock := tfb.updateExistingAndComputeMissingTxsInMiniBlockNoLock(miniBlock, uniqueTxHashes, txPool, method, txShardInfoObject)
		missingTxHashes = append(missingTxHashes, missingTxHashesInMiniBlock...)
		if len(missingTxHashes) > 0 {
			tfb.setMissingTxsForShardNoLock(miniBlock.SenderShardID, miniBlock.ReceiverShardID, missingTxHashes)
			missingTxsForShard[miniBlock.SenderShardID] = append(missingTxsForShard[miniBlock.SenderShardID], missingTxHashes...)
		}

		missingTxHashes = make([][]byte, 0)
	}

	return tfb.requestMissingTxsForShardNoLock(missingTxsForShard, onRequestTxs)
}

func (tfb *txsForBlock) updateExistingAndComputeMissingTxsInMiniBlockNoLock(
	miniBlock *block.MiniBlock,
	uniqueTxHashes map[string]struct{},
	txPool dataRetriever.ShardedDataCacherNotifier,
	method process.ShardedCacheSearchMethod,
	txShardInfoObject *process.TxShardInfo,
) [][]byte {
	missingTxHashes := make([][]byte, 0)
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
			missingTxHashes = append(missingTxHashes, txHash)
			tfb.numMissingTxs++
			log.Trace("missing tx",
				"miniblock type", miniBlock.Type,
				"sender", miniBlock.SenderShardID,
				"receiver", miniBlock.ReceiverShardID,
				"hash", txHash,
			)
			continue
		}

		tfb.txHashAndInfo[string(txHash)] = &process.TxInfo{Tx: tx, TxShardInfo: txShardInfoObject}
	}

	return missingTxHashes
}

// this method should be called only under the mutex protection: forBlock.mutTxsForBlock
func (tfb *txsForBlock) setMissingTxsForShardNoLock(
	senderShardID uint32,
	receiverShardID uint32,
	txHashes [][]byte,
) {
	txShardInfoToSet := &process.TxShardInfo{
		SenderShardID:   senderShardID,
		ReceiverShardID: receiverShardID,
	}

	for _, txHash := range txHashes {
		tfb.txHashAndInfo[string(txHash)] = &process.TxInfo{
			Tx:          nil,
			TxShardInfo: txShardInfoToSet,
		}
	}
}

// this method should be called only under the mutex protection: forBlock.mutTxsForBlock
func (tfb *txsForBlock) requestMissingTxsForShardNoLock(
	missingTxsForShard map[uint32][][]byte,
	onRequestTxs func(shardID uint32, txHashes [][]byte),
) int {
	requestedTxs := 0
	for shardID, txHashes := range missingTxsForShard {
		requestedTxs += len(txHashes)
		go func(providedShardID uint32, providedTxHashes [][]byte) {
			onRequestTxs(providedShardID, providedTxHashes)
		}(shardID, txHashes)
	}

	return requestedTxs
}

// WaitForRequestedData waits for the requested data to be received within the specified wait time.
func (tfb *txsForBlock) WaitForRequestedData(waitTime time.Duration) error {
	if !tfb.HasMissingTransactions() {
		core.EmptyChannel(tfb.chRcvAllTxs)
		return nil
	}

	select {
	case <-tfb.chRcvAllTxs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsInterfaceNil checks if the txsForBlock instance is nil.
func (tfb *txsForBlock) IsInterfaceNil() bool {
	return tfb == nil
}
