package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder/disabled"
)

// TrieCreateArgs holds arguments for calling the Create method on the TrieFactory
type TrieCreateArgs struct {
	MainStorer         storage.Storer
	CheckpointsStorer  storage.Storer
	PruningEnabled     bool
	CheckpointsEnabled bool
	SnapshotsEnabled   bool
	MaxTrieLevelInMem  uint
	IdleProvider       trie.IdleNodeProvider
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
		MainStorer:             args.MainStorer,
		CheckpointsStorer:      args.CheckpointsStorer,
		Marshalizer:            tc.marshalizer,
		Hasher:                 tc.hasher,
		GeneralConfig:          tc.trieStorageManagerConfig,
		CheckpointHashesHolder: tc.getCheckpointHashesHolder(args.CheckpointsEnabled),
		IdleProvider:           args.IdleProvider,
	}

	options := trie.StorageManagerOptions{
		PruningEnabled:     args.PruningEnabled,
		SnapshotsEnabled:   args.SnapshotsEnabled,
		CheckpointsEnabled: args.CheckpointsEnabled,
	}

	trieStorage, err := trie.CreateTrieStorageManager(
		storageManagerArgs,
		options,
	)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, args.MaxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

func (tc *trieCreator) getCheckpointHashesHolder(checkpointsEnabled bool) trie.CheckpointHashesHolder {
	if !checkpointsEnabled {
		return disabled.NewDisabledCheckpointHashesHolder()
	}

	return hashesHolder.NewCheckpointHashesHolder(
		tc.trieStorageManagerConfig.CheckpointHashesHolderMaxSize,
		uint64(tc.hasher.Size()),
	)
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

	checkpointsStorer, err := storageService.GetStorer(dataRetriever.UserAccountsCheckpointsUnit)
	if err != nil {
		return nil, nil, err
	}

	args := TrieCreateArgs{
		MainStorer:         mainStorer,
		CheckpointsStorer:  checkpointsStorer,
		PruningEnabled:     generalConfig.StateTriesConfig.AccountsStatePruningEnabled,
		CheckpointsEnabled: generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  generalConfig.StateTriesConfig.MaxStateTrieLevelInMemory,
		SnapshotsEnabled:   generalConfig.StateTriesConfig.SnapshotsEnabled,
		IdleProvider:       coreComponentsHolder.ProcessStatusHandler(),
	}
	userStorageManager, userAccountTrie, err := trFactory.Create(args)
	if err != nil {
		return nil, nil, err
	}

	trieContainer := state.NewDataTriesHolder()
	trieStorageManagers := make(map[string]common.StorageManager)

	trieContainer.Put([]byte(UserAccountTrie), userAccountTrie)
	trieStorageManagers[UserAccountTrie] = userStorageManager

	mainStorer, err = storageService.GetStorer(dataRetriever.PeerAccountsUnit)
	if err != nil {
		return nil, nil, err
	}

	checkpointsStorer, err = storageService.GetStorer(dataRetriever.PeerAccountsCheckpointsUnit)
	if err != nil {
		return nil, nil, err
	}

	args = TrieCreateArgs{
		MainStorer:         mainStorer,
		CheckpointsStorer:  checkpointsStorer,
		PruningEnabled:     generalConfig.StateTriesConfig.PeerStatePruningEnabled,
		CheckpointsEnabled: generalConfig.StateTriesConfig.CheckpointsEnabled,
		MaxTrieLevelInMem:  generalConfig.StateTriesConfig.MaxPeerTrieLevelInMemory,
		SnapshotsEnabled:   generalConfig.StateTriesConfig.SnapshotsEnabled,
		IdleProvider:       coreComponentsHolder.ProcessStatusHandler(),
	}
	peerStorageManager, peerAccountsTrie, err := trFactory.Create(args)
	if err != nil {
		return nil, nil, err
	}

	trieContainer.Put([]byte(PeerAccountTrie), peerAccountsTrie)
	trieStorageManagers[PeerAccountTrie] = peerStorageManager

	return trieContainer, trieStorageManagers, nil
}
