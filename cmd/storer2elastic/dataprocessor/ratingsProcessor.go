package dataprocessor

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factory2 "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
)

// RatingProcessorArgs holds the arguments needed for creating a new ratingsProcessor
type RatingProcessorArgs struct {
	ValidatorPubKeyConverter core.PubkeyConverter
	ShardCoordinator         sharding.Coordinator
	DbPathWithChainID        string
	GeneralConfig            config.Config
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	ElasticIndexer           indexer.Indexer
	DatabaseReader           DatabaseReaderHandler
}

type ratingsProcessor struct {
	validatorPubKeyConverter core.PubkeyConverter
	shardCoordinator         sharding.Coordinator
	generalConfig            config.Config
	dbPathWithChainID        string
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	elasticIndexer           indexer.Indexer
	databaseReader           DatabaseReaderHandler
	peerAdapter              state.AccountsAdapter
}

// NewRatingsProcessor will return a new instance of ratingsProcessor
func NewRatingsProcessor(args RatingProcessorArgs) (*ratingsProcessor, error) {
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return nil, errors.New("nil pub key converter")
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
	if check.IfNil(args.DatabaseReader) {
		return nil, ErrNilDatabaseReader
	}

	rp := &ratingsProcessor{
		validatorPubKeyConverter: args.ValidatorPubKeyConverter,
		shardCoordinator:         args.ShardCoordinator,
		generalConfig:            args.GeneralConfig,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		elasticIndexer:           args.ElasticIndexer,
		databaseReader:           args.DatabaseReader,
		dbPathWithChainID:        args.DbPathWithChainID,
	}

	err := rp.createPeerAdapter()
	if err != nil {
		return nil, err
	}

	return rp, nil
}

// IndexRatingsForEpochStartMetaBlock will index the ratings for an epoch start meta block
func (rp *ratingsProcessor) IndexRatingsForEpochStartMetaBlock(metaBlock *block.MetaBlock) error {
	rootHash := metaBlock.ValidatorStatsRootHash
	leaves, err := rp.peerAdapter.GetAllLeaves(rootHash)
	if err != nil {
		return err
	}

	validatorsData, err := getValidatorDataFromLeaves(leaves, rp.shardCoordinator, rp.marshalizer)
	if err != nil {
		return err
	}

	validatorsRatingInfo := make([]indexer.ValidatorRatingInfo, 0)
	for _, valsInShard := range validatorsData {
		for _, val := range valsInShard {
			validatorsRatingInfo = append(validatorsRatingInfo,
				indexer.ValidatorRatingInfo{
					PublicKey: rp.validatorPubKeyConverter.Encode(val.PublicKey),
					Rating:    float32(val.Rating),
				})
		}
	}

	log.Info("indexed validators rating", "num values", len(validatorsRatingInfo))
	rp.elasticIndexer.SaveValidatorsRating("0", validatorsRatingInfo)
	//ratingsBytes, err := rp.getRatingsFromStorage(rootHash)
	//if err != nil {
	//	return err
	//}
	//
	//var tr trie.CollapsedBn
	//err = rp.marshalizer.Unmarshal(&tr, ratingsBytes)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (rp *ratingsProcessor) getRatingsFromStorage(rootHash []byte) ([]byte, error) {
	staticDbs, err := rp.databaseReader.GetStaticDatabaseInfo()
	if err != nil {
		return nil, err
	}

	var metaStaticDb *databasereader.DatabaseInfo
	for _, db := range staticDbs {
		if db.Shard == core.MetachainShardId {
			metaStaticDb = db
			break
		}
	}

	if metaStaticDb == nil {
		return nil, errors.New("no metachain static DB found")
	}

	persisterPath := fmt.Sprintf("PeerAccountsTrie/TrieSnapshot/%d", metaStaticDb.Epoch)
	peerAccountsStorer, err := rp.databaseReader.LoadStaticPersister(metaStaticDb, persisterPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = peerAccountsStorer.Close()
		log.LogIfError(err)
	}()

	return peerAccountsStorer.Get(rootHash)
}

func (rp *ratingsProcessor) createPeerAdapter() error {
	pathTemplateForPruningStorer := filepath.Join(
		rp.dbPathWithChainID,
		fmt.Sprintf("%s_%s", "Epoch", core.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", "Shard", core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		rp.dbPathWithChainID,
		"Static",
		fmt.Sprintf("%s_%s", "Shard", core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)
	pathManager, err := pathmanager.NewPathManager(pathTemplateForPruningStorer, pathTemplateForStaticStorer)
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

	peerStMon, peerAccountsTrie, err := trieFactory.Create(
		rp.generalConfig.PeerAccountsTrieStorage,
		core.GetShardIDString(core.MetachainShardId),
		rp.generalConfig.StateTriesConfig.PeerStatePruningEnabled,
		rp.generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
	)
	if err != nil {
		return err
	}
	fmt.Println(peerStMon.IsInterfaceNil())
	peerAdapter, err := state.NewPeerAccountsDB(
		peerAccountsTrie,
		rp.hasher,
		rp.marshalizer,
		factory2.NewPeerAccountCreator(),
	)
	if err != nil {
		return err
	}

	rp.peerAdapter = peerAdapter
	return nil
}

// TODO: create a structure or use this function also in process/peer/process.go
func getValidatorDataFromLeaves(
	leaves map[string][]byte,
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
) (map[uint32][]*state.ValidatorInfo, error) {

	validators := make(map[uint32][]*state.ValidatorInfo, shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		validators[i] = make([]*state.ValidatorInfo, 0)
	}
	validators[core.MetachainShardId] = make([]*state.ValidatorInfo, 0)

	sliceLeaves := convertMapToSortedSlice(leaves)

	sort.Slice(sliceLeaves, func(i, j int) bool {
		return bytes.Compare(sliceLeaves[i], sliceLeaves[j]) < 0
	})

	for _, pa := range sliceLeaves {
		peerAccount, err := unmarshalPeer(pa, marshalizer)
		if err != nil {
			continue
			return nil, err
		}

		currentShardId := peerAccount.GetShardId()
		validatorInfoData := peerAccountToValidatorInfo(peerAccount)
		validators[currentShardId] = append(validators[currentShardId], validatorInfoData)
	}

	return validators, nil
}

func convertMapToSortedSlice(leaves map[string][]byte) [][]byte {
	newLeaves := make([][]byte, len(leaves))
	i := 0
	for _, pa := range leaves {
		newLeaves[i] = pa
		i++
	}

	return newLeaves
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

func peerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:  peerAccount.GetBLSPublicKey(),
		ShardId:    peerAccount.GetShardId(),
		Index:      peerAccount.GetIndexInList(),
		TempRating: peerAccount.GetTempRating(),
		Rating:     peerAccount.GetRating(),
	}
}
