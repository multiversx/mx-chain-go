package factory_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStateComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	args.ShardCoordinator = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewStateComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	args.Core = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, errors.ErrNilCoreComponents, err)
}

func TestNewStateComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)

	scf, err := factory.NewStateComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, scf)
}

func TestStateComponentsFactory_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)

	scf, _ := factory.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestStateComponentsFactory_CreateTriesShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)

	scf, _ := factory.NewStateComponentsFactory(args)

	tc, trieStorageManagers, err := scf.CreateTries()
	require.NoError(t, err)
	require.NotNil(t, trieStorageManagers)
	require.NotNil(t, tc)
}

// ------------ Test ManagedCoreComponents --------------------
func TestManagedStateComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	managedStateComponents, err := factory.NewManagedStateComponents(args)
	require.NoError(t, err)
	_ = args.Core.SetInternalMarshalizer(nil)
	err = managedStateComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedStateComponents.AccountsAdapter())
}

func TestManagedStateComponents_Create_ShouldWork(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	managedStateComponents, err := factory.NewManagedStateComponents(args)
	require.NoError(t, err)
	require.Nil(t, managedStateComponents.AccountsAdapter())
	require.Nil(t, managedStateComponents.PeerAccounts())
	require.Nil(t, managedStateComponents.TriesContainer())
	require.Nil(t, managedStateComponents.TrieStorageManagers())

	err = managedStateComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedStateComponents.AccountsAdapter())
	require.NotNil(t, managedStateComponents.PeerAccounts())
	require.NotNil(t, managedStateComponents.TriesContainer())
	require.NotNil(t, managedStateComponents.TrieStorageManagers())
}

func TestManagedStateComponents_Close(t *testing.T) {
	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	managedStateComponents, _ := factory.NewManagedStateComponents(args)
	err := managedStateComponents.Create()
	require.NoError(t, err)

	err = managedStateComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedStateComponents.AccountsAdapter())
}

// ------------ Test CoreComponents --------------------
func TestStateComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	scf, _ := factory.NewStateComponentsFactory(args)

	sc, _ := scf.Create()

	err := sc.Close()
	require.NoError(t, err)
}

func getStateArgs(coreComponents factory.CoreComponentsHolder) factory.StateComponentsFactoryArgs {
	storageManagers := make(map[string]data.StorageManager)
	storageManagers[trieFactory.UserAccountTrie] = &mock.StorageManagerStub{}
	storageManagers[trieFactory.PeerAccountTrie] = &mock.StorageManagerStub{}

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
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		Core:             coreComponents,
		TriesContainer: &mock.TriesHolderStub{
			GetCalled: func(bytes []byte) data.Trie {
				return &mock.TrieStub{}
			},
		},
		TrieStorageManagers: storageManagers,
	}

	return stateComponentsFactoryArgs
}

func getCoreComponents() factory.CoreComponentsHolder {
	coreArgs := getCoreArgs()
	coreComponents, err := factory.NewManagedCoreComponents(factory.CoreComponentsHandlerArgs(coreArgs))
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
