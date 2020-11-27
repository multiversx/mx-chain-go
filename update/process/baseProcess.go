package process

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

type baseProcessor struct {
	hasher             hashing.Hasher
	importHandler      update.ImportHandler
	marshalizer        marshal.Marshalizer
	pendingTxProcessor update.PendingTransactionProcessor
	shardCoordinator   sharding.Coordinator
	storage            dataRetriever.StorageService
	txCoordinator      process.TransactionCoordinator
	selfShardID        uint32
}

// CreatePostMiniBlocks will create all the post miniBlocks from the given miniBlocks info
func (b *baseProcessor) CreatePostMiniBlocks(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
	b.txCoordinator.CreateBlockStarted()

	body := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}

	for _, mbInfo := range mbsInfo {
		if mbInfo.ReceiverShardID != b.shardCoordinator.SelfId() {
			continue
		}

		miniBlock, err := b.pendingTxProcessor.ProcessTransactionsDstMe(mbInfo)
		if err != nil {
			return nil, nil, err
		}

		body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	}

	postProcessMiniBlocks := b.txCoordinator.CreatePostProcessMiniBlocks()
	postProcessMiniBlocksInfo := make([]*update.MbInfo, len(postProcessMiniBlocks))
	for index, postProcessMiniBlock := range postProcessMiniBlocks {
		mbHash, errCalculateHash := core.CalculateHash(b.marshalizer, b.hasher, postProcessMiniBlock)
		if errCalculateHash != nil {
			return nil, nil, errCalculateHash
		}

		body.MiniBlocks = append(body.MiniBlocks, postProcessMiniBlock)

		postProcessMiniBlockInfo, errCreate := b.createMiniBlockInfoForPostProcessMiniBlock(mbHash, postProcessMiniBlock)
		if errCreate != nil {
			return nil, nil, errCreate
		}

		postProcessMiniBlocksInfo[index] = postProcessMiniBlockInfo
	}

	_, err := b.pendingTxProcessor.Commit()
	if err != nil {
		return nil, nil, err
	}

	return body, postProcessMiniBlocksInfo, nil
}

func (b *baseProcessor) createMiniBlockInfoForPostProcessMiniBlock(
	mbHash []byte,
	mb *block.MiniBlock,
) (*update.MbInfo, error) {
	mapHashTx := b.txCoordinator.GetAllCurrentUsedTxs(mb.Type)
	txsInfo := make([]*update.TxInfo, len(mb.TxHashes))
	for index, txHash := range mb.TxHashes {
		tx, transactionFound := mapHashTx[string(txHash)]
		if !transactionFound {
			return nil, update.ErrPostProcessTransactionNotFound
		}

		txInfo := &update.TxInfo{
			TxHash: txHash,
			Tx:     tx,
		}

		txsInfo[index] = txInfo
	}

	return &update.MbInfo{
		MbHash:          mbHash,
		SenderShardID:   mb.SenderShardID,
		ReceiverShardID: mb.ReceiverShardID,
		Type:            mb.Type,
		TxsInfo:         txsInfo,
	}, nil
}

func (b *baseProcessor) createMiniBlockHeaders(body *block.Body) (int, []block.MiniBlockHeader, error) {
	if len(body.MiniBlocks) == 0 {
		return 0, nil, nil
	}

	totalTxCount := 0
	miniBlockHeaders := make([]block.MiniBlockHeader, len(body.MiniBlocks))

	for i := 0; i < len(body.MiniBlocks); i++ {
		txCount := len(body.MiniBlocks[i].TxHashes)
		totalTxCount += txCount

		miniBlockHash, err := core.CalculateHash(b.marshalizer, b.hasher, body.MiniBlocks[i])
		if err != nil {
			return 0, nil, err
		}

		miniBlockHeaders[i] = block.MiniBlockHeader{
			Hash:            miniBlockHash,
			SenderShardID:   body.MiniBlocks[i].SenderShardID,
			ReceiverShardID: body.MiniBlocks[i].ReceiverShardID,
			TxCount:         uint32(txCount),
			Type:            body.MiniBlocks[i].Type,
		}
	}

	return totalTxCount, miniBlockHeaders, nil
}

