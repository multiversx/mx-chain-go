package estimator

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/round"
)

func TestEstimatorCreation(t *testing.T) {
	t.Parallel()
	genesisTimeStampMs := uint64(1000)
	roundHandler := &round.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return genesisTimeStampMs + round*1000
		},
	}

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
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
	})

	t.Run("Custom config", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 120, MaxResultsPerBlock: 10}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		require.NotNil(t, erie, "NewExecutionResultInclusionEstimator should not return nil")
	})
}

func TestDecide(t *testing.T) {
	t.Parallel()

	genesisTimeStampMs := uint64(500)
	roundTime := uint64(100) // ms
	roundHandler := &round.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return genesisTimeStampMs + round*roundTime
		},
	}

	t.Run("Empty pending", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}
		var pending []data.ExecutionResultHandler
		currentHdrTsMs := uint64(1000 + 500) // 1000 ms + 500 ms margin
		wantAllowed := 0
		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		require.Equal(t, wantAllowed, got, fmt.Sprintf("Decide() = %d, want %d", got, wantAllowed))
	})

	t.Run("Accept all items", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110, MaxResultsPerBlock: 0}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}

		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 200}},
		}
		currentHdrTsMs := roundTime*3 + genesisTimeStampMs
		wantAllowed := 2

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Reject second item", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100_000_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 999_000_000}},
		}
		currentHdrTsMs := roundTime*3 + genesisTimeStampMs
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		if got != wantAllowed {
			t.Errorf("Decide() = %d, want %d", got, wantAllowed)
		}
	})

	t.Run("Allow only up to boundary (t_done == t_now)", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100_000_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 100_000_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 3, HeaderRound: 3, GasUsed: 100_000_000}},
		}
		currentHdrTsMs := (roundTime*2 + genesisTimeStampMs) * cfg.SafetyMargin / 100
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		require.Equal(t, wantAllowed, got, fmt.Sprintf("Decide() = %d, want %d", got, wantAllowed))
	})

	t.Run("Hit MaxResultsPerBlock cap", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110, MaxResultsPerBlock: 1}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 100}},
		}
		currentHdrTsMs := 2000 + genesisTimeStampMs
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		require.Equal(t, wantAllowed, got, fmt.Sprintf("Decide() = %d, want %d", got, wantAllowed))
	})

	t.Run("Genesis fallback", func(t *testing.T) {
		t.Parallel()

		genesisTimeStampMs := uint64(1_700_000_000_000)
		roundHandler := &round.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return genesisTimeStampMs + round*roundTime
			},
		}
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		var lastNotarised *LastExecutionResultForInclusion = nil
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}}, // after genesis
		}
		currentHdrTsMs := genesisTimeStampMs + 2*roundTime
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		require.Equal(t, wantAllowed, got, fmt.Sprintf("Decide() = %d, want %d", got, wantAllowed))
	})

	t.Run("Overflow protection", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110, MaxResultsPerBlock: 10}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 1000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: math.MaxUint64}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
		}
		currentHdrTsMs := uint64(1_700_000_000_000+1000) * 1_000_000
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, currentHdrTsMs)
		require.Equal(t, wantAllowed, got, fmt.Sprintf("Decide() = %d, want %d", got, wantAllowed))
	})
}

func TestOverflowProtection(t *testing.T) {
	t.Parallel()

	roundHandler := &round.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return round * 1000
		},
	}
	t.Run("gasUsed * t_gas overflows", func(t *testing.T) {
		cfg := config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
		}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)

		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 1000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: math.MaxUint64}}, // This will cause overflow
		}

		currentTime := uint64(1<<63 - 1)

		numAccepted := erie.Decide(nil, pending, currentTime)
		t.Log("num_accepted:", numAccepted)
		require.Equal(t, 1, numAccepted, "should only accept first result, then overflow")
	})
}

func TestDecide_EdgeCases(t *testing.T) {
	t.Parallel()
	_ = logger.SetLogLevel("*:DEBUG")

	genesisTimeStampMs := uint64(1000)
	roundTime := uint64(1000)
	roundHandler := &round.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return genesisTimeStampMs + round*roundTime
		},
	}
	cfg := config.ExecutionResultInclusionEstimatorConfig{
		SafetyMargin:       10,
		MaxResultsPerBlock: 10,
	}
	erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
	roundNow := uint64(3)
	now := genesisTimeStampMs + roundNow*roundTime

	t.Run("zero GasUsed", func(t *testing.T) {
		//t.Parallel()
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 0}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 0}},
		}

		// Ensure currentHdrTsMs is after all pending headers
		currentHdrTsMs := convertMsToNs(genesisTimeStampMs + 3*roundTime) // >  timestamp for HeaderNonce 2

		got := erie.Decide(nil, pending, currentHdrTsMs)
		require.Equal(t, 2, got)
	})

	t.Run("HeaderTime on genesis time", func(t *testing.T) {
		//t.Parallel()

		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)

		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 0, GasUsed: 100}},
		}
		got := erie.Decide(nil, pending, genesisTimeStampMs+roundTime)
		require.Equal(t, 1, got)
	})

	t.Run("HeaderRound before last notarised", func(t *testing.T) {
		//t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 2,
			ProposedInRound:  0,
		}
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
		}
		got := erie.Decide(lastNotarised, pending, genesisTimeStampMs+4*roundTime)
		require.Equal(t, 0, got)
	})

	t.Run("second execution result HeaderTimeMs after current", func(t *testing.T) {
		//t.Parallel()
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: roundNow + 1, GasUsed: 200}},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, 1, got) // should stop at 2nd
	})

	t.Run("all results after t_now", func(t *testing.T) {
		//t.Parallel()
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: roundNow + 1, GasUsed: 10_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: roundNow + 2, GasUsed: 10_000}},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, 0, got)
	})

	t.Run("non-monotonic in time", func(t *testing.T) {
		//t.Parallel()
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 2, GasUsed: 50}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 1, GasUsed: 20}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 3, HeaderRound: 3, GasUsed: 70}},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, got, 1)
	})

	t.Run("non-monotonic in nonce", func(t *testing.T) {
		//t.Parallel()
		pending := []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 50}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 20}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 4, HeaderRound: 3, GasUsed: 70}},
		}
		got := erie.Decide(nil, pending, now)
		require.Equal(t, got, 2)

		pending = []data.ExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 1, GasUsed: 20}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 2, GasUsed: 70}},
		}
		got = erie.Decide(nil, pending, now)
		require.Equal(t, got, 1)
	})
}

func BenchmarkDecideScaling_10(b *testing.B) {
	cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
	roundHandler := &round.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return round * 1000
		},
	}
	erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
	last := &LastExecutionResultForInclusion{
		NotarizedInRound: 0,
		ProposedInRound:  0,
	}

	b.ReportAllocs()

	for n := 0; n < 10; n++ {
		pendingSize := 12 * (1 << n) // 12, 24, 48, ...
		pending := make([]data.ExecutionResultHandler, pendingSize)
		for i := range pending {
			pending[i] = &block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{
				HeaderNonce: uint64(i + 1),
				HeaderRound: uint64(i + 1),
				GasUsed:     1,
			}}
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
