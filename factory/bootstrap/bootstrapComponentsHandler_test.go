package bootstrap_test

import (
	"errors"
	"testing"

	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory/bootstrap"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedBootstrapComponents --------------------
func TestNewManagedBootstrapComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, err := bootstrap.NewManagedBootstrapComponents(bcf)

	require.NotNil(t, mbc)
	require.Nil(t, err)
}

func TestNewBootstrapComponentsFactory_NilFactory(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	mbc, err := bootstrap.NewManagedBootstrapComponents(nil)

	require.Nil(t, mbc)
	require.Equal(t, errorsErd.ErrNilBootstrapComponentsFactory, err)
}

func TestManagedBootstrapComponents_CheckSubcomponentsNoCreate(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)
	err := mbc.CheckSubcomponents()

	require.Equal(t, errorsErd.ErrNilBootstrapComponentsHolder, err)
}

func TestManagedBootstrapComponents_Create(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)

	err := mbc.Create()
	require.Nil(t, err)

	err = mbc.CheckSubcomponents()
	require.Nil(t, err)
}

func TestManagedBootstrapComponents_CreateNilInternalMarshalizer(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)
	coreComponents.IntMarsh = nil

	err := mbc.Create()
	require.True(t, errors.Is(err, errorsErd.ErrBootstrapDataComponentsFactoryCreate))
}

func TestManagedBootstrapComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)

	_ = mbc.Create()
	require.NotNil(t, mbc.EpochBootstrapParams())

	_ = mbc.Close()
	require.Nil(t, mbc.EpochBootstrapParams())
}
