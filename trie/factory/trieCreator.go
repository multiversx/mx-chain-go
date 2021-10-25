package factory

import (
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
)

// TrieCreateArgs holds arguments for calling the Create method on the TrieFactory
type TrieCreateArgs struct {
	TrieStorageConfig  config.StorageConfig
	MainStorer         storage.Storer
	CheckpointsStorer  storage.Storer
	ShardID            string
	PruningEnabled     bool
	CheckpointsEnabled bool
	MaxTrieLevelInMem  uint
}

type trieCreator struct {
	snapshotDbCfg            config.DBConfig
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	pathManager              storage.PathManagerHandler
	trieStorageManagerConfig config.TrieStorageManagerConfig
}

var log = logger.GetOrCreate("trie")

// NewTrieFactory creates a new trie factory
func NewTrieFactory(
	args TrieFactoryArgs,
) (*trieCreator, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, trie.ErrNilHasher
	}
	if check.IfNil(args.PathManager) {
		return nil, trie.ErrNilPathManager
	}

	return &trieCreator{
		snapshotDbCfg:            args.SnapshotDbCfg,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		pathManager:              args.PathManager,
		trieStorageManagerConfig: args.TrieStorageManagerConfig,
	}, nil
}

// Create creates a new trie
func (tc *trieCreator) Create(args TrieCreateArgs) (common.StorageManager, common.Trie, error) {
	trieStoragePath, mainDb := path.Split(tc.pathManager.PathForStatic(args.ShardID, args.TrieStorageConfig.DB.FilePath))

	dbConfig := factory.GetDBFromConfig(args.TrieStorageConfig.DB)
	dbConfig.FilePath = path.Join(trieStoragePath, mainDb)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(args.TrieStorageConfig.Cache),
		dbConfig,
		factory.GetBloomFromConfig(args.TrieStorageConfig.Bloom),
	)
	if err != nil {
		return nil, nil, err
	}

	log.Debug("trie pruning status", "enabled", args.PruningEnabled)
	if !args.PruningEnabled {
		return tc.newTrieAndTrieStorageWithoutPruning(accountsTrieStorage, args.MaxTrieLevelInMem)
	}

	snapshotDbCfg := config.DBConfig{
		FilePath:          filepath.Join(trieStoragePath, tc.snapshotDbCfg.FilePath),
		Type:              tc.snapshotDbCfg.Type,
		BatchDelaySeconds: tc.snapshotDbCfg.BatchDelaySeconds,
		MaxBatchSize:      tc.snapshotDbCfg.MaxBatchSize,
		MaxOpenFiles:      tc.snapshotDbCfg.MaxOpenFiles,
	}

	checkpointHashesHolder := hashesHolder.NewCheckpointHashesHolder(
		tc.trieStorageManagerConfig.CheckpointHashesHolderMaxSize,
		uint64(tc.hasher.Size()),
	)
	storageManagerArgs := trie.NewTrieStorageManagerArgs{
		DB:                     accountsTrieStorage,
		MainStorer:             args.MainStorer,
		CheckpointsStorer:      args.CheckpointsStorer,
		Marshalizer:            tc.marshalizer,
		Hasher:                 tc.hasher,
		SnapshotDbConfig:       snapshotDbCfg,
		GeneralConfig:          tc.trieStorageManagerConfig,
		CheckpointHashesHolder: checkpointHashesHolder,
	}

	log.Debug("trie checkpoints status", "enabled", args.CheckpointsEnabled)
	if !args.CheckpointsEnabled {
		return tc.newTrieAndTrieStorageWithoutCheckpoints(storageManagerArgs, args.MaxTrieLevelInMem)
	}

	return tc.newTrieAndTrieStorage(storageManagerArgs, args.MaxTrieLevelInMem)
}

func (tc *trieCreator) newTrieAndTrieStorage(
	args trie.NewTrieStorageManagerArgs,
	maxTrieLevelInMem uint,
) (common.StorageManager, common.Trie, error) {
	trieStorage, err := trie.NewTrieStorageManager(args)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

func (tc *trieCreator) newTrieAndTrieStorageWithoutCheckpoints(
	args trie.NewTrieStorageManagerArgs,
	maxTrieLevelInMem uint,
) (common.StorageManager, common.Trie, error) {
	trieStorage, err := trie.NewTrieStorageManagerWithoutCheckpoints(args)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

func (tc *trieCreator) newTrieAndTrieStorageWithoutPruning(
	accountsTrieStorage common.DBWriteCacher,
	maxTrieLevelInMem uint,
) (common.StorageManager, common.Trie, error) {
	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *trieCreator) IsInterfaceNil() bool {
	return tc == nil
}

// CreateTriesComponentsForShardId creates the user and peer tries and trieStorageManagers
func CreateTriesComponentsForShardId(
	generalConfig config.Config,
	coreComponentsHolder coreComponentsHandler,
	shardId uint32,
	storageService dataRetriever.StorageService,
) (common.TriesHolder, map[string]common.StorageManager, error) {
	trieFactoryArgs := TrieFactoryArgs{
		SnapshotDbCfg:            generalConfig.TrieSnapshotDB,
		Marshalizer:              coreComponentsHolder.InternalMarshalizer(),
		Hasher:                   coreComponentsHolder.Hasher(),
		PathManager:              coreComponentsHolder.PathHandler(),
		TrieStorageManagerConfig: generalConfig.TrieStorageManagerConfig,
	}
	trFactory, err := NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	args := TrieCreateArgs{
		TrieStorageConfig:  generalConfig.AccountsTrieStorageOld,
		MainStorer:         storageService.GetStorer(dataRetriever.UserAccountsUnit),
		CheckpointsStorer:  storageService.GetStorer(dataRetriever.UserAccountsCheckpointsUnit),
		ShardID:            core.GetShardIDString(shardId),
		PruningEnabled:     generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		CheckpointsEnabled: generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
	}
	userStorageManager, userAccountTrie, err := trFactory.Create(args)
	if err != nil {
		return nil, nil, err
	}

	trieContainer := state.NewDataTriesHolder()
	trieStorageManagers := make(map[string]common.StorageManager)

	trieContainer.Put([]byte(UserAccountTrie), userAccountTrie)
	trieStorageManagers[UserAccountTrie] = userStorageManager

	args = TrieCreateArgs{
		TrieStorageConfig:  generalConfig.PeerAccountsTrieStorageOld,
		MainStorer:         storageService.GetStorer(dataRetriever.PeerAccountsUnit),
		CheckpointsStorer:  storageService.GetStorer(dataRetriever.PeerAccountsCheckpointsUnit),
		ShardID:            core.GetShardIDString(shardId),
		PruningEnabled:     generalConfig.StateTriesConfig.PeerStatePruningEnabled,
		CheckpointsEnabled: generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
	}
	peerStorageManager, peerAccountsTrie, err := trFactory.Create(args)
	if err != nil {
		return nil, nil, err
	}

	trieContainer.Put([]byte(PeerAccountTrie), peerAccountsTrie)
	trieStorageManagers[PeerAccountTrie] = peerStorageManager

	return trieContainer, trieStorageManagers, nil
}
