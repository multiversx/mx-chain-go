package economics_test

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/stretchr/testify/assert"
)

func createDummyEconomicsConfig() *config.ConfigEconomics {
	return &config.ConfigEconomics{
		EconomicsAddresses: config.EconomicsAddresses{
			CommunityAddress: "addr1",
			BurnAddress:      "addr2",
		},
		RewardsSettings: config.RewardsSettings{
			RewardsValue:        "1000",
			CommunityPercentage: 0.1,
			LeaderPercentage:    0.1,
			BurnPercentage:      0.8,
		},
		FeeSettings: config.FeeSettings{
			MinGasPrice:      "100",
			MinGasLimitForTx: "500",
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
		economicsConfig.FeeSettings.MinGasLimitForTx = minGasLimit
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

func TestEconomicsData_MinGasPrice(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasPrice = strconv.FormatUint(minGasPrice, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.MinGasPrice()
	assert.Equal(t, minGasPrice, value)
}

func TestEconomicsData_MinGasLimitForTx(t *testing.T) {
	t.Parallel()

	minGasLimitForTx := uint64(1000)
	economicsConfig := createDummyEconomicsConfig()
	economicsConfig.FeeSettings.MinGasLimitForTx = strconv.FormatUint(minGasLimitForTx, 10)
	economicsData, _ := economics.NewEconomicsData(economicsConfig)

	value := economicsData.MinGasLimitForTx()
	assert.Equal(t, minGasLimitForTx, value)
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
