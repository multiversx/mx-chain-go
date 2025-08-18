package estimator

import (
	"fmt"
	"math"
	"testing"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestEstimatorCreation(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Parallel()

	t.Run("Default config", func(t *testing.T) {
		cfg := Config{}
		erie := NewExecutionResultInclusionEstimator(cfg)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
	})

	t.Run("Custom config", func(t *testing.T) {
		cfg := Config{SafetyMargin: 120, MaxResultsPerBlock: 10}
		erie := NewExecutionResultInclusionEstimator(cfg)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
	})
	t.Run("SetTimePerGasUnit", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg)
		erie.SetTimePerGasUnit(2)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
		require.Equal(t, uint64(2), erie.tGas, "tGas should be set to 2")
	})
	t.Run("SetTimePerGasUnit with zero value", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg)
		erie.SetTimePerGasUnit(0)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
		require.Equal(t, uint64(1), erie.tGas, "tGas should default to 1 ns per gas unit when set to zero")
	})
}

func TestDecide(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Parallel()

	t.Run("Empty pending", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg)
		lastNotarised := &ExecutionResultMeta{HeaderTimeMs: 1000}
		pending := []ExecutionResultMeta{}
		currentHdrTsNs := uint64(1000+500) * 1_000_000 // 1000 ms + 500 ms margin
		wantAllowed := 0
		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Accept all items", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110, MaxResultsPerBlock: 0}
		erie := NewExecutionResultInclusionEstimator(cfg)
		lastNotarised := &ExecutionResultMeta{HeaderTimeMs: 1000}
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 200},
		}
		currentHdrTsNs := uint64(1000+500) * 1_000_000
		wantAllowed := 2

		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Reject second item", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg)
		lastNotarised := &ExecutionResultMeta{HeaderTimeMs: 1000}
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100_000_000},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 999_000_000},
		}
		currentHdrTsNs := uint64(1000+200) * 1_000_000
		wantAllowed := 1

		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Allow only up to boundary (t_done == t_now)", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg)
		lastNotarised := &ExecutionResultMeta{HeaderTimeMs: 1000}
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100_000_000}, // t_done with margin = 1232
			{HeaderNonce: 2, HeaderTimeMs: 1011, GasUsed: 100_000_000},
			{HeaderNonce: 3, HeaderTimeMs: 1012, GasUsed: 100_000_000},
		}
		currentHdrTsNs := uint64(1010*1.10) * 1_000_000
		wantAllowed := 3

		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Hit MaxResultsPerBlock cap", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110, MaxResultsPerBlock: 1}
		erie := NewExecutionResultInclusionEstimator(cfg)
		lastNotarised := &ExecutionResultMeta{HeaderTimeMs: 1000}
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 100},
		}
		currentHdrTsNs := uint64(1000+500) * 1_000_000
		wantAllowed := 1

		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Genesis fallback", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110, GenesisTimeMs: 1_700_000_000_000}
		erie := NewExecutionResultInclusionEstimator(cfg)
		var lastNotarised *ExecutionResultMeta = nil
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, HeaderTimeMs: 1_700_000_000_010, GasUsed: 100}, // after genesis
		}
		currentHdrTsNs := uint64(1_700_000_000_010) * 1_000_000
		wantAllowed := 1

		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		require.Equal(t, wantAllowed, got, "Decide() should allow %d result on genesis, got %d", wantAllowed, got)
	})

	t.Run("Overflow protection", func(t *testing.T) {
		cfg := Config{SafetyMargin: 110, MaxResultsPerBlock: 10}
		erie := NewExecutionResultInclusionEstimator(cfg)
		lastNotarised := &ExecutionResultMeta{HeaderTimeMs: 1000}
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
			{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 1 << 62},
			{HeaderNonce: 3, HeaderTimeMs: 1030, GasUsed: 100},
		}
		currentHdrTsNs := uint64(1_700_000_000_000+1000) * 1_000_000
		wantAllowed := 1

		got := erie.Decide(cfg, lastNotarised, pending, currentHdrTsNs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})
}

