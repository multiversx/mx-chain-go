package epochProviders

import (
	"testing"
	"time"

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
		false,
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
		false,
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
		false,
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
		false,
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

func TestCreateCurrentEpochProvider_RecoverySyncUsesObservedEpoch(t *testing.T) {
	chainParameterHandler := &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{RoundsPerEpoch: 1, RoundDuration: 1}
		},
	}
	provider, err := CreateCurrentEpochProvider(
		chainParameterHandler,
		1,
		true,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		3,
		true,
	)
	require.NoError(t, err)
	provider.EpochConfirmed(2241, uint64(time.Now().Add(-time.Hour).Unix()))
	require.False(t, provider.EpochIsActiveInNetwork(2241))
	syncProvider, ok := provider.(interface{ EpochIsActiveForSync(uint32) bool })
	require.True(t, ok)
	require.True(t, syncProvider.EpochIsActiveForSync(2241))
	require.False(t, syncProvider.EpochIsActiveForSync(2239))
}
