package roundActivation_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	mock2 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	mock3 "github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/roundActivation"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestNewRoundActivation(t *testing.T) {
	tests := []struct {
		args        func() (process.RoundHandler, sharding.Coordinator, config.RoundConfig)
		expectedErr error
	}{
		{
			args: func() (process.RoundHandler, sharding.Coordinator, config.RoundConfig) {
				return nil, &mock.ShardCoordinatorMock{}, config.RoundConfig{}
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() (process.RoundHandler, sharding.Coordinator, config.RoundConfig) {
				return &mock.RoundHandlerMock{}, nil, config.RoundConfig{}
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() (process.RoundHandler, sharding.Coordinator, config.RoundConfig) {
				return &mock.RoundHandlerMock{}, &mock.ShardCoordinatorMock{}, config.RoundConfig{
					EnableRoundsByName: []config.EnableRoundsByName{
						{Name: "Dummy1", Round: 1, Shard: 0},
						{Name: "Dummy1", Round: 2, Shard: 1},
					},
				}
			},
			expectedErr: process.ErrDuplicateRoundActivationName,
		},
		{
			args: func() (process.RoundHandler, sharding.Coordinator, config.RoundConfig) {
				return &mock.RoundHandlerMock{}, &mock.ShardCoordinatorMock{}, config.RoundConfig{}
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := roundActivation.NewRoundActivation(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestRoundActivation_IsEnabled(t *testing.T) {
	ra, _ := roundActivation.NewRoundActivation(
		&mock2.RoundHandlerStub{},
		&mock3.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 {
				return 1
			},
		},
		config.RoundConfig{
			EnableRoundsByName: []config.EnableRoundsByName{
				{Name: "Dummy1", Round: 1, Shard: 1},
				{Name: "Dummy2", Round: 2, Shard: 2},
			},
		},
	)

	require.True(t, ra.IsEnabled("Dummy1", 1))
	require.False(t, ra.IsEnabled("Dummy1", 2))
	require.False(t, ra.IsEnabled("Dummy2", 1))
	require.False(t, ra.IsEnabled("Dummy2", 2))
	require.False(t, ra.IsEnabled("Dummy3", 1))
}

func TestRoundActivation_IsEnabledInCurrentRound(t *testing.T) {
	ra, _ := roundActivation.NewRoundActivation(
		&mock2.RoundHandlerStub{
			RoundIndex: 1,
		},
		&mock3.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 {
				return 1
			},
		},
		config.RoundConfig{
			EnableRoundsByName: []config.EnableRoundsByName{
				{Name: "Dummy1", Round: 1, Shard: 1},
				{Name: "Dummy2", Round: 1, Shard: 1},
				{Name: "Dummy3", Round: 1, Shard: 2},
				{Name: "Dummy4", Round: 2, Shard: 1},
			},
		},
	)

	require.True(t, ra.IsEnabledInCurrentRound("Dummy1"))
	require.True(t, ra.IsEnabledInCurrentRound("Dummy2"))
	require.False(t, ra.IsEnabledInCurrentRound("Dummy3"))
	require.False(t, ra.IsEnabledInCurrentRound("Dummy4"))
	require.False(t, ra.IsEnabledInCurrentRound("Dummy5"))
}

func TestRoundActivation_IsInterfaceNil(t *testing.T) {
	ra, _ := roundActivation.NewRoundActivation(&mock.RoundHandlerMock{}, &mock.ShardCoordinatorMock{}, config.RoundConfig{})
	require.False(t, ra.IsInterfaceNil())
}
