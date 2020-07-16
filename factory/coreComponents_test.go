package factory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

const testHasher = "blake2b"
const testMarshalizer = "json"

func TestNewCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf := factory.NewCoreComponentsFactory(args)

	require.NotNil(t, ccf)
}

func TestCoreComponentsFactory_CreateCoreComponents_NoHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: "invalid_type",
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoInternalMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidInternalMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           "invalid_marshalizer_type",
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoVmMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidVmMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_NoTxSignMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		VmMarshalizer: config.TypeConfig{
			Type: testMarshalizer,
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponents_InvalidTxSignMarshalizerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           testMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: testHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: testMarshalizer,
		},
		TxSignMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, factory.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidValPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config.ValidatorPubkeyConverter.Type = "invalid"
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidAddrPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	args.Config.AddressPubkeyConverter.Type = "invalid"
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponents_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf := factory.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCoreComponentsFactory_CreateStorerTemplatePaths(t *testing.T) {
	t.Parallel()

	args := getCoreArgs()
	ccf := factory.NewCoreComponentsFactory(args)

	pathPruning, pathStatic := ccf.CreateStorerTemplatePaths()
	require.Equal(t, "home/db/undefined/Epoch_[E]/Shard_[S]/[I]", pathPruning)
	require.Equal(t, "home/db/undefined/Static/Shard_[S]/[I]", pathStatic)
}

func getCoreArgs() factory.CoreComponentsFactoryArgs {
	return factory.CoreComponentsFactoryArgs{
		Config: config.Config{
			PublicKeyPeerId: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			PublicKeyShardId: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			PeerIdShardId: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			PeerHonesty: config.CacheConfig{
				Type:     "LRU",
				Capacity: 5000,
				Shards:   16,
			},
			GeneralSettings: config.GeneralSettingsConfig{
				ChainID:               "undefined",
				MinTransactionVersion: 1,
			},
			Marshalizer: config.MarshalizerConfig{
				Type:           testMarshalizer,
				SizeCheckDelta: 0,
			},
			Hasher: config.TypeConfig{
				Type: testHasher,
			},
			VmMarshalizer: config.TypeConfig{
				Type: testMarshalizer,
			},
			TxSignMarshalizer: config.TypeConfig{
				Type: testMarshalizer,
			},
			AddressPubkeyConverter: config.PubkeyConfig{
				Length:          32,
				Type:            "bech32",
				SignatureLength: 0,
			},
			ValidatorPubkeyConverter: config.PubkeyConfig{
				Length:          96,
				Type:            "hex",
				SignatureLength: 48,
			},
			Consensus: config.ConsensusConfig{
				Type: "bls",
			},
			ValidatorStatistics: config.ValidatorStatisticsConfig{
				CacheRefreshIntervalInSec: uint32(100),
			},
			SoftwareVersionConfig: config.SoftwareVersionConfig{
				PollingIntervalInMinutes: 30,
			},
		},
		WorkingDirectory:    "home",
		GenesisTime:         time.Now(),
		ChanStopNodeProcess: make(chan endProcess.ArgEndProcess),
	}
}
