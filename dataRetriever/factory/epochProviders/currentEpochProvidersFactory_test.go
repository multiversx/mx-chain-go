package epochProviders

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCurrentEpochProvider_NilCurrentEpochProvider(t *testing.T) {
	t.Parallel()

	cnep, err := CreateCurrentEpochProvider(
		config.Config{
			StoragePruning: config.StoragePruningConfig{
				FullArchive: false,
			},
		},
		0,
		0,
	)

	assert.Nil(t, err)
	assert.IsType(t, &epochproviders.NilEpochProvider{}, cnep)
}

func TestCreateCurrentEpochProvider_ArithemticEpochProvider(t *testing.T) {
	t.Parallel()

	cnep, err := CreateCurrentEpochProvider(
		config.Config{
			StoragePruning: config.StoragePruningConfig{
				FullArchive: true,
			},
			EpochStartConfig: config.EpochStartConfig{
				RoundsPerEpoch: 1,
			},
		},
		1,
		1,
	)

	aep, _ := epochproviders.NewArithmeticEpochProvider(
		epochproviders.ArgArithmeticEpochProvider{
			RoundsPerEpoch:          1,
			RoundTimeInMilliseconds: 1,
			StartTime:               1,
		},
	)
	require.False(t, check.IfNil(aep))

	assert.Nil(t, err)
	assert.IsType(t, aep, cnep)
}
