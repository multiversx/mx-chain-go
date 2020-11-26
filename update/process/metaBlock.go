package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgsNewMetaBlockCreatorAfterHardfork defines the arguments structure for new metablock creator after hardfork
type ArgsNewMetaBlockCreatorAfterHardfork struct {
	ImportHandler     update.ImportHandler
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	ShardCoordinator  sharding.Coordinator
	ValidatorAccounts state.AccountsAdapter
}

type metaBlockCreator struct {
	importHandler     update.ImportHandler
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	validatorAccounts state.AccountsAdapter
}

// NewMetaBlockCreatorAfterHardfork creates the after hardfork metablock creator
func NewMetaBlockCreatorAfterHardfork(args ArgsNewMetaBlockCreatorAfterHardfork) (*metaBlockCreator, error) {
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
	if check.IfNil(args.ValidatorAccounts) {
		return nil, update.ErrNilAccounts
	}

	return &metaBlockCreator{
		importHandler:     args.ImportHandler,
		marshalizer:       args.Marshalizer,
		hasher:            args.Hasher,
		shardCoordinator:  args.ShardCoordinator,
		validatorAccounts: args.ValidatorAccounts,
	}, nil
}

// CreateBody will create a block body after hardfork import
func (m *metaBlockCreator) CreateBody() (*block.Body, []*update.MbInfo, error) {
	return &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}, nil, nil
}

// CreatePostMiniBlocks will create all the post miniBlocks from the given miniBlocks info
func (m *metaBlockCreator) CreatePostMiniBlocks(_ []*update.MbInfo) (*block.Body, []*update.MbInfo, error) {
	return &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}, nil, nil
}

// CreateBlock will create a block after hardfork import
func (m *metaBlockCreator) CreateBlock(
	_ *block.Body,
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

	accounts := m.importHandler.GetAccountsDBForShard(core.MetachainShardId)
	if check.IfNil(accounts) {
		return nil, update.ErrNilAccounts
	}

	rootHash, err := accounts.Commit()
	if err != nil {
		return nil, err
	}

	hardForkMeta := m.importHandler.GetHardForkMetaBlock()
	metaHdr := &block.MetaBlock{
		Nonce:                  nonce,
		Round:                  round,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: validatorRootHash,
		EpochStart:             hardForkMeta.EpochStart,
		ChainID:                []byte(chainID),
		SoftwareVersion:        []byte(""),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		Epoch:                  epoch,
		PubKeysBitmap:          []byte{1},
	}

	return metaHdr, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (m *metaBlockCreator) IsInterfaceNil() bool {
	return m == nil
}