func TestOverflowProtection(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Run("gasUsed * t_gas overflows", func(t *testing.T) {
		cfg := Config{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
			GenesisTimeMs:      0,
		}
		erie := NewExecutionResultInclusionEstimator(cfg)
		erie.SetTimePerGasUnit(2)

		pending := []ExecutionResultMeta{
			{GasUsed: math.MaxUint64, HeaderTimeMs: 1000}, // large enough to overflow
			{GasUsed: math.MaxUint64},
		}

		currentTime := uint64(1<<63 - 1) // large enough to fail if overflowed

		num_accepted := erie.Decide(cfg, nil, pending, currentTime)
		t.Log("num_accepted:", num_accepted)
		require.Equal(t, 0, num_accepted, "should not accept any results due to overflow")
	})

}

func TestDecide_EdgeCases(t *testing.T) {
	_ = logger.SetLogLevel("*:DEBUG")
	t.Parallel()
	cfg := Config{
		SafetyMargin:       10,
		MaxResultsPerBlock: 10,
		GenesisTimeMs:      1000,
	}
	erie := NewExecutionResultInclusionEstimator(cfg)
	now := uint64(2000)

	t.Run("zero GasUsed", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, GasUsed: 0, HeaderTimeMs: 1100},
			{HeaderNonce: 2, GasUsed: 0, HeaderTimeMs: 1200},
		}

		// Ensure currentHdrTsNs is after all pending headers
		currentHdrTsNs := convertMsToNs(1300) // > 1200

		got := erie.Decide(cfg, nil, pending, currentHdrTsNs)
		require.Equal(t, 2, got)
	})

	t.Run("large HeaderTimeMs gap", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, GasUsed: 100, HeaderTimeMs: 1000},
			{HeaderNonce: 2, GasUsed: 200, HeaderTimeMs: 10_000}, // large gap
		}
		got := erie.Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, 1, got) // should stop at 2nd
	})

	t.Run("all results after t_now", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{GasUsed: 10_000, HeaderTimeMs: 3000},
			{GasUsed: 20_000, HeaderTimeMs: 4000},
		}
		got := erie.Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, 0, got)
	})

	t.Run("non-monotonic in time", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, GasUsed: 50, HeaderTimeMs: 1100},
			{HeaderNonce: 2, GasUsed: 20, HeaderTimeMs: 1050},
			{HeaderNonce: 3, GasUsed: 70, HeaderTimeMs: 1150},
		}
		got := erie.Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, got, 1)
	})

	t.Run("non-monotonic in hash", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{HeaderNonce: 1, GasUsed: 10, HeaderTimeMs: 1100},
			{HeaderNonce: 4, GasUsed: 20, HeaderTimeMs: 1125},
			{HeaderNonce: 3, GasUsed: 70, HeaderTimeMs: 1150},
		}
		got := erie.Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, got, 2)
	})
}

func BenchmarkDecideScaling_10(b *testing.B) {
	cfg := Config{SafetyMargin: 110}
	erie := NewExecutionResultInclusionEstimator(cfg)
	last := &ExecutionResultMeta{HeaderTimeMs: 1}

	b.ReportAllocs()

	for n := range 10 {
		pendingSize := 12 * (1 << n) // 12, 24, 48, ...
		pending := make([]ExecutionResultMeta, pendingSize)
		for i := range pending {
			pending[i] = ExecutionResultMeta{
				HeaderNonce:  uint64(i + 1),
				HeaderTimeMs: last.HeaderTimeMs + uint64(i),
				GasUsed:      1,
			}
		}
		now := convertMsToNs(10000)

		b.Run(fmt.Sprintf("%d_results", pendingSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				erie.Decide(cfg, last, pending, now)
			}
		})
	}
}
