package process

import (
	"math/big"

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

// ArgsNewShardBlockCreatorAfterHardFork defines the arguments structure to create a new shard block creator
type ArgsNewShardBlockCreatorAfterHardFork struct {
	ShardCoordinator   sharding.Coordinator
	TxCoordinator      process.TransactionCoordinator
	PendingTxProcessor update.PendingTransactionProcessor
	ImportHandler      update.ImportHandler
	Marshalizer        marshal.Marshalizer
	Hasher             hashing.Hasher
	SelfShardID        uint32
	DataPool           dataRetriever.PoolsHolder
	Storage            dataRetriever.StorageService
}

type shardBlockCreator struct {
	shardCoordinator   sharding.Coordinator
	txCoordinator      process.TransactionCoordinator
	pendingTxProcessor update.PendingTransactionProcessor
	importHandler      update.ImportHandler
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	selfShardID        uint32
	dataPool           dataRetriever.PoolsHolder
	storage            dataRetriever.StorageService
}

// NewShardBlockCreatorAfterHardFork creates a shard block processor for the first block after hardfork
func NewShardBlockCreatorAfterHardFork(args ArgsNewShardBlockCreatorAfterHardFork) (*shardBlockCreator, error) {
	if check.IfNil(args.ImportHandler) {
		return nil, update.ErrNilImportHandler
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.TxCoordinator) {
		return nil, update.ErrNilTxCoordinator
	}
	if check.IfNil(args.PendingTxProcessor) {
		return nil, update.ErrNilPendingTxProcessor
	}
	if check.IfNil(args.DataPool) {
		return nil, update.ErrNilDataPoolHolder
	}
	if check.IfNil(args.Storage) {
		return nil, update.ErrNilStorage
	}

	return &shardBlockCreator{
		shardCoordinator:   args.ShardCoordinator,
		txCoordinator:      args.TxCoordinator,
		pendingTxProcessor: args.PendingTxProcessor,
		importHandler:      args.ImportHandler,
		marshalizer:        args.Marshalizer,
		hasher:             args.Hasher,
		selfShardID:        args.SelfShardID,
		dataPool:           args.DataPool,
		storage:            args.Storage,
	}, nil
}

// CreateBody will create a block body after hardfork import
func (s *shardBlockCreator) CreateBody() (*block.Body, []*update.MbInfo, error) {
	mbsInfo, err := s.getPendingMbsAndTxsInCorrectOrder()
	if err != nil {
		return nil, nil, err
	}

	return s.CreatePostMiniBlocks(mbsInfo)
}

// CreateBlock will create a block after hardfork import
func (s *shardBlockCreator) CreateBlock(
	body *block.Body,
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (data.HeaderHandler, error) {
	if len(chainID) == 0 {
		return nil, update.ErrEmptyChainID
	}

	rootHash, err := s.pendingTxProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	shardHeader := &block.Header{
		Nonce:           nonce,
		ShardID:         s.shardCoordinator.SelfId(),
		Round:           round,
		Epoch:           epoch,
		ChainID:         []byte(chainID),
		SoftwareVersion: []byte(""),
		RootHash:        rootHash,
		RandSeed:        rootHash,
		PrevHash:        rootHash,
		PrevRandSeed:    rootHash,
		AccumulatedFees: big.NewInt(0),
		PubKeysBitmap:   []byte{1},
	}

	shardHeader.ReceiptsHash, err = s.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, err
	}

	totalTxCount, miniBlockHeaders, err := s.createMiniBlockHeaders(body)
	if err != nil {
		return nil, err
	}

	shardHeader.MiniBlockHeaders = miniBlockHeaders
	shardHeader.TxCount = uint32(totalTxCount)

	metaBlockHash, err := core.CalculateHash(s.marshalizer, s.hasher, s.importHandler.GetHardForkMetaBlock())
	if err != nil {
		return nil, err
	}

	shardHeader.MetaBlockHashes = [][]byte{metaBlockHash}
	shardHeader.AccumulatedFees = big.NewInt(0)
	shardHeader.DeveloperFees = big.NewInt(0)

	s.saveAllCreatedDestMeTransactionsToCache(shardHeader, body)
	s.saveAllTransactionsToStorageIfSelfShard(shardHeader, body)

	return shardHeader, nil
}

// CreatePostMiniBlocks will create all the post miniBlocks from the given miniBlocks info
func (s *shardBlockCreator) CreatePostMiniBlocks(mbsInfo []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
	s.txCoordinator.CreateBlockStarted()

	body := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}

	for _, mbInfo := range mbsInfo {
		if mbInfo.ReceiverShardID != s.shardCoordinator.SelfId() {
			continue
		}

		miniBlock, err := s.pendingTxProcessor.ProcessTransactionsDstMe(mbInfo)
		if err != nil {
			return nil, nil, err
		}

		body.MiniBlocks = append(body.MiniBlocks, miniBlock)
	}

	postProcessMiniBlocks := s.txCoordinator.CreatePostProcessMiniBlocks()
	postProcessMiniBlocksInfo := make([]*update.MbInfo, len(postProcessMiniBlocks))
	for index, postProcessMiniBlock := range postProcessMiniBlocks {
		mbHash, errCalculateHash := core.CalculateHash(s.marshalizer, s.hasher, postProcessMiniBlock)
		if errCalculateHash != nil {
			return nil, nil, errCalculateHash
		}

		body.MiniBlocks = append(body.MiniBlocks, postProcessMiniBlock)

		postProcessMiniBlockInfo, errCreate := s.createMiniBlockInfoForPostProcessMiniBlock(mbHash, postProcessMiniBlock)
		if errCreate != nil {
			return nil, nil, errCreate
		}

		postProcessMiniBlocksInfo[index] = postProcessMiniBlockInfo
	}

	_, err := s.pendingTxProcessor.Commit()
	if err != nil {
		return nil, nil, err
	}

	return body, postProcessMiniBlocksInfo, nil
}

