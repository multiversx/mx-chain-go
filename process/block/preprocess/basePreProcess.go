package preprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TODO: increase code coverage with unit tests

const initialTxHashesSliceLen = 10

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx data.TransactionHandler
	*txShardInfo
}

type txsHashesInfo struct {
	txHashes        [][]byte
	receiverShardID uint32
}

type txsForBlock struct {
	missingTxs     int
	mutTxsForBlock sync.RWMutex
	txHashAndInfo  map[string]*txInfo
}

type basePreProcess struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	gasHandler       process.GasHandler
	economicsFee     process.FeeHandler
}

func (bpp *basePreProcess) removeDataFromPools(body *block.Body, miniBlockPool storage.Cacher, txPool dataRetriever.ShardedDataCacherNotifier, mbType block.Type) error {
	if miniBlockPool == nil || miniBlockPool.IsInterfaceNil() {
		return process.ErrNilMiniBlockPool
	}
	if txPool == nil || txPool.IsInterfaceNil() {
		return process.ErrNilTransactionPool
	}

	if body == nil {
		return process.ErrNilTxBlockBody
	}
	for i := 0; i < len(body.MiniBlocks); i++ {
		currentMiniBlock := body.MiniBlocks[i]
		if currentMiniBlock.Type != mbType {
			continue
		}

		strCache := process.ShardCacherIdentifier(currentMiniBlock.SenderShardID, currentMiniBlock.ReceiverShardID)
		txPool.RemoveSetOfDataFromPool(currentMiniBlock.TxHashes, strCache)

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

		if txInfoFromMap == nil || txInfoFromMap.tx == nil {
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
) error {

	for i := 0; i < len(txHashes); i++ {
		txHash := txHashes[i]

		forBlock.mutTxsForBlock.RLock()
		txInfoFromMap := forBlock.txHashAndInfo[string(txHash)]
		forBlock.mutTxsForBlock.RUnlock()

		if txInfoFromMap == nil || txInfoFromMap.tx == nil {
			log.Debug("missing transaction in saveTxsToStorage ", "type", dataUnit, "txHash", txHash)
			return process.ErrMissingTransaction
		}

		buff, err := bpp.marshalizer.Marshal(txInfoFromMap.tx)
		if err != nil {
			return err
		}

		errNotCritical := store.Put(dataUnit, txHash, buff)
		if errNotCritical != nil {
			log.Debug("store.Put",
				"error", errNotCritical.Error(),
				"dataUnit", dataUnit,
			)
		}
	}

	return nil
}

func (bpp *basePreProcess) baseReceivedTransaction(
	txHash []byte,
	forBlock *txsForBlock,
	txPool dataRetriever.ShardedDataCacherNotifier,
	blockType block.Type,
) bool {
	searchFirst := blockType == block.InvalidBlock
	forBlock.mutTxsForBlock.Lock()

	if forBlock.missingTxs > 0 {
		txInfoForHash := forBlock.txHashAndInfo[string(txHash)]
		if txInfoForHash != nil && txInfoForHash.txShardInfo != nil &&
			(txInfoForHash.tx == nil || txInfoForHash.tx.IsInterfaceNil()) {
			tx, _ := process.GetTransactionHandlerFromPool(
				txInfoForHash.senderShardID,
				txInfoForHash.receiverShardID,
				txHash,
				txPool,
				searchFirst)

			if tx != nil {
				forBlock.txHashAndInfo[string(txHash)].tx = tx
				forBlock.missingTxs--
			}
		}
		missingTxs := forBlock.missingTxs
		forBlock.mutTxsForBlock.Unlock()

		return missingTxs == 0
	}
	forBlock.mutTxsForBlock.Unlock()

	return false
}

func (bpp *basePreProcess) computeExistingAndMissing(
	body *block.Body,
	forBlock *txsForBlock,
	_ chan bool,
	currType block.Type,
	txPool dataRetriever.ShardedDataCacherNotifier,
) map[uint32][]*txsHashesInfo {

	searchFirst := currType == block.InvalidBlock
	missingTxsForShard := make(map[uint32][]*txsHashesInfo, len(body.MiniBlocks))
	if body == nil {
		return missingTxsForShard
	}

	txHashes := make([][]byte, 0, initialTxHashesSliceLen)
	forBlock.mutTxsForBlock.Lock()

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.Type != currType {
			continue
		}

		txShardInfoObject := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx, err := process.GetTransactionHandlerFromPool(
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
				txHash,
				txPool,
				searchFirst)

			if err != nil {
				txHashes = append(txHashes, txHash)
				forBlock.missingTxs++
				continue
			}

			forBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfoObject}
		}

		if len(txHashes) > 0 {
			tmp := &txsHashesInfo{
				txHashes:        sliceUtil.TrimSliceSliceByte(txHashes),
				receiverShardID: miniBlock.ReceiverShardID,
			}
			missingTxsForShard[miniBlock.SenderShardID] = append(missingTxsForShard[miniBlock.SenderShardID], tmp)
		}
		txHashes = txHashes[:0]
	}
	forBlock.mutTxsForBlock.Unlock()

	bpp.displayMissingTransactions(missingTxsForShard, currType)

	return missingTxsForShard
}

