package epochProviders

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func TestCreateCurrentEpochProvider_NilCurrentEpochProvider(t *testing.T) {
	t.Parallel()

	cnep, err := CreateCurrentEpochProvider(
		&chainParameters.ChainParametersHandlerStub{},
		0,
		false,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		3,
	)

	assert.Nil(t, err)
	assert.IsType(t, disabled.NewEpochProvider(), cnep)
}

func TestCreateCurrentEpochProvider_RegularNodeIgnoresZeroAssumedPersisters(t *testing.T) {
	t.Parallel()

	cnep, err := CreateCurrentEpochProvider(
		&chainParameters.ChainParametersHandlerStub{},
		0,
		false,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		0,
	)

	assert.Nil(t, err)
	assert.IsType(t, disabled.NewEpochProvider(), cnep)
}

func TestCreateCurrentEpochProvider_FullArchiveFallsBackOnZeroAssumedPersisters(t *testing.T) {
	t.Parallel()

	chainParameterHandler := &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{
				RoundsPerEpoch: 1,
				RoundDuration:  1,
			}
		},
	}
	cnep, err := CreateCurrentEpochProvider(
		chainParameterHandler,
		1,
		true,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		0,
	)

	assert.Nil(t, err)
	aep, _ := epochproviders.NewArithmeticEpochProvider(
		epochproviders.ArgArithmeticEpochProvider{
			StartTime:                       1,
			ChainParametersHandler:          chainParameterHandler,
			EnableEpochsHandler:             &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			AssumedPeersNumActivePersisters: defaultAssumedPeersNumActivePersisters,
		},
	)
	require.False(t, check.IfNil(aep))
	assert.IsType(t, aep, cnep)
}

func TestCreateCurrentEpochProvider_ArithmeticEpochProvider(t *testing.T) {
	t.Parallel()

	chainParameterHandler := &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{
				RoundsPerEpoch: 1,
				RoundDuration:  1,
			}
		},
	}
	cnep, err := CreateCurrentEpochProvider(
		chainParameterHandler,
		1,
		true,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		3,
	)
	require.Nil(t, err)

	aep, _ := epochproviders.NewArithmeticEpochProvider(
		epochproviders.ArgArithmeticEpochProvider{
			StartTime:                       1,
			ChainParametersHandler:          chainParameterHandler,
			EnableEpochsHandler:             &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			AssumedPeersNumActivePersisters: 3,
		},
	)
	require.False(t, check.IfNil(aep))
	assert.IsType(t, aep, cnep)
}
