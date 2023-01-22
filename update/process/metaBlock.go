package process

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/update"
)

// ArgsNewMetaBlockCreatorAfterHardFork defines the arguments structure for new metablock creator after hardfork
type ArgsNewMetaBlockCreatorAfterHardFork struct {
	Hasher             hashing.Hasher
	ImportHandler      update.ImportHandler
	Marshalizer        marshal.Marshalizer
	PendingTxProcessor update.PendingTransactionProcessor
	ShardCoordinator   sharding.Coordinator
	Storage            dataRetriever.StorageService
	TxCoordinator      process.TransactionCoordinator
	ValidatorAccounts  state.AccountsAdapter
	ReceiptsRepository receiptsRepository
	SelfShardID        uint32
}

type metaBlockCreator struct {
	*baseProcessor
	validatorAccounts state.AccountsAdapter
}

// NewMetaBlockCreatorAfterHardfork creates the after hardfork metablock creator
func NewMetaBlockCreatorAfterHardfork(args ArgsNewMetaBlockCreatorAfterHardFork) (*metaBlockCreator, error) {
	err := checkBlockCreatorAfterHardForkNilParameters(
		args.Hasher,
		args.ImportHandler,
		args.Marshalizer,
		args.PendingTxProcessor,
		args.ShardCoordinator,
		args.Storage,
		args.TxCoordinator,
		args.ReceiptsRepository,
	)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.ValidatorAccounts) {
		return nil, update.ErrNilAccounts
	}

	base := &baseProcessor{
		hasher:             args.Hasher,
		importHandler:      args.ImportHandler,
		marshalizer:        args.Marshalizer,
		pendingTxProcessor: args.PendingTxProcessor,
		shardCoordinator:   args.ShardCoordinator,
		storage:            args.Storage,
		txCoordinator:      args.TxCoordinator,
		receiptsRepository: args.ReceiptsRepository,
		selfShardID:        args.SelfShardID,
	}

	return &metaBlockCreator{
		baseProcessor:     base,
		validatorAccounts: args.ValidatorAccounts,
	}, nil
}

// CreateBlock will create a block after hardfork import
func (m *metaBlockCreator) CreateBlock(
	body *block.Body,
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (data.HeaderHandler, error) {
	if len(chainID) == 0 {
		return nil, update.ErrEmptyChainID
	}

	validatorRootHash, err := m.validatorAccounts.Commit()
	if err != nil {
		return nil, err
	}

	rootHash, err := m.pendingTxProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	hardForkMeta := m.importHandler.GetHardForkMetaBlock()
	epochStart, ok := hardForkMeta.GetEpochStartHandler().(*block.EpochStart)
	if !ok {
		return nil, update.ErrWrongTypeAssertion
	}

	metaHeader := &block.MetaBlock{
		Nonce:                  nonce,
		Round:                  round,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: validatorRootHash,
		EpochStart:             *epochStart,
		ChainID:                []byte(chainID),
		SoftwareVersion:        []byte(""),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		Epoch:                  epoch,
		PubKeysBitmap:          []byte{1},
	}

	metaHeader.ReceiptsHash, err = m.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, err
	}

	totalTxCount, miniBlockHeaders, err := m.createMiniBlockHeaders(body)
	if err != nil {
		return nil, err
	}

	metaHeader.MiniBlockHeaders = miniBlockHeaders
	metaHeader.TxCount = uint32(totalTxCount)

	m.saveAllBlockDataToStorageForSelfShard(metaHeader, body)

	return metaHeader, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (m *metaBlockCreator) IsInterfaceNil() bool {
	return m == nil
}
