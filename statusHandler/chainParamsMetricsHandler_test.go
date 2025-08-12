package statusHandler_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/statusHandler"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func TestNewChainParamsMetricsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil app status handler", func(t *testing.T) {
		t.Parallel()

		cpm, err := statusHandler.NewChainParamsMetricsHandler(nil)
		require.Nil(t, cpm)
		require.Equal(t, statusHandler.ErrNilAppStatusHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cpm, err := statusHandler.NewChainParamsMetricsHandler(&statusHandlerMock.AppStatusHandlerStub{})
		require.Nil(t, err)
		require.False(t, cpm.IsInterfaceNil())
	})
}

func TestChainParamsMetricsHandler_ChainParametersChanged(t *testing.T) {
	t.Parallel()

	numCalls := 0
	cpm, _ := statusHandler.NewChainParamsMetricsHandler(&statusHandlerMock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			numCalls++
		},
	})

	cpm.ChainParametersChanged(config.ChainParametersByEpochConfig{})

	require.Equal(t, 6, numCalls)
}
