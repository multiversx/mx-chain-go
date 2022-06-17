package blockAPI

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"

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
	logsFacade               LogsFacade
}

var log = logger.GetOrCreate("node/blockAPI")

func (bap *baseAPIBlockProcessor) getIntraMiniblocks(receiptsHash []byte, epoch uint32, options api.BlockQueryOptions) ([]*api.MiniBlock, error) {
	if bytes.Equal(bap.emptyReceiptsHash, receiptsHash) {
		return nil, nil
	}

	batchBytes, err := bap.getFromStorerWithEpoch(dataRetriever.ReceiptsUnit, receiptsHash, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w (receipts): %v, hash = %s", errCannotLoadMiniblocks, err, hex.EncodeToString(receiptsHash))
	}

	batchWithMbs := &batch.Batch{}
	err = bap.marshalizer.Unmarshal(batchWithMbs, batchBytes)
	if err != nil {
		return nil, fmt.Errorf("%w (receipts): %v, hash = %s", errCannotUnmarshalMiniblocks, err, hex.EncodeToString(receiptsHash))
	}

	return bap.extractMbsFromBatch(batchWithMbs, epoch, options)
}

func (bap *baseAPIBlockProcessor) extractMbsFromBatch(batchWithMbs *batch.Batch, epoch uint32, options api.BlockQueryOptions) ([]*api.MiniBlock, error) {
	mbs := make([]*api.MiniBlock, 0)
	for _, mbBytes := range batchWithMbs.Data {
		miniBlock := &block.MiniBlock{}
		err := bap.marshalizer.Unmarshal(miniBlock, mbBytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errCannotUnmarshalMiniblocks, err)
		}

		miniblockAPI, err := bap.prepareAPIMiniblock(miniBlock, epoch, options)
		if err != nil {
			return nil, err
		}

		mbs = append(mbs, miniblockAPI)
	}

	return mbs, nil
}

func (bap *baseAPIBlockProcessor) prepareAPIMiniblock(miniblock *block.MiniBlock, epoch uint32, options api.BlockQueryOptions) (*api.MiniBlock, error) {
	mbHash, err := core.CalculateHash(bap.marshalizer, bap.hasher, miniblock)
	if err != nil {
		return nil, err
	}

	miniblockAPI := &api.MiniBlock{
		Hash:             hex.EncodeToString(mbHash),
		Type:             miniblock.Type.String(),
		SourceShard:      miniblock.SenderShardID,
		DestinationShard: miniblock.ReceiverShardID,
		ProcessingType:   block.ProcessingType(miniblock.GetProcessingType()).String(),
		// TODO: Question for review: can / should we also return construction state?
	}

	if options.WithTransactions {
		err = bap.getAndAttachTxsToMbByEpoch(mbHash, miniblock, epoch, miniblockAPI, options)
		if err != nil {
			return nil, err
		}
	}

	return miniblockAPI, nil
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMb(mbHeader data.MiniBlockHeaderHandler, epoch uint32, apiMiniblock *api.MiniBlock, options api.BlockQueryOptions) error {
	miniblockHash := mbHeader.GetHash()
	mbBytes, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, miniblockHash, epoch)
	if err != nil {
		return fmt.Errorf("%w: %v, hash = %s", errCannotLoadMiniblocks, err, hex.EncodeToString(miniblockHash))
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, mbBytes)
	if err != nil {
		return fmt.Errorf("%w: %v, hash = %s", errCannotUnmarshalMiniblocks, err, hex.EncodeToString(miniblockHash))
	}

	return bap.getAndAttachTxsToMbByEpoch(miniblockHash, miniBlock, epoch, apiMiniblock, options)
}

func (bap *baseAPIBlockProcessor) getAndAttachTxsToMbByEpoch(miniblockHash []byte, miniBlock *block.MiniBlock, epoch uint32, apiMiniblock *api.MiniBlock, options api.BlockQueryOptions) error {
	var err error

	switch miniBlock.Type {
	case block.TxBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeNormal, dataRetriever.TransactionUnit)
	case block.RewardsBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeReward, dataRetriever.RewardTransactionUnit)
	case block.SmartContractResultBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeUnsigned, dataRetriever.UnsignedTransactionUnit)
	case block.InvalidBlock:
		apiMiniblock.Transactions, err = bap.getTxsFromMiniblock(miniBlock, miniblockHash, epoch, transaction.TxTypeInvalid, dataRetriever.TransactionUnit)
	case block.ReceiptBlock:
		apiMiniblock.Receipts, err = bap.getReceiptsFromMiniblock(miniBlock, epoch)
	}

	if err != nil {
		return err
	}

	if options.WithLogs {
		err := bap.logsFacade.IncludeLogsInTransactions(apiMiniblock.Transactions, miniBlock.TxHashes, epoch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bap *baseAPIBlockProcessor) getReceiptsFromMiniblock(miniblock *block.MiniBlock, epoch uint32) ([]*transaction.ApiReceipt, error) {
	storer := bap.store.GetStorer(dataRetriever.UnsignedTransactionUnit)
	start := time.Now()
	marshalledReceipts, err := storer.GetBulkFromEpoch(miniblock.TxHashes, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errCannotLoadReceipts, err)
	}
	log.Debug(fmt.Sprintf("GetBulkFromEpoch took %s", time.Since(start)))

	apiReceipts := make([]*transaction.ApiReceipt, 0)
	for _, pair := range marshalledReceipts {
		receipt, err := bap.txUnmarshaller.UnmarshalReceipt(pair.Value)
		if err != nil {
			return nil, fmt.Errorf("%w: %v, hash = %s", errCannotUnmarshalReceipts, err, hex.EncodeToString(pair.Key))
		}

		apiReceipts = append(apiReceipts, receipt)
	}

	return apiReceipts, nil
}

func (bap *baseAPIBlockProcessor) getTxsFromMiniblock(
	miniblock *block.MiniBlock,
	miniblockHash []byte,
	epoch uint32,
	txType transaction.TxType,
	unit dataRetriever.UnitType,
) ([]*transaction.ApiTransactionResult, error) {
	storer := bap.store.GetStorer(unit)
	start := time.Now()
	marshalledTxs, err := storer.GetBulkFromEpoch(miniblock.TxHashes, epoch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v, miniblock = %s", errCannotLoadTransactions, err, hex.EncodeToString(miniblockHash))
	}
	log.Debug(fmt.Sprintf("GetBulkFromEpoch took %s", time.Since(start)))

	start = time.Now()
	txs := make([]*transaction.ApiTransactionResult, 0)
	for _, pair := range marshalledTxs {
		tx, errUnmarshalTx := bap.txUnmarshaller.UnmarshalTransaction(epoch, pair.Value, txType)
		if errUnmarshalTx != nil {
			return nil, fmt.Errorf("%w: %v, miniblock = %s", errCannotUnmarshalTransactions, err, hex.EncodeToString(miniblockHash))
		}
		tx.Hash = hex.EncodeToString(pair.Key)
		tx.HashBytes = pair.Key
		tx.MiniBlockType = miniblock.Type.String()
		tx.MiniBlockHash = hex.EncodeToString(miniblockHash)
		tx.SourceShard = miniblock.SenderShardID
		tx.DestinationShard = miniblock.ReceiverShardID

		// TODO : should check if tx is reward reverted
		tx.Status, _ = bap.txStatusComputer.ComputeStatusWhenInStorageKnowingMiniblock(miniblock.Type, tx)

		txs = append(txs, tx)
	}
	log.Debug(fmt.Sprintf("UnmarshalTransactions took %s", time.Since(start)))

	return txs, nil
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
