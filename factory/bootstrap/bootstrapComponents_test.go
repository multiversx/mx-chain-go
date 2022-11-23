package bootstrap_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory/bootstrap"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------ Test BootstrapComponentsFactory --------------------
func TestNewBootstrapComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()

	bcf, err := bootstrap.NewBootstrapComponentsFactory(args)

	require.NotNil(t, bcf)
	require.Nil(t, err)
}

func TestNewBootstrapComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	args.CoreComponents = nil

	bcf, err := bootstrap.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCoreComponentsHolder, err)
}

func TestNewBootstrapComponentsFactory_NilCryptoComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	args.CryptoComponents = nil

	bcf, err := bootstrap.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCryptoComponentsHolder, err)
}

func TestNewBootstrapComponentsFactory_NilNetworkComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	args.NetworkComponents = nil

	bcf, err := bootstrap.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilNetworkComponentsHolder, err)
}

func TestNewBootstrapComponentsFactory_NilWorkingDir(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	args.WorkingDir = ""

	bcf, err := bootstrap.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrInvalidWorkingDir, err)
}

func TestBootstrapComponentsFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)

	bc, err := bcf.Create()

	require.Nil(t, err)
	require.NotNil(t, bc)
}

func TestBootstrapComponentsFactory_CreateBootstrapDataProviderCreationFail(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)

	coreComponents.IntMarsh = nil
	bc, err := bcf.Create()

	require.Nil(t, bc)
	require.True(t, errors.Is(err, errorsErd.ErrNewBootstrapDataProvider))
}

func TestBootstrapComponentsFactory_CreateEpochStartBootstrapCreationFail(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetBootStrapFactoryArgs()
	coreComponents := componentsMock.GetDefaultCoreComponents()
	args.CoreComponents = coreComponents

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)

	coreComponents.RatingHandler = nil
	bc, err := bcf.Create()

	require.Nil(t, bc)
	require.True(t, errors.Is(err, errorsErd.ErrNewEpochStartBootstrap))
}

func TestBootstrapComponentsFactory_CreateEpochStartBootstrapperShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should create a epoch start bootstrapper main chain instance", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetBootStrapFactoryArgs()
		args.ChainRunType = common.ChainRunTypeRegular

		bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
		bc, err := bcf.Create()

		require.NotNil(t, bc)
		assert.Nil(t, err)
	})

	t.Run("should create a epoch start bootstrapper sovereign chain instance", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetBootStrapFactoryArgs()
		args.ChainRunType = common.ChainRunTypeSovereign

		bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
		bc, err := bcf.Create()

		require.NotNil(t, bc)
		assert.Nil(t, err)
	})

	t.Run("should error when chain run type is not implemented", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetBootStrapFactoryArgs()
		args.ChainRunType = "X"

		bcf, _ := bootstrap.NewBootstrapComponentsFactory(args)
		bc, err := bcf.Create()

		assert.Nil(t, bc)
		require.True(t, errors.Is(err, errorsErd.ErrUnimplementedChainRunType))
	})
}
