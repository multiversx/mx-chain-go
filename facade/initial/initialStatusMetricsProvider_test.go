package initial

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewInitialStatusMetricsProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil statusHandler should error", func(t *testing.T) {
		t.Parallel()

		provider, err := NewInitialStatusMetricsProvider(nil)
		assert.Equal(t, facade.ErrNilStatusMetrics, err)
		assert.True(t, check.IfNil(provider))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not have panicked")
			}
		}()

		providedMetrics := map[string]interface{}{
			"key-1": uint64(10),
			"key-2": uint64(15),
			"key-3": uint64(20),
		}
		statusMetricsProvider := &testscommon.StatusMetricsStub{
			BootstrapMetricsCalled: func() (map[string]interface{}, error) {
				return providedMetrics, nil
			},
		}
		provider, err := NewInitialStatusMetricsProvider(statusMetricsProvider)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))

		testDisabledGetter(t, provider.StatusMetricsMapWithoutP2P)
		testDisabledGetter(t, provider.StatusP2pMetricsMap)
		testDisabledGetter(t, provider.EconomicsMetrics)
		testDisabledGetter(t, provider.ConfigMetrics)
		testDisabledGetter(t, provider.EnableEpochsMetrics)
		testDisabledGetter(t, provider.NetworkMetrics)
		testDisabledGetter(t, provider.RatingsMetrics)

		metrics, err := provider.StatusMetricsWithoutP2PPrometheusString()
		assert.Equal(t, errNodeStarting, err)
		assert.Equal(t, "", metrics)

		bootstrapMetrics, err := provider.BootstrapMetrics()
		assert.Nil(t, err)
		assert.Equal(t, providedMetrics, bootstrapMetrics)
	})
}

func testDisabledGetter(t *testing.T, getter func() (map[string]interface{}, error)) {
	metrics, err := getter()
	assert.Equal(t, errNodeStarting, err)
	assert.Equal(t, map[string]interface{}{}, metrics)
}
