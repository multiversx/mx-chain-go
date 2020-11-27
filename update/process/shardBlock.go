package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
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
	Hasher             hashing.Hasher
	ImportHandler      update.ImportHandler
	Marshalizer        marshal.Marshalizer
	PendingTxProcessor update.PendingTransactionProcessor
	ShardCoordinator   sharding.Coordinator
	Storage            dataRetriever.StorageService
	TxCoordinator      process.TransactionCoordinator
	SelfShardID        uint32
}

type shardBlockCreator struct {
	*baseProcessor
}

// NewShardBlockCreatorAfterHardFork creates a shard block processor for the first block after hardfork
func NewShardBlockCreatorAfterHardFork(args ArgsNewShardBlockCreatorAfterHardFork) (*shardBlockCreator, error) {
	err := checkBlockCreatorAfterHardForkNilParameters(
		args.Hasher,
		args.ImportHandler,
		args.Marshalizer,
		args.PendingTxProcessor,
		args.ShardCoordinator,
		args.Storage,
		args.TxCoordinator,
	)
	if err != nil {
		return nil, err
	}

	base := &baseProcessor{
		hasher:             args.Hasher,
		importHandler:      args.ImportHandler,
		marshalizer:        args.Marshalizer,
		pendingTxProcessor: args.PendingTxProcessor,
		shardCoordinator:   args.ShardCoordinator,
		storage:            args.Storage,
		txCoordinator:      args.TxCoordinator,
		selfShardID:        args.SelfShardID,
	}

	return &shardBlockCreator{
		baseProcessor: base,
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

	s.saveAllBlockDataToStorageForSelfShard(shardHeader, body)

	return shardHeader, nil
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

// IsInterfaceNil returns true if underlying object is nil
func (s *shardBlockCreator) IsInterfaceNil() bool {
	return s == nil
}
