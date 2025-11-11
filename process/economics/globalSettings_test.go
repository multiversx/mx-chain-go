package economics

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func createGlobalSettingsHandler() *globalSettingsHandler {
	tailInflationActivationEpoch := uint32(100)
	startYearInflation := 0.08757
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
	return gsh
}

func TestGlobalSettingsHandler_maxInflationRate(t *testing.T) {
	t.Parallel()

	tailInflationActivationEpoch := uint32(100)
	year1Inflation := 0.1
	minInflation := 0.01

	gsh := createGlobalSettingsHandler()
	gsh.yearSettings[1] = &config.YearSetting{Year: 1, MaximumInflation: year1Inflation}

	// Test before tail inflation activation
	rate := gsh.maxInflationRate(1, tailInflationActivationEpoch-1)
	require.Equal(t, year1Inflation, rate)

	// Test at tail inflation activation
	rate = gsh.maxInflationRate(1, tailInflationActivationEpoch)
	require.Equal(t, 0.08395550376084304, rate)

	// Test after tail inflation activation
	rate = gsh.maxInflationRate(1, tailInflationActivationEpoch+1)
	require.Equal(t, 0.08395550376084304, rate)

	// Test with a year that is not in the map
	rate = gsh.maxInflationRate(2, tailInflationActivationEpoch-1)
	require.Equal(t, minInflation, rate)
}

func TestGlobalSettingsMaxInflation(t *testing.T) {
	totalSupplyDay0 := 1000000.0 // Initial supply (not actually needed for Y, but included as per query)
	x := 0.05                    // Annual inflation rate as decimal (e.g., 5%)

	y := math.Pow(1+x, 1.0/365) - 1

	totalSupplyDay365 := totalSupplyDay0 * math.Pow(1+y, 365)

	totalSupplyOtherCompute := totalSupplyDay0 + totalSupplyDay0*x

	require.True(t, totalSupplyOtherCompute >= totalSupplyDay365)
	require.True(t, totalSupplyOtherCompute-totalSupplyDay365 < 1)
}

func TestGlobalSettingsMaxInflationRate_withSupply(t *testing.T) {
	gsh := createGlobalSettingsHandler()

	rate := gsh.maxInflationRate(1, 100)

	oneToken := 1000000000000000000.0 // 10^18
	totalSupplyDay0 := 28781358.0 * oneToken
	totalSupplyDay365Yearly := totalSupplyDay0 + totalSupplyDay0*gsh.startYearInflation

	dailyRate := rate / numberOfDaysInYear
	totalSupplyDay365Daily := totalSupplyDay0 * math.Pow(1+dailyRate, numberOfDaysInYear)

	require.True(t, totalSupplyDay365Yearly >= totalSupplyDay365Daily)
	require.True(t, totalSupplyDay365Yearly-totalSupplyDay365Daily < oneToken)
}
