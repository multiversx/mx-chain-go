package preprocess

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"sync"
)

type txShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type txInfo struct {
	tx data.TransactionHandler
	*txShardInfo
	has bool
}

type txsHashesInfo struct {
	txHashes        [][]byte
	receiverShardID uint32
}

type basePreProcess struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
}

func (bpp *basePreProcess) removeDataFromPools(body block.Body, miniBlockPool storage.Cacher, txPool dataRetriever.ShardedDataCacherNotifier, mbType block.Type) error {
	if miniBlockPool == nil {
		return process.ErrNilMiniBlockPool
	}
	if txPool == nil {
		return process.ErrNilTransactionPool
	}

	for i := 0; i < len(body); i++ {
		currentMiniBlock := body[i]
		if currentMiniBlock.Type != mbType {
			continue
		}

		strCache := process.ShardCacherIdentifier(currentMiniBlock.SenderShardID, currentMiniBlock.ReceiverShardID)
		txPool.RemoveSetOfDataFromPool(currentMiniBlock.TxHashes, strCache)

		buff, err := bpp.marshalizer.Marshal(currentMiniBlock)
		if err != nil {
			return err
		}

		miniBlockHash := bpp.hasher.Compute(string(buff))
		miniBlockPool.Remove(miniBlockHash)
	}

	return nil
}

func (bpp *basePreProcess) restoreMiniBlock(miniBlock *block.MiniBlock, miniBlockPool storage.Cacher, restoredHash []byte) error {
	miniBlockHash, err := core.CalculateHash(bpp.marshalizer, bpp.hasher, miniBlock)
	if err != nil {
		return err
	}

	miniBlockPool.Put(miniBlockHash, miniBlock)
	if miniBlock.SenderShardID != bpp.shardCoordinator.SelfId() {
		restoredHash = miniBlockHash
	}

	return err
}

func (bpp *basePreProcess) createMarshalizedData(txHashes [][]byte, mutForBlock *sync.RWMutex, currBlock map[string]*txInfo) ([][]byte, error) {
	mrsScrs := make([][]byte, 0)
	for _, txHash := range txHashes {
		mutForBlock.RLock()
		txInfo := currBlock[string(txHash)]
		mutForBlock.RUnlock()

		if txInfo == nil || txInfo.tx == nil {
			continue
		}

		txMrs, err := bpp.marshalizer.Marshal(txInfo.tx)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsScrs = append(mrsScrs, txMrs)
	}

	return mrsScrs, nil
}

func (bpp *basePreProcess) saveTxsToStorage(
	txHashes [][]byte,
	mutForBlock *sync.RWMutex,
	currBlock map[string]*txInfo,
	store dataRetriever.StorageService,
	dataUnit dataRetriever.UnitType,
) error {

	for j := 0; j < len(txHashes); j++ {
		txHash := txHashes[j]

		mutForBlock.RLock()
		txInfo := currBlock[string(txHash)]
		mutForBlock.RUnlock()

		if txInfo == nil || txInfo.tx == nil {
			return process.ErrMissingTransaction
		}

		buff, err := bpp.marshalizer.Marshal(txInfo.tx)
		if err != nil {
			return err
		}

		errNotCritical := store.Put(dataUnit, txHash, buff)
		if errNotCritical != nil {
			log.Error(errNotCritical.Error())
		}
	}

	return nil
}

func (bpp *basePreProcess) baseReceivedTransaction(
	txHash []byte,
	mutTxsForBlock *sync.RWMutex,
	missingTxs *int,
	txsForBlock map[string]*txInfo,
	txPool dataRetriever.ShardedDataCacherNotifier,
) bool {
	currMissingTxs := 0

	mutTxsForBlock.Lock()
	if *missingTxs > 0 {
		txInfoForHash := txsForBlock[string(txHash)]
		if txInfoForHash != nil &&
			txInfoForHash.txShardInfo != nil &&
			!txInfoForHash.has {
			tx := bpp.getTransactionFromPool(txInfoForHash.senderShardID, txInfoForHash.receiverShardID, txHash, txPool)
			if tx != nil {
				txsForBlock[string(txHash)].tx = tx
				txsForBlock[string(txHash)].has = true
				*missingTxs--
			}
		}

		currMissingTxs = *missingTxs
	}
	mutTxsForBlock.Unlock()

	return currMissingTxs == 0
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (bpp *basePreProcess) getTransactionFromPool(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
	txPool dataRetriever.ShardedDataCacherNotifier,
) data.TransactionHandler {
	strCache := process.ShardCacherIdentifier(senderShardID, destShardID)
	txStore := txPool.ShardDataStore(strCache)
	if txStore == nil {
		log.Error(process.ErrNilStorage.Error())
		return nil
	}

	val, ok := txStore.Peek(txHash)
	if !ok {
		log.Debug(process.ErrTxNotFound.Error())
		return nil
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		log.Error(process.ErrInvalidTxInPool.Error())
		return nil
	}

	return tx
}
