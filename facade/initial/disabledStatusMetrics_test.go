package initial

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

	cfg, err := dsm.ConfigMetrics()
	require.Empty(t, cfg)
	require.Equal(t, errNodeStarting, err)

	ecCfg, err := dsm.EconomicsMetrics()
	require.Empty(t, ecCfg)
	require.Equal(t, errNodeStarting, err)

	nMetrics, err := dsm.NetworkMetrics()
	require.Empty(t, nMetrics)
	require.Equal(t, errNodeStarting, err)

	stMetrics, err := dsm.StatusMetricsMapWithoutP2P()
	require.Empty(t, stMetrics)
	require.Equal(t, errNodeStarting, err)

	p2pMetrics, err := dsm.StatusP2pMetricsMap()
	require.Empty(t, p2pMetrics)
	require.Equal(t, errNodeStarting, err)

	enableMetrics, err := dsm.EnableEpochsMetrics()
	require.Empty(t, enableMetrics)
	require.Equal(t, errNodeStarting, err)

	promString, err := dsm.StatusMetricsWithoutP2PPrometheusString()
	require.Empty(t, promString)
	require.Equal(t, errNodeStarting, err)
}
