package estimator

import (
	"fmt"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

	genesisTimeStampMs := uint64(1000)
	roundTime := uint64(100)
	roundHandler := &round.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return genesisTimeStampMs + round*roundTime
		},
	}
	defaultCfg := config.ExecutionResultInclusionEstimatorConfig{
		SafetyMargin:       110,
		MaxResultsPerBlock: 10,
	}
	defaultErie := NewExecutionResultInclusionEstimator(defaultCfg, roundHandler)
	roundNow := uint64(3)

	t.Run("Empty pending", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}
		var pending []data.BaseExecutionResultHandler
		wantAllowed := 0
		got := erie.Decide(lastNotarised, pending, roundNow)
		require.Equal(t, wantAllowed, got, fmt.Sprintf("Decide() = %d, want %d", got, wantAllowed))
	})

	t.Run("Accept all items", func(t *testing.T) {
		t.Parallel()
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 1,
			ProposedInRound:  0,
		}

		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 200}},
		}
		wantAllowed := 2

		got := defaultErie.Decide(lastNotarised, pending, roundNow)
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
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100_000_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 999_000_000}},
		}
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, roundNow)
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
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100_000_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 100_000_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 3, HeaderRound: 3, GasUsed: 100_000_000}},
		}
		wantAllowed := 1

		got := erie.Decide(lastNotarised, pending, roundNow)
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
		pending := []data.BaseExecutionResultHandler{
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
		pending := []data.BaseExecutionResultHandler{
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
		pending := []data.BaseExecutionResultHandler{
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

	t.Run("overflow detected in block transactions time estimation - gasUsed * t_gas overflows", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
		}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)

		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 1000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: math.MaxUint64}}, // This will cause overflow
		}

		currentRound := uint64(1<<63 - 1)

		numAccepted := erie.Decide(nil, pending, currentRound)
		t.Log("num_accepted:", numAccepted)
		require.Equal(t, 1, numAccepted, "should only accept first result, then overflow")
	})

	t.Run("overflow detected in estimated time with margin", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
		}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: math.MaxUint64 / erie.tGas}}, // This will bring estimatedTime close to max
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 2}},                          // This will cause overflow
		}
		currentRound := uint64(1<<63 - 1)
		numAccepted := erie.Decide(nil, pending, currentRound)
		t.Log("num_accepted:", numAccepted)
		require.Equal(t, 0, numAccepted, "should overflow from the first result")
	})

	t.Run("overflow detected in total estimated time - accumulated estimatedTime overflows", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 0,
		}
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: uint64(math.MaxUint64 - 3),
			ProposedInRound:  0,
		}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: uint64(math.MaxUint64 - 2), GasUsed: math.MaxUint64 / 1000}}, // This will bring estimatedTime close to max in margin calculation
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 1}},
		}
		currentRound := uint64(math.MaxUint64 - 1)
		numAccepted := erie.Decide(lastNotarised, pending, currentRound)
		t.Log("num_accepted:", numAccepted)
		require.Equal(t, 0, numAccepted, "should overflow from the first result")
	})
}

func TestDecide_EdgeCases(t *testing.T) {
	t.Parallel()

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

	t.Run("zero GasUsed", func(t *testing.T) {
		t.Parallel()
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 0}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 0}},
		}

		got := erie.Decide(nil, pending, roundNow)
		require.Equal(t, 2, got)
	})

	t.Run("HeaderTime on genesis time", func(t *testing.T) {
		t.Parallel()

		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 0, GasUsed: 100}},
		}

		got := erie.Decide(nil, pending, roundNow)
		require.Equal(t, 0, got)
	})

	t.Run("HeaderRound before last notarised", func(t *testing.T) {
		t.Parallel()
		cfg := config.ExecutionResultInclusionEstimatorConfig{SafetyMargin: 110}
		erie := NewExecutionResultInclusionEstimator(cfg, roundHandler)
		lastNotarised := &LastExecutionResultForInclusion{
			NotarizedInRound: 3,
			ProposedInRound:  2,
		}
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
		}
		round := uint64(4)
		got := erie.Decide(lastNotarised, pending, round)
		require.Equal(t, 0, got)
	})

	t.Run("second execution result HeaderTimeMs after current", func(t *testing.T) {
		t.Parallel()
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 100}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: roundNow + 1, GasUsed: 200}},
		}
		got := erie.Decide(nil, pending, roundNow)
		require.Equal(t, 1, got) // should stop at 2nd
	})

	t.Run("all results after t_now", func(t *testing.T) {
		t.Parallel()
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: roundNow + 1, GasUsed: 10_000}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: roundNow + 2, GasUsed: 10_000}},
		}
		got := erie.Decide(nil, pending, roundNow)
		require.Equal(t, 0, got)
	})

	t.Run("non-monotonic in round", func(t *testing.T) {
		t.Parallel()
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 2, GasUsed: 50}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 1, GasUsed: 20}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 3, HeaderRound: 3, GasUsed: 70}},
		}
		got := erie.Decide(nil, pending, roundNow)
		require.Equal(t, got, 1)
	})

	t.Run("non-monotonic in nonce", func(t *testing.T) {
		t.Parallel()
		pending := []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 1, GasUsed: 50}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 2, GasUsed: 20}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 4, HeaderRound: 3, GasUsed: 70}},
		}
		got := erie.Decide(nil, pending, roundNow)
		require.Equal(t, got, 2)

		pending = []data.BaseExecutionResultHandler{
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 2, HeaderRound: 1, GasUsed: 20}},
			&block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{HeaderNonce: 1, HeaderRound: 2, GasUsed: 70}},
		}
		got = erie.Decide(nil, pending, roundNow)
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
		pending := make([]data.BaseExecutionResultHandler, pendingSize)
		for i := range pending {
			pending[i] = &block.ExecutionResult{BaseExecutionResult: &block.BaseExecutionResult{
				HeaderNonce: uint64(i + 1),
				HeaderRound: uint64(i + 1),
				GasUsed:     1,
			}}
		}
		now := uint64(100)

		b.Run(fmt.Sprintf("%d_results", pendingSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				erie.Decide(last, pending, now)
			}
		})
	}
}
