package economics_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDummyEconomicsConfig() *config.EconomicsConfig {
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
			MinGasPrice:             "18446744073709551615",
			MinGasLimit:             "500",
			GasPerDataByte:          "1",
			GasPriceModifier:        1.0,
		},
	}
}

func createArgsForEconomicsData() economics.ArgsNewEconomicsData {
	args := economics.ArgsNewEconomicsData{
		Economics:                      createDummyEconomicsConfig(),
		PenalizedTooMuchGasEnableEpoch: 0,
		EpochNotifier:                  &mock.EpochNotifierStub{},
	}
	return args
}

func TestNewEconomicsData_InvalidMaxGasLimitPerBlockShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	badGasLimitPerBlock := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
		"10000000000000000000000000000000000000000000000000000000000000",
	}

	for _, gasLimitPerBlock := range badGasLimitPerBlock {
		args.Economics.FeeSettings.MaxGasLimitPerBlock = gasLimitPerBlock
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMaxGasLimitPerBlock, err)
	}

}

func TestNewEconomicsData_InvalidMinGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	badGasPrice := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
		"10000000000000000000000000000000000000000000000000000000000000",
	}

	for _, gasPrice := range badGasPrice {
		args.Economics.FeeSettings.MinGasPrice = gasPrice
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMinimumGasPrice, err)
	}

}

func TestNewEconomicsData_InvalidMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	bagMinGasLimit := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
		"10000000000000000000000000000000000000000000000000000000000000",
	}

	for _, minGasLimit := range bagMinGasLimit {
		args.Economics.FeeSettings.MinGasLimit = minGasLimit
		_, err := economics.NewEconomicsData(args)
		assert.Equal(t, process.ErrInvalidMinimumGasLimitForTx, err)
	}

}

func TestNewEconomicsData_InvalidLeaderPercentageShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	args.Economics.RewardsSettings.LeaderPercentage = -0.1

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	args.EpochNotifier = nil

	_, err := economics.NewEconomicsData(args)
	assert.Equal(t, process.ErrNilEpochNotifier, err)

}

func TestNewEconomicsData_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	economicsData, _ := economics.NewEconomicsData(args)
	assert.NotNil(t, economicsData)
}

func TestEconomicsData_LeaderPercentage(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	leaderPercentage := 0.40
	args.Economics.RewardsSettings.LeaderPercentage = leaderPercentage
	economicsData, _ := economics.NewEconomicsData(args)

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_ComputeMoveBalanceFeeNoTxData(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	cost := economicsData.ComputeMoveBalanceFee(tx)

	expectedCost := big.NewInt(0).SetUint64(gasPrice)
	expectedCost.Mul(expectedCost, big.NewInt(0).SetUint64(minGasLimit))
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_ComputeMoveBalanceFeeWithTxData(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	txData := "text to be notarized"
	args.Economics.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: minGasLimit,
		Data:     []byte(txData),
		Value:    big.NewInt(0),
	}

	cost := economicsData.ComputeMoveBalanceFee(tx)

	expectedGasLimit := big.NewInt(0).SetUint64(minGasLimit)
	expectedGasLimit.Add(expectedGasLimit, big.NewInt(int64(len(txData))))
	expectedCost := big.NewInt(0).SetUint64(gasPrice)
	expectedCost.Mul(expectedCost, expectedGasLimit)
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_ComputeTxFeeShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	gasPrice := uint64(500)
	gasLimit := uint64(20)
	minGasLimit := uint64(10)
	args.Economics.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	args.Economics.FeeSettings.GasPriceModifier = 0.01
	args.PenalizedTooMuchGasEnableEpoch = 1
	args.GasPriceModifierEnableEpoch = 2
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	cost := economicsData.ComputeTxFee(tx)
	expectedCost := core.SafeMul(minGasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	economicsData.EpochConfirmed(1)

	cost = economicsData.ComputeTxFee(tx)
	expectedCost = core.SafeMul(gasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	economicsData.EpochConfirmed(2)
	cost = economicsData.ComputeTxFee(tx)
	assert.Equal(t, big.NewInt(5050), cost)
}

func TestEconomicsData_TxWithLowerGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice - 1,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrInsufficientGasPriceInTx, err)
}

func TestEconomicsData_TxWithLowerGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit - 1,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrInsufficientGasLimitInTx, err)
}

func TestEconomicsData_TxWithHigherGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit + 1,
		Data:     []byte("1"),
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)
}

func TestEconomicsData_TxWithWithMinGasPriceAndLimitShouldWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 1
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit,
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Nil(t, err)
}

func TestEconomicsData_TxWithWithMoreGasLimitThanMaximumPerBlockShouldNotWork(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := uint64(42)
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)

	tx := &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock,
		Value:    big.NewInt(0),
	}
	err := economicsData.CheckValidityTxValues(tx)
	require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)

	tx = &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock + 1,
		Value:    big.NewInt(0),
	}
	err = economicsData.CheckValidityTxValues(tx)
	require.Equal(t, process.ErrMoreGasThanGasLimitPerBlock, err)

	tx = &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock - 1,
		Value:    big.NewInt(0),
	}
	err = economicsData.CheckValidityTxValues(tx)
	require.Nil(t, err)
}

func TestEconomicsData_TxWithWithMoreValueThanGenesisSupplyShouldError(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 42
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: minGasLimit + 1,
		Value:    big.NewInt(0).Add(economicsData.GenesisTotalSupply(), big.NewInt(1)),
	}

	err := economicsData.CheckValidityTxValues(tx)
	assert.Equal(t, err, process.ErrTxValueTooBig)

	tx.Value.SetBytes([]byte("99999999999999999999999999999999999999999999999999999999999999999999999999999999999999"))
	err = economicsData.CheckValidityTxValues(tx)
	assert.Equal(t, err, process.ErrTxValueOutOfBounds)
}

func TestEconomicsData_SCRWithNotEnoughMoveBalanceShouldNotError(t *testing.T) {
	t.Parallel()

	args := createArgsForEconomicsData()
	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 42
	args.Economics.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	args.Economics.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	args.Economics.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(args)
	scr := &smartContractResult.SmartContractResult{
		GasPrice: minGasPrice + 1,
		GasLimit: minGasLimit + 1,
		Value:    big.NewInt(0).Add(economicsData.GenesisTotalSupply(), big.NewInt(1)),
	}

	err := economicsData.CheckValidityTxValues(scr)
	assert.Equal(t, err, process.ErrTxValueTooBig)

	scr.Value.SetBytes([]byte("99999999999999999999999999999999999999999999999999999999999999999999999999999999999999"))
	err = economicsData.CheckValidityTxValues(scr)
	assert.Equal(t, err, process.ErrTxValueOutOfBounds)

	scr.Value.Set(big.NewInt(10))
	scr.GasLimit = 1
	err = economicsData.CheckValidityTxValues(scr)
	assert.Nil(t, err)

	moveBalanceFee := economicsData.ComputeMoveBalanceFee(scr)
	assert.Equal(t, moveBalanceFee, big.NewInt(0))
}
