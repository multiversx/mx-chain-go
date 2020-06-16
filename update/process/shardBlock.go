package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
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
}

type shardBlockCreator struct {
	shardCoordinator   sharding.Coordinator
	txCoordinator      process.TransactionCoordinator
	pendingTxProcessor update.PendingTransactionProcessor
	importHandler      update.ImportHandler
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
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

	return &shardBlockCreator{
		shardCoordinator:   args.ShardCoordinator,
		txCoordinator:      args.TxCoordinator,
		pendingTxProcessor: args.PendingTxProcessor,
		importHandler:      args.ImportHandler,
		marshalizer:        args.Marshalizer,
		hasher:             args.Hasher,
	}, nil
}

// CreateNewBlock will create a new block after hardfork import
func (s *shardBlockCreator) CreateNewBlock(
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (data.HeaderHandler, data.BodyHandler, error) {
	if len(chainID) == 0 {
		return nil, nil, update.ErrEmptyChainID
	}

	blockBody, err := s.createBody()
	if err != nil {
		return nil, nil, err
	}

	rootHash, err := s.pendingTxProcessor.RootHash()
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	totalTxCount, miniBlockHeaders, err := s.createMiniBlockHeaders(blockBody)
	if err != nil {
		return nil, nil, err
	}

	shardHeader.MiniBlockHeaders = miniBlockHeaders
	shardHeader.TxCount = uint32(totalTxCount)

	metaBlockHash, err := core.CalculateHash(s.marshalizer, s.hasher, s.importHandler.GetHardForkMetaBlock())
	if err != nil {
		return nil, nil, err
	}

	shardHeader.MetaBlockHashes = [][]byte{metaBlockHash}
	shardHeader.AccumulatedFees = big.NewInt(0)
	shardHeader.DeveloperFees = big.NewInt(0)

	return shardHeader, blockBody, nil
}

func (s *shardBlockCreator) createBody() (*block.Body, error) {
	mapTxs := s.importHandler.GetTransactions()

	s.txCoordinator.CreateBlockStarted()

	dstMeMiniBlocks, err := s.pendingTxProcessor.ProcessTransactionsDstMe(mapTxs)
	if err != nil {
		return nil, err
	}

	postProcessMiniBlocks := s.txCoordinator.CreatePostProcessMiniBlocks()

	return &block.Body{
		MiniBlocks: append(dstMeMiniBlocks, postProcessMiniBlocks...),
	}, nil
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

// IsInterfaceNil returns true if underlying object is nil
func (s *shardBlockCreator) IsInterfaceNil() bool {
	return s == nil
}
