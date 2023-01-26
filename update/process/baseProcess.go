package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/logging"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/update"
)

type baseProcessor struct {
	hasher             hashing.Hasher
	importHandler      update.ImportHandler
	marshalizer        marshal.Marshalizer
	pendingTxProcessor update.PendingTransactionProcessor
	shardCoordinator   sharding.Coordinator
	storage            dataRetriever.StorageService
	txCoordinator      process.TransactionCoordinator
	receiptsRepository receiptsRepository
	selfShardID        uint32
}

// CreateBody will create a block body after hardfork import
func (b *baseProcessor) CreateBody() (*block.Body, []*update.MbInfo, error) {
	mbsInfo, err := b.getPendingMbsAndTxsInCorrectOrder()
	if err != nil {
		return nil, nil, err
	}

	return b.CreatePostMiniBlocks(mbsInfo)
}

func (b *baseProcessor) getPendingMbsAndTxsInCorrectOrder() ([]*update.MbInfo, error) {
	hardForkMetaBlock := b.importHandler.GetHardForkMetaBlock()
	unFinishedMetaBlocks := b.importHandler.GetUnFinishedMetaBlocks()
	pendingMiniBlocks, err := update.GetPendingMiniBlocks(hardForkMetaBlock, unFinishedMetaBlocks)
	if err != nil {
		return nil, err
	}

	importedMiniBlocksMap := b.importHandler.GetMiniBlocks()
	if len(importedMiniBlocksMap) != len(pendingMiniBlocks) {
		return nil, update.ErrWrongImportedMiniBlocksMap
	}

	numPendingTransactions := 0
	importedTransactionsMap := b.importHandler.GetTransactions()
	mbsInfo := make([]*update.MbInfo, len(pendingMiniBlocks))

	for mbIndex, pendingMiniBlock := range pendingMiniBlocks {
		miniBlock, miniBlockFound := importedMiniBlocksMap[string(pendingMiniBlock.GetHash())]
		if !miniBlockFound {
			return nil, update.ErrMiniBlockNotFoundInImportedMap
		}

		txsInfo, errGetTxsInfoFromMiniBlock := b.getTxsInfoFromMiniBlock(miniBlock, importedTransactionsMap)
		if errGetTxsInfoFromMiniBlock != nil {
			return nil, errGetTxsInfoFromMiniBlock
		}

		numPendingTransactions += len(miniBlock.TxHashes)
		mbsInfo[mbIndex] = &update.MbInfo{
			MbHash:          pendingMiniBlock.GetHash(),
			SenderShardID:   pendingMiniBlock.GetSenderShardID(),
			ReceiverShardID: pendingMiniBlock.GetReceiverShardID(),
			Type:            block.Type(pendingMiniBlock.GetTypeInt32()),
			TxsInfo:         txsInfo,
		}
	}

	if len(importedTransactionsMap) != numPendingTransactions {
		return nil, update.ErrWrongImportedTransactionsMap
	}

	return mbsInfo, nil
}

func (b *baseProcessor) getTxsInfoFromMiniBlock(
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) ([]*update.TxInfo, error) {
	txsInfo := make([]*update.TxInfo, len(miniBlock.TxHashes))
	for txIndex, txHash := range miniBlock.TxHashes {
		tx, transactionFound := mapHashTx[string(txHash)]
		if !transactionFound {
			return nil, update.ErrTransactionNotFoundInImportedMap
		}

		txsInfo[txIndex] = &update.TxInfo{
			TxHash: txHash,
			Tx:     tx,
		}
	}

	return txsInfo, nil
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
		log.Warn("saveAllBlockDataToStorageForSelfShard", "error", update.ErrNilHeaderHandler)
		return
	}
	if check.IfNil(body) {
		log.Warn("saveAllBlockDataToStorageForSelfShard", "error", update.ErrNilBlockBody)
		return
	}
	if headerHandler.GetShardID() != b.selfShardID {
		return
	}

	b.saveMiniBlocks(headerHandler, body)
	b.saveReceipts(headerHandler)
	b.saveTransactions(body)
}

