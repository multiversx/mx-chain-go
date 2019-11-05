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
	stakeValue          *big.Int
	unBoundPeriod       uint64
}

const float64EqualityThreshold = 1e-9

// NewEconomicsData will create and object with information about economics parameters
func NewEconomicsData(economics *config.ConfigEconomics) (*EconomicsData, error) {
	//TODO check what happens if addresses are wrong
	rewardsValue, minGasPrice, minGasLimit, stakeValue, unBoundPeriod, err := convertValues(economics)
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
		stakeValue:          stakeValue,
		unBoundPeriod:       unBoundPeriod,
	}, nil
}

func convertValues(economics *config.ConfigEconomics) (*big.Int, uint64, uint64, *big.Int, uint64, error) {
	conversionBase := 10
	bitConversionSize := 64

	rewardsValue := new(big.Int)
	rewardsValue, ok := rewardsValue.SetString(economics.RewardsSettings.RewardsValue, conversionBase)
	if !ok {
		return nil, 0, 0, nil, 0, process.ErrInvalidRewardsValue
	}

	minGasPrice, err := strconv.ParseUint(economics.FeeSettings.MinGasPrice, conversionBase, bitConversionSize)
	if err != nil {
		return nil, 0, 0, nil, 0, process.ErrInvalidMinimumGasPrice
	}

	minGasLimit, err := strconv.ParseUint(economics.FeeSettings.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return nil, 0, 0, nil, 0, process.ErrInvalidMinimumGasLimitForTx
	}

	stakeValue := new(big.Int)
	stakeValue, ok = stakeValue.SetString(economics.ValidatorSettings.StakeValue, conversionBase)
	if !ok {
		return nil, 0, 0, nil, 0, process.ErrInvalidRewardsValue
	}

	unBoundPeriod, err := strconv.ParseUint(economics.FeeSettings.MinGasLimit, conversionBase, bitConversionSize)
	if err != nil {
		return nil, 0, 0, nil, 0, process.ErrInvalidUnboundPeriod
	}

	return rewardsValue, minGasPrice, minGasLimit, stakeValue, unBoundPeriod, nil
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

// ComputeFee computes the provided transaction's fee
func (ed *EconomicsData) ComputeFee(tx process.TransactionWithFeeHandler) *big.Int {
	gasPrice := big.NewInt(0).SetUint64(tx.GetGasPrice())
	gasLimit := big.NewInt(0).SetUint64(ed.ComputeGasLimit(tx))

	return gasPrice.Mul(gasPrice, gasLimit)
}

// CheckValidityTxValues checks if the provided transaction is economically correct
func (ed *EconomicsData) CheckValidityTxValues(tx process.TransactionWithFeeHandler) error {
	if ed.minGasPrice > tx.GetGasPrice() {
		return process.ErrInsufficientGasPriceInTx
	}

	requiredGasLimit := ed.ComputeGasLimit(tx)
	if requiredGasLimit > tx.GetGasLimit() {
		return process.ErrInsufficientGasLimitInTx
	}

	return nil
}

// ComputeGasLimit returns the gas limit need by the provided transaction in order to be executed
func (ed *EconomicsData) ComputeGasLimit(tx process.TransactionWithFeeHandler) uint64 {
	gasLimit := ed.minGasLimit

	//TODO: change this method of computing the gas limit of a notarizing tx
	// it should follow an exponential curve as to disincentivise notarizing large data
	// also, take into account if destination address is 0000...00000 as this will be a SC deploy tx
	gasLimit += uint64(len(tx.GetData()))

	return gasLimit
}

// CommunityAddress will return community address
func (ed *EconomicsData) CommunityAddress() string {
	return ed.communityAddress
}

// BurnAddress will return burn address
func (ed *EconomicsData) BurnAddress() string {
	return ed.burnAddress
}

// StakeValue will return the minimum stake value
func (ed *EconomicsData) StakeValue() *big.Int {
	return ed.stakeValue
}

// UnBoundPeriod will return the unbound period
func (ed *EconomicsData) UnBoundPeriod() uint64 {
	return ed.unBoundPeriod
}

// IsInterfaceNil returns true if there is no value under the interface
func (ed *EconomicsData) IsInterfaceNil() bool {
	if ed == nil {
		return true
	}
	return false
}
