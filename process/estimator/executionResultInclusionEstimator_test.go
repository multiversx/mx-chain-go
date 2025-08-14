package estimator

import (
	"fmt"
	"math"
	"testing"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestEIE(t *testing.T) {
	t.Parallel()

}

func TestDecide(t *testing.T) {

	logger.SetLogLevel("*:DEBUG")

	tests := []struct {
		name           string
		cfg            Config
		lastNotarised  *ExecutionResultMeta
		pending        []ExecutionResultMeta
		currentHdrTsNs uint64
		wantAllowed    int
	}{
		{
			name: "Accept all items",
			cfg:  Config{SafetyMargin: 110, MaxResultsPerBlock: 0},
			lastNotarised: &ExecutionResultMeta{
				HeaderTimeMs: 1000,
			},
			pending: []ExecutionResultMeta{
				{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
				{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 200},
			},
			currentHdrTsNs: (1000 + 500) * 1_000_000, // 500ms later
			wantAllowed:    2,
		},
		{
			name: "Reject second item",
			cfg:  Config{SafetyMargin: 110},
			lastNotarised: &ExecutionResultMeta{
				HeaderTimeMs: 1000,
			},
			pending: []ExecutionResultMeta{
				{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100_000_000},
				{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 999_000_000},
			},
			currentHdrTsNs: (1000 + 200) * 1_000_000, // 200ms later
			wantAllowed:    1,
		},
		{
			name: "Allow only up to boundary (t_done == t_now)",
			cfg:  Config{SafetyMargin: 110},
			lastNotarised: &ExecutionResultMeta{
				HeaderTimeMs: 1000,
			},
			pending: []ExecutionResultMeta{
				{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100_000_000}, // t_done = 1120
				{HeaderNonce: 2, HeaderTimeMs: 1011, GasUsed: 100_000_000}, // t_done = 1121
				{HeaderNonce: 3, HeaderTimeMs: 1012, GasUsed: 100_000_000}, // t_done = 1122
			},
			currentHdrTsNs: 1120 * 1_000_000, // exactly matches t_done of first item
			wantAllowed:    1,
		},
		{
			name: "Hit MaxResultsPerBlock cap",
			cfg:  Config{SafetyMargin: 110, MaxResultsPerBlock: 1},
			lastNotarised: &ExecutionResultMeta{
				HeaderTimeMs: 1000,
			},
			pending: []ExecutionResultMeta{
				{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
				{HeaderNonce: 2, HeaderTimeMs: 1020, GasUsed: 100},
			},
			currentHdrTsNs: (1000 + 500) * 1_000_000,
			wantAllowed:    1,
		},
		{
			name:          "Genesis fallback",
			cfg:           Config{SafetyMargin: 110, GenesisTimeMs: 1_700_000_000_000},
			lastNotarised: nil,
			pending: []ExecutionResultMeta{
				{HeaderNonce: 1, HeaderTimeMs: 1010, GasUsed: 100},
			},
			currentHdrTsNs: (1_700_000_000_000 + 1000) * 1_000_000,
			wantAllowed:    1,
		},
		{
			name: "Overflow protection",
			cfg:  Config{SafetyMargin: 110, MaxResultsPerBlock: 10},
			lastNotarised: &ExecutionResultMeta{
				HeaderTimeMs: 1000,
			},
			pending: []ExecutionResultMeta{
				{HeaderNonce: 1, HeaderTimeMs: 1020, GasUsed: 100},
				{HeaderNonce: 2, HeaderTimeMs: 1010, GasUsed: 1 << 62},
				{HeaderNonce: 3, HeaderTimeMs: 1020, GasUsed: 100},
			},
			currentHdrTsNs: (1_700_000_000_000 + 1000) * 1_000_000,
			wantAllowed:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decide(tt.cfg, tt.lastNotarised, tt.pending, tt.currentHdrTsNs)
			if got != tt.wantAllowed {
				t.Errorf("Decide() = %d, want %d", got, tt.wantAllowed)
			}
		})
	}
}

func TestOverflowProtection(t *testing.T) {
	t.Run("gasUsed * t_gas overflows", func(t *testing.T) {
		cfg := Config{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
			GenesisTimeMs:      0,
		}

		pending := []ExecutionResultMeta{
			{GasUsed: math.MaxUint64},
		}

		currentTime := uint64(1<<63 - 1) // large enough to fail if overflowed

		num_accepted := Decide(cfg, nil, pending, currentTime)
		t.Log("num_accepted:", num_accepted)
		require.Equal(t, 0, num_accepted, "should not accept any results due to overflow")
	})

	t.Run("estimatedTime * margin overflows", func(t *testing.T) {
		cfg := Config{
			SafetyMargin:       math.MaxUint64 / 2, // absurd margin
			MaxResultsPerBlock: 0,
			GenesisTimeMs:      0,
		}

		pending := []ExecutionResultMeta{
			{GasUsed: 2}, // small gas, but huge margin should trigger overflow
		}

		currentTime := uint64(1<<63 - 1)

		num_accepted := Decide(cfg, nil, pending, currentTime)
		t.Log("num_accepted:", num_accepted)
		require.Equal(t, 0, num_accepted, "should not accept any results due to overflow")
	})
}

func TestDecide_EdgeCases(t *testing.T) {
	cfg := Config{
		SafetyMargin:       10,
		MaxResultsPerBlock: 10,
		GenesisTimeMs:      1000,
	}
	now := uint64(2000)

	t.Run("zero GasUsed", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{GasUsed: 0, HeaderTimeMs: 1100},
			{GasUsed: 0, HeaderTimeMs: 1200},
		}
		got := Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, 2, got)
	})

	t.Run("large HeaderTimeMs gap", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{GasUsed: 100, HeaderTimeMs: 1000},
			{GasUsed: 200, HeaderTimeMs: 10_000}, // large gap
		}
		got := Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, 1, got) // should stop at 2nd
	})

	t.Run("all results after t_now", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{GasUsed: 10_000, HeaderTimeMs: 3000},
			{GasUsed: 20_000, HeaderTimeMs: 4000},
		}
		got := Decide(cfg, nil, pending, now*1_000_000)
		require.Equal(t, 0, got)
	})

	t.Run("mixed order and values", func(t *testing.T) {
		pending := []ExecutionResultMeta{
			{GasUsed: 50, HeaderTimeMs: 1100},
			{GasUsed: 20, HeaderTimeMs: 1050},
			{GasUsed: 70, HeaderTimeMs: 1150},
		}
		got := Decide(cfg, nil, pending, now*1_000_000)
		require.GreaterOrEqual(t, got, 1)
	})
}

