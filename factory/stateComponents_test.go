package factory_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewStateComponentsFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getStateArgs()
	args.PathManager = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, factory.ErrNilPathManager, err)
}

func TestNewStateComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getStateArgs()
	args.ShardCoordinator = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewStateComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()

	args := getStateArgs()
	args.Core = nil

	scf, err := factory.NewStateComponentsFactory(args)
	require.Nil(t, scf)
	require.Equal(t, factory.ErrNilCoreComponents, err)
}

func TestNewStateComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getStateArgs()

	scf, err := factory.NewStateComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, scf)
}

func TestStateComponentsFactory_Create_InvalidValidatorPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getStateArgs()
	args.Config.ValidatorPubkeyConverter = config.PubkeyConfig{}

	scf, _ := factory.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.True(t, errors.Is(err, factory.ErrPubKeyConverterCreation))
	require.Nil(t, res)
}

func TestStateComponentsFactory_Create_InvalidAddressPubKeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getStateArgs()
	args.Config.AddressPubkeyConverter = config.PubkeyConfig{}

	scf, _ := factory.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.True(t, errors.Is(err, factory.ErrPubKeyConverterCreation))
	require.Nil(t, res)
}

func TestStateComponentsFactory_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getStateArgs()

	scf, _ := factory.NewStateComponentsFactory(args)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

func getStateArgs() factory.StateComponentsFactoryArgs {
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
		},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		PathManager:      &mock.PathManagerStub{},
		Core:             getCoreComponents(),
		Tries:            getTriesComponents(),
	}
}

func getCoreComponents() *factory.CoreComponents {
	ccf := factory.NewCoreComponentsFactory(getCoreArgs())
	cc, _ := ccf.Create()
	return cc
}

func getTriesComponents() *factory.TriesComponents {
	tcf, _ := factory.NewTriesComponentsFactory(getTriesArgs())
	tc, _ := tcf.Create()
	return tc
}
