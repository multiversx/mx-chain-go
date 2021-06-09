package factory_test

import (
	"errors"
	"testing"

	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedBootstrapComponents --------------------
func TestNewManagedBootstrapComponents(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	bcf, _ := factory.NewBootstrapComponentsFactory(args)
	mbc, err := factory.NewManagedBootstrapComponents(bcf)

	require.NotNil(t, mbc)
	require.Nil(t, err)
}

func TestNewBootstrapComponentsFactory_NilFactory(t *testing.T) {
	t.Parallel()

	mbc, err := factory.NewManagedBootstrapComponents(nil)

	require.Nil(t, mbc)
	require.Equal(t, errorsErd.ErrNilBootstrapComponentsFactory, err)
}

func TestManagedBootstrapComponents_CheckSubcomponents_NoCreate(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	bcf, _ := factory.NewBootstrapComponentsFactory(args)
	mbc, _ := factory.NewManagedBootstrapComponents(bcf)
	err := mbc.CheckSubcomponents()

	require.Equal(t, errorsErd.ErrNilBootstrapComponentsHolder, err)
}

func TestManagedBootstrapComponents_Create(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	bcf, _ := factory.NewBootstrapComponentsFactory(args)
	mbc, _ := factory.NewManagedBootstrapComponents(bcf)

	err := mbc.Create()
	require.Nil(t, err)

	err = mbc.CheckSubcomponents()
	require.Nil(t, err)
}

func TestManagedBootstrapComponents_Create_NilInternalMarshalizer(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	coreComponents := getDefaultCoreComponents()
	args.CoreComponents = coreComponents
	bcf, _ := factory.NewBootstrapComponentsFactory(args)
	mbc, _ := factory.NewManagedBootstrapComponents(bcf)
	coreComponents.IntMarsh = nil

	err := mbc.Create()
	require.True(t, errors.Is(err, errorsErd.ErrBootstrapDataComponentsFactoryCreate))
}

func TestManagedBootstrapComponents_Close(t *testing.T) {
	t.Parallel()
	args := getBootStrapArgs()

	bcf, _ := factory.NewBootstrapComponentsFactory(args)
	mbc, _ := factory.NewManagedBootstrapComponents(bcf)

	_ = mbc.Create()
	require.NotNil(t, mbc.EpochBootstrapParams())

	_ = mbc.Close()
	require.Nil(t, mbc.EpochBootstrapParams())
}
