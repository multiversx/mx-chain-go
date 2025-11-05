package economics

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGlobalSettingsHandler_maxInflationRate(t *testing.T) {
	t.Parallel()

	tailInflationActivationEpoch := uint32(100)
	startYearInflation := 0.02
	year1Inflation := 0.1
	minInflation := 0.01
	economics := config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "2000000000000000000000",
			MinimumInflation:   minInflation,
			YearSettings: []*config.YearSetting{
				{
					Year:             1,
					MaximumInflation: 0.1,
				},
			},
			TailInflation: config.TailInflationSettings{
				EnableEpoch:        tailInflationActivationEpoch,
				StartYearInflation: startYearInflation,
				DecayPercentage:    0,
				MinimumInflation:   minInflation,
			},
		},
		RewardsSettings: config.RewardsSettings{
			RewardsConfigByEpoch: []config.EpochRewardSettings{
				{
					LeaderPercentage:                 0.1,
					DeveloperPercentage:              0.1,
					ProtocolSustainabilityPercentage: 0.1,
					ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
					TopUpGradientPoint:               "300000000000000000000",
					TopUpFactor:                      0.25,
					EpochEnable:                      0,
					EcosystemGrowthPercentage:        0.0,
					EcosystemGrowthAddress:           "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
					GrowthDividendPercentage:         0.0,
					GrowthDividendAddress:            "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
				},
			},
		},
		FeeSettings: config.FeeSettings{},
	}

	gsh, _ := newGlobalSettingsHandler(&economics)
	gsh.yearSettings[1] = &config.YearSetting{Year: 1, MaximumInflation: year1Inflation}

	// Test before tail inflation activation
	rate := gsh.maxInflationRate(1, tailInflationActivationEpoch-1)
	require.Equal(t, year1Inflation, rate)

	// Test at tail inflation activation
	rate = gsh.maxInflationRate(1, tailInflationActivationEpoch)
	require.Equal(t, startYearInflation, rate)

	// Test after tail inflation activation
	rate = gsh.maxInflationRate(1, tailInflationActivationEpoch+1)
	require.Equal(t, startYearInflation, rate)

	// Test with a year that is not in the map
	rate = gsh.maxInflationRate(2, tailInflationActivationEpoch-1)
	require.Equal(t, minInflation, rate)
}