func (b *baseProcessor) saveAllBlockDataToStorageForSelfShard(
	headerHandler data.HeaderHandler,
	body *block.Body,
) {
	if check.IfNil(headerHandler) {
		log.Warn("SaveAllBlockDataToStorageForSelfShard", "error", update.ErrNilHeaderHandler)
		return
	}
	if body == nil {
		log.Warn("SaveAllBlockDataToStorageForSelfShard", "error", update.ErrNilBlockBody)
		return
	}
	if headerHandler.GetShardID() != b.selfShardID {
		return
	}

	miniBlockHeadersHashes := headerHandler.GetMiniBlockHeadersHashes()
	mapBlockTypesTxs := make(map[block.Type]map[string]data.TransactionHandler)
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if _, ok := mapBlockTypesTxs[miniBlock.Type]; !ok {
			mapBlockTypesTxs[miniBlock.Type] = b.txCoordinator.GetAllCurrentUsedTxs(miniBlock.Type)
		}

		marshalizedMiniBlock, errNotCritical := b.marshalizer.Marshal(miniBlock)
		if errNotCritical != nil {
			log.Warn("SaveAllBlockDataToStorageForSelfShard.Marshal -> MiniBlock",
				"error", errNotCritical.Error())
			continue
		}

		errNotCritical = b.storage.Put(dataRetriever.MiniBlockUnit, miniBlockHeadersHashes[i], marshalizedMiniBlock)
		if errNotCritical != nil {
			log.Warn("SaveAllBlockDataToStorageForSelfShard.Put -> MiniBlockUnit",
				"error", errNotCritical.Error())
		}
	}

	marshalizedReceipts, errNotCritical := b.txCoordinator.CreateMarshalizedReceipts()
	if errNotCritical != nil {
		log.Warn("SaveAllBlockDataToStorageForSelfShard.CreateMarshalizedReceipts",
			"error", errNotCritical.Error())
	} else {
		if len(marshalizedReceipts) > 0 {
			errNotCritical = b.storage.Put(dataRetriever.ReceiptsUnit, headerHandler.GetReceiptsHash(), marshalizedReceipts)
			if errNotCritical != nil {
				log.Warn("SaveAllBlockDataToStorageForSelfShard.Put -> ReceiptsUnit",
					"error", errNotCritical.Error())
			}
		}
	}

	mapTxs := b.importHandler.GetTransactions()
	for _, miniBlock := range body.MiniBlocks {
		for _, txHash := range miniBlock.TxHashes {
			tx, ok := mapTxs[string(txHash)]
			if !ok {
				log.Warn("SaveAllBlockDataToStorageForSelfShard",
					"error", update.ErrTransactionNotFoundInImportedMap)
				continue
			}

			unitType := getUnitTypeFromMiniBlockType(miniBlock.Type)
			marshaledData, errNotCritical := b.marshalizer.Marshal(tx)
			if errNotCritical != nil {
				log.Warn("SaveAllBlockDataToStorageForSelfShard.Marshal -> Transaction",
					"error", errNotCritical.Error())
				continue
			}

			errNotCritical = b.storage.Put(unitType, txHash, marshaledData)
			if errNotCritical != nil {
				log.Warn("SaveAllBlockDataToStorageForSelfShard.Put -> Transaction",
					"error", errNotCritical.Error())
			}
		}
	}
}

func getUnitTypeFromMiniBlockType(mbType block.Type) dataRetriever.UnitType {
	unitType := dataRetriever.TransactionUnit
	switch mbType {
	case block.TxBlock:
		unitType = dataRetriever.TransactionUnit
	case block.RewardsBlock:
		unitType = dataRetriever.RewardTransactionUnit
	case block.SmartContractResultBlock:
		unitType = dataRetriever.UnsignedTransactionUnit
	}

	return unitType
}

func checkBlockCreatorAfterHardForkNilParameters(
	hasher hashing.Hasher,
	importHandler update.ImportHandler,
	marshalizer marshal.Marshalizer,
	pendingTxProcessor update.PendingTransactionProcessor,
	shardCoordinator sharding.Coordinator,
	storage dataRetriever.StorageService,
	txCoordinator process.TransactionCoordinator,
) error {
	if check.IfNil(hasher) {
		return update.ErrNilHasher
	}
	if check.IfNil(importHandler) {
		return update.ErrNilImportHandler
	}
	if check.IfNil(marshalizer) {
		return update.ErrNilMarshalizer
	}
	if check.IfNil(pendingTxProcessor) {
		return update.ErrNilPendingTxProcessor
	}
	if check.IfNil(shardCoordinator) {
		return update.ErrNilShardCoordinator
	}
	if check.IfNil(storage) {
		return update.ErrNilStorage
	}
	if check.IfNil(txCoordinator) {
		return update.ErrNilTxCoordinator
	}

	return nil
}
