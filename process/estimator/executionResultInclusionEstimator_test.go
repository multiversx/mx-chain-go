package estimator

import "testing"

func TestEIE(t *testing.T) {
	t.Parallel()

}

func TestDecide(t *testing.T) {
	tests := []struct {
		name           string
		cfg            Config
		lastNotarised  *ExecutionResultMeta
		pending        []ExecutionResultMeta
		currentHdrTsNs uint64
		wantAllowed    int
	}{
		{
			name: "All fit",
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
			name: "Second result too late",
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
			name: "Hit MaxResultsPerBlock",
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

func BenchmarkDecide(b *testing.B) {
	cfg := Config{SafetyMargin: 110}
	last := &ExecutionResultMeta{HeaderTimeMs: 1000}

	pending := make([]ExecutionResultMeta, 12)
	for i := range pending {
		pending[i] = ExecutionResultMeta{
			HeaderNonce:  uint64(i + 1),
			HeaderTimeMs: 1000 + uint64(i)*10,
			GasUsed:      5000,
		}
	}
	now := (1000 + 600) * 1_000_000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decide(cfg, last, pending, uint64(now))
	}
}
