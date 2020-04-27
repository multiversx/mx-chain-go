package factory_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

const okHasher = "blake2b"
const okMarshalizer = "json"

func TestNewCoreComponentsFactory_NilConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = nil
	ccf, err := factory.NewCoreComponentsFactory(args)

	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilConfiguration, err)
}

func TestNewCoreComponentsFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.PathManager = nil
	ccf, err := factory.NewCoreComponentsFactory(args)

	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilPathManager, err)
}

func TestNewCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	ccf, err := factory.NewCoreComponentsFactory(args)

	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestCoreComponentsFactory_CreateCoreComponents_NoHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           okMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           okMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: "invalid_type",
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoInternalMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Hasher: config.TypeConfig{
			Type: okHasher,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidInternalMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           "invalid_marshalizer_type",
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: okHasher,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoVmMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Hasher: config.TypeConfig{
			Type: okHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           okMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidVmMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           okMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: okHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoTxSignMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Hasher: config.TypeConfig{
			Type: okHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           okMarshalizer,
			SizeCheckDelta: 0,
		},
		VmMarshalizer: config.TypeConfig{
			Type: okMarshalizer,
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidTxSignMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config = &config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           okMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: okHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: okMarshalizer,
		},
		TxSignMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidPeerAccountsTrieStorageShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config.PeerAccountsTrieStorage = config.StorageConfig{}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.Error(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidAccountsTrieStorageeShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Config.AccountsTrieStorage = config.StorageConfig{}
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.Error(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponents_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	ccf, _ := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func getArgs() factory.CoreComponentsFactoryArgs {
	return factory.CoreComponentsFactoryArgs{
		Config: &config.Config{
			Marshalizer: config.MarshalizerConfig{
				Type:           okMarshalizer,
				SizeCheckDelta: 0,
			},
			Hasher: config.TypeConfig{
				Type: okHasher,
			},
			VmMarshalizer: config.TypeConfig{
				Type: okMarshalizer,
			},
			TxSignMarshalizer: config.TypeConfig{
				Type: okMarshalizer,
			},
			EvictionWaitingList: config.EvictionWaitingListConfig{
				Size: 10,
				DB:   getDBCfg(),
			},
			TrieSnapshotDB: getDBCfg(),
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:   10,
				SnapshotsBufferLen: 10,
				MaxSnapshots:       10,
			},
			AccountsTrieStorage: config.StorageConfig{
				Cache: getCacheCfg(),
				DB:    getDBCfg(),
				Bloom: config.BloomFilterConfig{},
			},
			PeerAccountsTrieStorage: config.StorageConfig{
				Cache: getCacheCfg(),
				DB:    getDBCfg(),
				Bloom: config.BloomFilterConfig{},
			},
			StateTriesConfig: config.StateTriesConfig{
				CheckpointRoundsModulus:     5,
				AccountsStatePruningEnabled: false,
				PeerStatePruningEnabled:     false,
			},
		},
		PathManager: &mock.PathManagerStub{},
		ShardId:     "0",
		ChainID:     []byte("chainID"),
	}
}

func getCacheCfg() config.CacheConfig {
	return config.CacheConfig{
		Type:        "LRU",
		Size:        10,
		SizeInBytes: 10,
		Shards:      1,
	}
}

func getDBCfg() config.DBConfig {
	return config.DBConfig{
		FilePath:          "",
		Type:              string(storageUnit.MemoryDB),
		BatchDelaySeconds: 10,
		MaxBatchSize:      10,
		MaxOpenFiles:      10,
	}
}
