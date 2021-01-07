package fees

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/stretchr/testify/require"
)

func createDummyEconomicsConfig(gasPriceModifier float64) *config.EconomicsConfig {
	return &config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			GenesisTotalSupply: "2000000000000000000000",
			MinimumInflation:   0,
			YearSettings: []*config.YearSetting{
				{
					Year:             0,
					MaximumInflation: 0.01,
				},
			},
		},
		RewardsSettings: config.RewardsSettings{
			LeaderPercentage:                 0.1,
			ProtocolSustainabilityPercentage: 0.1,
			ProtocolSustainabilityAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			TopUpGradientPoint:               "300000000000000000000",
			TopUpFactor:                      0.25,
		},
		FeeSettings: config.FeeSettings{
			MaxGasLimitPerBlock:     "100000",
			MaxGasLimitPerMetaBlock: "1000000",
			MinGasPrice:             "1000000000",
			MinGasLimit:             "50000",
			GasPerDataByte:          "1500",
			GasPriceModifier:        gasPriceModifier,
		},
	}
}

func createArgsForEconomicsData(gasPriceModifier float64) economics.ArgsNewEconomicsData {
	args := economics.ArgsNewEconomicsData{
		Economics:                      createDummyEconomicsConfig(gasPriceModifier),
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
	}
	return args
}

func TestFeesProcessor_MoveBalanceNoData(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	feesProcessor := NewFeesProcessor(economicData)

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 50000,
	}

	gasUsed := feesProcessor.ComputeMoveBalanceGasUsed(tx)
	require.Equal(t, uint64(50000), gasUsed)

	fee := feesProcessor.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)
	require.Equal(t, big.NewInt(50000000000000), fee)
}

func TestFeesProcessor_MoveBalanceWithData(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	feesProcessor := NewFeesProcessor(economicData)

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("hello"),
	}

	gasUsed := feesProcessor.ComputeMoveBalanceGasUsed(tx)
	require.Equal(t, uint64(57500), gasUsed)

	fee := feesProcessor.ComputeTxFeeBasedOnGasUsed(tx, gasUsed)
	require.Equal(t, big.NewInt(57500000000000), fee)
}

func TestFeesProcessor_ComputeGasUsedAndFeeBasedOnRefundValue(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(1))
	feesProcessor := NewFeesProcessor(economicData)

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("hello"),
	}

	refundValue := "42500000000000"

	gasUsed, fee := feesProcessor.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, uint64(57500), gasUsed)
	require.Equal(t, big.NewInt(57500000000000), fee)
}

func TestFeesProcessor_ComputeGasUsedAndFeeBasedOnRefundValueFeeMultiplier(t *testing.T) {
	t.Parallel()

	economicData, _ := economics.NewEconomicsData(createArgsForEconomicsData(0.5))
	feesProcessor := NewFeesProcessor(economicData)

	tx := &transaction.Transaction{
		GasPrice: 1000000000,
		GasLimit: 100000,
		Data:     []byte("hello"),
	}

	refundValue := "40000000000000"

	gasUsed, fee := feesProcessor.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
	require.Equal(t, uint64(77500), gasUsed)
	require.Equal(t, big.NewInt(67500000000000), fee)
}
