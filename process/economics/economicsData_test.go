package economics_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/stretchr/testify/assert"
)

func TestEconomicsData_ShouldWork(t *testing.T) {
	t.Parallel()

	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{})
	assert.NotNil(t, economicsData)
}

func TestEconomicsData_RewardsValue(t *testing.T) {
	t.Parallel()

	rewardsValue := uint64(100)
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		RewardsSettings: config.RewardsSettings{
			RewardsValue: rewardsValue,
		},
	})

	value := economicsData.RewardsValue()
	assert.Equal(t, rewardsValue, value)
}

func TestEconomicsData_CommunityPercentage(t *testing.T) {
	t.Parallel()

	communityPercentage := 0.50
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		RewardsSettings: config.RewardsSettings{
			CommunityPercentage: communityPercentage,
		},
	})

	value := economicsData.CommunityPercentage()
	assert.Equal(t, communityPercentage, value)
}

func TestEconomicsData_LeaderPercentage(t *testing.T) {
	t.Parallel()

	leaderPercentage := 0.40
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		RewardsSettings: config.RewardsSettings{
			LeaderPercentage: leaderPercentage,
		},
	})

	value := economicsData.LeaderPercentage()
	assert.Equal(t, leaderPercentage, value)
}

func TestEconomicsData_BurnPercentage(t *testing.T) {
	t.Parallel()

	burnPercentage := 0.41
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		RewardsSettings: config.RewardsSettings{
			BurnPercentage: burnPercentage,
		},
	})

	value := economicsData.BurnPercentage()
	assert.Equal(t, burnPercentage, value)
}

func TestEconomicsData_MinGasPrice(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(500)
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		FeeSettings: config.FeeSettings{
			MinGasPrice: minGasPrice,
		},
	})

	value := economicsData.MinGasPrice()
	assert.Equal(t, minGasPrice, value)
}

func TestEconomicsData_MinGasLimitForTx(t *testing.T) {
	t.Parallel()

	minGasLimitForTx := uint64(1500)
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		FeeSettings: config.FeeSettings{
			MinGasLimitForTx: minGasLimitForTx,
		},
	})

	value := economicsData.MinGasLimitForTx()
	assert.Equal(t, minGasLimitForTx, value)
}

func TestEconomicsData_MinTxFee(t *testing.T) {
	t.Parallel()

	minTxFee := uint64(502)
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		FeeSettings: config.FeeSettings{
			MinTxFee: minTxFee,
		},
	})

	value := economicsData.MinTxFee()
	assert.Equal(t, minTxFee, value)
}

func TestEconomicsData_CommunityAddress(t *testing.T) {
	t.Parallel()

	communityAddress := "addr1"
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		EconomicsAddresses: config.EconomicsAddresses{
			CommunityAddress: communityAddress,
		},
	})

	value := economicsData.CommunityAddress()
	assert.Equal(t, communityAddress, value)
}

func TestEconomicsData_BurnAddress(t *testing.T) {
	t.Parallel()

	burnAddress := "addr2"
	economicsData := economics.NewEconomicsData(&config.ConfigEconomics{
		EconomicsAddresses: config.EconomicsAddresses{
			BurnAddress: burnAddress,
		},
	})

	value := economicsData.BurnAddress()
	assert.Equal(t, burnAddress, value)
}
