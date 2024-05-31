package core_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	"github.com/multiversx/mx-chain-go/state"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	require.NotNil(t, ccf)
}

func TestCoreComponentsFactory_CreateCoreComponentsNoHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: "invalid_type",
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoInternalMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidInternalMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           "invalid_marshalizer_type",
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoVmMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidVmMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoTxSignMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		VmMarshalizer: config.TypeConfig{
			Type: componentsMock.TestMarshalizer,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidTxSignMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Marshalizer: config.MarshalizerConfig{
			Type:           componentsMock.TestMarshalizer,
			SizeCheckDelta: 0,
		},
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
		VmMarshalizer: config.TypeConfig{
			Type: componentsMock.TestMarshalizer,
		},
		TxSignMarshalizer: config.TypeConfig{
			Type: "invalid",
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidTxSignHasherConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.TxSignHasher = config.TypeConfig{
		Type: "invalid",
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsMx.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidValPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.ValidatorPubkeyConverter.Type = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidAddrPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.AddressPubkeyConverter.Type = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsNilChanStopNodeProcessShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.ChanStopNodeProcess = nil
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidRoundConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.RoundConfig = config.RoundConfig{}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidGenesisMaxNumberOfShardsShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.GeneralSettings.GenesisMaxNumberOfShards = 0
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidEconomicsConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.EconomicsConfig = config.EconomicsConfig{}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidRatingsConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.RatingsConfig = config.RatingsConfig{}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidHardforkPubKeyShouldErr(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.Hardfork.PublicKeyToListenFrom = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.NotNil(t, err)
}

func TestCoreComponentsFactory_CreateCoreComponentsShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestCoreComponentsFactory_CreateCoreComponentsShouldWorkAfterHardfork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	args.Config.Hardfork.AfterHardFork = true
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

// ------------ Test CoreComponents --------------------
func TestCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)
	cc, _ := ccf.Create()
	err := cc.Close()

	require.NoError(t, err)
}
