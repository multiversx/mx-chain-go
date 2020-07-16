package factory_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
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

// ------------ Test ManagedCoreComponents --------------------
func TestManagedCoreComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
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

func TestManagedCoreComponents_Create_ShouldWork(t *testing.T) {
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

func TestManagedCoreComponents_Close(t *testing.T) {
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
func TestCoreComponents_Close_ShouldWork(t *testing.T) {
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

func getCoreArgs() factory.CoreComponentsFactoryArgs {
	return factory.CoreComponentsFactoryArgs{
		Config: config.Config{
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
		WorkingDirectory: "home",
	}
}
