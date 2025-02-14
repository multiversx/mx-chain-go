package chaos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChaosConfig_verify(t *testing.T) {
	t.Run("with valid configuration", func(t *testing.T) {
		config := &chaosConfig{
			Failures: []failureDefinition{
				{
					Name:     string(failureShouldSkipSendingBlock),
					Triggers: []string{"true"},
				},
			},
		}

		err := config.verify()
		require.NoError(t, err)
	})

	t.Run("with unknown failure entries", func(t *testing.T) {
		config := &chaosConfig{
			Failures: []failureDefinition{
				{
					Name: "unknown",
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "unknown failure: unknown")
	})

	t.Run("with failure entries without triggers", func(t *testing.T) {
		config := &chaosConfig{
			Failures: []failureDefinition{
				{
					Name: string(failureShouldSkipSendingBlock),
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure shouldSkipSendingBlock has no triggers")
	})

	t.Run("with failure entries that require parameters", func(t *testing.T) {
		config := &chaosConfig{
			Failures: []failureDefinition{
				{
					Name:     string(failureShouldDelayBroadcastingFinalBlockAsLeader),
					Triggers: []string{"true"},
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure shouldDelayBroadcastingFinalBlockAsLeader requires the parameter 'duration'")

		config = &chaosConfig{
			Failures: []failureDefinition{
				{
					Name:     string(failureShouldDelayLeaderSignature),
					Triggers: []string{"true"},
				},
			},
		}

		err = config.verify()
		require.ErrorContains(t, err, "failure shouldDelayLeaderSignature requires the parameter 'duration'")
	})
}
