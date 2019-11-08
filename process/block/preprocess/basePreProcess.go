package preprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
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
}

func (bpp *basePreProcess) removeDataFromPools(body block.Body, miniBlockPool storage.Cacher, txPool dataRetriever.ShardedDataCacherNotifier, mbType block.Type) error {
	if miniBlockPool == nil || miniBlockPool.IsInterfaceNil() {
		return process.ErrNilMiniBlockPool
	}
	if txPool == nil || txPool.IsInterfaceNil() {
		return process.ErrNilTransactionPool
	}

	for i := 0; i < len(body); i++ {
		currentMiniBlock := body[i]
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
	mrsTxs := make([][]byte, 0)
	for _, txHash := range txHashes {
		forBlock.mutTxsForBlock.RLock()
		txInfo := forBlock.txHashAndInfo[string(txHash)]
		forBlock.mutTxsForBlock.RUnlock()

		if txInfo == nil || txInfo.tx == nil {
			continue
		}

		txMrs, err := bpp.marshalizer.Marshal(txInfo.tx)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsTxs = append(mrsTxs, txMrs)
	}

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
		txInfo := forBlock.txHashAndInfo[string(txHash)]
		forBlock.mutTxsForBlock.RUnlock()

		if txInfo == nil || txInfo.tx == nil {
			return process.ErrMissingTransaction
		}

		buff, err := bpp.marshalizer.Marshal(txInfo.tx)
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
) bool {
	forBlock.mutTxsForBlock.Lock()

	if forBlock.missingTxs > 0 {
		txInfoForHash := forBlock.txHashAndInfo[string(txHash)]
		if txInfoForHash != nil && txInfoForHash.txShardInfo != nil &&
			(txInfoForHash.tx == nil || txInfoForHash.tx.IsInterfaceNil()) {
			tx, _ := process.GetTransactionHandlerFromPool(
				txInfoForHash.senderShardID,
				txInfoForHash.receiverShardID,
				txHash,
				txPool)

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
	body block.Body,
	forBlock *txsForBlock,
	chRcvAllTxs chan bool,
	currType block.Type,
	txPool dataRetriever.ShardedDataCacherNotifier,
) map[uint32][]*txsHashesInfo {

	missingTxsForShard := make(map[uint32][]*txsHashesInfo, 0)

	forBlock.mutTxsForBlock.Lock()
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != currType {
			continue
		}

		txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}
		txHashes := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx, err := process.GetTransactionHandlerFromPool(
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
				txHash,
				txPool)

			if err != nil {
				txHashes = append(txHashes, txHash)
				forBlock.missingTxs++
				continue
			}

			forBlock.txHashAndInfo[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfo}
		}

		if len(txHashes) > 0 {
			missingTxsForShard[miniBlock.SenderShardID] = append(missingTxsForShard[miniBlock.SenderShardID],
				&txsHashesInfo{txHashes: txHashes, receiverShardID: miniBlock.ReceiverShardID})
		}
	}
	forBlock.mutTxsForBlock.Unlock()

	return missingTxsForShard
}

func (bpp *basePreProcess) isTxAlreadyProcessed(txHash []byte, forBlock *txsForBlock) bool {
	forBlock.mutTxsForBlock.RLock()
	_, txAlreadyProcessed := forBlock.txHashAndInfo[string(txHash)]
	forBlock.mutTxsForBlock.RUnlock()

	return txAlreadyProcessed
}
