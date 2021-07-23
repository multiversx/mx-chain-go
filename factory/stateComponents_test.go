package factory_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/stretchr/testify/require"
)

func TestNewStateComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)
	args.ShardCoordinator = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewStateComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)
	args.Core = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, errors.ErrNilCoreComponents, err)
}

func TestNewStateComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)

	scf, err := factory.NewStateComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, scf)
}

func TestStateComponentsFactory_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)

	scf, _ := factory.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

// ------------ Test StateComponents --------------------
func TestStateComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)
	scf, _ := factory.NewStateComponentsFactory(args)

	sc, _ := scf.Create()

	err := sc.Close()
	require.NoError(t, err)
}

func getStateArgs(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.StateComponentsFactoryArgs {
	memDBUsers := mock.NewMemDbMock()
	memdbPeers := mock.NewMemDbMock()
	storageManagerUser, _ := trie.NewTrieStorageManagerWithoutPruning(memDBUsers)
	storageManagerPeer, _ := trie.NewTrieStorageManagerWithoutPruning(memdbPeers)

	trieStorageManagers := make(map[string]temporary.StorageManager)
	trieStorageManagers[trieFactory.UserAccountTrie] = storageManagerUser
	trieStorageManagers[trieFactory.PeerAccountTrie] = storageManagerPeer

	triesHolder := state.NewDataTriesHolder()
	trieUsers, _ := trie.NewTrie(storageManagerUser, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	triePeers, _ := trie.NewTrie(storageManagerPeer, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	triesHolder.Put([]byte(trieFactory.UserAccountTrie), trieUsers)
	triesHolder.Put([]byte(trieFactory.PeerAccountTrie), triePeers)

	stateComponentsFactoryArgs := factory.StateComponentsFactoryArgs{
		Config: config.Config{
			AddressPubkeyConverter: config.PubkeyConfig{
				Length:          32,
				Type:            "hex",
				SignatureLength: 0,
			},
			ValidatorPubkeyConverter: config.PubkeyConfig{
				Length:          96,
				Type:            "hex",
				SignatureLength: 0,
			},
			StateTriesConfig: config.StateTriesConfig{
				CheckpointRoundsModulus:     5,
				AccountsStatePruningEnabled: true,
				PeerStatePruningEnabled:     true,
				MaxStateTrieLevelInMemory:   5,
				MaxPeerTrieLevelInMemory:    5,
			},
			EvictionWaitingList: config.EvictionWaitingListConfig{
				Size: 100,
				DB: config.DBConfig{
					FilePath:          "EvictionWaitingList",
					Type:              "MemoryDB",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			TrieSnapshotDB: config.DBConfig{
				FilePath:          "TrieSnapshot",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
			AccountsTrieStorage: config.StorageConfig{
				Cache: config.CacheConfig{
					Capacity: 10000,
					Type:     "LRU",
					Shards:   1,
				},
				DB: config.DBConfig{
					FilePath:          "AccountsTrie/MainDB",
					Type:              "MemoryDB",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			PeerAccountsTrieStorage: config.StorageConfig{
				Cache: config.CacheConfig{
					Capacity: 10000,
					Type:     "LRU",
					Shards:   1,
				},
				DB: config.DBConfig{
					FilePath:          "PeerAccountsTrie/MainDB",
					Type:              "MemoryDB",
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:   1000,
				SnapshotsBufferLen: 10,
				MaxSnapshots:       2,
			},
		},
		ShardCoordinator:    shardCoordinator,
		Core:                coreComponents,
		TriesContainer:      triesHolder,
		TrieStorageManagers: trieStorageManagers,
	}

	return stateComponentsFactoryArgs
}

func getCoreComponents() factory.CoreComponentsHolder {
	coreArgs := getCoreArgs()
	coreComponentsFactory, _ := factory.NewCoreComponentsFactory(coreArgs)
	coreComponents, err := factory.NewManagedCoreComponents(coreComponentsFactory)
	if err != nil {
		fmt.Println("getCoreComponents NewManagedCoreComponents", "error", err.Error())
		return nil
	}
	err = coreComponents.Create()
	if err != nil {
		fmt.Println("getCoreComponents Create", "error", err.Error())
	}
	return coreComponents
}
