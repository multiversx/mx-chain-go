package estimator

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestEstimatorCreation(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Parallel()

	t.Run("Nil interface", func(t *testing.T) {
		t.Parallel()
		var erie *ExecutionResultInclusionEstimator
		require.True(t, erie.IsInterfaceNil(), "IsInterfaceNil() should return true for nil interface")

		erie = &ExecutionResultInclusionEstimator{}
		require.False(t, erie.IsInterfaceNil(), "IsInterfaceNil() should return false for non-nil interface")
	})

	t.Run("Default config", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{}
		erie := NewExecutionResultInclusionEstimator(cfg, 1000)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
	})

	t.Run("Custom config", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 120, MaxResultsPerBlock: 10}
		erie := NewExecutionResultInclusionEstimator(cfg, 1000)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
	})
}

func TestDecide(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Parallel()

	t.Run("Empty pending", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, 500)
		lastNotarised := &ExecutionResultMetaData{HeaderTimeMs: 1000}
		pending := []ExecutionResultMetaData{}
		currentHdrTsMs := uint64(1000 + 500) // 1000 ms + 500 ms margin
		wantAllowed := 0
		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Accept all items", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110, MaxResultsPerBlock: 0}
		erie := NewExecutionResultInclusionEstimator(cfg, 500)
		lastNotarised := &ExecutionResultMetaData{HeaderTimeMs: 1000}
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 200},
		}
		currentHdrTsMs := uint64(1000 + 500)
		wantAllowed := 2

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Reject second item", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, 500)
		lastNotarised := &ExecutionResultMetaData{HeaderTimeMs: 1000}
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100_000_000},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 999_000_000},
		}
		currentHdrTsMs := uint64(1000 + 200)
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Allow only up to boundary (t_done == t_now)", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, 500)
		lastNotarised := &ExecutionResultMetaData{HeaderTimeMs: 1000}
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100_000_000},
			{HeaderNonce: 2, HeaderTimeMs: 1011, GasUsed: 100_000_000},
			{HeaderNonce: 3, HeaderTimeMs: 1012, GasUsed: 100_000_000},
		}
		currentHdrTsMs := uint64(1010 * 110 / 100)
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Hit MaxResultsPerBlock cap", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110, MaxResultsPerBlock: 1}
		erie := NewExecutionResultInclusionEstimator(cfg, 500)
		lastNotarised := &ExecutionResultMetaData{HeaderTimeMs: 1000}
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 100},
		}
		currentHdrTsMs := uint64(1000 + 500)
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Genesis fallback", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, 1_700_000_000_000)
		var lastNotarised *ExecutionResultMetaData = nil
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: 1_700_000_000_010, GasUsed: 100}, // after genesis
		}
		currentHdrTsMs := uint64(1_700_000_000_010)
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		require.Equal(t, wantAllowed, got, "Decide() should allow %d result on genesis, got %d", wantAllowed, got)
	})

	t.Run("Overflow protection", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110, MaxResultsPerBlock: 10}
		erie := NewExecutionResultInclusionEstimator(cfg, 500)
		lastNotarised := &ExecutionResultMetaData{HeaderTimeMs: 1000}
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: 1000, GasUsed: 1000},
			{HeaderNonce: 2, HeaderTimeMs: 2000, GasUsed: math.MaxUint64},
			{HeaderNonce: 3, HeaderTimeMs: 3000, GasUsed: 100},
		}
		currentHdrTsMs := uint64(1_700_000_000_000+1000) * 1_000_000
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})
}

func TestOverflowProtection(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Run("gasUsed * t_gas overflows", func(t *testing.T) {
		cfg := config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
		}
		erie := NewExecutionResultInclusionEstimator(cfg, 0)

		pending := []ExecutionResultMetaData{
			{GasUsed: 1000, HeaderTimeMs: 1000, HeaderNonce: 1},
			{GasUsed: math.MaxUint64, HeaderTimeMs: 2000, HeaderNonce: 2}, // This will cause overflow
		}

		currentTime := uint64(1<<63 - 1)

		num_accepted := erie.Decide(nil, pending, currentTime)
		t.Log("num_accepted:", num_accepted)
		require.Equal(t, 1, num_accepted, "should only accept first result, then overflow")
	})

}

