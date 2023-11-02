package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie"
)

// TrieCreateArgs holds arguments for calling the Create method on the TrieFactory
type TrieCreateArgs struct {
	MainStorer          storage.Storer
	PruningEnabled      bool
	SnapshotsEnabled    bool
	MaxTrieLevelInMem   uint
	IdleProvider        trie.IdleNodeProvider
	Identifier          string
	EnableEpochsHandler common.EnableEpochsHandler
	StatsCollector      common.StateStatisticsHandler
}

type trieCreator struct {
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	pathManager              storage.PathManagerHandler
	trieStorageManagerConfig config.TrieStorageManagerConfig
}

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
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		pathManager:              args.PathManager,
		trieStorageManagerConfig: args.TrieStorageManagerConfig,
	}, nil
}

// Create creates a new trie
func (tc *trieCreator) Create(args TrieCreateArgs) (common.StorageManager, common.Trie, error) {
	storageManagerArgs := trie.NewTrieStorageManagerArgs{
		MainStorer:     args.MainStorer,
		Marshalizer:    tc.marshalizer,
		Hasher:         tc.hasher,
		GeneralConfig:  tc.trieStorageManagerConfig,
		IdleProvider:   args.IdleProvider,
		Identifier:     args.Identifier,
		StatsCollector: args.StatsCollector,
	}

	options := trie.StorageManagerOptions{
		PruningEnabled:   args.PruningEnabled,
		SnapshotsEnabled: args.SnapshotsEnabled,
	}

	trieStorage, err := trie.CreateTrieStorageManager(
		storageManagerArgs,
		options,
	)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, args.EnableEpochsHandler, args.MaxTrieLevelInMem)
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
	storageService dataRetriever.StorageService,
	stateStatsHandler common.StateStatisticsHandler,
) (common.TriesHolder, map[string]common.StorageManager, error) {
	trieFactoryArgs := TrieFactoryArgs{
		Marshalizer:              coreComponentsHolder.InternalMarshalizer(),
		Hasher:                   coreComponentsHolder.Hasher(),
		PathManager:              coreComponentsHolder.PathHandler(),
		TrieStorageManagerConfig: generalConfig.TrieStorageManagerConfig,
	}
	trFactory, err := NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	mainStorer, err := storageService.GetStorer(dataRetriever.UserAccountsUnit)
	if err != nil {
		return nil, nil, err
	}

	args := TrieCreateArgs{
		MainStorer:          mainStorer,
		PruningEnabled:      generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		MaxTrieLevelInMem:   generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
		SnapshotsEnabled:    generalConfig.StateTriesConfig.SnapshotsEnabled,
		IdleProvider:        coreComponentsHolder.ProcessStatusHandler(),
		Identifier:          dataRetriever.UserAccountsUnit.String(),
		EnableEpochsHandler: coreComponentsHolder.EnableEpochsHandler(),
		StatsCollector:      stateStatsHandler,
	}
	userStorageManager, userAccountTrie, err := trFactory.Create(args)
	if err != nil {
		return nil, nil, err
	}

	trieContainer := state.NewDataTriesHolder()
	trieStorageManagers := make(map[string]common.StorageManager)

	trieContainer.Put([]byte(dataRetriever.UserAccountsUnit.String()), userAccountTrie)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = userStorageManager

	mainStorer, err = storageService.GetStorer(dataRetriever.PeerAccountsUnit)
	if err != nil {
		return nil, nil, err
	}

	args = TrieCreateArgs{
		MainStorer:          mainStorer,
		PruningEnabled:      generalConfig.StateTriesConfig.PeerStatePruningEnabled,
		MaxTrieLevelInMem:   generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
		SnapshotsEnabled:    generalConfig.StateTriesConfig.SnapshotsEnabled,
		IdleProvider:        coreComponentsHolder.ProcessStatusHandler(),
		Identifier:          dataRetriever.PeerAccountsUnit.String(),
		EnableEpochsHandler: coreComponentsHolder.EnableEpochsHandler(),
		StatsCollector:      stateStatsHandler,
	}
	peerStorageManager, peerAccountsTrie, err := trFactory.Create(args)
	if err != nil {
		return nil, nil, err
	}

	trieContainer.Put([]byte(dataRetriever.PeerAccountsUnit.String()), peerAccountsTrie)
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = peerStorageManager

	return trieContainer, trieStorageManagers, nil
}
