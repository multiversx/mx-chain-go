package economics

import (
	"math"
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

// EconomicsData will store information about economics
type EconomicsData struct {
	rewardsValue        *big.Int
	communityPercentage float64
	leaderPercentage    float64
	burnPercentage      float64
	minGasPrice         uint64
	minGasLimit         uint64
	communityAddress    string
	burnAddress         string
}

const float64EqualityThreshold = 1e-9

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(economics *config.ConfigEconomics) (*EconomicsData, error) {
	//TODO check what happens if addresses are wrong
	rewardsValue, minGasPrice, minGasLimit, err := convertValues(economics)
	if err != nil {
		return nil, err
	}

	notGreaterThanZero := rewardsValue.Cmp(big.NewInt(0))
	if notGreaterThanZero < 0 {
		return nil, process.ErrInvalidRewardsValue
	}

	err = checkValues(economics)
	if err != nil {
		return nil, err
	}

	return &EconomicsData{
		rewardsValue:        rewardsValue,
		communityPercentage: economics.RewardsSettings.CommunityPercentage,
		leaderPercentage:    economics.RewardsSettings.LeaderPercentage,
		burnPercentage:      economics.RewardsSettings.BurnPercentage,
		minGasPrice:         minGasPrice,
		minGasLimit:         minGasLimit,
		communityAddress:    economics.EconomicsAddresses.CommunityAddress,
		burnAddress:         economics.EconomicsAddresses.BurnAddress,
	}, nil
}

func convertValues(economics *config.ConfigEconomics) (*big.Int, uint64, uint64, error) {
	conversionBase := 10
	bitConversionSize := 64

	rewardsValue := new(big.Int)
	rewardsValue, ok := rewardsValue.SetString(economics.RewardsSettings.RewardsValue, conversionBase)
	if !ok {
		return nil, 0, 0, process.ErrInvalidRewardsValue
	}

	minGasPrice, err := strconv.ParseUint(economics.FeeSettings.MinGasPrice, conversionBase, bitConversionSize)
	if err != nil {
		return nil, 0, 0, process.ErrInvalidMinimumGasPrice
	}

	minGasLimit, err := strconv.ParseUint(economics.FeeSettings.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return nil, 0, 0, process.ErrInvalidMinimumGasLimitForTx
	}

	return rewardsValue, minGasPrice, minGasLimit, nil
}

func checkValues(economics *config.ConfigEconomics) error {
	if isPercentageInvalid(economics.RewardsSettings.BurnPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.CommunityPercentage) ||
		isPercentageInvalid(economics.RewardsSettings.LeaderPercentage) {
		return process.ErrInvalidRewardsPercentages
	}

	sumPercentage := economics.RewardsSettings.BurnPercentage
	sumPercentage += economics.RewardsSettings.CommunityPercentage
	sumPercentage += economics.RewardsSettings.LeaderPercentage
	isEqualsToOne := math.Abs(sumPercentage-1.0) <= float64EqualityThreshold
	if !isEqualsToOne {
		return process.ErrInvalidRewardsPercentages
	}

	return nil
}

func isPercentageInvalid(percentage float64) bool {
	isLessThanZero := percentage < 0.0
	isGreaterThanOne := percentage > 1.0
	if isLessThanZero || isGreaterThanOne {
		return true
	}
	return false
}

// RewardsValue will return rewards value
func (ed *EconomicsData) RewardsValue() *big.Int {
	return ed.rewardsValue
}

// CommunityPercentage will return community reward percentage
func (ed *EconomicsData) CommunityPercentage() float64 {
	return ed.communityPercentage
}

// LeaderPercentage will return leader reward percentage
func (ed *EconomicsData) LeaderPercentage() float64 {
	return ed.leaderPercentage
}

// BurnPercentage will return burn percentage
func (ed *EconomicsData) BurnPercentage() float64 {
	return ed.burnPercentage
}

// MinGasPrice will return minimum gas price
func (ed *EconomicsData) MinGasPrice() uint64 {
	return ed.minGasPrice
}

// MinGasLimit will return minimum gas limit
func (ed *EconomicsData) MinGasLimit() uint64 {
	return ed.minGasLimit
}

// CommunityAddress will return community address
func (ed *EconomicsData) CommunityAddress() string {
	return ed.communityAddress
}

// BurnAddress will return burn address
func (ed *EconomicsData) BurnAddress() string {
	return ed.burnAddress
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	if ed == nil {
		return true
	}
	return false
}
