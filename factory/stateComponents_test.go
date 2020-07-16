package factory_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
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
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewStateComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	args := getStateArgs(coreComponents)
	args.Core = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, factory.ErrNilCoreComponents, err)
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
func TestManagedCStateComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	coreArgs := getCoreArgs()
	coreArgs.Config.Marshalizer = config.MarshalizerConfig{
		Type:           "invalid_marshalizer_type",
		SizeCheckDelta: 0,
	}
	managedCoreComponents, err := factory.NewManagedCoreComponents(factory.CoreComponentsHandlerArgs(coreArgs))
	require.NoError(t, err)
	err = managedCoreComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedCoreComponents.StatusHandler())
}

func TestStateComponents_Create_ShouldWork(t *testing.T) {
	coreArgs := getCoreArgs()
	managedCoreComponents, err := factory.NewManagedCoreComponents(factory.CoreComponentsHandlerArgs(coreArgs))
	require.NoError(t, err)
	require.Nil(t, managedCoreComponents.Hasher())
	require.Nil(t, managedCoreComponents.InternalMarshalizer())
	require.Nil(t, managedCoreComponents.VmMarshalizer())
	require.Nil(t, managedCoreComponents.TxMarshalizer())
	require.Nil(t, managedCoreComponents.Uint64ByteSliceConverter())
	require.Nil(t, managedCoreComponents.AddressPubKeyConverter())
	require.Nil(t, managedCoreComponents.ValidatorPubKeyConverter())
	require.Nil(t, managedCoreComponents.StatusHandler())
	require.Nil(t, managedCoreComponents.PathHandler())
	require.Equal(t, "", managedCoreComponents.ChainID())
	require.Nil(t, managedCoreComponents.AddressPubKeyConverter())

	err = managedCoreComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedCoreComponents.Hasher())
	require.NotNil(t, managedCoreComponents.InternalMarshalizer())
	require.NotNil(t, managedCoreComponents.VmMarshalizer())
	require.NotNil(t, managedCoreComponents.TxMarshalizer())
	require.NotNil(t, managedCoreComponents.Uint64ByteSliceConverter())
	require.NotNil(t, managedCoreComponents.AddressPubKeyConverter())
	require.NotNil(t, managedCoreComponents.ValidatorPubKeyConverter())
	require.NotNil(t, managedCoreComponents.StatusHandler())
	require.NotNil(t, managedCoreComponents.PathHandler())
	require.NotEqual(t, "", managedCoreComponents.ChainID())
	require.NotNil(t, managedCoreComponents.AddressPubKeyConverter())
}

func TestStateComponents_Close(t *testing.T) {
	coreArgs := getCoreArgs()
	managedCoreComponents, _ := factory.NewManagedCoreComponents(factory.CoreComponentsHandlerArgs(coreArgs))
	err := managedCoreComponents.Create()
	require.NoError(t, err)

	closed := false
	statusHandlerMock := &mock.AppStatusHandlerMock{
		CloseCalled: func() {
			closed = true
		},
	}
	_ = managedCoreComponents.SetStatusHandler(statusHandlerMock)
	err = managedCoreComponents.Close()
	require.NoError(t, err)
	require.True(t, closed)
	require.Nil(t, managedCoreComponents.StatusHandler())
}

// ------------ Test CoreComponents --------------------
func TestStateComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf := factory.NewCoreComponentsFactory(args)
	cc, _ := ccf.Create()

	closeCalled := false
	statusHandler := &mock.AppStatusHandlerMock{
		CloseCalled: func() {
			closeCalled = true
		},
	}
	cc.SetStatusHandler(statusHandler)

	err := cc.Close()

	require.NoError(t, err)
	require.True(t, closeCalled)
}

func getStateArgs(coreComponents factory.CoreComponentsHolder) factory.StateComponentsFactoryArgs {
	return factory.StateComponentsFactoryArgs{
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
	}
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
