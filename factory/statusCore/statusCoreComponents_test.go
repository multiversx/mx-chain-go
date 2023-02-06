package statusCore_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	errErd "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCoreComponentsFactory(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(nil)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errErd.ErrNilCoreComponents, err)
		require.Nil(t, sccf)
	})
	t.Run("nil economics data should error", func(t *testing.T) {
		t.Parallel()

		coreComp := &mock.CoreComponentsStub{
			EconomicsDataField: nil,
		}

		args := componentsMock.GetStatusCoreArgs(coreComp)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errErd.ErrNilEconomicsData, err)
		require.Nil(t, sccf)
	})
	t.Run("nil genesis node setup should error", func(t *testing.T) {
		t.Parallel()

		coreComp := &mock.CoreComponentsStub{
			EconomicsDataField:     &economicsmocks.EconomicsHandlerStub{},
			GenesisNodesSetupField: nil,
		}

		args := componentsMock.GetStatusCoreArgs(coreComp)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errErd.ErrNilGenesisNodesSetupHandler, err)
		require.Nil(t, sccf)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		coreComp := &mock.CoreComponentsStub{
			EconomicsDataField:       &economicsmocks.EconomicsHandlerStub{},
			GenesisNodesSetupField:   &testscommon.NodesSetupStub{},
			InternalMarshalizerField: nil,
		}

		args := componentsMock.GetStatusCoreArgs(coreComp)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errErd.ErrNilMarshalizer, err)
		require.Nil(t, sccf)
	})
	t.Run("nil slice converter should error", func(t *testing.T) {
		t.Parallel()

		coreComp := &mock.CoreComponentsStub{
			EconomicsDataField:            &economicsmocks.EconomicsHandlerStub{},
			GenesisNodesSetupField:        &testscommon.NodesSetupStub{},
			InternalMarshalizerField:      &testscommon.MarshalizerStub{},
			Uint64ByteSliceConverterField: nil,
		}

		args := componentsMock.GetStatusCoreArgs(coreComp)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errErd.ErrNilUint64ByteSliceConverter, err)
		require.Nil(t, sccf)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Nil(t, err)
		require.NotNil(t, sccf)
	})
}

func TestStatusCoreComponentsFactory_InvalidValueShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	args.Config = config.Config{
		ResourceStats: config.ResourceStatsConfig{
			RefreshIntervalInSec: 0,
		},
	}
	sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.Nil(t, err)

	cc, err := sccf.Create()
	require.Nil(t, cc)
	require.True(t, errors.Is(err, statistics.ErrInvalidRefreshIntervalValue))
}

func TestStatusCoreComponentsFactory_CreateStatusCoreComponentsShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.Nil(t, err)

	cc, err := sccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

// ------------ Test CoreComponents --------------------
func TestStatusCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	args := componentsMock.GetStatusCoreArgs(componentsMock.GetCoreComponents())
	sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.Nil(t, err)
	cc, err := sccf.Create()
	require.NoError(t, err)

	err = cc.Close()
	require.NoError(t, err)
}