func (bpp *basePreProcess) displayMissingTransactions(
	missingTxsFromShard map[uint32][]*txsHashesInfo,
	currType block.Type,
) {

	for shard, txHashInfoSlice := range missingTxsFromShard {
		for _, txHashInfo := range txHashInfoSlice {
			for _, hash := range txHashInfo.txHashes {
				log.Trace("missing txs",
					"block type", currType.String(),
					"sender shard id", shard,
					"receiver shard id", txHashInfo.receiverShardID,
					"hash", hash,
				)
			}
		}
	}
}

func (bpp *basePreProcess) isTxAlreadyProcessed(txHash []byte, forBlock *txsForBlock) bool {
	forBlock.mutTxsForBlock.RLock()
	_, txAlreadyProcessed := forBlock.txHashAndInfo[string(txHash)]
	forBlock.mutTxsForBlock.RUnlock()

	return txAlreadyProcessed
}

func (bpp *basePreProcess) computeGasConsumed(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
	gasConsumedByMiniBlockInSenderShard *uint64,
	gasConsumedByMiniBlockInReceiverShard *uint64,
) error {

	gasConsumedByTxInSenderShard, gasConsumedByTxInReceiverShard, err := bpp.computeGasConsumedByTx(
		senderShardId,
		receiverShardId,
		tx,
		txHash)
	if err != nil {
		return err
	}

	gasConsumedByTxInSelfShard := uint64(0)
	if bpp.shardCoordinator.SelfId() == senderShardId {
		gasConsumedByTxInSelfShard = gasConsumedByTxInSenderShard

		if *gasConsumedByMiniBlockInReceiverShard+gasConsumedByTxInReceiverShard > bpp.economicsFee.MaxGasLimitPerBlock() {
			return process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
		}
	} else {
		gasConsumedByTxInSelfShard = gasConsumedByTxInReceiverShard

		if *gasConsumedByMiniBlockInSenderShard+gasConsumedByTxInSenderShard > bpp.economicsFee.MaxGasLimitPerBlock() {
			return process.ErrMaxGasLimitPerMiniBlockInSenderShardIsReached
		}
	}

	if bpp.gasHandler.TotalGasConsumed()+gasConsumedByTxInSelfShard > bpp.economicsFee.MaxGasLimitPerBlock() {
		return process.ErrMaxGasLimitPerBlockInSelfShardIsReached
	}

	*gasConsumedByMiniBlockInSenderShard += gasConsumedByTxInSenderShard
	*gasConsumedByMiniBlockInReceiverShard += gasConsumedByTxInReceiverShard
	bpp.gasHandler.SetGasConsumed(gasConsumedByTxInSelfShard, txHash)

	return nil
}

func (bpp *basePreProcess) computeGasConsumedByTx(
	senderShardId uint32,
	receiverShardId uint32,
	tx data.TransactionHandler,
	txHash []byte,
) (uint64, uint64, error) {

	txGasLimitInSenderShard, txGasLimitInReceiverShard, err := bpp.gasHandler.ComputeGasConsumedByTx(
		senderShardId,
		receiverShardId,
		tx)
	if err != nil {
		return 0, 0, err
	}

	if core.IsSmartContractAddress(tx.GetRecvAddress()) {
		txGasRefunded := bpp.gasHandler.GasRefunded(txHash)

		if txGasLimitInReceiverShard < txGasRefunded {
			return 0, 0, process.ErrInsufficientGasLimitInTx
		}

		txGasLimitInReceiverShard -= txGasRefunded

		if senderShardId == receiverShardId {
			txGasLimitInSenderShard -= txGasRefunded
		}
	}

	return txGasLimitInSenderShard, txGasLimitInReceiverShard, nil
}
