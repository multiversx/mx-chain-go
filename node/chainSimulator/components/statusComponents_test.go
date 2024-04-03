package components

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	mxErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func TestCreateStatusComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateStatusComponents(0, &statusHandler.AppStatusHandlerStub{}, 5, config.ExternalConfig{})
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateStatusComponents(0, nil, 5, config.ExternalConfig{})
		require.Equal(t, core.ErrNilAppStatusHandler, err)
		require.Nil(t, comp)
	})
}

func TestStatusComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *statusComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateStatusComponents(0, &statusHandler.AppStatusHandlerStub{}, 5, config.ExternalConfig{})
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestStatusComponentsHolder_Getters(t *testing.T) {
	t.Parallel()

	comp, err := CreateStatusComponents(0, &statusHandler.AppStatusHandlerStub{}, 5, config.ExternalConfig{})
	require.NoError(t, err)

	require.NotNil(t, comp.OutportHandler())
	require.NotNil(t, comp.SoftwareVersionChecker())
	require.NotNil(t, comp.ManagedPeersMonitor())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())

	require.Nil(t, comp.Close())
}
func TestStatusComponentsHolder_SetForkDetector(t *testing.T) {
	t.Parallel()

	comp, err := CreateStatusComponents(0, &statusHandler.AppStatusHandlerStub{}, 5, config.ExternalConfig{})
	require.NoError(t, err)

	err = comp.SetForkDetector(nil)
	require.Equal(t, process.ErrNilForkDetector, err)

	err = comp.SetForkDetector(&mock.ForkDetectorStub{})
	require.NoError(t, err)

	require.Nil(t, comp.Close())
}

func TestStatusComponentsHolder_StartPolling(t *testing.T) {
	t.Parallel()

	t.Run("nil fork detector should error", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateStatusComponents(0, &statusHandler.AppStatusHandlerStub{}, 5, config.ExternalConfig{})
		require.NoError(t, err)

		err = comp.StartPolling()
		require.Equal(t, process.ErrNilForkDetector, err)
	})
	t.Run("NewAppStatusPolling failure should error", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateStatusComponents(0, &statusHandler.AppStatusHandlerStub{}, 0, config.ExternalConfig{})
		require.NoError(t, err)

		err = comp.SetForkDetector(&mock.ForkDetectorStub{})
		require.NoError(t, err)

		err = comp.StartPolling()
		require.Equal(t, mxErrors.ErrStatusPollingInit, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHighestNonce := uint64(123)
		providedStatusPollingIntervalSec := 1
		wasSetUInt64ValueCalled := atomic.Flag{}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				require.Equal(t, common.MetricProbableHighestNonce, key)
				require.Equal(t, providedHighestNonce, value)
				wasSetUInt64ValueCalled.SetValue(true)
			},
		}
		comp, err := CreateStatusComponents(0, appStatusHandler, providedStatusPollingIntervalSec, config.ExternalConfig{})
		require.NoError(t, err)

		forkDetector := &mock.ForkDetectorStub{
			ProbableHighestNonceCalled: func() uint64 {
				return providedHighestNonce
			},
		}
		err = comp.SetForkDetector(forkDetector)
		require.NoError(t, err)

		err = comp.StartPolling()
		require.NoError(t, err)

		time.Sleep(time.Duration(providedStatusPollingIntervalSec+1) * time.Second)
		require.True(t, wasSetUInt64ValueCalled.IsSet())

		require.Nil(t, comp.Close())
	})
}
