package dataprocessor

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/state"
	stateFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
)

// RatingProcessorArgs holds the arguments needed for creating a new ratingsProcessor
type RatingProcessorArgs struct {
	ValidatorPubKeyConverter core.PubkeyConverter
	GenesisNodesConfig       sharding.GenesisNodesSetupHandler
	ShardCoordinator         sharding.Coordinator
	DbPathWithChainID        string
	GeneralConfig            config.Config
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	ElasticIndexer           StorageDataIndexer
	RatingsConfig            config.RatingsConfig
}

type ratingsProcessor struct {
	validatorPubKeyConverter core.PubkeyConverter
	shardCoordinator         sharding.Coordinator
	generalConfig            config.Config
	dbPathWithChainID        string
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	elasticIndexer           StorageDataIndexer
	peerAdapter              state.AccountsAdapter
	genesisNodesConfig       sharding.GenesisNodesSetupHandler
	ratingsConfig            config.RatingsConfig
}

// NewRatingsProcessor will return a new instance of ratingsProcessor
func NewRatingsProcessor(args RatingProcessorArgs) (*ratingsProcessor, error) {
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return nil, ErrNilPubKeyConverter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.ElasticIndexer) {
		return nil, ErrNilElasticIndexer
	}
	if check.IfNil(args.GenesisNodesConfig) {
		return nil, ErrNilGenesisNodesSetup
	}

	rp := &ratingsProcessor{
		validatorPubKeyConverter: args.ValidatorPubKeyConverter,
		shardCoordinator:         args.ShardCoordinator,
		generalConfig:            args.GeneralConfig,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		elasticIndexer:           args.ElasticIndexer,
		dbPathWithChainID:        args.DbPathWithChainID,
		genesisNodesConfig:       args.GenesisNodesConfig,
		ratingsConfig:            args.RatingsConfig,
	}

	err := rp.createPeerAdapter()
	if err != nil {
		return nil, err
	}

	return rp, nil
}

// IndexRatingsForEpochStartMetaBlock will index the ratings for an epoch start meta block
func (rp *ratingsProcessor) IndexRatingsForEpochStartMetaBlock(metaBlock *block.MetaBlock) error {
	if metaBlock.GetNonce() == 0 {
		rp.indexRating(0, rp.getGenesisRating())
		return nil
	}

	rootHash := metaBlock.ValidatorStatsRootHash
	ctx := context.Background()
	leaves, err := rp.peerAdapter.GetAllLeaves(rootHash, ctx)
	if err != nil {
		log.Error("ratingsProcessor -> GetAllLeaves error", "error", err, "root hash", rootHash)
		// don't return error because if the trie is prunned this kind of data will be available only for the last 3 epochs
		return nil
	}

	validatorsRatingData, err := rp.getValidatorsRatingFromLeaves(leaves)
	if err != nil {
		return err
	}

	rp.indexRating(metaBlock.GetEpoch(), validatorsRatingData)

	return nil
}

func (rp *ratingsProcessor) indexRating(epoch uint32, validatorsRatingData map[uint32][]*indexer.ValidatorRatingInfo) {
	for shardID, validators := range validatorsRatingData {
		index := fmt.Sprintf("%d_%d", shardID, epoch)
		rp.elasticIndexer.SaveValidatorsRating(index, validators)
		log.Info("indexed validators rating", "shard ID", shardID, "num peers", len(validators))
	}
}

func (rp *ratingsProcessor) createPeerAdapter() error {
	pathManager, err := storageFactory.CreatePathManagerFromSinglePathString(rp.dbPathWithChainID)
	if err != nil {
		return err
	}
	trieFactoryArgs := factory.TrieFactoryArgs{
		EvictionWaitingListCfg:   rp.generalConfig.EvictionWaitingList,
		SnapshotDbCfg:            rp.generalConfig.TrieSnapshotDB,
		Marshalizer:              rp.marshalizer,
		Hasher:                   rp.hasher,
		PathManager:              pathManager,
		TrieStorageManagerConfig: rp.generalConfig.TrieStorageManagerConfig,
	}
	trieFactory, err := factory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return err
	}

	_, peerAccountsTrie, err := trieFactory.Create(
		rp.generalConfig.PeerAccountsTrieStorage,
		core.GetShardIDString(core.MetachainShardId),
		rp.generalConfig.StateTriesConfig.PeerStatePruningEnabled,
		rp.generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
	)
	if err != nil {
		return err
	}

	peerAdapter, err := state.NewPeerAccountsDB(
		peerAccountsTrie,
		rp.hasher,
		rp.marshalizer,
		stateFactory.NewPeerAccountCreator(),
	)
	if err != nil {
		return err
	}

	rp.peerAdapter = peerAdapter
	return nil
}

func (rp *ratingsProcessor) getValidatorsRatingFromLeaves(leavesChannel chan core.KeyValueHolder) (map[uint32][]*indexer.ValidatorRatingInfo, error) {
	validatorsRatingInfo := make(map[uint32][]*indexer.ValidatorRatingInfo)
	for pa := range leavesChannel {
		peerAccount, err := unmarshalPeer(pa.Value(), rp.marshalizer)
		if err != nil {
			continue
		}

		validatorsRatingInfo[peerAccount.GetShardId()] = append(validatorsRatingInfo[peerAccount.GetShardId()],
			&indexer.ValidatorRatingInfo{
				PublicKey: rp.validatorPubKeyConverter.Encode(peerAccount.GetBLSPublicKey()),
				Rating:    float32(peerAccount.GetRating()) * 100 / float32(rp.ratingsConfig.General.MaxRating),
			})
	}

	return validatorsRatingInfo, nil
}

func (rp *ratingsProcessor) getGenesisRating() map[uint32][]*indexer.ValidatorRatingInfo {
	ratingsForGenesis := make(map[uint32][]*indexer.ValidatorRatingInfo)

	eligible, waiting := rp.genesisNodesConfig.InitialNodesInfo()
	for shardID, nodesInShard := range eligible {
		for _, node := range nodesInShard {
			ratingsForGenesis[shardID] = append(ratingsForGenesis[shardID], &indexer.ValidatorRatingInfo{
				PublicKey: rp.validatorPubKeyConverter.Encode(node.PubKeyBytes()),
				Rating:    float32(node.GetInitialRating()) * 100 / float32(rp.ratingsConfig.General.MaxRating),
			})
		}
	}
	for shardID, nodesInShard := range waiting {
		for _, node := range nodesInShard {
			ratingsForGenesis[shardID] = append(ratingsForGenesis[shardID], &indexer.ValidatorRatingInfo{
				PublicKey: rp.validatorPubKeyConverter.Encode(node.PubKeyBytes()),
				Rating:    float32(node.GetInitialRating()) * 100 / float32(rp.ratingsConfig.General.MaxRating),
			})
		}
	}

	return ratingsForGenesis
}

// IsInterfaceNil returns true if there is no value under the interface
func (rp *ratingsProcessor) IsInterfaceNil() bool {
	return rp == nil
}

func unmarshalPeer(pa []byte, marshalizer marshal.Marshalizer) (state.PeerAccountHandler, error) {
	peerAccount := state.NewEmptyPeerAccount()
	err := marshalizer.Unmarshal(peerAccount, pa)
	if err != nil {
		log.Error("cannot unmarshal peer account", "error", err)
		return nil, err
	}

	return peerAccount, nil
}