func (b *baseProcessor) saveMiniBlocks(headerHandler data.HeaderHandler, body *block.Body) {
	miniBlockHeadersHashes := headerHandler.GetMiniBlockHeadersHashes()
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]

		marshalizedMiniBlock, errNotCritical := b.marshalizer.Marshal(miniBlock)
		if errNotCritical != nil {
			log.Warn("saveMiniBlocks.Marshal -> MiniBlock",
				"error", errNotCritical.Error())
			continue
		}

		errNotCritical = b.storage.Put(dataRetriever.MiniBlockUnit, miniBlockHeadersHashes[i], marshalizedMiniBlock)
		if errNotCritical != nil {
			logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
				"saveMiniBlocks.Put -> MiniBlockUnit",
				"err", errNotCritical)
		}
	}
}

func (b *baseProcessor) saveReceipts(headerHandler data.HeaderHandler) {
	headerHash, errNotCritical := core.CalculateHash(b.marshalizer, b.hasher, headerHandler)
	if errNotCritical != nil {
		log.Warn("saveReceipts(), error on CalculateHash(header)", "error", errNotCritical.Error())
		return
	}

	receiptsHolder := holders.NewReceiptsHolder(b.txCoordinator.GetCreatedInShardMiniBlocks())
	errNotCritical = b.receiptsRepository.SaveReceipts(receiptsHolder, headerHandler, headerHash)
	if errNotCritical != nil {
		logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
			"saveReceipts(), error on receiptsRepository.SaveReceipts()",
			"err", errNotCritical)
	}
}

func (b *baseProcessor) saveTransactions(body *block.Body) {
	mapTxs := b.importHandler.GetTransactions()
	for _, miniBlock := range body.MiniBlocks {
		for _, txHash := range miniBlock.TxHashes {
			tx, ok := mapTxs[string(txHash)]
			if !ok {
				log.Warn("saveTransactions",
					"error", update.ErrTransactionNotFoundInImportedMap)
				continue
			}

			unitType, errNotCritical := getUnitTypeFromMiniBlockType(miniBlock.Type)
			if errNotCritical != nil {
				log.Warn("saveTransactions.getUnitTypeFromMiniBlockType",
					"error", errNotCritical.Error())
				continue
			}

			marshaledData, errNotCritical := b.marshalizer.Marshal(tx)
			if errNotCritical != nil {
				log.Warn("saveTransactions.Marshal -> Transaction",
					"error", errNotCritical.Error())
				continue
			}

			errNotCritical = b.storage.Put(unitType, txHash, marshaledData)
			if errNotCritical != nil {
				logging.LogErrAsWarnExceptAsDebugIfClosingError(log, errNotCritical,
					"saveTransactions.Put -> Transaction",
					"err", errNotCritical)
			}
		}
	}
}

func getUnitTypeFromMiniBlockType(mbType block.Type) (dataRetriever.UnitType, error) {
	var err error
	unitType := dataRetriever.TransactionUnit
	switch mbType {
	case block.TxBlock:
		unitType = dataRetriever.TransactionUnit
	case block.RewardsBlock:
		unitType = dataRetriever.RewardTransactionUnit
	case block.SmartContractResultBlock:
		unitType = dataRetriever.UnsignedTransactionUnit
	default:
		err = update.ErrInvalidMiniBlockType
	}

	return unitType, err
}

func checkBlockCreatorAfterHardForkNilParameters(
	hasher hashing.Hasher,
	importHandler update.ImportHandler,
	marshalizer marshal.Marshalizer,
	pendingTxProcessor update.PendingTransactionProcessor,
	shardCoordinator sharding.Coordinator,
	storage dataRetriever.StorageService,
	txCoordinator process.TransactionCoordinator,
	receiptsRepository receiptsRepository,
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
	if check.IfNil(receiptsRepository) {
		return update.ErrNilReceiptsRepository
	}

	return nil
}
