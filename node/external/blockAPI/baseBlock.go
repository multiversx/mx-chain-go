package blockAPI

import (
	"bytes"
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/api/shared/logging"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
)

// BlockStatus is the status of a block
type BlockStatus string

const (
	// BlockStatusOnChain represents the identifier for an on-chain block
	BlockStatusOnChain = "on-chain"
	// BlockStatusReverted represent the identifier for a reverted block
	BlockStatusReverted = "reverted"
)

type baseAPIBlockProcessor struct {
	hasDbLookupExtensions    bool
	selfShardID              uint32
	emptyReceiptsHash        []byte
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	historyRepo              dblookupext.HistoryRepository
	hasher                   hashing.Hasher
	addressPubKeyConverter   core.PubkeyConverter
	txStatusComputer         transaction.StatusComputerHandler
	txUnmarshaller           TransactionUnmarshaller
}

var log = logger.GetOrCreate("node/blockAPI")

func (bap *baseAPIBlockProcessor) getIntraMiniblocks(receiptsHash []byte, epoch uint32, withTxs bool) []*api.MiniBlock {
	if bytes.Equal(bap.emptyReceiptsHash, receiptsHash) {
		return nil
	}

	batchBytes, err := bap.getFromStorerWithEpoch(dataRetriever.ReceiptsUnit, receiptsHash, epoch)
	if err != nil {
		log.Warn("cannot get miniblock from receipts storage",
			"hash", receiptsHash,
			"error", err.Error())
		return nil
	}

	batchWithMbs := &batch.Batch{}
	err = bap.marshalizer.Unmarshal(batchWithMbs, batchBytes)
	if err != nil {
		log.Warn("cannot unmarshal batch",
			"hash", receiptsHash,
			"error", err.Error())
		return nil
	}

	return bap.extractMbsFromBatch(batchWithMbs, epoch, withTxs)
}

func (bap *baseAPIBlockProcessor) extractMbsFromBatch(batchWithMbs *batch.Batch, epoch uint32, withTxs bool) []*api.MiniBlock {
	mbs := make([]*api.MiniBlock, 0)
	for _, mbBytes := range batchWithMbs.Data {
		miniBlock := &block.MiniBlock{}
		err := bap.marshalizer.Unmarshal(miniBlock, mbBytes)
		if err != nil {
			continue
		}

		miniblockAPI, ok := bap.prepareAPIMiniblock(miniBlock, epoch, withTxs)
		if !ok {
			continue
		}

		mbs = append(mbs, miniblockAPI)
	}

	return mbs
}

func (bap *baseAPIBlockProcessor) prepareAPIMiniblock(miniblock *block.MiniBlock, epoch uint32, withTxs bool) (*api.MiniBlock, bool) {
	mbHash, err := core.CalculateHash(bap.marshalizer, bap.hasher, miniblock)
	if err != nil {
		log.Warn("cannot compute miniblock's hash", "error", err.Error())
		return nil, false
	}

	miniblockAPI := &api.MiniBlock{
		Hash:             hex.EncodeToString(mbHash),
		Type:             miniblock.Type.String(),
		SourceShard:      miniblock.SenderShardID,
		DestinationShard: miniblock.ReceiverShardID,
	}
	if withTxs {
		bap.getAndAttachTxsToMbByEpoch(mbHash, miniblock, epoch, miniblockAPI)
	}

	return miniblockAPI, true
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMb(mbHeader data.MiniBlockHeaderHandler, epoch uint32, apiMiniblock *api.MiniBlock) {
	miniblockHash := mbHeader.GetHash()
	mbBytes, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, miniblockHash, epoch)
	if err != nil {
		log.Warn("cannot get miniblock from storage",
			"hash", hex.EncodeToString(miniblockHash),
			"error", err.Error())
		return
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, mbBytes)
	if err != nil {
		log.Warn("cannot unmarshal miniblock",
			"hash", miniblockHash,
			"error", err.Error())
		return
	}

	bap.getAndAttachTxsToMbByEpoch(miniblockHash, miniBlock, epoch, apiMiniblock)
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMbByEpoch(miniblockHash []byte, miniBlock *block.MiniBlock, epoch uint32, apiMiniblock *api.MiniBlock) {
	switch miniBlock.Type {
	case block.TxBlock:
		apiMiniblock.Transactions = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeNormal, dataRetriever.TransactionUnit)
	case block.RewardsBlock:
		apiMiniblock.Transactions = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeReward, dataRetriever.RewardTransactionUnit)
	case block.SmartContractResultBlock:
		apiMiniblock.Transactions = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeUnsigned, dataRetriever.UnsignedTransactionUnit)
	case block.InvalidBlock:
		apiMiniblock.Transactions = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeInvalid, dataRetriever.TransactionUnit)
	case block.ReceiptBlock:
		apiMiniblock.Receipts = bap.getReceiptsFromMiniblock(miniBlock, epoch)
	default:
		return
	}
}

