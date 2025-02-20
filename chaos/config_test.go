package chaos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChaosConfig_verify(t *testing.T) {
	t.Run("with valid configuration", func(t *testing.T) {
		config := &chaosConfig{
			SelectedProfileName: "dummy",
			Profiles: []chaosProfile{
				{
					Name: "dummy",
				},
			},
		}

		err := config.verify()
		require.NoError(t, err)
	})

	t.Run("with no selected profile", func(t *testing.T) {
		config := &chaosConfig{
			SelectedProfileName: "",
		}

		err := config.verify()
		require.ErrorContains(t, err, "no selected profile")
	})

	t.Run("with undefined selected profile", func(t *testing.T) {
		config := &chaosConfig{
			SelectedProfileName: "dummy",
		}

		err := config.verify()
		require.ErrorContains(t, err, "selected profile does not exist: dummy")
	})
}

func TestChaosProfile_verify(t *testing.T) {
	t.Run("with valid configuration", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name:     "foo",
					Type:     "panic",
					OnPoints: []string{"epochConfirmed"},
					Triggers: []string{"true"},
				},
			},
		}

		err := config.verify()
		require.NoError(t, err)
	})

	t.Run("with failure without fail type", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name: "foo",
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure 'foo' has no fail type")
	})

	t.Run("with failure with unknown fail type", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name: "foo",
					Type: "bar",
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure 'foo' has unknown fail type: 'bar'")
	})

	t.Run("with failure without points", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name: "foo",
					Type: "panic",
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure 'foo' has no points configured")
	})

	t.Run("with failure with unknown points", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name:     "foo",
					Type:     "panic",
					OnPoints: []string{"bar"},
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure 'foo' has unknown activation point: 'bar'")
	})

	t.Run("with failure without triggers", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name:     "foo",
					Type:     "panic",
					OnPoints: []string{"epochConfirmed"},
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure 'foo' has no triggers")
	})

	t.Run("with failure with fail type that require parameters", func(t *testing.T) {
		config := &chaosProfile{
			Failures: []failureDefinition{
				{
					Name:     "foo",
					Type:     "sleep",
					OnPoints: []string{"epochConfirmed"},
					Triggers: []string{"true"},
				},
			},
		}

		err := config.verify()
		require.ErrorContains(t, err, "failure 'foo', with fail type 'sleep', requires the parameter 'duration'")
	})
}
