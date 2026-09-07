package economics

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
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

	chainParams := config.ChainParametersByEpochConfig{RoundsPerEpoch: 14400, RoundDuration: 6000}
	chainParamsHolder := &chainParameters.ChainParametersHandlerStub{
		ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
			return chainParams, nil
		},
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return chainParams
		},
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return []config.ChainParametersByEpochConfig{chainParams}
		},
	}

	gsh, _ := newGlobalSettingsHandler(&economics, chainParamsHolder)
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
	require.Equal(t, year1Inflation, rate)

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

	rate := gsh.maxInflationRate(1, 101)

	oneToken := 1000000000000000000.0 // 10^18
	totalSupplyDay0 := 28781358.0 * oneToken
	totalSupplyDay365Yearly := totalSupplyDay0 + totalSupplyDay0*gsh.startYearInflation

	dailyRate := rate / float64(numDaysInYear)
	totalSupplyDay365Daily := totalSupplyDay0 * math.Pow(1+dailyRate, float64(numDaysInYear))

	require.True(t, totalSupplyDay365Yearly >= totalSupplyDay365Daily)
	require.True(t, totalSupplyDay365Yearly-totalSupplyDay365Daily < oneToken)
}

func TestNewGlobalSettingsHandler_ZeroEpochDuration(t *testing.T) {
	t.Parallel()

	economics := &config.EconomicsConfig{}
	chainParams := config.ChainParametersByEpochConfig{RoundsPerEpoch: 100, RoundDuration: 0}
	chainParamsHolder := &chainParameters.ChainParametersHandlerStub{
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return []config.ChainParametersByEpochConfig{chainParams}
		},
	}

	_, err := newGlobalSettingsHandler(economics, chainParamsHolder)
	require.Error(t, err)
}

func TestNewGlobalSettingsHandler_Sorting(t *testing.T) {
	t.Parallel()

	economics := &config.EconomicsConfig{}
	chainParams1 := config.ChainParametersByEpochConfig{RoundsPerEpoch: 100, RoundDuration: 1000, EnableEpoch: 10}
	chainParams2 := config.ChainParametersByEpochConfig{RoundsPerEpoch: 200, RoundDuration: 2000, EnableEpoch: 20}
	chainParamsHolder := &chainParameters.ChainParametersHandlerStub{
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return []config.ChainParametersByEpochConfig{chainParams1, chainParams2}
		},
	}

	gsh, err := newGlobalSettingsHandler(economics, chainParamsHolder)
	require.NoError(t, err)

	require.Equal(t, uint32(20), gsh.inflationForEpoch[0].enableEpoch)
	require.Equal(t, uint32(10), gsh.inflationForEpoch[1].enableEpoch)
}

func TestNewGlobalSettingsHandler_MultipleChainParameters(t *testing.T) {
	t.Parallel()

	economics := &config.EconomicsConfig{}
	economics.GlobalSettings.TailInflation.StartYearInflation = 0.1

	chainParams1 := config.ChainParametersByEpochConfig{RoundsPerEpoch: 14400, RoundDuration: 6000, EnableEpoch: 10}
	chainParams2 := config.ChainParametersByEpochConfig{RoundsPerEpoch: 7200, RoundDuration: 3000, EnableEpoch: 20}
	chainParamsHolder := &chainParameters.ChainParametersHandlerStub{
		AllChainParametersCalled: func() []config.ChainParametersByEpochConfig {
			return []config.ChainParametersByEpochConfig{chainParams1, chainParams2}
		},
	}

	gsh, err := newGlobalSettingsHandler(economics, chainParamsHolder)
	require.NoError(t, err)

	require.Equal(t, uint32(20), gsh.inflationForEpoch[0].enableEpoch)
	require.InDelta(t, 0.0953137, gsh.inflationForEpoch[0].maxInflation, 0.000001)

	require.Equal(t, uint32(10), gsh.inflationForEpoch[1].enableEpoch)
	require.InDelta(t, 0.0953227, gsh.inflationForEpoch[1].maxInflation, 0.000001)
}
