package epochProviders

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCurrentEpochProvider_NilCurrentEpochProvider(t *testing.T) {
	t.Parallel()

	cnep, err := CreateCurrentEpochProvider(
		config.Config{},
		0,
		0,
		false,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	assert.Nil(t, err)
	assert.IsType(t, disabled.NewEpochProvider(), cnep)
}

func TestCreateCurrentEpochProvider_ArithmeticEpochProvider(t *testing.T) {
	t.Parallel()

	cnep, err := CreateCurrentEpochProvider(
		config.Config{
			EpochStartConfig: config.EpochStartConfig{
				RoundsPerEpoch: 1,
			},
		},
		1,
		1,
		true,
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)
	require.Nil(t, err)

	aep, _ := epochproviders.NewArithmeticEpochProvider(
		epochproviders.ArgArithmeticEpochProvider{
			RoundsPerEpoch:          1,
			RoundTimeInMilliseconds: 1,
			StartTime:               1,
			EnableEpochsHandler:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		},
	)
	require.False(t, check.IfNil(aep))
	assert.IsType(t, aep, cnep)
}
