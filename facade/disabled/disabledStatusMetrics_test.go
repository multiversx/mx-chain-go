package disabled

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDisabledStatusMetricsHandler(t *testing.T) {
	t.Parallel()

	dsm := NewDisabledStatusMetricsHandler()
	require.NotNil(t, dsm)
}

func TestDisabledStatusMetricsHandler_AllMethods(t *testing.T) {
	t.Parallel()

	dsm := NewDisabledStatusMetricsHandler()
	expectedMap := getReturnMap()
	require.Equal(t, expectedMap, dsm.ConfigMetrics())
	require.Equal(t, expectedMap, dsm.EconomicsMetrics())
	require.Equal(t, expectedMap, dsm.NetworkMetrics())
	require.Equal(t, expectedMap, dsm.StatusMetricsMapWithoutP2P())
	require.Equal(t, expectedMap, dsm.StatusP2pMetricsMap())
	require.Equal(t, expectedMap, dsm.EnableEpochMetrics())
	require.Equal(t, responseValue, dsm.StatusMetricsWithoutP2PPrometheusString())
}
