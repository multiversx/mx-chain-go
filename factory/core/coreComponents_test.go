package core_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	coreComp "github.com/ElrondNetwork/elrond-go/factory/core"
	"github.com/ElrondNetwork/elrond-go/state"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewCoreComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	require.NotNil(t, ccf)
}

func TestCoreComponentsFactory_CreateCoreComponentsNoHasherConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidHasherConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrHasherCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoInternalMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCoreArgs()
	args.Config = config.Config{
		Hasher: config.TypeConfig{
			Type: componentsMock.TestHasher,
		},
	}
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidInternalMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoVmMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidVmMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsNoTxSignMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidTxSignMarshallerConfigShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

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
	require.True(t, errors.Is(err, errorsErd.ErrMarshalizerCreation))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidValPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCoreArgs()
	args.Config.ValidatorPubkeyConverter.Type = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsInvalidAddrPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCoreArgs()
	args.Config.AddressPubkeyConverter.Type = "invalid"
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, state.ErrInvalidPubkeyConverterType))
}

func TestCoreComponentsFactory_CreateCoreComponentsShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

// ------------ Test CoreComponents --------------------
func TestCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetCoreArgs()
	ccf, _ := coreComp.NewCoreComponentsFactory(args)
	cc, _ := ccf.Create()
	err := cc.Close()

	require.NoError(t, err)
}