func (s *shardBlockCreator) getPendingMbsAndTxsInCorrectOrder() ([]*update.MbInfo, error) {
	hardForkMetaBlock := s.importHandler.GetHardForkMetaBlock()
	unFinishedMetaBlocks := s.importHandler.GetUnFinishedMetaBlocks()
	pendingMiniBlocks, err := update.GetPendingMiniBlocks(hardForkMetaBlock, unFinishedMetaBlocks)
	if err != nil {
		return nil, err
	}

	importedMiniBlocksMap := s.importHandler.GetMiniBlocks()
	if len(importedMiniBlocksMap) != len(pendingMiniBlocks) {
		return nil, update.ErrWrongImportedMiniBlocksMap
	}

	numPendingTransactions := 0
	importedTransactionsMap := s.importHandler.GetTransactions()
	mbsInfo := make([]*update.MbInfo, len(pendingMiniBlocks))

	for mbIndex, pendingMiniBlock := range pendingMiniBlocks {
		miniBlock, miniBlockFound := importedMiniBlocksMap[string(pendingMiniBlock.Hash)]
		if !miniBlockFound {
			return nil, update.ErrMiniBlockNotFoundInImportedMap
		}

		numPendingTransactions += len(miniBlock.TxHashes)
		txsInfo := make([]*update.TxInfo, len(miniBlock.TxHashes))
		for txIndex, txHash := range miniBlock.TxHashes {
			tx, transactionFound := importedTransactionsMap[string(txHash)]
			if !transactionFound {
				return nil, update.ErrTransactionNotFoundInImportedMap
			}

			txsInfo[txIndex] = &update.TxInfo{
				TxHash: txHash,
				Tx:     tx,
			}
		}

		mbsInfo[mbIndex] = &update.MbInfo{
			MbHash:          pendingMiniBlock.Hash,
			SenderShardID:   pendingMiniBlock.SenderShardID,
			ReceiverShardID: pendingMiniBlock.ReceiverShardID,
			Type:            pendingMiniBlock.Type,
			TxsInfo:         txsInfo,
		}
	}

	if len(importedTransactionsMap) != numPendingTransactions {
		return nil, update.ErrWrongImportedTransactionsMap
	}

	return mbsInfo, nil
}