func (bap *baseAPIBlockProcessor) getReceiptsFromMiniblock(miniblock *block.MiniBlock, epoch uint32) []*transaction.ApiReceipt {
	storer := bap.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	start := time.Now()
	marshalizedReceipts, err := storer.GetBulkFromEpoch(miniblock.TxHashes, epoch)
	if err != nil {
		log.Warn("cannot get receipts from storage", "error", err.Error())
		return []*transaction.ApiReceipt{}
	}
	logging.LogAPIActionDurationIfNeeded(start, "GetBulkFromEpoch")

	apiReceipts := make([]*transaction.ApiReceipt, 0)
	for recHash, recBytes := range marshalizedReceipts {
		receipt, errUnmarshal := bap.txUnmarshaller.UnmarshalReceipt(recBytes)
		if errUnmarshal != nil {
			log.Warn("cannot unmarshal receipt",
				"hash", []byte(recHash),
				"error", errUnmarshal.Error())
			continue
		}

		apiReceipts = append(apiReceipts, receipt)
	}

	return apiReceipts
}

func (bap *baseAPIBlockProcessor) getTxsFromMiniblock(
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
	logging.LogAPIActionDurationIfNeeded(start, "GetBulkFromEpoch")

	start = time.Now()
	txs := make([]*transaction.ApiTransactionResult, 0)
	for txHash, txBytes := range marshalizedTxs {
		tx, errUnmarshalTx := bap.txUnmarshaller.UnmarshalTransaction(txBytes, txType)
		if errUnmarshalTx != nil {
			log.Warn("cannot unmarshal transaction",
				"hash", []byte(txHash),
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
	logging.LogAPIActionDurationIfNeeded(start, "UnmarshalTransactions")

	return txs
}

func (bap *baseAPIBlockProcessor) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
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

func (bap *baseAPIBlockProcessor) getFromStorerWithEpoch(unit dataRetriever.UnitType, key []byte, epoch uint32) ([]byte, error) {
	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBlockProcessor) computeBlockStatus(storerUnit dataRetriever.UnitType, blockAPI *api.Block) (string, error) {
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

func (bap *baseAPIBlockProcessor) computeStatusAndPutInBlock(blockAPI *api.Block, storerUnit dataRetriever.UnitType) (*api.Block, error) {
	blockStatus, err := bap.computeBlockStatus(storerUnit, blockAPI)
	if err != nil {
		return nil, err
	}

	blockAPI.Status = blockStatus

	return blockAPI, nil
}

func (bap *baseAPIBlockProcessor) getBlockHeaderHashAndBytesByRound(round uint64, blockUnitType dataRetriever.UnitType) (headerHash []byte, blockBytes []byte, err error) {
	roundToByteSlice := bap.uint64ByteSliceConverter.ToByteSlice(round)
	headerHash, err = bap.store.Get(dataRetriever.RoundHdrHashDataUnit, roundToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	blockBytes, err = bap.getFromStorer(blockUnitType, headerHash)
	if err != nil {
		return nil, nil, err
	}

	return headerHash, blockBytes, nil
}
