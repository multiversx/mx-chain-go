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

func createDummyEconomicsConfig() *config.ConfigEconomics {
	return &config.ConfigEconomics{
		EconomicsAddresses: config.EconomicsAddresses{
			CommunityAddress: "addr1",
			BurnAddress:      "addr2",
		},
		RewardsSettings: config.RewardsSettings{
			RewardsValue:        "1000000000000000000000000000000000",
			CommunityPercentage: 0.1,
			LeaderPercentage:    0.1,
			BurnPercentage:      0.8,
		},
		FeeSettings: config.FeeSettings{
			MaxGasLimitPerBlock: "100000",
			MinGasPrice:         "18446744073709551615",
			MinGasLimit:         "500",
		},
		ValidatorSettings: config.ValidatorSettings{
			StakeValue:    "500000000",
			UnBoundPeriod: "100000",
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

func TestNewEconomicsData_InvalidRewardsValueShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	badRewardsValues := []string{
		"-1",
		"-100000000000000000000",
		"badValue",
		"",
		"#########",
		"11112S",
		"1111O0000",
		"10ERD",
	}

	for _, rewardsValue := range badRewardsValues {
		economicsConfig.RewardsSettings.RewardsValue = rewardsValue
		_, err := economics.NewEconomicsData(economicsConfig)
		assert.Equal(t, process.ErrInvalidRewardsValue, err)
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

func TestNewEconomicsData_InvalidBurnPercentageShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.BurnPercentage = -1.0
	economicsConfig.RewardsSettings.CommunityPercentage = 0.1
	economicsConfig.RewardsSettings.LeaderPercentage = 0.1

	_, err := economics.NewEconomicsData(economicsConfig)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_InvalidCommunityPercentageShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.BurnPercentage = 0.1
	economicsConfig.RewardsSettings.CommunityPercentage = -0.1
	economicsConfig.RewardsSettings.LeaderPercentage = 0.1

	_, err := economics.NewEconomicsData(economicsConfig)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_InvalidLeaderPercentageShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.BurnPercentage = 0.1
	economicsConfig.RewardsSettings.CommunityPercentage = 0.1
	economicsConfig.RewardsSettings.LeaderPercentage = -0.1

	_, err := economics.NewEconomicsData(economicsConfig)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}
func TestNewEconomicsData_InvalidRewardsPercentageSumShouldErr(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.BurnPercentage = 0.5
	economicsConfig.RewardsSettings.CommunityPercentage = 0.2
	economicsConfig.RewardsSettings.LeaderPercentage = 0.5

	_, err := economics.NewEconomicsData(economicsConfig)
	assert.Equal(t, process.ErrInvalidRewardsPercentages, err)

}

func TestNewEconomicsData_ShouldWork(t *testing.T) {
	t.Parallel()

	economicsConfig := createDummyEconomicsConfig()
	economicsData, _ := economics.NewEconomicsData(economicsConfig)
	assert.NotNil(t, economicsData)
}

func TestEconomicsData_RewardsValue(t *testing.T) {
	t.Parallel()

	rewardsValue := int64(100)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.RewardsValue = strconv.FormatInt(rewardsValue, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.RewardsValue()
	assert.Equal(t, big.NewInt(rewardsValue), value)
}

func TestEconomicsData_CommunityPercentage(t *testing.T) {
	t.Parallel()

	communityPercentage := 0.50
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.CommunityPercentage = communityPercentage
	economicsConfig.RewardsSettings.BurnPercentage = 0.2
	economicsConfig.RewardsSettings.LeaderPercentage = 0.3
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.CommunityPercentage()
	assert.Equal(t, communityPercentage, value)
}

func TestEconomicsData_LeaderPercentage(t *testing.T) {
	t.Parallel()

	leaderPercentage := 0.40
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.CommunityPercentage = 0.30
	economicsConfig.RewardsSettings.BurnPercentage = 0.30
	economicsConfig.RewardsSettings.LeaderPercentage = leaderPercentage
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_BurnPercentage(t *testing.T) {
	t.Parallel()

	burnPercentage := 0.41
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.RewardsSettings.BurnPercentage = burnPercentage
	economicsConfig.RewardsSettings.CommunityPercentage = 0.29
	economicsConfig.RewardsSettings.LeaderPercentage = 0.3
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.BurnPercentage()
	assert.Equal(t, burnPercentage, value)
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
		Data:     txData,
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
		Data:     "1",
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

func TestEconomicsData_CommunityAddress(t *testing.T) {
	t.Parallel()

	communityAddress := "addr1"
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.EconomicsAddresses.CommunityAddress = communityAddress
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.CommunityAddress()
	assert.Equal(t, communityAddress, value)
}

func TestEconomicsData_BurnAddress(t *testing.T) {
	t.Parallel()

	burnAddress := "addr2"
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.EconomicsAddresses.BurnAddress = burnAddress
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.BurnAddress()
	assert.Equal(t, burnAddress, value)
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
