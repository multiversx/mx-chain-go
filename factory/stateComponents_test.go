package factory_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/stretchr/testify/require"
)

func TestNewStateComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)

	scf, err := factory.NewStateComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, scf)
}

func TestStateComponentsFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getStateArgs(coreComponents, shardCoordinator)

	scf, _ := factory.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

// ------------ Test StateComponents --------------------
func TestStateComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[trieFactory.UserAccountTrie] = storageManagerUser
	trieStorageManagers[trieFactory.PeerAccountTrie] = storageManagerPeer

	triesHolder := state.NewDataTriesHolder()
	trieUsers, _ := trie.NewTrie(storageManagerUser, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	triePeers, _ := trie.NewTrie(storageManagerPeer, coreComponents.InternalMarshalizer(), coreComponents.Hasher(), 5)
	triesHolder.Put([]byte(trieFactory.UserAccountTrie), trieUsers)
	triesHolder.Put([]byte(trieFactory.PeerAccountTrie), triePeers)

	stateComponentsFactoryArgs := factory.StateComponentsFactoryArgs{
		Config:           getGeneralConfig(),
		ShardCoordinator: shardCoordinator,
		Core:             coreComponents,
		StorageService:   disabled.NewChainStorer(),
		ProcessingMode:   common.Normal,
		ChainHandler:     &testscommon.ChainHandlerStub{},
	}

	return stateComponentsFactoryArgs
}

func getGeneralConfig() config.Config {
	return config.Config{
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
			HashesSize:     100,
			RootHashesSize: 100,
			DB: config.DBConfig{
				FilePath:          "EvictionWaitingList",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
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
		AccountsTrieCheckpointsStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "AccountsTrieCheckpoints",
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
		PeerAccountsTrieCheckpointsStorage: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
			DB: config.DBConfig{
				FilePath:          "PeerAccountsTrieCheckpoints",
				Type:              "MemoryDB",
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TrieStorageManagerConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:      1000,
			SnapshotsBufferLen:    10,
			SnapshotsGoroutineNum: 1,
		},
		VirtualMachine: config.VirtualMachineServicesConfig{
			Querying: config.QueryVirtualMachineConfig{
				NumConcurrentVMs: 1,
				VirtualMachineConfig: config.VirtualMachineConfig{
					ArwenVersions: []config.ArwenVersionByEpoch{
						{StartEpoch: 0, Version: "v0.3"},
					},
				},
			},
			Execution: config.VirtualMachineConfig{
				ArwenVersions: []config.ArwenVersionByEpoch{
					{StartEpoch: 0, Version: "v0.3"},
				},
			},
			GasConfig: config.VirtualMachineGasConfig{
				ShardMaxGasPerVmQuery: 1_500_000_000,
				MetaMaxGasPerVmQuery:  0,
			},
		},
		SmartContractsStorageForSCQuery: config.StorageConfig{
			Cache: config.CacheConfig{
				Capacity: 10000,
				Type:     "LRU",
				Shards:   1,
			},
		},
		SmartContractDataPool: config.CacheConfig{
			Capacity: 10000,
			Type:     "LRU",
			Shards:   1,
		},
		PeersRatingConfig: config.PeersRatingConfig{
			TopRatedCacheCapacity: 1000,
			BadRatedCacheCapacity: 1000,
		},
		BuiltInFunctions: config.BuiltInFunctionsConfig{
			AutomaticCrawlerAddress:       "erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c",
			MaxNumAddressesInTransferRole: 100,
		},
	}
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
