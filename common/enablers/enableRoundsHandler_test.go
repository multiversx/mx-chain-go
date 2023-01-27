package enablers

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/assert"
)

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
		assert.True(t, strings.Contains(err.Error(), disableAsyncCallV1))
	})
	t.Run("invalid round string", func(t *testing.T) {
		t.Parallel()

		cfg := config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				disableAsyncCallV1: {
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
				disableAsyncCallV1: {
					Round:   "0",
					Options: []string{"string 1", "string 2"},
				},
			},
		}

		handler, err := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})

		assert.False(t, check.IfNil(handler))
		assert.Nil(t, err)
	})
	t.Run("should work: round non-zero", func(t *testing.T) {
		t.Parallel()

		cfg := config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				disableAsyncCallV1: {
					Round:   "445",
					Options: nil,
				},
			},
		}

		handler, err := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})

		assert.False(t, check.IfNil(handler))
		assert.Nil(t, err)
	})
}

func TestFlagsHolder_DisableAsyncCallV1Enabled(t *testing.T) {
	t.Parallel()

	t.Run("should work: config round 0", func(t *testing.T) {
		t.Parallel()

		cfg := config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				disableAsyncCallV1: {
					Round:   "0",
					Options: nil,
				},
			},
		}

		handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
		assert.True(t, handler.IsDisableAsyncCallV1Enabled()) // check round not called

		handler.RoundConfirmed(0, 0)
		assert.True(t, handler.IsDisableAsyncCallV1Enabled())

		handler.RoundConfirmed(1, 0)
		assert.True(t, handler.IsDisableAsyncCallV1Enabled())
	})
	t.Run("should work: config round 1", func(t *testing.T) {
		t.Parallel()

		cfg := config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				disableAsyncCallV1: {
					Round:   "1",
					Options: nil,
				},
			},
		}

		handler, _ := NewEnableRoundsHandler(cfg, &epochNotifier.RoundNotifierStub{})
		assert.False(t, handler.IsDisableAsyncCallV1Enabled()) // check round not called
		handler.RoundConfirmed(0, 0)
		assert.False(t, handler.IsDisableAsyncCallV1Enabled())

		handler.RoundConfirmed(1, 0)
		assert.True(t, handler.IsDisableAsyncCallV1Enabled())

		handler.RoundConfirmed(2, 0)
		assert.True(t, handler.IsDisableAsyncCallV1Enabled())

		handler.RoundConfirmed(0, 0)
		assert.False(t, handler.IsDisableAsyncCallV1Enabled())

		handler.RoundConfirmed(2, 0)
		assert.True(t, handler.IsDisableAsyncCallV1Enabled())
	})
}
