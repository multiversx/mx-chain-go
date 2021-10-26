package roundActivation_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	mock2 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/roundActivation"
	"github.com/stretchr/testify/require"
)

func TestNewRoundActivation(t *testing.T) {
	tests := []struct {
		args        func() (process.RoundHandler, config.RoundConfig)
		expectedErr error
	}{
		{
			args: func() (process.RoundHandler, config.RoundConfig) {
				return nil, config.RoundConfig{}
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() (process.RoundHandler, config.RoundConfig) {
				return &mock.RoundHandlerMock{}, config.RoundConfig{
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
	ra, _ := roundActivation.NewRoundActivation(
		&mock2.RoundHandlerStub{
			RoundIndex: 1000,
		},
		config.RoundConfig{
			ActivationDummy1: config.ActivationRoundByName{
				Name:  "Fix1",
				Round: 1000,
			},
		},
	)

	require.True(t, ra.IsEnabled("Fix1", 1000))
	require.False(t, ra.IsEnabled("Fix1", 1001))
	require.False(t, ra.IsEnabled("Fix2", 1000))
	require.False(t, ra.IsEnabled("Fix2", 1001))
}

func TestRoundActivation_IsEnabledInCurrentRound(t *testing.T) {
	ra, _ := roundActivation.NewRoundActivation(
		&mock2.RoundHandlerStub{
			RoundIndex: 1000,
		},
		config.RoundConfig{
			ActivationDummy1: config.ActivationRoundByName{
				Name:  "Fix1",
				Round: 1000,
			},
		},
	)

	require.True(t, ra.IsEnabledInCurrentRound("Fix1"))
	require.False(t, ra.IsEnabledInCurrentRound("Fix2"))
}

func TestRoundActivation_IsInterfaceNil(t *testing.T) {
	ra, _ := roundActivation.NewRoundActivation(
		&mock.RoundHandlerMock{},
		config.RoundConfig{
			ActivationDummy1: config.ActivationRoundByName{
				Name:  "Fix1",
				Round: 0,
			},
		})
	require.False(t, ra.IsInterfaceNil())
}
