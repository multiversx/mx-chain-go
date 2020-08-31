package economics_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
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
		},
		FeeSettings: config.FeeSettings{
			MaxGasLimitPerBlock:     "100000",
			MaxGasLimitPerMetaBlock: "1000000",
			MinGasPrice:             "18446744073709551615",
			MinGasLimit:             "500",
			GasPerDataByte:          "1",
			DataLimitForBaseCalc:    "100000000",
		},
	}
}

func TestNewEconomicsData_InvalidMaxGasLimitPerBlockShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
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
		economicsConfig.FeeSettings.MaxGasLimitPerBlock = gasLimitPerBlock
		_, err := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
		assert.Equal(t, process.ErrInvalidMaxGasLimitPerBlock, err)
	}

}

func TestNewEconomicsData_InvalidMinGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
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
		economicsConfig.FeeSettings.MinGasPrice = gasPrice
		_, err := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
		assert.Equal(t, process.ErrInvalidMinimumGasPrice, err)
	}

}

func TestNewEconomicsData_InvalidMinGasLimitShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
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
		economicsConfig.FeeSettings.MinGasLimit = minGasLimit
		_, err := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
		assert.Equal(t, process.ErrInvalidMinimumGasLimitForTx, err)
	}

}

func TestNewEconomicsData_InvalidLeaderPercentageShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.LeaderPercentage = -0.1

	_, err := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()

	_, err := economics.NewEconomicsData(economicsConfig, 0, nil)
	assert.Equal(t, process.ErrNilEpochNotifier, err)

}

func TestNewEconomicsData_ShouldWork(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
	assert.NotNil(t, economicsData)
}

func TestEconomicsData_LeaderPercentage(t *testing.T) {
	t.Parallel()

	leaderPercentage := 0.40
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.LeaderPercentage = leaderPercentage
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_ComputeMoveBalanceFeeNoTxData(t *testing.T) {
	t.Parallel()

	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
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

	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	txData := "text to be notarized"
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
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

func TestEconomicsData_EstimateMoveBalanceFeeShouldWork(t *testing.T) {
	t.Parallel()

	gasPrice := uint64(500)
	gasLimit := uint64(20)
	minGasLimit := uint64(10)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 1, &mock.EpochNotifierStub{})
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	cost := economicsData.EstimateMoveBalanceFee(tx)
	expectedCost := core.SafeMul(minGasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)

	economicsData.EpochConfirmed(1)

	cost = economicsData.EstimateMoveBalanceFee(tx)
	expectedCost = core.SafeMul(gasLimit, gasPrice)
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_TxWithLowerGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
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

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
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

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit + 1,
		Data:     []byte("1"),
		Value:    big.NewInt(0),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrHigherGasLimitRequiredInTx, err)
}

func TestEconomicsData_TxWithWithMinGasPriceAndLimitShouldWork(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 1
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
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

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := uint64(42)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})

	tx := &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock,
		Value:    big.NewInt(0),
	}
	err := economicsData.CheckValidityTxValues(tx)
	require.Equal(t, process.ErrHigherGasLimitRequiredInTx, err)

	tx = &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: maxGasLimitPerBlock + 1,
		Value:    big.NewInt(0),
	}
	err = economicsData.CheckValidityTxValues(tx)
	require.Equal(t, process.ErrHigherGasLimitRequiredInTx, err)

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

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 42
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig, 0, &mock.EpochNotifierStub{})
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