func BenchmarkDecideScaling_10(b *testing.B) {
	cfg := Config{SafetyMargin: 110}
	last := &ExecutionResultMeta{HeaderTimeMs: 1000}

	b.ReportAllocs()

	for n := range 10 {
		pendingSize := 12 * (1 << n) // 12, 24, 48, ...
		pending := make([]ExecutionResultMeta, pendingSize)
		for i := range pending {
			pending[i] = ExecutionResultMeta{
				HeaderNonce:  uint64(i + 1),
				HeaderTimeMs: 1000 + uint64(i)*10,
				GasUsed:      5000,
			}
		}
		now := convertMsToNs(1500)

		b.Run(fmt.Sprintf("%d_results", pendingSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Decide(cfg, last, pending, now)
			}
		})
	}
}

func BenchmarkDecideScaling_100(b *testing.B) {
	cfg := Config{SafetyMargin: 110}
	last := &ExecutionResultMeta{HeaderTimeMs: 1000}

	b.ReportAllocs()

	for n := range 100 {
		pendingSize := 12 * (1 << n) // 12, 24, 48, ...
		pending := make([]ExecutionResultMeta, pendingSize)
		for i := range pending {
			pending[i] = ExecutionResultMeta{
				HeaderNonce:  uint64(i + 1),
				HeaderTimeMs: 1000 + uint64(i)*10,
				GasUsed:      5000,
			}
		}
		now := convertMsToNs(1500)

		b.Run(fmt.Sprintf("%d_results", pendingSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Decide(cfg, last, pending, now)
			}
		})
	}
}
