package statusCore_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/statusCore"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/factory"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCoreComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		args := componentsMock.GetStatusCoreArgs(testscommon.GetGeneralConfig(), nil)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errorsMx.ErrNilCoreComponents, err)
		require.Nil(t, sccf)
	})
	t.Run("nil economics data should error", func(t *testing.T) {
		t.Parallel()

		coreComp := &mock.CoreComponentsStub{
			EconomicsDataField: nil,
		}

		args := componentsMock.GetStatusCoreArgs(testscommon.GetGeneralConfig(), coreComp)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Equal(t, errorsMx.ErrNilEconomicsData, err)
		require.Nil(t, sccf)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := testscommon.GetGeneralConfig()
		args := componentsMock.GetStatusCoreArgs(cfg, componentsMock.GetCoreComponents(cfg))
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		assert.Nil(t, err)
		require.NotNil(t, sccf)
	})
}

func TestStatusCoreComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("NewResourceMonitor fails should error", func(t *testing.T) {
		t.Parallel()

		cfg := testscommon.GetGeneralConfig()
		coreComp := componentsMock.GetCoreComponents(cfg)
		args := componentsMock.GetStatusCoreArgs(cfg, coreComp)
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
	})
	t.Run("NewPersistentStatusHandler fails should error", func(t *testing.T) {
		t.Parallel()

		cfg := testscommon.GetGeneralConfig()
		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(componentsMock.GetCoreComponents(cfg))
		coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
			return nil
		}
		args := componentsMock.GetStatusCoreArgs(cfg, coreCompStub)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.Error(t, err)
		require.Nil(t, cc)
	})
	t.Run("SetStatusHandler fails should error", func(t *testing.T) {
		t.Parallel()

		cfg := testscommon.GetGeneralConfig()
		expectedErr := errors.New("expected error")
		coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(componentsMock.GetCoreComponents(cfg))
		coreCompStub.EconomicsDataCalled = func() process.EconomicsDataHandler {
			return &economicsmocks.EconomicsHandlerStub{
				SetStatusHandlerCalled: func(statusHandler core.AppStatusHandler) error {
					return expectedErr
				},
			}
		}
		args := componentsMock.GetStatusCoreArgs(cfg, coreCompStub)
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.Equal(t, expectedErr, err)
		require.Nil(t, cc)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := testscommon.GetGeneralConfig()
		args := componentsMock.GetStatusCoreArgs(cfg, componentsMock.GetCoreComponents(cfg))
		args.Config.ResourceStats.Enabled = true // coverage
		sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
		require.Nil(t, err)

		cc, err := sccf.Create()
		require.NoError(t, err)
		require.NotNil(t, cc)
		require.NoError(t, cc.Close())
	})
}

// ------------ Test CoreComponents --------------------
func TestStatusCoreComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()

	cfg := testscommon.GetGeneralConfig()
	coreComp := componentsMock.GetCoreComponents(cfg)
	args := componentsMock.GetStatusCoreArgs(cfg, coreComp)
	sccf, err := statusCore.NewStatusCoreComponentsFactory(args)
	require.Nil(t, err)
	cc, err := sccf.Create()
	require.NoError(t, err)

	err = cc.Close()
	require.NoError(t, err)
}
