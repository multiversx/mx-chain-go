package process

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgsNewMetaBlockCreatorAfterHardfork defines the arguments structure for new metablock creator after hardfork
type ArgsNewMetaBlockCreatorAfterHardfork struct {
	ImportHandler    update.ImportHandler
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}

type metaBlockCreator struct {
	importHandler    update.ImportHandler
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
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

	return &metaBlockCreator{
		importHandler:    args.ImportHandler,
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

// CreateNewBlock will create a new block after hardfork import
func (m *metaBlockCreator) CreateNewBlock(
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (data.HeaderHandler, data.BodyHandler, error) {

	validatorAccounts := m.importHandler.GetValidatorAccountsDB()
	if check.IfNil(validatorAccounts) {
		return nil, nil, update.ErrNilAccounts
	}
	validatorRootHash, err := validatorAccounts.RootHash()
	if err != nil {
		return nil, nil, err
	}

	accounts := m.importHandler.GetAccountsDBForShard(core.MetachainShardId)
	if check.IfNil(accounts) {
		return nil, nil, update.ErrNilAccounts
	}
	rootHash, err := accounts.RootHash()
	if err != nil {
		return nil, nil, err
	}

	blockBody := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}
	metaHdr := &block.MetaBlock{
		Nonce:                  nonce,
		Round:                  round,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: validatorRootHash,
		EpochStart:             block.EpochStart{},
		ChainID:                []byte(chainID),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		Epoch:                  epoch,
	}

	return metaHdr, blockBody, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (m *metaBlockCreator) IsInterfaceNil() bool {
	return m == nil
}