func (s *shardBlockCreator) createMiniBlockHeaders(body *block.Body) (int, []block.MiniBlockHeader, error) {
	if len(body.MiniBlocks) == 0 {
		return 0, nil, nil
	}

	totalTxCount := 0
	miniBlockHeaders := make([]block.MiniBlockHeader, len(body.MiniBlocks))

	for i := 0; i < len(body.MiniBlocks); i++ {
		txCount := len(body.MiniBlocks[i].TxHashes)
		totalTxCount += txCount

		miniBlockHash, err := core.CalculateHash(s.marshalizer, s.hasher, body.MiniBlocks[i])
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

func (s *shardBlockCreator) saveAllTransactionsToStorageIfSelfShard(
	shardHeader *block.Header,
	body *block.Body,
) {
	if shardHeader.GetShardID() != s.selfShardID {
		return
	}

	mapBlockTypesTxs := make(map[block.Type]map[string]data.TransactionHandler)
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if _, ok := mapBlockTypesTxs[miniBlock.Type]; !ok {
			mapBlockTypesTxs[miniBlock.Type] = s.txCoordinator.GetAllCurrentUsedTxs(miniBlock.Type)
		}

		marshalizedMiniBlock, errNotCritical := s.marshalizer.Marshal(miniBlock)
		if errNotCritical != nil {
			log.Warn("saveAllTransactionsToStorageIfSelfShard.Marshal", "error", errNotCritical.Error())
			continue
		}

		errNotCritical = s.storage.Put(dataRetriever.MiniBlockUnit, shardHeader.MiniBlockHeaders[i].Hash, marshalizedMiniBlock)
		if errNotCritical != nil {
			log.Warn("saveAllTransactionsToStorageIfSelfShard.Put -> MiniBlockUnit", "error", errNotCritical.Error())
		}
	}

	marshalizedReceipts, errNotCritical := s.txCoordinator.CreateMarshalizedReceipts()
	if errNotCritical != nil {
		log.Warn("saveAllTransactionsToStorageIfSelfShard.CreateMarshalizedReceipts", "error", errNotCritical.Error())
	} else {
		if len(marshalizedReceipts) > 0 {
			errNotCritical = s.storage.Put(dataRetriever.ReceiptsUnit, shardHeader.GetReceiptsHash(), marshalizedReceipts)
			if errNotCritical != nil {
				log.Warn("saveAllTransactionsToStorageIfSelfShard.Put -> ReceiptsUnit", "error", errNotCritical.Error())
			}
		}
	}

	mapTxs := s.importHandler.GetTransactions()
	// save transactions from imported map
	for _, miniBlock := range body.MiniBlocks {
		for _, txHash := range miniBlock.TxHashes {
			tx, ok := mapTxs[string(txHash)]
			if !ok {
				log.Warn("miniblock destination me in genesis block should contain only imported txs")
				continue
			}

			unitType := getUnitTypeFromMiniBlockType(miniBlock.Type)
			marshaledData, errNotCritical := s.marshalizer.Marshal(tx)
			if errNotCritical != nil {
				log.Warn("saveAllTransactionsToStorageIfSelfShard.Marshal", "error", errNotCritical.Error())
				continue
			}

			errNotCritical = s.storage.Put(unitType, txHash, marshaledData)
			if errNotCritical != nil {
				log.Warn("saveAllTransactionsToStorageIfSelfShard.Put -> Transaction", "error", errNotCritical.Error())
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

func (s *shardBlockCreator) saveAllCreatedDestMeTransactionsToCache(
	shardHeader *block.Header,
	body *block.Body,
) {
	// no need to save from me, only from other genesis blocks which has results towards me
	if shardHeader.GetShardID() == s.selfShardID {
		return
	}

	mapBlockTypesTxs := make(map[block.Type]map[string]data.TransactionHandler)
	crossMiniBlocksToMe := make([]*block.MiniBlock, 0)
	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		isCrossShardDstMe := miniBlock.SenderShardID != s.selfShardID &&
			miniBlock.ReceiverShardID == s.selfShardID
		if !isCrossShardDstMe {
			continue
		}

		_ = s.dataPool.MiniBlocks().Put(shardHeader.MiniBlockHeaders[i].Hash, miniBlock, miniBlock.Size())
		crossMiniBlocksToMe = append(crossMiniBlocksToMe, miniBlock)

		if _, ok := mapBlockTypesTxs[miniBlock.Type]; !ok {
			mapBlockTypesTxs[miniBlock.Type] = s.txCoordinator.GetAllCurrentUsedTxs(miniBlock.Type)
		}
	}

	for _, miniBlock := range crossMiniBlocksToMe {
		for _, txHash := range miniBlock.TxHashes {
			tx := mapBlockTypesTxs[miniBlock.Type][string(txHash)]
			s.saveTxToCache(tx, txHash, miniBlock)
		}
	}
}

// with the current design only smart contract results are possible for this scenario - but wanted to leave it open,
// to see if other scenarios are needed as well
func (s *shardBlockCreator) saveTxToCache(
	tx data.TransactionHandler,
	txHash []byte,
	miniBlock *block.MiniBlock,
) {
	if check.IfNil(tx) {
		log.Warn("missing transaction in saveTxToCache shard genesis block", "hash", txHash, "type", miniBlock.Type)
		return
	}

	var chosenCache dataRetriever.ShardedDataCacherNotifier
	strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
	switch miniBlock.Type {
	case block.TxBlock:
		chosenCache = s.dataPool.Transactions()
	case block.RewardsBlock:
		chosenCache = s.dataPool.RewardTransactions()
	case block.SmartContractResultBlock:
		chosenCache = s.dataPool.UnsignedTransactions()
	default:
		log.Warn("invalid miniblock type to save into cache", miniBlock.Type)
		return
	}

	chosenCache.AddData(txHash, tx, tx.Size(), strCache)
}

func (s *shardBlockCreator) createMiniBlockInfoForPostProcessMiniBlock(
	mbHash []byte,
	mb *block.MiniBlock,
) (*update.MbInfo, error) {
	mapHashTx := s.txCoordinator.GetAllCurrentUsedTxs(mb.Type)
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

// IsInterfaceNil returns true if underlying object is nil
func (s *shardBlockCreator) IsInterfaceNil() bool {
	return s == nil
}