func TestDecide_EdgeCases(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Parallel()
	cfg := config.ExecutionResultInclusionEstimatorConfig{
		SafetyMargin:       10,
		MaxResultsPerBlock: 10,
	}
	erie := NewExecutionResultInclusionEstimator(cfg, 1000)
	now := uint64(2000)

	t.Run("zero GasUsed", func(t *testing.T) {
		t.Parallel()
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, GasUsed: 0, HeaderTimeMs: 1100},
			{HeaderNonce: 2, GasUsed: 0, HeaderTimeMs: 1200},
		}

		// Ensure currentHdrTsMs is after all pending headers
		currentHdrTsMs := convertMsToNs(1300) // > 1200

		got := erie.Decide(nil, pending, currentHdrTsMs)
		require.Equal(t, 2, got)
	})

	t.Run("HeaderTime before genesis", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, 1_700_000_000_000)
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, HeaderTimeMs: erie.genesisTimeMs - 1, GasUsed: 100},
		}
		got := erie.Decide(nil, pending, erie.genesisTimeMs+1000)
		require.Equal(t, 0, got)
	})

	t.Run("HeaderTime before last notarised", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, 1_700_000_000_000)
		lastNotarised := &ExecutionResultMetaData{HeaderNonce: 1, HeaderTimeMs: erie.genesisTimeMs + 1000, GasUsed: 100}
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 2, HeaderTimeMs: lastNotarised.HeaderTimeMs - 1, GasUsed: 50},
		}
		got := erie.Decide(lastNotarised, pending, lastNotarised.HeaderTimeMs+1000)
		require.Equal(t, 0, got)
	})

	t.Run("large HeaderTimeMs gap", func(t *testing.T) {
		t.Parallel()
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, GasUsed: 100, HeaderTimeMs: 1000},
			{HeaderNonce: 2, GasUsed: 200, HeaderTimeMs: 10_000}, // large gap
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, 1, got) // should stop at 2nd
	})

	t.Run("all results after t_now", func(t *testing.T) {
		t.Parallel()
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, GasUsed: 10_000, HeaderTimeMs: 3000},
			{HeaderNonce: 2, GasUsed: 20_000, HeaderTimeMs: 4000},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, 0, got)
	})

	t.Run("non-monotonic in time", func(t *testing.T) {
		t.Parallel()
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, GasUsed: 50, HeaderTimeMs: 1100},
			{HeaderNonce: 2, GasUsed: 20, HeaderTimeMs: 1050},
			{HeaderNonce: 3, GasUsed: 70, HeaderTimeMs: 1150},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, got, 1)
	})

	t.Run("non-monotonic in hash", func(t *testing.T) {
		t.Parallel()
		pending := []ExecutionResultMetaData{
			{HeaderNonce: 1, GasUsed: 10, HeaderTimeMs: 1100},
			{HeaderNonce: 2, GasUsed: 20, HeaderTimeMs: 1125},
			{HeaderNonce: 4, GasUsed: 70, HeaderTimeMs: 1150},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, got, 2)

		pending = []ExecutionResultMetaData{
			{HeaderNonce: 2, GasUsed: 10, HeaderTimeMs: 1100},
			{HeaderNonce: 1, GasUsed: 20, HeaderTimeMs: 1125},
		}
		got = erie.Decide(nil, pending, now)
		require.Equal(t, got, 1)
	})
}

func BenchmarkDecideScaling_10(b *testing.B) {
	cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
	erie := NewExecutionResultInclusionEstimator(cfg, 0)
	last := &ExecutionResultMetaData{HeaderTimeMs: 1}

	b.ReportAllocs()

	for n := range 10 {
		pendingSize := 12 * (1 << n) // 12, 24, 48, ...
		pending := make([]ExecutionResultMetaData, pendingSize)
		for i := range pending {
			pending[i] = ExecutionResultMetaData{
				HeaderNonce:  uint64(i + 1),
				HeaderTimeMs: last.HeaderTimeMs + uint64(i),
				GasUsed:      1,
			}
		}
		now := uint64(10000)

		b.Run(fmt.Sprintf("%d_results", pendingSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				erie.Decide(last, pending, now)
			}
		})
	}
}
