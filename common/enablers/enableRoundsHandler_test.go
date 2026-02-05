package enablers

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createEnableRoundsConfig() config.RoundConfig {
	return config.RoundConfig{
		RoundActivations: map[string]config.ActivationRoundByName{
			string(common.DisableAsyncCallV1Flag): {
				Round:   "1",
				Options: nil,
			},
			string(common.SupernovaRoundFlag): {
				Round:   "2",
				Options: nil,
			},
		},
	}
}

func TestNewEnableRoundsHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid config: empty (unloaded) round config", func(t *testing.T) {
		t.Parallel()

		handler, err := NewEnableRoundsHandler(
			config.RoundConfig{},
			&epochNotifier.RoundNotifierStub{},
		)

		assert.True(t, check.IfNil(handler))
		assert.True(t, errors.Is(err, errMissingRoundActivation))
		assert.True(t, strings.Contains(err.Error(), string(common.DisableAsyncCallV1Flag)))
	})

	t.Run("invalid round string", func(t *testing.T) {
		t.Parallel()

		cfg := config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				string(common.DisableAsyncCallV1Flag): {
					Round:   "0",
					Options: []string{"string 1", "string 2"},
				},
				string(common.SupernovaRoundFlag): {
					Round:   "[invalid round]",
					Options: []string{"string 1", "string 2"},
				},
			},
		}

		handler, err := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})

		assert.True(t, check.IfNil(handler))
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "invalid syntax while trying to convert"))
		assert.True(t, strings.Contains(err.Error(), "[invalid round]"))
	})

	t.Run("should work: round 0", func(t *testing.T) {
		t.Parallel()

		cfg := config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				string(common.DisableAsyncCallV1Flag): {
					Round:   "0",
					Options: []string{"string 1", "string 2"},
				},
				string(common.SupernovaRoundFlag): {
					Round:   "0",
					Options: []string{"string 1", "string 2"},
				},
			},
		}

		handler, err := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})

	t.Run("should work: round non-zero", func(t *testing.T) {
		t.Parallel()

		cfg := createEnableRoundsConfig()
		handler, err := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(handler))
	})
}

func TestEnableRoundsHandler_GetCurrentRound(t *testing.T) {
	t.Parallel()

	cfg := createEnableRoundsConfig()
	handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
	require.NotNil(t, handler)

	currentRound := uint64(123)
	handler.RoundConfirmed(currentRound, 0)

	require.Equal(t, currentRound, handler.GetCurrentRound())
}

func TestEnableRoundsHandler_IsFlagDefined(t *testing.T) {
	t.Parallel()

	cfg := createEnableRoundsConfig()
	handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
	require.NotNil(t, handler)

	require.True(t, handler.IsFlagDefined(common.SupernovaRoundFlag))
	require.False(t, handler.IsFlagDefined("new flag"))
}

func TestEnableRoundsHandler_IsFlagEnabledInRound(t *testing.T) {
	t.Parallel()

	cfg := createEnableRoundsConfig()
	handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
	require.NotNil(t, handler)

	require.False(t, handler.IsFlagEnabledInRound("not defined", 10))

	require.True(t, handler.IsFlagEnabledInRound(common.SupernovaRoundFlag, 2))
	require.True(t, handler.IsFlagEnabledInRound(common.SupernovaRoundFlag, 3))
	require.False(t, handler.IsFlagEnabledInRound(common.SupernovaRoundFlag, 1))
}

func TestEnableRoundsHandler_IsFlagEnabled(t *testing.T) {
	t.Parallel()

	cfg := createEnableRoundsConfig()
	handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
	require.NotNil(t, handler)

	handler.RoundConfirmed(0, 0)

	require.False(t, handler.IsFlagEnabled(common.DisableAsyncCallV1Flag))
	require.False(t, handler.IsFlagEnabled(common.SupernovaRoundFlag))

	handler.RoundConfirmed(math.MaxUint32, 0)

	require.True(t, handler.IsFlagEnabled(common.DisableAsyncCallV1Flag))
	require.True(t, handler.IsFlagEnabled(common.SupernovaRoundFlag))
}

func TestEnableRoundsHandler_GetActivationRound(t *testing.T) {
	t.Parallel()

	cfg := createEnableRoundsConfig()
	handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
	require.NotNil(t, handler)

	require.Equal(t, uint64(0), handler.GetActivationRound("not defined"))

	require.Equal(t, uint64(1), handler.GetActivationRound(common.DisableAsyncCallV1Flag))
	require.Equal(t, uint64(2), handler.GetActivationRound(common.SupernovaRoundFlag))
}
