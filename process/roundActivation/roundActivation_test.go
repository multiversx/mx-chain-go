package roundActivation_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/roundActivation"
	"github.com/stretchr/testify/require"
)

func TestNewRoundActivation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() config.RoundConfig
		expectedErr error
	}{
		{
			args: func() config.RoundConfig {
				return config.RoundConfig{
					ActivationDummy1: config.ActivationRoundByName{
						Name:  "",
						Round: 0,
					},
				}
			},
			expectedErr: process.ErrNilActivationRoundName,
		},
	}

	for _, currTest := range tests {
		_, err := roundActivation.NewRoundActivation(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestRoundActivation_IsEnabled(t *testing.T) {
	t.Parallel()

	ra, _ := roundActivation.NewRoundActivation(
		config.RoundConfig{
			ActivationDummy1: config.ActivationRoundByName{
				Name:  "Fix1",
				Round: 1000,
			},
		},
	)

	require.True(t, ra.IsEnabled("Fix1", 1000))
	require.True(t, ra.IsEnabled("Fix1", 1001))
	require.True(t, ra.IsEnabled("Fix1", 2000))

	require.False(t, ra.IsEnabled("Fix1", 100))
	require.False(t, ra.IsEnabled("Fix1", 998))
	require.False(t, ra.IsEnabled("Fix1", 999))

	require.False(t, ra.IsEnabled("Fix2", 999))
	require.False(t, ra.IsEnabled("Fix2", 1000))
	require.False(t, ra.IsEnabled("Fix2", 1001))
}

func TestRoundActivation_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ra, _ := roundActivation.NewRoundActivation(
		config.RoundConfig{
			ActivationDummy1: config.ActivationRoundByName{
				Name:  "Fix1",
				Round: 0,
			},
		})
	require.False(t, ra.IsInterfaceNil())
}
