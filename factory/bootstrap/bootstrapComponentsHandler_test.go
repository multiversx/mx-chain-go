package bootstrap_test

import (
	"errors"
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/bootstrap"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedBootstrapComponents --------------------
func TestNewManagedBootstrapComponents(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, err := bootstrap.NewManagedBootstrapComponents(bcf)

	require.NotNil(t, mbc)
	require.Nil(t, err)
}

func TestNewBootstrapComponentsFactory_NilFactory(t *testing.T) {
	t.Parallel()

	mbc, err := bootstrap.NewManagedBootstrapComponents(nil)

	require.Nil(t, mbc)
	require.Equal(t, errorsMx.ErrNilBootstrapComponentsFactory, err)
}

func TestManagedBootstrapComponents_MethodsNoCreate(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)
	err := mbc.CheckSubcomponents()
	require.Equal(t, errorsMx.ErrNilBootstrapComponentsHolder, err)

	assert.Nil(t, mbc.EpochStartBootstrapper())
	assert.Nil(t, mbc.EpochBootstrapParams())
	assert.Nil(t, mbc.Close())
	assert.Equal(t, factory.BootstrapComponentsName, mbc.String())
}

func TestManagedBootstrapComponents_MethodsCreate(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)

	err := mbc.Create()
	require.Nil(t, err)

	err = mbc.CheckSubcomponents()
	require.Nil(t, err)

	assert.NotNil(t, mbc.EpochStartBootstrapper())
	params := mbc.EpochBootstrapParams()
	require.NotNil(t, mbc)
	assert.Equal(t, uint32(0), params.Epoch())
	assert.Equal(t, uint32(0), params.SelfShardID())
	assert.Equal(t, uint32(2), params.NumOfShards())
	assert.Nil(t, params.NodesConfig())

	assert.Nil(t, mbc.Close())
	assert.Equal(t, factory.BootstrapComponentsName, mbc.String())
}

func TestManagedBootstrapComponents_CreateNilInternalMarshalizer(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetBootStrapFactoryArgs()
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)
	coreComponents.IntMarsh = nil

	err := mbc.Create()
	require.True(t, errors.Is(err, errorsMx.ErrBootstrapDataComponentsFactoryCreate))
}

func TestManagedBootstrapComponents_Close(t *testing.T) {
	t.Parallel()

	args := componentsMock.GetBootStrapFactoryArgs()

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ := bootstrap.NewManagedBootstrapComponents(bcf)

	_ = mbc.Create()
	require.NotNil(t, mbc.EpochBootstrapParams())

	_ = mbc.Close()
	require.Nil(t, mbc.EpochBootstrapParams())
}

func TestManagedBootstrapComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	mbc, _ := bootstrap.NewManagedBootstrapComponents(nil)
	require.True(t, mbc.IsInterfaceNil())

	args := componentsMock.GetBootStrapFactoryArgs()
	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
	mbc, _ = bootstrap.NewManagedBootstrapComponents(bcf)
	require.False(t, mbc.IsInterfaceNil())
}
