package blockAPI

import (
	"encoding/hex"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// BlockStatus is the status of a block
type BlockStatus string

const (
	// BlockStatusOnChain represents the identifier for an on-chain block
	BlockStatusOnChain = "on-chain"
	// BlockStatusReverted represent the identifier for a reverted block
	BlockStatusReverted = "reverted"
)

type baseAPIBockProcessor struct {
	hasDbLookupExtensions    bool
	selfShardID              uint32
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	historyRepo              dblookupext.HistoryRepository
	unmarshalTx              func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	txStatusComputer		 transaction.StatusComputerHandler
}

var log = logger.GetOrCreate("node/blockAPI")

func (bap *baseAPIBockProcessor) getTxsByMb(mbHeader *block.MiniBlockHeader, epoch uint32) []*transaction.ApiTransactionResult {
	miniblockHash := mbHeader.Hash
	mbBytes, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, miniblockHash, epoch)
	if err != nil {
		log.Warn("cannot get miniblock from storage",
			"hash", hex.EncodeToString(miniblockHash),
			"error", err.Error())
		return nil
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, mbBytes)
	if err != nil {
		log.Warn("cannot unmarshal miniblock",
			"hash", hex.EncodeToString(miniblockHash),
			"error", err.Error())
		return nil
	}

	switch miniBlock.Type {
	case block.TxBlock:
		return bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeNormal, dataRetriever.TransactionUnit)
	case block.RewardsBlock:
		return bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeReward, dataRetriever.RewardTransactionUnit)
	case block.SmartContractResultBlock:
		return bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeUnsigned, dataRetriever.UnsignedTransactionUnit)
	case block.InvalidBlock:
		return bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeInvalid, dataRetriever.TransactionUnit)
	default:
		return nil
	}
}

func (bap *baseAPIBockProcessor) getTxsFromMiniblock(
	miniblock *block.MiniBlock,
	miniblockHash []byte,
	epoch uint32,
	txType transaction.TxType,
	unit dataRetriever.UnitType,
) []*transaction.ApiTransactionResult {
	storer := bap.store.GetStorer(unit)
	start := time.Now()
	marshalizedTxs, err := storer.GetBulkFromEpoch(miniblock.TxHashes, epoch)
	if err != nil {
		log.Warn("cannot get from storage transactions",
			"error", err.Error())
		return []*transaction.ApiTransactionResult{}
	}
	log.Debug(fmt.Sprintf("GetBulkFromEpoch took %s", time.Since(start)))

	start = time.Now()
	txs := make([]*transaction.ApiTransactionResult, 0)
	for txHash, txBytes := range marshalizedTxs {
		tx, errUnmarshalTx := bap.unmarshalTx(txBytes, txType)
		if errUnmarshalTx != nil {
			log.Warn("cannot unmarshal transaction",
				"hash", hex.EncodeToString([]byte(txHash)),
				"error", errUnmarshalTx.Error())
			continue
		}
		tx.Hash = hex.EncodeToString([]byte(txHash))
		tx.MiniBlockType = miniblock.Type.String()
		tx.MiniBlockHash = hex.EncodeToString(miniblockHash)
		tx.SourceShard = miniblock.SenderShardID
		tx.DestinationShard = miniblock.ReceiverShardID

		// TODO : should check if tx is reward reverted
		tx.Status, _ = bap.txStatusComputer.ComputeStatusWhenInStorageKnowingMiniblock(miniblock.Type, tx)

		txs = append(txs, tx)
	}
	log.Debug(fmt.Sprintf("UnmarshalTransactions took %s", time.Since(start)))

	return txs
}

func (bap *baseAPIBockProcessor) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
	if !bap.hasDbLookupExtensions {
		return bap.store.Get(unit, key)
	}

	epoch, err := bap.historyRepo.GetEpochByHash(key)
	if err != nil {
		return nil, err
	}

	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBockProcessor) getFromStorerWithEpoch(unit dataRetriever.UnitType, key []byte, epoch uint32) ([]byte, error) {
	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBockProcessor) computeBlockStatus(storerUnit dataRetriever.UnitType, blockAPI *api.Block) (string, error) {
	nonceToByteSlice := bap.uint64ByteSliceConverter.ToByteSlice(blockAPI.Nonce)
	headerHash, err := bap.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return "", err
	}

	if hex.EncodeToString(headerHash) != blockAPI.Hash {
		return BlockStatusReverted, err
	}

	return BlockStatusOnChain, nil
}

func (bap *baseAPIBockProcessor) computeStatusAndPutInBlock(blockAPI *api.Block, storerUnit dataRetriever.UnitType) (*api.Block, error) {
	blockStatus, err := bap.computeBlockStatus(storerUnit, blockAPI)
	if err != nil {
		return nil, err
	}

	blockAPI.Status = blockStatus

	return blockAPI, nil
}
