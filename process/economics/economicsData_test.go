package economics_test

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/stretchr/testify/assert"
)

const (
	validatorIncreaseRatingStep = uint32(2)
	validatorDecreaseRatingStep = uint32(4)
	proposerIncreaseRatingStep  = uint32(1)
	proposerDecreaseRatingStep  = uint32(2)
)

func createDummyEconomicsConfig() *config.EconomicsConfig {
	return &config.EconomicsConfig{
		GlobalSettings: config.GlobalSettings{
			TotalSupply:      "2000000000000000000000",
			MinimumInflation: 0,
			MaximumInflation: 0.5,
		},
		RewardsSettings: config.RewardsSettings{
			LeaderPercentage: 0.1,
		},
		FeeSettings: config.FeeSettings{
			MaxGasLimitPerBlock:  "100000",
			MinGasPrice:          "18446744073709551615",
			MinGasLimit:          "500",
			GasPerDataByte:       "1",
			DataLimitForBaseCalc: "100000000",
		},
		ValidatorSettings: config.ValidatorSettings{
			GenesisNodePrice: "500000000",
			UnBoundPeriod:    "100000",
		},
		RatingSettings: config.RatingSettings{
			StartRating:                 50,
			MaxRating:                   100,
			MinRating:                   1,
			ProposerDecreaseRatingStep:  proposerDecreaseRatingStep,
			ProposerIncreaseRatingStep:  proposerIncreaseRatingStep,
			ValidatorDecreaseRatingStep: validatorDecreaseRatingStep,
			ValidatorIncreaseRatingStep: validatorIncreaseRatingStep,
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
		_, err := economics.NewEconomicsData(economicsConfig)
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
		_, err := economics.NewEconomicsData(economicsConfig)
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
		_, err := economics.NewEconomicsData(economicsConfig)
		assert.Equal(t, process.ErrInvalidMinimumGasLimitForTx, err)
	}

}

func TestNewEconomicsData_InvalidLeaderPercentageShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.LeaderPercentage = -0.1

	_, err := economics.NewEconomicsData(economicsConfig)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_ShouldWork(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	assert.NotNil(t, economicsData)
}

func TestEconomicsData_LeaderPercentage(t *testing.T) {
	t.Parallel()

	leaderPercentage := 0.40
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.LeaderPercentage = leaderPercentage
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_ComputeFeeNoTxData(t *testing.T) {
	t.Parallel()

	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: minGasLimit,
	}

	cost := economicsData.ComputeFee(tx)

	expectedCost := big.NewInt(0).SetUint64(gasPrice)
	expectedCost.Mul(expectedCost, big.NewInt(0).SetUint64(minGasLimit))
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_ComputeFeeWithTxData(t *testing.T) {
	t.Parallel()

	gasPrice := uint64(500)
	minGasLimit := uint64(12)
	txData := "text to be notarized"
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasLimit = strconv.FormatUint(minGasLimit, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: gasPrice,
		GasLimit: minGasLimit,
		Data:     []byte(txData),
	}

	cost := economicsData.ComputeFee(tx)

	expectedGasLimit := big.NewInt(0).SetUint64(minGasLimit)
	expectedGasLimit.Add(expectedGasLimit, big.NewInt(int64(len(txData))))
	expectedCost := big.NewInt(0).SetUint64(gasPrice)
	expectedCost.Mul(expectedCost, expectedGasLimit)
	assert.Equal(t, expectedCost, cost)
}

func TestEconomicsData_TxWithLowerGasPriceShouldErr(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice - 1,
		GasLimit: minGasLimit,
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
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit - 1,
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
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit + 1,
		Data:     []byte("1"),
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Equal(t, process.ErrHigherGasLimitRequiredInTx, err)
}

func TestEconomicsData_TxWithWithEqualGasPriceLimitShouldWork(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice,
		GasLimit: minGasLimit,
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Nil(t, err)
}

func TestEconomicsData_TxWithWithMoreGasPriceLimitShouldWork(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	minGasLimit := uint64(12)
	maxGasLimitPerBlock := minGasLimit + 1
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MaxGasLimitPerBlock = fmt.Sprintf("%d", maxGasLimitPerBlock)
	economicsConfig.FeeSettings.MinGasPrice = fmt.Sprintf("%d", minGasPrice)
	economicsConfig.FeeSettings.MinGasLimit = fmt.Sprintf("%d", minGasLimit)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	tx := &transaction.Transaction{
		GasPrice: minGasPrice + 1,
		GasLimit: minGasLimit + 1,
	}

	err := economicsData.CheckValidityTxValues(tx)

	assert.Nil(t, err)
}

func TestEconomicsData_RatingsDataMinGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RatingSettings.MinRating = 10
	economicsConfig.RatingSettings.MaxRating = 8
	economicsData, err := economics.NewEconomicsData(economicsConfig)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrMaxRatingIsSmallerThanMinRating, err)
}

func TestEconomicsData_RatingsDataMinSmallerThanOne(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RatingSettings.MinRating = 0
	economicsConfig.RatingSettings.MaxRating = 8
	economicsData, err := economics.NewEconomicsData(economicsConfig)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrMinRatingSmallerThanOne, err)
}

func TestEconomicsData_RatingsStartGreaterMaxShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RatingSettings.MinRating = 10
	economicsConfig.RatingSettings.MaxRating = 100
	economicsConfig.RatingSettings.StartRating = 110
	economicsData, err := economics.NewEconomicsData(economicsConfig)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrStartRatingNotBetweenMinAndMax, err)
}

func TestEconomicsData_RatingsStartLowerMinShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RatingSettings.MinRating = 10
	economicsConfig.RatingSettings.MaxRating = 100
	economicsConfig.RatingSettings.StartRating = 5
	economicsData, err := economics.NewEconomicsData(economicsConfig)

	assert.Nil(t, economicsData)
	assert.Equal(t, process.ErrStartRatingNotBetweenMinAndMax, err)
}

func TestEconomicsData_RatingsCorrectValues(t *testing.T) {
	t.Parallel()

	minRating := uint32(10)
	maxRating := uint32(100)
	startRating := uint32(50)

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RatingSettings.MinRating = minRating
	economicsConfig.RatingSettings.MaxRating = maxRating
	economicsConfig.RatingSettings.StartRating = startRating
	economicsConfig.RatingSettings.ProposerDecreaseRatingStep = proposerDecreaseRatingStep
	economicsConfig.RatingSettings.ProposerIncreaseRatingStep = proposerIncreaseRatingStep
	economicsConfig.RatingSettings.ValidatorIncreaseRatingStep = validatorIncreaseRatingStep
	economicsConfig.RatingSettings.ValidatorDecreaseRatingStep = validatorDecreaseRatingStep

	economicsData, err := economics.NewEconomicsData(economicsConfig)

	assert.Nil(t, err)
	assert.NotNil(t, economicsData)
	assert.Equal(t, startRating, economicsData.RatingsData().StartRating())
	assert.Equal(t, minRating, economicsData.RatingsData().MinRating())
	assert.Equal(t, maxRating, economicsData.RatingsData().MaxRating())
	assert.Equal(t, validatorIncreaseRatingStep, economicsData.RatingsData().ValidatorIncreaseRatingStep())
	assert.Equal(t, validatorDecreaseRatingStep, economicsData.RatingsData().ValidatorDecreaseRatingStep())
	assert.Equal(t, proposerIncreaseRatingStep, economicsData.RatingsData().ProposerIncreaseRatingStep())
	assert.Equal(t, proposerDecreaseRatingStep, economicsData.RatingsData().ProposerDecreaseRatingStep())
}
